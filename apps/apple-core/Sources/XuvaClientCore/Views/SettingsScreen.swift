import SwiftUI

struct SettingsScreen: View {
    @EnvironmentObject private var store: XuvaClientStore
    let dismiss: () -> Void

    var body: some View {
        NavigationStack {
            List {
                Section("Server") {
                    LabeledContent("Address", value: store.serverText)
                    LabeledContent("Status", value: connectionLabel)
                }
                Section {
                    Button(role: .destructive) {
                        store.resetConnection()
                        dismiss()
                    } label: {
                        Label("Sign Out", systemImage: "rectangle.portrait.and.arrow.right")
                    }
                }
                Section("About") {
                    LabeledContent("Version", value: appVersion)
                    LabeledContent("Build", value: buildNumber)
                }
            }
            .navigationTitle("Settings")
            #if os(tvOS)
            .toolbar {
                ToolbarItem(placement: .topBarTrailing) {
                    Button("Done", action: dismiss)
                }
            }
            #else
            .toolbar {
                ToolbarItem(placement: .cancellationAction) {
                    Button("Done", action: dismiss)
                }
            }
            #endif
        }
    }

    private var connectionLabel: String {
        switch store.connectionState {
        case .paired:   return "Connected"
        case .pairing:  return "Pairing…"
        case .connected: return "Connected (unpaired)"
        case .needsAuthCredential: return "Auth required"
        case .idle:     return "Not connected"
        }
    }

    private var appVersion: String {
        Bundle.main.infoDictionary?["CFBundleShortVersionString"] as? String ?? "—"
    }

    private var buildNumber: String {
        Bundle.main.infoDictionary?["CFBundleVersion"] as? String ?? "—"
    }
}
