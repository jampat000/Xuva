import SwiftUI

public struct PairingScreen: View {
    @EnvironmentObject private var store: XuvaClientStore
    @StateObject private var discovery = XuvaDiscovery()
    @State private var showManualEntry = false
    @State private var discoveryTimedOut = false
    @State private var showDiagLog = false
    @State private var showQREntry = false
    @State private var qrCodeText = ""
    #if !os(tvOS)
    @State private var showQRScanner = false
    #endif

    public init() {}

    public var body: some View {
        GeometryReader { geometry in
            let viewport = geometry.size
            ScrollView {
                VStack(alignment: .leading, spacing: viewport.width < 700 ? 22 : 32) {
                    XuvaLogo(viewport: viewport)
                    Text("Connect to Xuva")
                        .font(.system(size: XuvaScale.heroTitleSize(viewport) * 0.55, weight: .semibold, design: .default))
                        .tracking(XuvaScale.heroTitleSize(viewport) * 0.55 * -0.045)
                        .foregroundStyle(XuvaTheme.text)
                        .lineLimit(2)
                        .minimumScaleFactor(0.6)
                        .frame(maxWidth: XuvaScale.heroContentMaxWidth(viewport), alignment: .leading)
                    Text(introCopy)
                        .font(.system(size: XuvaScale.bodyFontSize(viewport)))
                        .foregroundStyle(XuvaTheme.muted)
                        .frame(maxWidth: XuvaScale.heroContentMaxWidth(viewport), alignment: .leading)

                    if store.bootstrap == nil && store.pairing == nil {
                        // Stage 1: pick a discovered server or fall back to manual URL.
                        discoverySection(viewport: viewport)
                        qrSection(viewport: viewport)
                        manualSection(viewport: viewport)
                    } else {
                        // Stage 2: pairing card with code.
                        pairingCard(viewport: viewport)
                    }

                    connectionHint(viewport: viewport)

                    // Error is rendered by the top toast in XuvaRootView
                    // (ErrorToast, auto-dismisses after 7s). Don't duplicate
                    // it inline here — real-device photos showed "the request
                    // timed out" appearing twice: once at top, once below.

                    // Version label — long-press (2 s on tvOS) to open the diagnostic log.
                    versionLabel(viewport: viewport)
                }
                .padding(.horizontal, XuvaScale.safeHorizontal(viewport))
                .padding(.vertical, viewport.width < 700 ? 36 : viewport.height * 0.08)
                .frame(maxWidth: .infinity, alignment: .leading)
            }
        }
        .sheet(isPresented: $showDiagLog) { DiagnosticLogView() }
        #if !os(tvOS)
        .sheet(isPresented: $showQRScanner) {
            QRScannerSheet { scanned in
                handleScannedURL(scanned)
                showQRScanner = false
            }
        }
        #endif
        .task(id: store.pairing?.stableID) {
            await pollPairingWhilePending()
        }
        .onAppear {
            discovery.start()
            Task {
                try? await Task.sleep(nanoseconds: 2_000_000_000)
                if discovery.servers.isEmpty {
                    discoveryTimedOut = true
                    showManualEntry = true
                }
            }
        }
        .onDisappear { discovery.stop() }
    }

    // MARK: – Discovery

    @ViewBuilder
    private func discoverySection(viewport: CGSize) -> some View {
        let labelSize = XuvaScale.eyebrowFontSize(viewport) + 1
        VStack(alignment: .leading, spacing: 14) {
            HStack(spacing: 10) {
                Image(systemName: "antenna.radiowaves.left.and.right")
                    .foregroundStyle(XuvaTheme.primaryGlow)
                Text("Servers on your network")
                    .font(.system(size: labelSize, weight: .semibold))
                    .tracking(labelSize * 0.20)
                    .textCase(.uppercase)
                    .foregroundStyle(XuvaTheme.mutedText)
                if discovery.isBrowsing && discovery.servers.isEmpty {
                    ProgressView()
                        #if !os(tvOS)
                        .controlSize(.small)
                        #endif
                        .tint(XuvaTheme.mutedText)
                }
                Spacer()
                refreshDiscoveryButton(viewport: viewport)
            }
            if discovery.servers.isEmpty {
                emptyDiscoveryHint(viewport: viewport)
            } else {
                VStack(alignment: .leading, spacing: 10) {
                    ForEach(discovery.servers) { server in
                        discoveryRow(server: server, viewport: viewport)
                    }
                }
            }
        }
        .frame(maxWidth: XuvaScale.heroContentMaxWidth(viewport), alignment: .leading)
    }

    @ViewBuilder
    private func refreshDiscoveryButton(viewport: CGSize) -> some View {
        // tvOS focus is rect-based: from the QR/manual sections pressing UP
        // should land on this refresh chip (it's the only focusable thing in
        // the discovery header). We keep it visible even when the list has
        // servers so users can re-scan after fixing a network/firewall.
        Button {
            store.clearError()
            discoveryTimedOut = false
            discovery.stop()
            discovery.start()
            Task {
                try? await Task.sleep(nanoseconds: 2_000_000_000)
                if discovery.servers.isEmpty {
                    discoveryTimedOut = true
                }
            }
        } label: {
            HStack(spacing: 6) {
                Image(systemName: "arrow.clockwise")
                    .font(.system(size: 13, weight: .semibold))
                Text("Refresh")
                    .font(.system(size: 13, weight: .semibold))
                    .tracking(0.4)
                    .textCase(.uppercase)
            }
            .foregroundStyle(XuvaTheme.text)
            .padding(.horizontal, 16)
            .padding(.vertical, 10)
            .background(XuvaTheme.surface.opacity(0.74), in: Capsule(style: .continuous))
            .overlay(Capsule(style: .continuous).stroke(XuvaTheme.hairline))
        }
        .buttonStyle(.plain)
        .xuvaFocused(radius: 999)
    }

    @ViewBuilder
    private func discoveryRow(server: DiscoveredServer, viewport: CGSize) -> some View {
        Button {
            Task { await store.selectDiscoveredServer(server) }
        } label: {
            HStack(spacing: 16) {
                ZStack {
                    RoundedRectangle(cornerRadius: 12, style: .continuous)
                        .fill(XuvaTheme.action.opacity(0.18))
                    Image(systemName: "tv.fill")
                        .font(.system(size: 20, weight: .semibold))
                        .foregroundStyle(XuvaTheme.focus)
                }
                .frame(width: 44, height: 44)

                VStack(alignment: .leading, spacing: 2) {
                    Text(server.name)
                        .font(.system(size: XuvaScale.bodyFontSize(viewport), weight: .semibold))
                        .foregroundStyle(XuvaTheme.text)
                    Text(server.baseURL.absoluteString)
                        .font(.system(size: XuvaScale.metaFontSize(viewport) - 2, design: .monospaced))
                        .foregroundStyle(XuvaTheme.mutedText)
                        .lineLimit(1)
                }
                Spacer()
                Image(systemName: "chevron.right")
                    .foregroundStyle(XuvaTheme.mutedText)
            }
            .padding(.horizontal, 18)
            .padding(.vertical, 14)
            .background(XuvaTheme.surface.opacity(0.74), in: RoundedRectangle(cornerRadius: 18, style: .continuous))
            .overlay(RoundedRectangle(cornerRadius: 18, style: .continuous).stroke(XuvaTheme.hairline))
        }
        .buttonStyle(.plain)
        .xuvaFocused(radius: 18)
    }

    @ViewBuilder
    private func emptyDiscoveryHint(viewport: CGSize) -> some View {
        VStack(alignment: .leading, spacing: 6) {
            Text(discoveryTimedOut ? "Couldn't find Xuva on this network" : "Looking for Xuva servers on this network…")
                .font(.system(size: XuvaScale.bodyFontSize(viewport)))
                .foregroundStyle(XuvaTheme.muted)
            Text(discoveryTimedOut
                 ? "Your network may block local discovery. Enter your server's address below to continue."
                 : "Make sure your Xuva server is running and connected to the same network.")
                .font(.system(size: XuvaScale.metaFontSize(viewport)))
                .foregroundStyle(XuvaTheme.mutedText)
        }
        .padding(20)
        .frame(maxWidth: .infinity, alignment: .leading)
        .background(XuvaTheme.surface.opacity(0.55), in: RoundedRectangle(cornerRadius: 16, style: .continuous))
        .overlay(RoundedRectangle(cornerRadius: 16, style: .continuous).stroke(XuvaTheme.hairline))
    }

    // MARK: – Manual URL (collapsed by default)

    @ViewBuilder
    private func manualSection(viewport: CGSize) -> some View {
        VStack(alignment: .leading, spacing: 12) {
            Button {
                showManualEntry.toggle()
            } label: {
                HStack(spacing: 8) {
                    Image(systemName: showManualEntry ? "chevron.down" : "chevron.right")
                        .font(.system(size: 13, weight: .semibold))
                    Text(showManualEntry ? "Hide manual address" : "Enter address manually")
                        .font(.system(size: XuvaScale.metaFontSize(viewport), weight: .medium))
                }
                .foregroundStyle(XuvaTheme.mutedText)
            }
            .buttonStyle(.plain)

            if showManualEntry {
                VStack(alignment: .leading, spacing: 12) {
                    serverURLControl(viewport: viewport)
                    Button {
                        Task {
                            await store.connect()
                            if store.errorMessage == nil { await store.startPairing() }
                        }
                    } label: {
                        buttonLabel(title: store.isBusy ? "Connecting…" : "Connect", systemImage: store.isBusy ? "hourglass" : "play.fill")
                    }
                    .buttonStyle(XuvaPrimaryButtonStyle(viewport: viewport))
                    .xuvaDefaultKeyboardAction()
                    .disabled(store.isBusy)
                }
                .padding(.top, 4)
            }
        }
        .frame(maxWidth: XuvaScale.heroContentMaxWidth(viewport), alignment: .leading)
    }

    private func serverURLControl(viewport: CGSize) -> some View {
        TextField("http://10.0.0.x:8097", text: $store.serverText)
            .autocorrectionDisabled()
            .textInputAutocapitalization(.never)
            .keyboardType(.URL)
            .textContentType(.URL)
            .submitLabel(.go)
            .onSubmit {
                Task {
                    await store.connect()
                    if store.errorMessage == nil { await store.startPairing() }
                }
            }
            .font(.system(size: XuvaScale.bodyFontSize(viewport), weight: .medium))
            .foregroundStyle(XuvaTheme.text)
            .padding(.horizontal, 22)
            .frame(height: XuvaScale.buttonHeight(viewport))
            .background(Color.white.opacity(0.06), in: Capsule(style: .continuous))
            .overlay(Capsule(style: .continuous).stroke(XuvaTheme.hairline))
    }

    // MARK: – Pairing card

    private func pairingCard(viewport: CGSize) -> some View {
        VStack(alignment: .leading, spacing: 20) {
            Label(store.bootstrap?.server?.name ?? "Xuva connected", systemImage: "checkmark.seal.fill")
                .font(.system(size: XuvaScale.bodyFontSize(viewport), weight: .bold))
                .foregroundStyle(XuvaTheme.good)

            if let code = store.pairing?.code {
                VStack(alignment: .leading, spacing: 8) {
                    Text("Pairing code")
                        .font(.system(size: XuvaScale.eyebrowFontSize(viewport), weight: .semibold))
                        .tracking(XuvaScale.eyebrowFontSize(viewport) * 0.20)
                        .textCase(.uppercase)
                        .foregroundStyle(XuvaTheme.mutedText)
                    Text(code)
                        .font(.system(size: XuvaScale.heroTitleSize(viewport) * 0.85, weight: .black, design: .rounded))
                        .tracking(10)
                        .foregroundStyle(XuvaTheme.text)
                }
                Text("Open Xuva on your computer → Settings → Devices → Approve.")
                    .font(.system(size: XuvaScale.bodyFontSize(viewport)))
                    .foregroundStyle(XuvaTheme.muted)
                HStack(spacing: 10) {
                    ProgressView()
                        #if !os(tvOS)
                        .controlSize(.small)
                        #endif
                        .tint(XuvaTheme.primaryGlow)
                    Text("Waiting for approval…")
                        .font(.system(size: XuvaScale.metaFontSize(viewport), weight: .semibold))
                        .foregroundStyle(XuvaTheme.primaryGlow)
                }
            } else {
                Text("Generating pairing code…")
                    .font(.system(size: XuvaScale.bodyFontSize(viewport)))
                    .foregroundStyle(XuvaTheme.muted)
            }

            HStack(spacing: 12) {
                if store.pairing != nil {
                    Button {
                        Task { await store.pollPairingOnce() }
                    } label: {
                        buttonLabel(title: "Check approval", systemImage: "arrow.clockwise")
                    }
                    .buttonStyle(XuvaSecondaryButtonStyle(viewport: viewport))
                    .disabled(store.isBusy)
                }
                Button {
                    store.resetConnection()
                } label: {
                    buttonLabel(title: "Pick a different server", systemImage: "arrow.counterclockwise")
                }
                .buttonStyle(XuvaSecondaryButtonStyle(viewport: viewport))
                .disabled(store.isBusy)
            }
        }
        .padding(32)
        .frame(maxWidth: XuvaScale.heroContentMaxWidth(viewport), alignment: .leading)
        .background(XuvaTheme.surface.opacity(0.78), in: RoundedRectangle(cornerRadius: 24, style: .continuous))
        .overlay(RoundedRectangle(cornerRadius: 24, style: .continuous).stroke(XuvaTheme.hairline))
    }

    @ViewBuilder
    private func connectionHint(viewport: CGSize) -> some View {
        if store.connectionState == .needsAuthCredential {
            Label(
                "Saved access was rejected. Reset and pair again.",
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

    // MARK: – QR / code entry

    @ViewBuilder
    private func qrSection(viewport: CGSize) -> some View {
        VStack(alignment: .leading, spacing: 12) {
            Button {
                showQREntry.toggle()
                if !showQREntry { qrCodeText = "" }
            } label: {
                HStack(spacing: 8) {
                    Image(systemName: showQREntry ? "chevron.down" : "chevron.right")
                        .font(.system(size: 13, weight: .semibold))
                    Text(showQREntry ? "Hide QR / code pairing" : "Use QR code or pair code")
                        .font(.system(size: XuvaScale.metaFontSize(viewport), weight: .medium))
                }
                .foregroundStyle(XuvaTheme.mutedText)
            }
            .buttonStyle(.plain)

            if showQREntry {
                VStack(alignment: .leading, spacing: 12) {
                    #if !os(tvOS)
                    Button {
                        showQRScanner = true
                    } label: {
                        buttonLabel(title: "Scan QR code", systemImage: "qrcode.viewfinder")
                    }
                    .buttonStyle(XuvaPrimaryButtonStyle(viewport: viewport))
                    #endif

                    VStack(alignment: .leading, spacing: 8) {
                        Text("Or enter the 8-character pair code")
                            .font(.system(size: XuvaScale.metaFontSize(viewport)))
                            .foregroundStyle(XuvaTheme.muted)
                        HStack(spacing: 10) {
                            TextField("XXXXXXXX", text: $qrCodeText)
                                .autocorrectionDisabled()
                                .textInputAutocapitalization(.characters)
                                .font(.system(size: XuvaScale.bodyFontSize(viewport), weight: .medium, design: .monospaced))
                                .foregroundStyle(XuvaTheme.text)
                                .padding(.horizontal, 22)
                                .frame(height: XuvaScale.buttonHeight(viewport))
                                .background(Color.white.opacity(0.06), in: Capsule(style: .continuous))
                                .overlay(Capsule(style: .continuous).stroke(XuvaTheme.hairline))
                            Button {
                                Task { await store.claimQRToken(qrCodeText.trimmingCharacters(in: .whitespacesAndNewlines)) }
                            } label: {
                                buttonLabel(title: store.isBusy ? "Pairing…" : "Pair", systemImage: store.isBusy ? "hourglass" : "link")
                            }
                            .buttonStyle(XuvaPrimaryButtonStyle(viewport: viewport))
                            .disabled(qrCodeText.trimmingCharacters(in: .whitespacesAndNewlines).count < 6 || store.isBusy)
                        }
                    }
                }
                .padding(.top, 4)
            }
        }
        .frame(maxWidth: XuvaScale.heroContentMaxWidth(viewport), alignment: .leading)
    }

    #if !os(tvOS)
    private func handleScannedURL(_ urlString: String) {
        // Expected format: http[s]://host/api/pairing/qr/{TOKEN}/claim
        guard let url = URL(string: urlString),
              let host = url.host,
              url.path.contains("/api/pairing/qr/") else { return }
        let components = url.path.components(separatedBy: "/")
        guard let tokenIdx = components.firstIndex(of: "qr"), tokenIdx + 1 < components.count else { return }
        let token = components[tokenIdx + 1]
        guard !token.isEmpty else { return }
        // Set the server URL from the scanned QR first, then claim
        let scheme = url.scheme ?? "http"
        let port = url.port.map { ":\($0)" } ?? ""
        store.serverText = "\(scheme)://\(host)\(port)"
        Task {
            await store.connect()
            guard store.errorMessage == nil else { return }
            await store.claimQRToken(token)
        }
    }
    #endif

    @ViewBuilder
    private func versionLabel(viewport: CGSize) -> some View {
        let appVersion = (Bundle.main.infoDictionary?["CFBundleShortVersionString"] as? String) ?? "?"
        let buildNum   = (Bundle.main.infoDictionary?["CFBundleVersion"] as? String) ?? "?"
        // CFBundleGitSHA is injected at build time by the "Stamp Build Version
        // from Git" build phase. Including the short SHA in the on-screen
        // version label is the cheapest way to confirm a deploy actually
        // landed — the user can read it off the TV and match it to a commit.
        let sha = (Bundle.main.infoDictionary?["CFBundleGitSHA"] as? String) ?? ""
        let label = sha.isEmpty
            ? "Xuva \(appVersion) (\(buildNum))"
            : "Xuva \(appVersion) (\(buildNum) · \(sha))"
        Text(label)
            .font(.system(size: XuvaScale.metaFontSize(viewport)))
            .foregroundStyle(XuvaTheme.mutedText.opacity(0.45))
            .frame(maxWidth: XuvaScale.heroContentMaxWidth(viewport), alignment: .leading)
            #if os(tvOS)
            .onLongPressGesture(minimumDuration: 2) { showDiagLog = true }
            #else
            .onLongPressGesture { showDiagLog = true }
            #endif
    }

    private var introCopy: String {
        "Pick your Xuva server below — we'll auto-request a pairing code. Approve it once from your computer and the library opens straight to movies and shows."
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
