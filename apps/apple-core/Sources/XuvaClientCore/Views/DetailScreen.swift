import SwiftUI

public struct DetailScreen: View {
    @EnvironmentObject private var store: XuvaClientStore

    public init() {}

    public var body: some View {
        if let detail = store.selectedDetail {
            DetailContentView(detail: detail)
                .transition(.opacity)
        } else {
            VStack(spacing: 16) {
                ProgressView()
                    .controlSize(.large)
                    .tint(.white)
                Text("Loading title…")
                    .font(.system(size: XuvaScale.bodyFontSize()))
                    .foregroundStyle(XuvaTheme.muted)
            }
            .frame(maxWidth: .infinity, maxHeight: .infinity)
            .background(XuvaTheme.background)
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
            ZStack(alignment: .top) {
                backdropLayer(geometry: geometry)
                ScrollView {
                    VStack(alignment: .leading, spacing: 0) {
                        topBar
                            .padding(.top, XuvaScale.safeTop)
                        heroBlock(viewport: geometry.size)
                        Spacer().frame(height: XuvaScale.sectionSpacing)
                        bodySections(viewport: geometry.size)
                    }
                    .padding(.bottom, 120)
                }
            }
            .background(XuvaTheme.background)
        }
        .onAppear { focus = .play }
    }

    @ViewBuilder
    private func backdropLayer(geometry: GeometryProxy) -> some View {
        let height = geometry.size.height * XuvaScale.detailBackdropFraction() + 80
        ZStack {
            RemoteImage(urlString: detail.displayBackdropURL, aspectRatio: 16 / 9)
                .frame(width: geometry.size.width, height: height)
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
        .frame(width: geometry.size.width, height: height)
        .ignoresSafeArea()
    }

    private var topBar: some View {
        HStack(spacing: 16) {
            Button {
                store.backToHome()
            } label: {
                Image(systemName: "chevron.left")
            }
            .buttonStyle(XuvaIconButtonStyle())
            .focused($focus, equals: .back)

            XuvaLogo()
            Spacer()
        }
        .padding(.horizontal, XuvaScale.safeHorizontal)
        .frame(height: XuvaScale.navBarHeight())
    }

    @ViewBuilder
    private func heroBlock(viewport: CGSize) -> some View {
        let isCompact = viewport.width < 700
        VStack(alignment: .leading, spacing: 22) {
            HStack(spacing: 14) {
                Rectangle()
                    .fill(XuvaTheme.text.opacity(0.40))
                    .frame(width: 36, height: 1)
                Text(eyebrow)
                    .font(.system(size: XuvaScale.eyebrowFontSize(), weight: .semibold))
                    .tracking(5.6)
                    .foregroundStyle(XuvaTheme.mutedText)
            }

            if let logo = detail.displayLogoURL, !logo.isEmpty {
                RemoteLogo(
                    urlString: logo,
                    fallbackTitle: detail.displayTitle,
                    maxWidth: XuvaScale.heroLogoMaxWidth(viewportWidth: viewport.width),
                    maxHeight: XuvaScale.heroLogoMaxHeight(viewportWidth: viewport.width)
                )
            } else {
                Text(detail.displayTitle)
                    .font(.system(size: XuvaScale.heroTitleSize(viewportWidth: viewport.width), weight: .bold))
                    .foregroundStyle(XuvaTheme.text)
                    .lineLimit(2)
                    .minimumScaleFactor(0.6)
                    .frame(maxWidth: viewport.width * 0.6, alignment: .leading)
            }

            metaLine
            if let tagline = detail.displayTagline, !tagline.isEmpty {
                Text(tagline)
                    .italic()
                    .font(.system(size: XuvaScale.bodyFontSize() - 2))
                    .foregroundStyle(XuvaTheme.mutedText)
            }
            Text(detail.displayOverview)
                .font(.system(size: XuvaScale.bodyFontSize()))
                .foregroundStyle(XuvaTheme.secondaryText)
                .lineLimit(isCompact ? 6 : 4)
                .frame(maxWidth: viewport.width * 0.6, alignment: .leading)
            actionBar
        }
        .padding(.horizontal, XuvaScale.safeHorizontal)
        .padding(.top, viewport.height * 0.10)
        .frame(maxWidth: .infinity, alignment: .leading)
    }

    @ViewBuilder
    private var actionBar: some View {
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
            .buttonStyle(XuvaPrimaryButtonStyle())
            .focused($focus, equals: .play)

            if let trailer = detail.displayTrailerPath, !trailer.isEmpty {
                Button {
                    // Trailer playback can route through the same player using a synthetic playback
                    Task { await playTrailer() }
                } label: {
                    Label("Trailer", systemImage: "film")
                }
                .buttonStyle(XuvaSecondaryButtonStyle())
                .focused($focus, equals: .trailer)
            }

            Button {
                // Reserved for watchlist; currently no-op
            } label: {
                Image(systemName: "plus")
            }
            .buttonStyle(XuvaIconButtonStyle())
            .focused($focus, equals: .add)

            RouteBadge(decision: routeDecision)
        }
    }

    private var metaLine: some View {
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
        .font(.system(size: XuvaScale.metaFontSize(), weight: .semibold))
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
        VStack(alignment: .leading, spacing: XuvaScale.sectionSpacing) {
            if let versions = detail.versions, !versions.isEmpty {
                versionsSection(versions: versions)
            }
            if !detail.audioTracks.isEmpty || !detail.subtitleTracks.isEmpty {
                tracksSection
            }
            if !detail.displayDirectors.isEmpty {
                creditsSection
            }
        }
    }

    @ViewBuilder
    private func versionsSection(versions: [MediaVersion]) -> some View {
        SectionContainer(title: "Versions", subtitle: "Choose source quality") {
            ScrollView(.horizontal, showsIndicators: false) {
                HStack(spacing: 18) {
                    ForEach(versions, id: \.stableID) { version in
                        VersionCard(version: version, isSelected: version.stableID == selectedVersion?.stableID) {
                            selectedVersionID = version.stableID
                        }
                    }
                }
                .padding(.vertical, 12)
            }
        }
    }

    @ViewBuilder
    private var tracksSection: some View {
        SectionContainer(title: "Audio & Subtitles", subtitle: "Language and captions") {
            HStack(alignment: .top, spacing: 18) {
                TrackStack(
                    title: "Audio",
                    systemImage: "speaker.wave.2",
                    tracks: detail.audioTracks,
                    selectedTrackID: $selectedAudioID,
                    allowsNone: false
                )
                TrackStack(
                    title: "Subtitles",
                    systemImage: "captions.bubble",
                    tracks: detail.subtitleTracks,
                    selectedTrackID: $selectedSubtitleID,
                    allowsNone: true
                )
            }
        }
    }

    @ViewBuilder
    private var creditsSection: some View {
        SectionContainer(title: "Direction", subtitle: "Credits") {
            HStack(spacing: 10) {
                ForEach(detail.displayDirectors.prefix(6), id: \.self) { person in
                    MediaPill(text: person, systemImage: "person.fill", tint: XuvaTheme.secondaryText)
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
        // Trailer is a server-served MP4 path; route through the same player as a one-off URL
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
    let content: () -> Content

    init(title: String, subtitle: String, @ViewBuilder content: @escaping () -> Content) {
        self.title = title
        self.subtitle = subtitle
        self.content = content
    }

    var body: some View {
        VStack(alignment: .leading, spacing: 18) {
            VStack(alignment: .leading, spacing: 4) {
                Text(title)
                    .font(.system(size: XuvaScale.sectionTitleSize(), weight: .bold))
                    .foregroundStyle(XuvaTheme.text)
                Text(subtitle)
                    .font(.system(size: XuvaScale.metaFontSize() - 2, weight: .medium))
                    .foregroundStyle(XuvaTheme.mutedText)
            }
            content()
        }
        .padding(.horizontal, XuvaScale.safeHorizontal)
    }
}

private struct VersionCard: View {
    let version: MediaVersion
    let isSelected: Bool
    let select: () -> Void

    var body: some View {
        Button(action: select) {
            VStack(alignment: .leading, spacing: 12) {
                HStack(spacing: 12) {
                    Image(systemName: "film.stack")
                        .font(.system(size: XuvaScale.platform == .tv ? 22 : 16, weight: .semibold))
                        .foregroundStyle(XuvaTheme.action)
                    Spacer()
                    if isSelected {
                        MediaPill(text: "Selected", systemImage: "checkmark.circle.fill", tint: XuvaTheme.focus)
                    } else {
                        RouteBadge(decision: version.decision)
                    }
                }
                Text(qualityTitle)
                    .font(.system(size: XuvaScale.platform == .tv ? 24 : 17, weight: .bold))
                    .foregroundStyle(XuvaTheme.text)
                    .lineLimit(2)
                Text(subtitleParts.joined(separator: " · "))
                    .font(.system(size: XuvaScale.platform == .tv ? 18 : 13))
                    .foregroundStyle(XuvaTheme.muted)
                    .lineLimit(2)
                Spacer(minLength: 0)
                if let bitrate = version.displayBitrate {
                    Text(bitrate)
                        .font(.system(size: XuvaScale.platform == .tv ? 16 : 12, weight: .semibold))
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
        #if os(tvOS)
        return 460
        #else
        return 300
        #endif
    }

    private var cardHeight: CGFloat {
        #if os(tvOS)
        return 220
        #else
        return 178
        #endif
    }
}

private struct TrackStack: View {
    let title: String
    let systemImage: String
    let tracks: [MediaTrack]
    @Binding var selectedTrackID: String?
    let allowsNone: Bool

    var body: some View {
        VStack(alignment: .leading, spacing: 12) {
            Label(title, systemImage: systemImage)
                .font(.system(size: XuvaScale.platform == .tv ? 24 : 17, weight: .bold))
                .foregroundStyle(XuvaTheme.text)
            if allowsNone {
                TrackButton(title: "Off", subtitle: "No subtitle track", isSelected: selectedTrackID == nil) {
                    selectedTrackID = nil
                }
            }
            if tracks.isEmpty {
                Text("No tracks available")
                    .font(.system(size: XuvaScale.platform == .tv ? 18 : 14))
                    .foregroundStyle(XuvaTheme.muted)
            } else {
                ForEach(tracks.prefix(6), id: \.stableID) { track in
                    TrackButton(
                        title: track.title ?? track.language?.uppercased() ?? "Track \(track.index ?? 0)",
                        subtitle: [track.codec?.uppercased(), track.channels.map { "\($0)ch" }, track.external == true ? "External" : nil, track.default == true ? "Default" : nil, track.forced == true ? "Forced" : nil].compactMap { $0 }.joined(separator: " · "),
                        isSelected: selectedTrackID == track.stableID || (!allowsNone && selectedTrackID == nil && track.default == true)
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
    let select: () -> Void

    var body: some View {
        Button(action: select) {
            HStack(spacing: 12) {
                Circle()
                    .fill(isSelected ? XuvaTheme.focus : XuvaTheme.muted.opacity(0.42))
                    .frame(width: 9, height: 9)
                VStack(alignment: .leading, spacing: 3) {
                    Text(title)
                        .font(.system(size: XuvaScale.platform == .tv ? 20 : 15, weight: .semibold))
                        .foregroundStyle(XuvaTheme.text)
                        .lineLimit(1)
                    if !subtitle.isEmpty {
                        Text(subtitle)
                            .font(.system(size: XuvaScale.platform == .tv ? 16 : 12))
                            .foregroundStyle(XuvaTheme.muted)
                            .lineLimit(1)
                    }
                }
                Spacer()
                if isSelected {
                    Image(systemName: "checkmark")
                        .font(.system(size: XuvaScale.platform == .tv ? 18 : 13, weight: .bold))
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
