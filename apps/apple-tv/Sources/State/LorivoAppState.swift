import Foundation

@MainActor
final class LorivoAppState: ObservableObject {
    @Published var serverURL = "http://127.0.0.1:8097"
    @Published var bootstrap: ClientBootstrap?
    @Published var pairingRequest: PairingRequest?
    @Published var pairedDeviceID = ""
    @Published var home: TVHomeResponse?
    @Published var connectionState = "Not connected"
    @Published var errorMessage = ""
    @Published var focusedPoster: MediaPoster = MediaPoster.samples[0]

    var isPaired: Bool {
        bootstrap != nil && !pairedDeviceID.isEmpty
    }

    func connectAndStartPairing() async {
        errorMessage = ""
        connectionState = "Connecting"
        do {
            let client = try LorivoAPIClient(serverURL: serverURL)
            bootstrap = try await client.bootstrap()
            pairingRequest = try await client.createPairingRequest(deviceName: "Apple TV")
            connectionState = "Enter code in Lorivo Settings"
            await pollPairing(client: client, id: pairingRequest?.id ?? "")
        } catch {
            if pairedDeviceID.isEmpty {
                bootstrap = nil
            }
            connectionState = "Server unreachable"
            errorMessage = error.localizedDescription
        }
    }

    func pollPairing(client: LorivoAPIClient, id: String) async {
        guard !id.isEmpty else { return }
        for _ in 0..<120 {
            do {
                try await Task.sleep(nanoseconds: 2_000_000_000)
                let latest = try await client.pairingStatus(id: id)
                pairingRequest = latest
                if latest.isApproved, let deviceID = latest.deviceId {
                    pairedDeviceID = deviceID
                    connectionState = "Paired"
                    errorMessage = ""
                    await loadHome(client: client)
                    return
                }
                if latest.isClosed {
                    connectionState = latest.status.capitalized
                    errorMessage = latest.status == "approved" ? "" : "Pairing \(latest.status). Start a new request from this Apple TV."
                    return
                }
            } catch {
                errorMessage = error.localizedDescription
            }
        }
        connectionState = "Pairing timed out"
        errorMessage = "Start a new pairing request from this Apple TV."
    }

    func resetPairing() {
        pairingRequest = nil
        pairedDeviceID = ""
        connectionState = "Not connected"
        errorMessage = ""
    }

    func loadHome(client: LorivoAPIClient? = nil) async {
        do {
            let api: LorivoAPIClient
            if let client {
                api = client
            } else {
                api = try LorivoAPIClient(serverURL: serverURL)
            }
            home = try await api.home()
            focusedPoster = home?.hero.posterModel() ?? focusedPoster
        } catch {
            errorMessage = error.localizedDescription
        }
    }
}
