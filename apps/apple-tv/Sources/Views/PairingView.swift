import SwiftUI

struct PairingView: View {
    @EnvironmentObject private var appState: LorivoAppState
    @FocusState private var focusedField: Bool

    var body: some View {
        VStack(alignment: .leading, spacing: 34) {
            VStack(alignment: .leading, spacing: 10) {
                Text("Lorivo")
                    .font(.system(size: 72, weight: .bold))
                    .foregroundStyle(LorivoTheme.text)
                Text("Private Cinema")
                    .font(.system(size: 28, weight: .semibold))
                    .foregroundStyle(LorivoTheme.soft)
            }

            VStack(alignment: .leading, spacing: 18) {
                Text(appState.pairingRequest == nil ? "Connect to your local server" : "Approve this Apple TV")
                    .font(.system(size: 38, weight: .semibold))
                    .foregroundStyle(LorivoTheme.text)
                Text(appState.pairingRequest == nil ? "Use the URL shown by the Lorivo desktop app or web admin." : "Open Lorivo Settings, choose Devices, then approve this pairing code.")
                    .font(.system(size: 24, weight: .regular))
                    .foregroundStyle(LorivoTheme.soft)
            }

            if let request = appState.pairingRequest {
                VStack(alignment: .leading, spacing: 14) {
                    Text(request.code ?? "------")
                        .font(.system(size: 92, weight: .black, design: .monospaced))
                        .foregroundStyle(LorivoTheme.text)
                        .tracking(10)
                    Text("Pairing request \(request.status)")
                        .font(.system(size: 22, weight: .semibold))
                        .foregroundStyle(LorivoTheme.soft)
                }
                .padding(.horizontal, 34)
                .padding(.vertical, 26)
                .background(LorivoTheme.graphite, in: RoundedRectangle(cornerRadius: LorivoTheme.panelRadius))
                .overlay(RoundedRectangle(cornerRadius: LorivoTheme.panelRadius).stroke(LorivoTheme.focus.opacity(0.36), lineWidth: 2))
            } else {
                TextField("http://lorivo.local:8097", text: $appState.serverURL)
                    .textContentType(.URL)
                    .font(.system(size: 28, weight: .medium))
                    .foregroundStyle(LorivoTheme.text)
                    .padding(.horizontal, 24)
                    .frame(width: 760, height: 72)
                    .background(LorivoTheme.graphite, in: RoundedRectangle(cornerRadius: LorivoTheme.cardRadius))
                    .overlay(RoundedRectangle(cornerRadius: LorivoTheme.cardRadius).stroke(focusedField ? LorivoTheme.focus : .clear, lineWidth: 3))
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
                .tint(LorivoTheme.amber)

                RouteBadge(text: appState.connectionState, tone: appState.connectionState == "Paired" ? LorivoTheme.green : LorivoTheme.focus)
            }

            if !appState.errorMessage.isEmpty {
                Text(appState.errorMessage)
                    .font(.system(size: 22, weight: .medium))
                    .foregroundStyle(LorivoTheme.red)
            }
        }
        .padding(.leading, LorivoTheme.horizontalMargin)
        .frame(maxWidth: .infinity, maxHeight: .infinity, alignment: .leading)
        .background(
            LinearGradient(colors: [LorivoTheme.carbon, LorivoTheme.cinema], startPoint: .topLeading, endPoint: .bottomTrailing)
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
