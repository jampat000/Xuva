import SwiftUI

public struct DetailScreen: View {
    @EnvironmentObject private var store: XuvaClientStore

    public init() {}

    public var body: some View {
        if let detail = store.selectedDetail {
            DetailContentView(detail: detail)
                .transition(.opacity)
        } else {
            GeometryReader { geometry in
                VStack(spacing: 16) {
                    ProgressView()
                        .controlSize(.large)
                        .tint(.white)
                    Text("Loading title…")
                        .font(.system(size: XuvaScale.bodyFontSize(geometry.size)))
                        .foregroundStyle(XuvaTheme.muted)
                }
                .frame(maxWidth: .infinity, maxHeight: .infinity)
                .background(XuvaTheme.background)
            }
        }
    }
}

private struct DetailContentView: View {
    @EnvironmentObject private var store: XuvaClientStore
    let detail: DetailResponse
    @State private var selectedVersionID: String?
    @State private var selectedAudioID: String?
    @State private var selectedSubtitleID: String?
    @FocusState private var focus: DetailFocus?

    var body: some View {
        GeometryReader { geometry in
            let viewport = geometry.size
            ZStack(alignment: .top) {
                backdropLayer(viewport: viewport)
                ScrollView {
                    VStack(alignment: .leading, spacing: 0) {
                        topBar(viewport: viewport)
                            .padding(.top, XuvaScale.safeTop(viewport))
                        heroBlock(viewport: viewport)
                        Spacer().frame(height: XuvaScale.sectionSpacing(viewport))
                        bodySections(viewport: viewport)
                    }
                    .padding(.bottom, viewport.height * 0.12)
                }
            }
            .background(XuvaTheme.background)
        }
        .defaultFocus($focus, .play)
        .onAppear {
            focus = .play
            if UserDefaults.standard.bool(forKey: "xuva.dev.autoPlayOnDetail") {
                Task {
                    try? await Task.sleep(nanoseconds: 800_000_000)
                    await store.play()
                }
            }
        }
    }

    @ViewBuilder
    private func backdropLayer(viewport: CGSize) -> some View {
        let height = viewport.height * XuvaScale.detailBackdropFraction(viewport) + 80
        ZStack {
            RemoteImage(urlString: detail.displayBackdropURL, aspectRatio: 16 / 9)
                .frame(width: viewport.width, height: height)
                .clipped()
                .opacity(0.5)
            LinearGradient(
                colors: [.clear, XuvaTheme.background.opacity(0.75), XuvaTheme.background],
                startPoint: .top,
                endPoint: .bottom
            )
            LinearGradient(
                colors: [XuvaTheme.background.opacity(0.85), XuvaTheme.background.opacity(0.20), .clear],
                startPoint: .leading,
                endPoint: .trailing
            )
        }
        .frame(width: viewport.width, height: height)
        .ignoresSafeArea()
    }

    private func topBar(viewport: CGSize) -> some View {
        HStack(spacing: 16) {
            Button {
                store.backToHome()
            } label: {
                Image(systemName: "chevron.left")
            }
            .buttonStyle(XuvaIconButtonStyle(viewport: viewport))
            .focused($focus, equals: .back)

            XuvaLogo(viewport: viewport)
            Spacer()
        }
        .padding(.horizontal, XuvaScale.safeHorizontal(viewport))
        .frame(height: XuvaScale.navBarHeight(viewport))
    }

    @ViewBuilder
    private func heroBlock(viewport: CGSize) -> some View {
        let isCompact = viewport.width < 600
        VStack(alignment: .leading, spacing: isCompact ? 14 : 22) {
            HStack(spacing: 14) {
                Rectangle()
                    .fill(XuvaTheme.text.opacity(0.40))
                    .frame(width: 36, height: 1)
                Text(eyebrow)
                    .font(.system(size: XuvaScale.eyebrowFontSize(viewport), weight: .semibold))
                    .tracking(5.6)
                    .foregroundStyle(XuvaTheme.mutedText)
            }

            if let logo = detail.displayLogoURL, !logo.isEmpty {
                RemoteLogo(
                    urlString: logo,
                    fallbackTitle: detail.displayTitle,
                    maxWidth: XuvaScale.heroLogoMaxWidth(viewport),
                    maxHeight: XuvaScale.heroLogoMaxHeight(viewport)
                )
            } else {
                Text(detail.displayTitle)
                    .font(.system(size: XuvaScale.heroTitleSize(viewport), weight: .bold))
                    .foregroundStyle(XuvaTheme.text)
                    .lineLimit(2)
                    .minimumScaleFactor(0.6)
                    .frame(maxWidth: XuvaScale.heroContentMaxWidth(viewport), alignment: .leading)
            }

            metaLine(viewport: viewport)
            if let tagline = detail.displayTagline, !tagline.isEmpty {
                Text(tagline)
                    .italic()
                    .font(.system(size: XuvaScale.bodyFontSize(viewport) - 2))
                    .foregroundStyle(XuvaTheme.mutedText)
            }
            Text(detail.displayOverview)
                .font(.system(size: XuvaScale.bodyFontSize(viewport)))
                .foregroundStyle(XuvaTheme.secondaryText)
                .lineLimit(isCompact ? 6 : 4)
                .frame(maxWidth: XuvaScale.heroContentMaxWidth(viewport), alignment: .leading)
            actionBar(viewport: viewport)
        }
        .padding(.horizontal, XuvaScale.safeHorizontal(viewport))
        .padding(.top, viewport.height * 0.06)
        .frame(maxWidth: .infinity, alignment: .leading)
    }

    @ViewBuilder
    private func actionBar(viewport: CGSize) -> some View {
        HStack(spacing: 14) {
            Button {
                Task {
                    await store.play(
                        version: selectedVersion,
                        audioTrack: selectedAudioTrack,
                        subtitleTrack: selectedSubtitleTrack
                    )
                }
            } label: {
                Label("Play", systemImage: "play.fill")
            }
            .buttonStyle(XuvaPrimaryButtonStyle(viewport: viewport))
            .focused($focus, equals: .play)

            if let trailer = detail.displayTrailerPath, !trailer.isEmpty {
                Button {
                    Task { await playTrailer() }
                } label: {
                    Label("Trailer", systemImage: "film")
                }
                .buttonStyle(XuvaSecondaryButtonStyle(viewport: viewport))
                .focused($focus, equals: .trailer)
            }

            Button {
                // Watchlist placeholder
            } label: {
                Image(systemName: "plus")
            }
            .buttonStyle(XuvaIconButtonStyle(viewport: viewport))
            .focused($focus, equals: .add)

            RouteBadge(decision: routeDecision, viewport: viewport)
        }
    }

    private func metaLine(viewport: CGSize) -> some View {
        HStack(spacing: 12) {
            ForEach(metaParts, id: \.self) { part in
                Text(part)
                if part != metaParts.last {
                    Circle()
                        .fill(XuvaTheme.muted.opacity(0.45))
                        .frame(width: 5, height: 5)
                }
            }
        }
        .font(.system(size: XuvaScale.metaFontSize(viewport), weight: .semibold))
        .foregroundStyle(XuvaTheme.secondaryText)
    }

    private var metaParts: [String] {
        var parts: [String] = []
        if let year = detail.displayYear { parts.append(String(year)) }
        if let runtime = detail.displayRuntime { parts.append(runtime) }
        if let rating = detail.displayRating, rating > 0 { parts.append(String(format: "★ %.1f", rating)) }
        if let cr = detail.displayContentRating, !cr.isEmpty { parts.append(cr) }
        let genres = detail.displayGenres.prefix(2).joined(separator: " / ")
        if !genres.isEmpty { parts.append(genres) }
        return parts
    }

    @ViewBuilder
    private func bodySections(viewport: CGSize) -> some View {
        VStack(alignment: .leading, spacing: XuvaScale.sectionSpacing(viewport)) {
            if let versions = detail.versions, !versions.isEmpty {
                versionsSection(versions: versions, viewport: viewport)
            }
            if !detail.audioTracks.isEmpty || !detail.subtitleTracks.isEmpty {
                tracksSection(viewport: viewport)
            }
            if !detail.displayDirectors.isEmpty {
                creditsSection(viewport: viewport)
            }
        }
    }

    @ViewBuilder
    private func versionsSection(versions: [MediaVersion], viewport: CGSize) -> some View {
        SectionContainer(title: "Versions", subtitle: "Choose source quality", viewport: viewport) {
            ScrollView(.horizontal, showsIndicators: false) {
                HStack(spacing: 18) {
                    ForEach(versions, id: \.stableID) { version in
                        VersionCard(version: version, viewport: viewport, isSelected: version.stableID == selectedVersion?.stableID) {
                            selectedVersionID = version.stableID
                        }
                    }
                }
                .padding(.vertical, 12)
            }
        }
    }

    @ViewBuilder
    private func tracksSection(viewport: CGSize) -> some View {
        SectionContainer(title: "Audio & Subtitles", subtitle: "Language and captions", viewport: viewport) {
            HStack(alignment: .top, spacing: 18) {
                TrackStack(
                    title: "Audio",
                    systemImage: "speaker.wave.2",
                    tracks: detail.audioTracks,
                    selectedTrackID: $selectedAudioID,
                    allowsNone: false,
                    viewport: viewport
                )
                TrackStack(
                    title: "Subtitles",
                    systemImage: "captions.bubble",
                    tracks: detail.subtitleTracks,
                    selectedTrackID: $selectedSubtitleID,
                    allowsNone: true,
                    viewport: viewport
                )
            }
        }
    }

    @ViewBuilder
    private func creditsSection(viewport: CGSize) -> some View {
        SectionContainer(title: "Direction", subtitle: "Credits", viewport: viewport) {
            HStack(spacing: 10) {
                ForEach(detail.displayDirectors.prefix(6), id: \.self) { person in
                    MediaPill(text: person, systemImage: "person.fill", tint: XuvaTheme.secondaryText, viewport: viewport)
                }
            }
        }
    }

    private var routeDecision: PlaybackDecision? {
        selectedVersion?.decision ?? detail.versions?.first?.decision
    }

    private var selectedVersion: MediaVersion? {
        let versions = detail.versions ?? []
        if let selectedVersionID, let version = versions.first(where: { $0.stableID == selectedVersionID }) {
            return version
        }
        return versions.first
    }

    private var selectedAudioTrack: MediaTrack? {
        let tracks = detail.audioTracks
        if let selectedAudioID, let track = tracks.first(where: { $0.stableID == selectedAudioID }) {
            return track
        }
        return tracks.first(where: { $0.default == true }) ?? tracks.first
    }

    private var selectedSubtitleTrack: MediaTrack? {
        guard let selectedSubtitleID else { return nil }
        return detail.subtitleTracks.first(where: { $0.stableID == selectedSubtitleID })
    }

    private var eyebrow: String {
        if let kind = detail.kind?.lowercased() {
            if kind.contains("series") { return "TV SERIES" }
            if kind.contains("episode") { return "EPISODE" }
        }
        return "FEATURE FILM"
    }

    private func playTrailer() async {
        guard let trailer = detail.displayTrailerPath, !trailer.isEmpty,
              let api = store.api,
              let url = api.resolvedURL(trailer) else { return }
        store.playback = PlaybackStartResponse(
            sessionId: nil,
            heartbeatUrl: nil,
            stopUrl: nil,
            playbackStateUrl: nil,
            heartbeatIntervalMs: nil,
            decision: nil,
            route: PlaybackRoute(url: url.absoluteString, manifestUrl: nil, protocolValue: "direct", route: "trailer", status: "ready", decision: nil),
            mediaSourceId: nil,
            deviceId: nil,
            defaultSubtitlesEnabled: false
        )
        store.screen = .player
    }
}

private enum DetailFocus: Hashable {
    case back
    case play
    case trailer
    case add
    case version(String)
}

private struct SectionContainer<Content: View>: View {
    let title: String
    let subtitle: String
    let viewport: CGSize
    let content: () -> Content

    init(title: String, subtitle: String, viewport: CGSize, @ViewBuilder content: @escaping () -> Content) {
        self.title = title
        self.subtitle = subtitle
        self.viewport = viewport
        self.content = content
    }

    var body: some View {
        VStack(alignment: .leading, spacing: 18) {
            VStack(alignment: .leading, spacing: 4) {
                Text(title)
                    .font(.system(size: XuvaScale.sectionTitleSize(viewport), weight: .bold))
                    .foregroundStyle(XuvaTheme.text)
                Text(subtitle)
                    .font(.system(size: XuvaScale.metaFontSize(viewport) - 2, weight: .medium))
                    .foregroundStyle(XuvaTheme.mutedText)
            }
            content()
        }
        .padding(.horizontal, XuvaScale.safeHorizontal(viewport))
    }
}

private struct VersionCard: View {
    let version: MediaVersion
    let viewport: CGSize
    let isSelected: Bool
    let select: () -> Void

    var body: some View {
        Button(action: select) {
            VStack(alignment: .leading, spacing: 12) {
                HStack(spacing: 12) {
                    Image(systemName: "film.stack")
                        .font(.system(size: XuvaScale.bodyFontSize(viewport) * 0.9, weight: .semibold))
                        .foregroundStyle(XuvaTheme.action)
                    Spacer()
                    if isSelected {
                        MediaPill(text: "Selected", systemImage: "checkmark.circle.fill", tint: XuvaTheme.focus, viewport: viewport)
                    } else {
                        RouteBadge(decision: version.decision, viewport: viewport)
                    }
                }
                Text(qualityTitle)
                    .font(.system(size: XuvaScale.bodyFontSize(viewport) + 2, weight: .bold))
                    .foregroundStyle(XuvaTheme.text)
                    .lineLimit(2)
                Text(subtitleParts.joined(separator: " · "))
                    .font(.system(size: XuvaScale.metaFontSize(viewport) - 2))
                    .foregroundStyle(XuvaTheme.muted)
                    .lineLimit(2)
                Spacer(minLength: 0)
                if let bitrate = version.displayBitrate {
                    Text(bitrate)
                        .font(.system(size: XuvaScale.metaFontSize(viewport) - 4, weight: .semibold))
                        .foregroundStyle(XuvaTheme.mutedText)
                }
            }
            .padding(22)
            .frame(width: cardWidth, height: cardHeight, alignment: .topLeading)
            .background(isSelected ? XuvaTheme.focus.opacity(0.12) : XuvaTheme.elevated.opacity(0.74), in: RoundedRectangle(cornerRadius: 18, style: .continuous))
            .overlay(RoundedRectangle(cornerRadius: 18, style: .continuous).stroke(isSelected ? XuvaTheme.focus.opacity(0.42) : XuvaTheme.hairline))
        }
        .buttonStyle(.plain)
        .xuvaFocused(radius: 18)
    }

    private var qualityTitle: String {
        if let q = version.qualityLabel, !q.isEmpty { return q }
        if let res = version.displayResolution { return res }
        return version.name ?? "Original"
    }

    private var subtitleParts: [String] {
        [version.displayResolution, version.displayVideoCodec, version.displayAudioSummary, version.displayDuration]
            .compactMap { $0 }
            .filter { !$0.isEmpty }
    }

    private var cardWidth: CGFloat {
        XuvaScale.clamped(280, viewport.width * 0.28, 540)
    }

    private var cardHeight: CGFloat {
        XuvaScale.clamped(168, viewport.height * 0.20, 240)
    }
}

private struct TrackStack: View {
    let title: String
    let systemImage: String
    let tracks: [MediaTrack]
    @Binding var selectedTrackID: String?
    let allowsNone: Bool
    let viewport: CGSize

    var body: some View {
        VStack(alignment: .leading, spacing: 12) {
            Label(title, systemImage: systemImage)
                .font(.system(size: XuvaScale.bodyFontSize(viewport) + 2, weight: .bold))
                .foregroundStyle(XuvaTheme.text)
            if allowsNone {
                TrackButton(title: "Off", subtitle: "No subtitle track", isSelected: selectedTrackID == nil, viewport: viewport) {
                    selectedTrackID = nil
                }
            }
            if tracks.isEmpty {
                Text("No tracks available")
                    .font(.system(size: XuvaScale.metaFontSize(viewport)))
                    .foregroundStyle(XuvaTheme.muted)
            } else {
                ForEach(tracks.prefix(6), id: \.stableID) { track in
                    TrackButton(
                        title: track.title ?? track.language?.uppercased() ?? "Track \(track.index ?? 0)",
                        subtitle: [track.codec?.uppercased(), track.channels.map { "\($0)ch" }, track.external == true ? "External" : nil, track.default == true ? "Default" : nil, track.forced == true ? "Forced" : nil].compactMap { $0 }.joined(separator: " · "),
                        isSelected: selectedTrackID == track.stableID || (!allowsNone && selectedTrackID == nil && track.default == true),
                        viewport: viewport
                    ) {
                        selectedTrackID = track.stableID
                    }
                }
            }
        }
        .padding(24)
        .frame(maxWidth: 520, minHeight: 180, alignment: .topLeading)
        .background(XuvaTheme.elevated.opacity(0.62), in: RoundedRectangle(cornerRadius: 18, style: .continuous))
        .overlay(RoundedRectangle(cornerRadius: 18, style: .continuous).stroke(XuvaTheme.hairline))
    }
}

private struct TrackButton: View {
    let title: String
    let subtitle: String
    let isSelected: Bool
    let viewport: CGSize
    let select: () -> Void

    var body: some View {
        Button(action: select) {
            HStack(spacing: 12) {
                Circle()
                    .fill(isSelected ? XuvaTheme.focus : XuvaTheme.muted.opacity(0.42))
                    .frame(width: 9, height: 9)
                VStack(alignment: .leading, spacing: 3) {
                    Text(title)
                        .font(.system(size: XuvaScale.bodyFontSize(viewport) - 2, weight: .semibold))
                        .foregroundStyle(XuvaTheme.text)
                        .lineLimit(1)
                    if !subtitle.isEmpty {
                        Text(subtitle)
                            .font(.system(size: XuvaScale.metaFontSize(viewport) - 2))
                            .foregroundStyle(XuvaTheme.muted)
                            .lineLimit(1)
                    }
                }
                Spacer()
                if isSelected {
                    Image(systemName: "checkmark")
                        .font(.system(size: XuvaScale.metaFontSize(viewport), weight: .bold))
                        .foregroundStyle(XuvaTheme.focus)
                }
            }
            .padding(.horizontal, 14)
            .padding(.vertical, 12)
            .background(isSelected ? XuvaTheme.focus.opacity(0.10) : Color.white.opacity(0.035), in: RoundedRectangle(cornerRadius: 12, style: .continuous))
            .overlay(RoundedRectangle(cornerRadius: 12, style: .continuous).stroke(isSelected ? XuvaTheme.focus.opacity(0.34) : Color.clear))
        }
        .buttonStyle(.plain)
        .xuvaFocused(radius: 12)
    }
}
