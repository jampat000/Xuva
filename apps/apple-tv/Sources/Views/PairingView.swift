import SwiftUI

struct PairingView: View {
    @EnvironmentObject private var appState: VyrdenAppState
    @FocusState private var focusedField: Bool

    var body: some View {
        VStack(alignment: .leading, spacing: 34) {
            VStack(alignment: .leading, spacing: 10) {
                Text("Vyrden")
                    .font(.system(size: 72, weight: .bold))
                    .foregroundStyle(VyrdenTheme.text)
                Text("Private Cinema")
                    .font(.system(size: 28, weight: .semibold))
                    .foregroundStyle(VyrdenTheme.soft)
            }

            VStack(alignment: .leading, spacing: 18) {
                Text(appState.pairingRequest == nil ? "Connect to your local server" : "Approve this Apple TV")
                    .font(.system(size: 38, weight: .semibold))
                    .foregroundStyle(VyrdenTheme.text)
                Text(appState.pairingRequest == nil ? "Use the URL shown by the Vyrden desktop app or web admin." : "Open Vyrden Settings, choose Devices, then approve this pairing code.")
                    .font(.system(size: 24, weight: .regular))
                    .foregroundStyle(VyrdenTheme.soft)
            }

            if let request = appState.pairingRequest {
                VStack(alignment: .leading, spacing: 14) {
                    Text(request.code ?? "------")
                        .font(.system(size: 92, weight: .black, design: .monospaced))
                        .foregroundStyle(VyrdenTheme.text)
                        .tracking(10)
                    Text("Pairing request \(request.status)")
                        .font(.system(size: 22, weight: .semibold))
                        .foregroundStyle(VyrdenTheme.soft)
                }
                .padding(.horizontal, 34)
                .padding(.vertical, 26)
                .background(VyrdenTheme.graphite, in: RoundedRectangle(cornerRadius: VyrdenTheme.panelRadius))
                .overlay(RoundedRectangle(cornerRadius: VyrdenTheme.panelRadius).stroke(VyrdenTheme.focus.opacity(0.36), lineWidth: 2))
            } else {
                TextField("http://vyrden.local:8097", text: $appState.serverURL)
                    .textContentType(.URL)
                    .font(.system(size: 28, weight: .medium))
                    .foregroundStyle(VyrdenTheme.text)
                    .padding(.horizontal, 24)
                    .frame(width: 760, height: 72)
                    .background(VyrdenTheme.graphite, in: RoundedRectangle(cornerRadius: VyrdenTheme.cardRadius))
                    .overlay(RoundedRectangle(cornerRadius: VyrdenTheme.cardRadius).stroke(focusedField ? VyrdenTheme.focus : .clear, lineWidth: 3))
                    .focused($focusedField)
            }

            HStack(spacing: 18) {
                Button {
                    if appState.pairingRequest == nil {
                        Task { await appState.connectAndStartPairing() }
                    } else {
                        appState.resetPairing()
                    }
                } label: {
                    Text(actionTitle)
                        .font(.system(size: 26, weight: .bold))
                        .padding(.horizontal, 34)
                        .padding(.vertical, 18)
                }
                .buttonStyle(.borderedProminent)
                .tint(VyrdenTheme.amber)

                RouteBadge(text: appState.connectionState, tone: appState.connectionState == "Paired" ? VyrdenTheme.green : VyrdenTheme.focus)
            }

            if !appState.errorMessage.isEmpty {
                Text(appState.errorMessage)
                    .font(.system(size: 22, weight: .medium))
                    .foregroundStyle(VyrdenTheme.red)
            }
        }
        .padding(.leading, VyrdenTheme.horizontalMargin)
        .frame(maxWidth: .infinity, maxHeight: .infinity, alignment: .leading)
        .background(
            LinearGradient(colors: [VyrdenTheme.carbon, VyrdenTheme.cinema], startPoint: .topLeading, endPoint: .bottomTrailing)
        )
    }

    private var actionTitle: String {
        if appState.connectionState == "Connecting" {
            return "Connecting"
        }
        if appState.pairingRequest != nil {
            return "Start Over"
        }
        return "Connect"
    }
}
