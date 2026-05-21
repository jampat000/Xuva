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

                VStack(alignment: .leading, spacing: isCompact ? 20 : 24) {
                    XuvaLogo()
                    Text("Connect to Xuva")
                        .font(.system(size: titleSize(for: geometry.size), weight: .bold, design: .rounded))
                        .foregroundStyle(XuvaTheme.text)
                        .lineLimit(2)
                        .minimumScaleFactor(0.72)
                        .frame(maxWidth: isCompact ? .infinity : 760, alignment: .leading)
                    Text(introCopy)
                        .font(isCompact ? .body : .title2.weight(.medium))
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
                .padding(.vertical, isCompact ? 48 : 86)
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
        TextField("Xuva address", text: $store.serverText)
            .autocorrectionDisabled()
            .textInputAutocapitalization(.never)
            .keyboardType(.URL)
            .textContentType(.URL)
            .submitLabel(.go)
            .onSubmit {
                Task { await store.connect() }
            }
            .font(.system(size: tvControlFontSize, weight: .medium))
            .foregroundStyle(XuvaTheme.text)
            .padding(.horizontal, 22)
            .frame(height: tvControlHeight)
            .background(Color.white.opacity(0.06), in: Capsule(style: .continuous))
            .overlay(Capsule(style: .continuous).stroke(XuvaTheme.hairline))
    }

    private var pairingCard: some View {
        VStack(alignment: .leading, spacing: 18) {
            Label(store.bootstrap?.server?.name ?? "Xuva found", systemImage: "checkmark.seal.fill")
                .font(.headline)
                .foregroundStyle(XuvaTheme.good)

            if let code = store.pairing?.code {
                Text(code)
                    .font(.system(size: codeSize, weight: .black, design: .rounded))
                    .tracking(8)
                Text("Approve this code in Xuva. The library opens automatically.")
                    .foregroundStyle(XuvaTheme.muted)
                Text("Waiting for approval")
                    .font(.caption.weight(.semibold))
                    .foregroundStyle(XuvaTheme.primaryGlow)
            } else {
                Text("Create a pairing code for this Apple TV.")
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
                    buttonLabel(title: "Home", systemImage: "house.fill")
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
        .background(XuvaTheme.surface.opacity(0.78), in: RoundedRectangle(cornerRadius: 22, style: .continuous))
        .overlay(RoundedRectangle(cornerRadius: 22, style: .continuous).stroke(XuvaTheme.hairline))
    }

    @ViewBuilder
    private var connectionHint: some View {
        if store.connectionState == .needsAuthCredential {
            Label(
                "Saved access was rejected. Reset this Apple TV and pair again from Xuva.",
                systemImage: "lock.shield"
            )
            .font(.callout)
            .foregroundStyle(XuvaTheme.warn)
            .padding(16)
            .background(XuvaTheme.warn.opacity(0.12), in: RoundedRectangle(cornerRadius: 16, style: .continuous))
            .frame(maxWidth: 760, alignment: .leading)
        }
    }

    private func buttonLabel(title: String, systemImage: String) -> some View {
        Label(title, systemImage: systemImage)
            .labelStyle(.titleAndIcon)
            .lineLimit(1)
            .fixedSize(horizontal: true, vertical: false)
    }

    private var codeSize: CGFloat {
        #if os(tvOS)
        return 72
        #else
        return 56
        #endif
    }

    private var tvControlHeight: CGFloat {
        #if os(tvOS)
        return 50
        #else
        return 58
        #endif
    }

    private var tvControlFontSize: CGFloat {
        #if os(tvOS)
        return 28
        #else
        return 22
        #endif
    }

    private func titleSize(for size: CGSize) -> CGFloat {
        #if os(tvOS)
        return size.width > 900 ? 54 : 42
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
