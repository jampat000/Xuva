import Foundation
import Network
import SwiftUI
import os

/// A single Xuva server discovered via Bonjour (`_xuva._tcp.local.`).
public struct DiscoveredServer: Identifiable, Equatable, Hashable {
    public let id: String
    public let name: String
    public let baseURL: URL
    public let hostName: String?

    public init(id: String, name: String, baseURL: URL, hostName: String?) {
        self.id = id
        self.name = name
        self.baseURL = baseURL
        self.hostName = hostName
    }
}

/// Discovers Xuva servers on the local network. The server advertises itself
/// via mDNS in apps/.../server/internal/discovery/service.go with TXT records
/// `app=xuva`, `serverName=…`, `hostName=…`, `api=/api/client/bootstrap`.
@MainActor
public final class XuvaDiscovery: ObservableObject {
    @Published public private(set) var servers: [DiscoveredServer] = []
    @Published public private(set) var isBrowsing: Bool = false
    private let logger = Logger(subsystem: "com.xuva.client", category: "discovery")

    private var browser: NWBrowser?
    private var resolvers: [String: NWConnection] = [:]
    private var pending: [String: NWBrowser.Result] = [:]

    public init() {}

    public func start() {
        guard browser == nil else { return }
        logger.debug("start() — looking for _xuva._tcp on .local")
        let params = NWParameters()
        params.includePeerToPeer = true
        let descriptor = NWBrowser.Descriptor.bonjourWithTXTRecord(type: "_xuva._tcp", domain: nil)
        let nb = NWBrowser(for: descriptor, using: params)
        browser = nb
        isBrowsing = true

        nb.stateUpdateHandler = { [weak self] state in
            Task { @MainActor in
                guard let self else { return }
                switch state {
                case .ready:
                    self.isBrowsing = true
                case .failed(let err):
                    self.logger.error("NWBrowser failed: \(err)")
                    self.isBrowsing = false
                case .cancelled:
                    self.isBrowsing = false
                default:
                    break
                }
            }
        }

        nb.browseResultsChangedHandler = { [weak self] results, _ in
            Task { @MainActor in
                self?.handleResults(Array(results))
            }
        }
        nb.start(queue: .main)
    }

    public func stop() {
        browser?.cancel()
        browser = nil
        for (_, c) in resolvers { c.cancel() }
        resolvers.removeAll()
        pending.removeAll()
        isBrowsing = false
    }

    private func handleResults(_ results: [NWBrowser.Result]) {
        var nextServers: [DiscoveredServer] = []
        var seen = Set<String>()
        for result in results {
            guard case .service(let name, _, _, _) = result.endpoint else { continue }
            let id = name
            if seen.contains(id) { continue }
            seen.insert(id)
            if let server = serverFromResult(result, id: id) {
                nextServers.append(server)
            } else {
                // Need a TCP resolve to get the address — kick off and wait
                pending[id] = result
                resolveEndpoint(result, id: id)
            }
        }
        servers = nextServers.sorted { $0.name.localizedCaseInsensitiveCompare($1.name) == .orderedAscending }
    }

    private func serverFromResult(_ result: NWBrowser.Result, id: String) -> DiscoveredServer? {
        guard case .bonjour(let txtRecord) = result.metadata else { return nil }
        let port = portFromTXT(txtRecord)
        guard let host = hostFromTXT(txtRecord), let port else { return nil }
        let scheme = "http"
        guard let baseURL = URL(string: "\(scheme)://\(host):\(port)") else { return nil }
        let serverName = txtRecord["serverName"] ?? id
        return DiscoveredServer(id: id, name: serverName, baseURL: baseURL, hostName: txtRecord["hostName"])
    }

    private func hostFromTXT(_ txt: NWTXTRecord) -> String? {
        // Prefer explicit web URL; fall back to hostName + .local for mDNS.
        if let web = txt["web"], let url = URL(string: web), let host = url.host {
            return host
        }
        if let host = txt["hostName"], !host.isEmpty {
            if host.contains(".") { return host }
            return host + ".local"
        }
        return nil
    }

    private func portFromTXT(_ txt: NWTXTRecord) -> Int? {
        // The actual port is on the endpoint; TXT doesn't carry it. Connect
        // briefly to discover the resolved port.
        if let web = txt["web"], let url = URL(string: web), let port = url.port {
            return port
        }
        return nil
    }

    /// Fall back to resolving the endpoint to a real host:port when the TXT
    /// record didn't carry an explicit `web=` URL.
    private func resolveEndpoint(_ result: NWBrowser.Result, id: String) {
        guard resolvers[id] == nil else { return }
        let params = NWParameters.tcp
        let conn = NWConnection(to: result.endpoint, using: params)
        resolvers[id] = conn
        conn.stateUpdateHandler = { [weak self] state in
            switch state {
            case .ready:
                let host = conn.currentPath?.remoteEndpoint.flatMap { ep -> String? in
                    if case .hostPort(let h, _) = ep { return "\(h)" }
                    return nil
                }
                let port = conn.currentPath?.remoteEndpoint.flatMap { ep -> Int? in
                    if case .hostPort(_, let p) = ep { return Int(p.rawValue) }
                    return nil
                }
                Task { @MainActor in
                    self?.completeResolve(id: id, host: host, port: port)
                }
            case .failed(_), .cancelled:
                Task { @MainActor in
                    self?.resolvers.removeValue(forKey: id)
                }
            default:
                break
            }
        }
        conn.start(queue: .main)
    }

    private func completeResolve(id: String, host: String?, port: Int?) {
        resolvers[id]?.cancel()
        resolvers.removeValue(forKey: id)
        guard let result = pending.removeValue(forKey: id) else { return }
        guard case .bonjour(let txt) = result.metadata else { return }
        let resolvedHost = host ?? hostFromTXT(txt)
        guard let h = resolvedHost,
              let p = port ?? portFromTXT(txt),
              let url = URL(string: "http://\(h):\(p)") else { return }
        let serverName = txt["serverName"] ?? id
        let server = DiscoveredServer(id: id, name: serverName, baseURL: url, hostName: txt["hostName"])
        var next = servers
        if !next.contains(where: { $0.id == id }) {
            next.append(server)
            servers = next.sorted { $0.name.localizedCaseInsensitiveCompare($1.name) == .orderedAscending }
        }
    }
}
