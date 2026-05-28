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
/// `app=xuva`, `serverName=…`, `hostName=…`, `api=/api/client/bootstrap`,
/// and optionally `web=http://<ip-or-host>:<port>`.
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
        // print() (vs os.Logger) surfaces to `devicectl ... launch --console`
        // captures so discovery can be diagnosed from a real Apple TV. Keep
        // these in production builds — they're cheap, off the hot path, and
        // pay for themselves the moment something breaks on a user's LAN.
        print("[XUVA-DISC] start() — looking for _xuva._tcp on .local")
        let params = NWParameters()
        params.includePeerToPeer = true
        let descriptor = NWBrowser.Descriptor.bonjourWithTXTRecord(type: "_xuva._tcp", domain: nil)
        let nb = NWBrowser(for: descriptor, using: params)
        browser = nb
        isBrowsing = true

        nb.stateUpdateHandler = { [weak self] state in
            print("[XUVA-DISC] state -> \(state)")
            Task { @MainActor in
                guard let self else { return }
                switch state {
                case .ready:
                    self.isBrowsing = true
                case .failed(let err):
                    print("[XUVA-DISC] NWBrowser FAILED: \(err)")
                    self.logger.error("NWBrowser failed: \(err)")
                    self.isBrowsing = false
                case .cancelled:
                    self.isBrowsing = false
                default:
                    break
                }
            }
        }

        nb.browseResultsChangedHandler = { [weak self] results, changes in
            print("[XUVA-DISC] results=\(results.count) changes=\(changes.count)")
            for r in results {
                print("[XUVA-DISC]   endpoint=\(r.endpoint) metadata=\(r.metadata)")
            }
            Task { @MainActor in
                self?.handleResults(Array(results))
            }
        }
        nb.start(queue: .main)
        print("[XUVA-DISC] NWBrowser.start() invoked")
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
        // Preserve previously-resolved entries across browse refreshes. The
        // earlier implementation wiped `servers` to whatever could be derived
        // synchronously from TXT, throwing away any entry that took an async
        // NWConnection resolve to determine. That was the discovery bug.
        let existing = Dictionary(uniqueKeysWithValues: servers.map { ($0.id, $0) })
        var nextServers: [DiscoveredServer] = []
        var seen = Set<String>()
        for result in results {
            guard case .service(let name, _, _, _) = result.endpoint else { continue }
            let id = name
            if seen.contains(id) { continue }
            seen.insert(id)
            if let server = serverFromResult(result, id: id) {
                print("[XUVA-DISC] fastpath id=\(id) baseURL=\(server.baseURL)")
                nextServers.append(server)
            } else if let prior = existing[id] {
                // Keep the previously-resolved entry rather than wiping it.
                print("[XUVA-DISC] keep prior id=\(id) baseURL=\(prior.baseURL)")
                nextServers.append(prior)
            } else {
                // Need a TCP resolve to discover the actual host:port.
                print("[XUVA-DISC] resolving id=\(id) (no usable TXT data)")
                pending[id] = result
                resolveEndpoint(result, id: id)
            }
        }
        servers = nextServers.sorted { $0.name.localizedCaseInsensitiveCompare($1.name) == .orderedAscending }
        print("[XUVA-DISC] servers now=\(servers.map { "\($0.id)→\($0.baseURL)" })")
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
        // Prefer explicit web URL — but if its host is a bare hostname (no
        // dot, not an IP literal) it won't resolve from tvOS without a
        // `.local` suffix for mDNS. The server sometimes broadcasts
        // `web=http://DESKTOP-XYZ:8097` when no usable IPv4 was picked.
        if let web = txt["web"], let url = URL(string: web), let host = url.host {
            if host.contains(".") || isIPLiteral(host) {
                return host
            }
            return host + ".local"
        }
        if let host = txt["hostName"], !host.isEmpty {
            if host.contains(".") || isIPLiteral(host) { return host }
            return host + ".local"
        }
        return nil
    }

    private func isIPLiteral(_ s: String) -> Bool {
        if IPv4Address(s) != nil { return true }
        if IPv6Address(s) != nil { return true }
        return false
    }

    private func portFromTXT(_ txt: NWTXTRecord) -> Int? {
        // The actual port is on the endpoint; TXT only carries it when the
        // server includes a `web=…:<port>` entry. If the URL is https with
        // the default port, `url.port` is nil and we fall back to async
        // NWConnection resolve.
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
            print("[XUVA-DISC] resolve[\(id)] state -> \(state)")
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
                print("[XUVA-DISC] resolve[\(id)] ready host=\(host ?? "nil") port=\(port.map(String.init) ?? "nil")")
                Task { @MainActor in
                    self?.completeResolve(id: id, host: host, port: port)
                }
            case .failed(let err):
                print("[XUVA-DISC] resolve[\(id)] failed: \(err)")
                Task { @MainActor in
                    self?.resolvers.removeValue(forKey: id)
                }
            case .cancelled:
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
              let url = URL(string: "http://\(h):\(p)") else {
            print("[XUVA-DISC] completeResolve[\(id)] dropped — host=\(host ?? "nil") port=\(port.map(String.init) ?? "nil")")
            return
        }
        let serverName = txt["serverName"] ?? id
        let server = DiscoveredServer(id: id, name: serverName, baseURL: url, hostName: txt["hostName"])
        print("[XUVA-DISC] completeResolve[\(id)] -> \(url)")
        var next = servers
        if !next.contains(where: { $0.id == id }) {
            next.append(server)
            servers = next.sorted { $0.name.localizedCaseInsensitiveCompare($1.name) == .orderedAscending }
        }
    }
}
