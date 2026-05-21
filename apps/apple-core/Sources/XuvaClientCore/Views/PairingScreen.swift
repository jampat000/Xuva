import SwiftUI

public struct PairingScreen: View {
    @EnvironmentObject private var store: XuvaClientStore
    @FocusState private var focusedControl: PairingFocus?

    public init() {}

    public var body: some View {
        pairingBody
    }

    private var pairingBody: some View {
        GeometryReader { geometry in
            ScrollView {
                let isCompact = geometry.size.width < 700
                let controls: AnyLayout = if isCompact {
                    AnyLayout(VStackLayout(alignment: .leading, spacing: 14))
                } else {
                    AnyLayout(HStackLayout(spacing: 14))
                }

                VStack(alignment: .leading, spacing: isCompact ? 22 : 28) {
                    XuvaLogo()
                    Text("Pair this device to your Xuva library")
                        .font(.system(size: titleSize(for: geometry.size), weight: .bold, design: .rounded))
                        .foregroundStyle(XuvaTheme.text)
                        .lineLimit(3)
                        .minimumScaleFactor(0.72)
                        .frame(maxWidth: isCompact ? .infinity : 760, alignment: .leading)
                    Text("Connect to your local server, approve the device in Xuva, then browse and play from the couch. No admin surfaces live in this app.")
                        .font(isCompact ? .body : .title3)
                        .foregroundStyle(XuvaTheme.muted)
                        .frame(maxWidth: isCompact ? .infinity : 720, alignment: .leading)

                    controls {
                        serverURLControl

                        Button {
                            Task { await store.connect() }
                        } label: {
                            buttonLabel(title: store.isBusy ? "Connecting..." : "Connect", systemImage: store.isBusy ? "hourglass" : "play.fill")
                        }
                        .xuvaTVPrimaryActionStyle()
                        .focused($focusedControl, equals: .connect)
                        .xuvaDefaultKeyboardAction()
                        .disabled(store.isBusy)
                    }
                    .frame(maxWidth: isCompact ? .infinity : 1280, alignment: .leading)

                    if store.bootstrap != nil {
                        pairingCard
                    }

                    connectionHint

                    if let error = store.errorMessage {
                        Text(error)
                            .font(.callout)
                            .foregroundStyle(XuvaTheme.danger)
                            .padding(16)
                            .background(XuvaTheme.danger.opacity(0.12), in: RoundedRectangle(cornerRadius: 16, style: .continuous))
                    }
                }
                .padding(.horizontal, horizontalPadding(for: geometry.size))
                .padding(.vertical, isCompact ? 48 : 72)
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

    @ViewBuilder
    private var serverURLControl: some View {
        TextField("Server URL", text: $store.serverText)
            .autocorrectionDisabled()
            .textInputAutocapitalization(.never)
            .keyboardType(.URL)
            .textContentType(.URL)
            .submitLabel(.go)
            .onSubmit {
                Task { await store.connect() }
            }
            .font(.title3)
            .foregroundStyle(XuvaTheme.text)
            .padding(.horizontal, 18)
            .frame(height: 58)
            .background(XuvaTheme.surface, in: RoundedRectangle(cornerRadius: 8, style: .continuous))
            .overlay(RoundedRectangle(cornerRadius: 8, style: .continuous).stroke(XuvaTheme.hairline))
    }

    private var pairingCard: some View {
        VStack(alignment: .leading, spacing: 18) {
            Label(store.bootstrap?.server?.name ?? "Xuva server found", systemImage: "checkmark.seal.fill")
                .font(.headline)
                .foregroundStyle(XuvaTheme.good)

            if let code = store.pairing?.code {
                Text(code)
                    .font(.system(size: codeSize, weight: .black, design: .rounded))
                    .tracking(8)
                Text("Approve this code in the Xuva web app, then the client will continue to the media library.")
                    .foregroundStyle(XuvaTheme.muted)
                Text("Waiting for approval...")
                    .font(.caption.weight(.semibold))
                    .foregroundStyle(XuvaTheme.primaryGlow)
            } else {
                Text("Create a local pairing code for this device.")
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
                .xuvaTVPrimaryActionStyle()
                .focused($focusedControl, equals: .pair)
                .xuvaDefaultKeyboardAction()
                .disabled(store.isBusy)

                Button {
                    Task { await store.loadHome() }
                } label: {
                    buttonLabel(title: "Try home", systemImage: "house.fill")
                }
                .xuvaTVSecondaryActionStyle()
                .focused($focusedControl, equals: .home)
                .disabled(store.isBusy)

                Button {
                    store.resetConnection()
                } label: {
                    buttonLabel(title: "Reset", systemImage: "arrow.counterclockwise")
                }
                .xuvaTVSecondaryActionStyle()
                .disabled(store.isBusy)
            }
        }
        .padding(28)
        .frame(maxWidth: 760, alignment: .leading)
        .background(XuvaTheme.surface.opacity(0.86), in: RoundedRectangle(cornerRadius: 10, style: .continuous))
        .overlay(RoundedRectangle(cornerRadius: 10).stroke(XuvaTheme.hairline))
    }

    @ViewBuilder
    private var connectionHint: some View {
        if store.connectionState == .needsAuthCredential {
            Label(
                "This server is protected and did not accept the saved native credential. Reset this device and pair again from the Xuva web app.",
                systemImage: "lock.shield"
            )
            .font(.callout)
            .foregroundStyle(XuvaTheme.warn)
            .padding(16)
            .background(XuvaTheme.warn.opacity(0.12), in: RoundedRectangle(cornerRadius: 10, style: .continuous))
            .frame(maxWidth: 760, alignment: .leading)
        }
    }

    private func buttonLabel(title: String, systemImage: String) -> some View {
        Label(title, systemImage: systemImage)
            .labelStyle(.titleAndIcon)
            .lineLimit(1)
            .minimumScaleFactor(0.82)
    }

    private var codeSize: CGFloat {
        #if os(tvOS)
        return 72
        #else
        return 56
        #endif
    }

    private func titleSize(for size: CGSize) -> CGFloat {
        #if os(tvOS)
        return size.width > 900 ? 64 : 46
        #else
        return size.width > 700 ? 46 : 40
        #endif
    }

    private func horizontalPadding(for size: CGSize) -> CGFloat {
        #if os(tvOS)
        return size.width > 900 ? 96 : 44
        #else
        return size.width > 700 ? 44 : 28
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
    func xuvaTVPrimaryActionStyle() -> some View {
        self.buttonStyle(XuvaPrimaryButtonStyle())
    }

    @ViewBuilder
    func xuvaTVSecondaryActionStyle() -> some View {
        self.buttonStyle(XuvaSecondaryButtonStyle())
    }

    @ViewBuilder
    func xuvaDefaultKeyboardAction() -> some View {
        #if os(tvOS)
        self
        #else
        self.keyboardShortcut(.defaultAction)
        #endif
    }
}
