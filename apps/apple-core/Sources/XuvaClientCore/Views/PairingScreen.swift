import SwiftUI

public struct PairingScreen: View {
    @EnvironmentObject private var store: XuvaClientStore
    @FocusState private var focusedControl: PairingFocus?

    public init() {}

    public var body: some View {
        GeometryReader { geometry in
            let viewport = geometry.size
            let isCompact = viewport.width < 700
            let controls: AnyLayout = isCompact
                ? AnyLayout(VStackLayout(alignment: .leading, spacing: 14))
                : AnyLayout(HStackLayout(spacing: 14))

            ScrollView {
                VStack(alignment: .leading, spacing: isCompact ? 18 : 28) {
                    XuvaLogo(viewport: viewport)
                    Text("Connect to Xuva")
                        .font(.system(size: XuvaScale.heroTitleSize(viewport) * 0.55, weight: .bold))
                        .foregroundStyle(XuvaTheme.text)
                        .lineLimit(2)
                        .minimumScaleFactor(0.6)
                        .frame(maxWidth: XuvaScale.heroContentMaxWidth(viewport), alignment: .leading)
                    Text(introCopy)
                        .font(.system(size: XuvaScale.bodyFontSize(viewport)))
                        .foregroundStyle(XuvaTheme.muted)
                        .frame(maxWidth: XuvaScale.heroContentMaxWidth(viewport), alignment: .leading)

                    controls {
                        serverURLControl(viewport: viewport)

                        Button {
                            Task { await store.connect() }
                        } label: {
                            buttonLabel(title: store.isBusy ? "Connecting..." : "Connect", systemImage: store.isBusy ? "hourglass" : "play.fill")
                        }
                        .buttonStyle(XuvaPrimaryButtonStyle(viewport: viewport))
                        .focused($focusedControl, equals: .connect)
                        .xuvaDefaultKeyboardAction()
                        .disabled(store.isBusy)
                    }
                    .frame(maxWidth: .infinity, alignment: .leading)

                    if store.bootstrap != nil {
                        pairingCard(viewport: viewport)
                    }

                    connectionHint(viewport: viewport)

                    if let error = store.errorMessage {
                        Text(error)
                            .font(.system(size: XuvaScale.metaFontSize(viewport)))
                            .foregroundStyle(XuvaTheme.danger)
                            .padding(16)
                            .background(XuvaTheme.danger.opacity(0.12), in: RoundedRectangle(cornerRadius: 16, style: .continuous))
                    }
                }
                .padding(.horizontal, XuvaScale.safeHorizontal(viewport))
                .padding(.vertical, isCompact ? 36 : viewport.height * 0.10)
                .frame(maxWidth: .infinity, alignment: .leading)
            }
        }
        .onAppear {
            focusedControl = .connect
        }
        .task(id: store.pairing?.stableID) {
            await pollPairingWhilePending()
        }
    }

    private func serverURLControl(viewport: CGSize) -> some View {
        TextField("Xuva address", text: $store.serverText)
            .autocorrectionDisabled()
            .textInputAutocapitalization(.never)
            .keyboardType(.URL)
            .textContentType(.URL)
            .submitLabel(.go)
            .onSubmit {
                Task { await store.connect() }
            }
            .font(.system(size: XuvaScale.bodyFontSize(viewport), weight: .medium))
            .foregroundStyle(XuvaTheme.text)
            .padding(.horizontal, 22)
            .frame(height: XuvaScale.buttonHeight(viewport))
            .background(Color.white.opacity(0.06), in: Capsule(style: .continuous))
            .overlay(Capsule(style: .continuous).stroke(XuvaTheme.hairline))
    }

    private func pairingCard(viewport: CGSize) -> some View {
        VStack(alignment: .leading, spacing: 18) {
            Label(store.bootstrap?.server?.name ?? "Xuva found", systemImage: "checkmark.seal.fill")
                .font(.system(size: XuvaScale.bodyFontSize(viewport), weight: .bold))
                .foregroundStyle(XuvaTheme.good)

            if let code = store.pairing?.code {
                Text(code)
                    .font(.system(size: XuvaScale.heroTitleSize(viewport) * 0.78, weight: .black))
                    .tracking(8)
                Text("Approve this code in Xuva. The library opens automatically.")
                    .font(.system(size: XuvaScale.bodyFontSize(viewport)))
                    .foregroundStyle(XuvaTheme.muted)
                Text("Waiting for approval")
                    .font(.system(size: XuvaScale.metaFontSize(viewport), weight: .semibold))
                    .foregroundStyle(XuvaTheme.primaryGlow)
            } else {
                Text("Create a pairing code for this device.")
                    .font(.system(size: XuvaScale.bodyFontSize(viewport)))
                    .foregroundStyle(XuvaTheme.muted)
            }

            HStack(spacing: 12) {
                Button {
                    Task {
                        if store.pairing == nil {
                            await store.startPairing()
                        } else {
                            await store.pollPairingOnce()
                        }
                    }
                } label: {
                    buttonLabel(
                        title: store.isBusy ? "Working..." : (store.pairing == nil ? "Create pairing code" : "Check approval"),
                        systemImage: store.isBusy ? "hourglass" : "key.fill"
                    )
                }
                .buttonStyle(XuvaPrimaryButtonStyle(viewport: viewport))
                .focused($focusedControl, equals: .pair)
                .xuvaDefaultKeyboardAction()
                .disabled(store.isBusy)

                Button {
                    Task { await store.loadHome() }
                } label: {
                    buttonLabel(title: "Home", systemImage: "house.fill")
                }
                .buttonStyle(XuvaSecondaryButtonStyle(viewport: viewport))
                .focused($focusedControl, equals: .home)
                .disabled(store.isBusy)

                Button {
                    store.resetConnection()
                } label: {
                    buttonLabel(title: "Reset", systemImage: "arrow.counterclockwise")
                }
                .buttonStyle(XuvaSecondaryButtonStyle(viewport: viewport))
                .disabled(store.isBusy)
            }
        }
        .padding(28)
        .frame(maxWidth: XuvaScale.heroContentMaxWidth(viewport), alignment: .leading)
        .background(XuvaTheme.surface.opacity(0.78), in: RoundedRectangle(cornerRadius: 22, style: .continuous))
        .overlay(RoundedRectangle(cornerRadius: 22, style: .continuous).stroke(XuvaTheme.hairline))
    }

    @ViewBuilder
    private func connectionHint(viewport: CGSize) -> some View {
        if store.connectionState == .needsAuthCredential {
            Label(
                "Saved access was rejected. Reset this device and pair again from Xuva.",
                systemImage: "lock.shield"
            )
            .font(.system(size: XuvaScale.metaFontSize(viewport)))
            .foregroundStyle(XuvaTheme.warn)
            .padding(16)
            .background(XuvaTheme.warn.opacity(0.12), in: RoundedRectangle(cornerRadius: 16, style: .continuous))
            .frame(maxWidth: XuvaScale.heroContentMaxWidth(viewport), alignment: .leading)
        }
    }

    private func buttonLabel(title: String, systemImage: String) -> some View {
        Label(title, systemImage: systemImage)
            .labelStyle(.titleAndIcon)
            .lineLimit(1)
            .fixedSize(horizontal: true, vertical: false)
    }

    private var introCopy: String {
        #if os(tvOS)
        return "Approve this Apple TV once, then your library opens straight to movies and shows."
        #else
        return "Pair this device once, then your library opens straight to movies and shows."
        #endif
    }

    private func pollPairingWhilePending() async {
        guard store.pairing?.stableID.isEmpty == false else { return }
        while !Task.isCancelled {
            let status = store.pairing?.status?.lowercased()
            if status == "approved" || status == "denied" || status == "expired" {
                return
            }
            try? await Task.sleep(nanoseconds: 2_000_000_000)
            await store.pollPairingOnce()
        }
    }
}

private enum PairingFocus {
    case connect
    case pair
    case home
}

private extension View {
    @ViewBuilder
    func xuvaDefaultKeyboardAction() -> some View {
        #if os(tvOS)
        self
        #else
        self.keyboardShortcut(.defaultAction)
        #endif
    }
}
