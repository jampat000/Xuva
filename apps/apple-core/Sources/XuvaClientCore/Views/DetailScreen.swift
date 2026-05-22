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
    @EnvironmentObject private var watchlist: XuvaWatchlist
    let detail: DetailResponse
    @State private var selectedVersionID: String?
    @State private var selectedAudioID: String?
    @State private var selectedSubtitleID: String?
    @State private var selectedSeasonNumber: Int?

    var body: some View {
        GeometryReader { geometry in
            let viewport = geometry.size
            ScrollView {
                VStack(alignment: .leading, spacing: 0) {
                    // Back button is first focusable item — no spacer so tvOS
                    // focus engine can reach it immediately on load.
                    backLink(viewport: viewport)
                        .padding(.horizontal, XuvaScale.safeHorizontal(viewport))
                        .padding(.top, topBarPadding(viewport))
                        .padding(.bottom, 24)
                        .focusSection()
                    twoColumn(viewport: viewport)
                        .padding(.horizontal, XuvaScale.safeHorizontal(viewport))
                        .padding(.bottom, viewport.height * 0.10)
                        .focusSection()
                }
            }
            // Backdrop is a fixed (non-scrolling) background layer aligned to
            // the top — visually identical to the old ZStack approach but the
            // focus engine now sees only the ScrollView, not a nested ZStack.
            .background(alignment: .top) { backdrop(viewport: viewport) }
            .background(XuvaTheme.background)
        }
        .ignoresSafeArea(.container, edges: .top)
        .onAppear {
            if UserDefaults.standard.bool(forKey: "xuva.dev.autoPlayOnDetail") {
                Task {
                    try? await Task.sleep(nanoseconds: 800_000_000)
                    await store.play()
                }
            }
        }
        #if os(tvOS)
        .onExitCommand { store.backToHome() }
        #endif
    }

    private func topBarPadding(_ viewport: CGSize) -> CGFloat {
        #if os(tvOS)
        return max(viewport.height * 0.07, 64)
        #else
        return 60
        #endif
    }

    // MARK: – Backdrop (top 60vh with 3 gradients, mirrors web)

    @ViewBuilder
    private func backdrop(viewport: CGSize) -> some View {
        let height = max(viewport.height * 0.60, 480)
        ZStack {
            RemoteImage(urlString: detail.displayBackdropURL, aspectRatio: 16 / 9)
                .frame(width: viewport.width, height: height)
                .clipped()
            // right→transparent fade (matches `from-background via-background/70 to-transparent`)
            LinearGradient(
                colors: [XuvaTheme.background, XuvaTheme.background.opacity(0.70), .clear],
                startPoint: .leading,
                endPoint: .trailing
            )
            // bottom→clear fade (matches `from-background to-transparent`)
            LinearGradient(
                colors: [XuvaTheme.background, .clear],
                startPoint: .bottom,
                endPoint: .center
            )
            // top edge subtle dark (matches `from-background/80 to-transparent`)
            LinearGradient(
                colors: [XuvaTheme.background.opacity(0.80), .clear],
                startPoint: .top,
                endPoint: UnitPoint(x: 0.5, y: 0.2)
            )
        }
        .frame(width: viewport.width, height: height)
        .frame(maxWidth: .infinity, alignment: .top)
    }

    // MARK: – Back link

    @ViewBuilder
    private func backLink(viewport: CGSize) -> some View {
        Button {
            store.backToHome()
        } label: {
            HStack(spacing: 6) {
                Image(systemName: "chevron.left")
                Text(backTitle)
                    .font(.system(size: XuvaScale.metaFontSize(viewport), weight: .medium))
            }
            .foregroundStyle(XuvaTheme.mutedText)
        }
        .buttonStyle(XuvaNakedButtonStyle())
        .xuvaFocused(radius: 8)
    }

    private var backTitle: String {
        if let kind = detail.kind?.lowercased(), kind.contains("series") { return "Back to TV" }
        return "Back to Movies"
    }

    // MARK: – Two-column hero (poster | content)

    @ViewBuilder
    private func twoColumn(viewport: CGSize) -> some View {
        let isCompact = viewport.width < 700
        let posterW: CGFloat = isCompact ? viewport.width * 0.35 : XuvaScale.clamped(220, viewport.width * 0.175, 360)
        let posterH = posterW * 1.5
        // Explicit branches instead of AnyLayout — AnyLayout is opaque to the
        // tvOS focus engine so swipe-down from the back button can't reach the
        // Play button inside contentColumn.
        if isCompact {
            VStack(alignment: .leading, spacing: 28) {
                poster(width: posterW, height: posterH)
                contentColumn(viewport: viewport)
            }
        } else {
            HStack(alignment: .top, spacing: viewport.width * 0.035) {
                poster(width: posterW, height: posterH)
                contentColumn(viewport: viewport)
            }
        }
    }

    @ViewBuilder
    private func poster(width: CGFloat, height: CGFloat) -> some View {
        ZStack {
            LinearGradient(
                colors: [
                    Color(red: 0.118, green: 0.227, blue: 0.373),
                    Color(red: 0.059, green: 0.090, blue: 0.165)
                ],
                startPoint: .topLeading,
                endPoint: .bottomTrailing
            )
            RemoteImage(urlString: detail.displayPosterURL, aspectRatio: 2 / 3)
                .frame(width: width, height: height)
                .clipped()
        }
        .frame(width: width, height: height)
        .clipShape(RoundedRectangle(cornerRadius: 18, style: .continuous))
        .shadow(color: .black.opacity(0.50), radius: 38, y: 28)
    }

    @ViewBuilder
    private func contentColumn(viewport: CGSize) -> some View {
        VStack(alignment: .leading, spacing: 0) {
            metaStrip(viewport: viewport)
                .padding(.bottom, 16)
            heroTitle(viewport: viewport)
            if let tagline = detail.displayTagline, !tagline.isEmpty {
                Text(tagline)
                    .italic()
                    .font(.system(size: XuvaScale.bodyFontSize(viewport) - 3))
                    .foregroundStyle(XuvaTheme.mutedText)
                    .padding(.top, 8)
            }
            if !detail.displayOverview.isEmpty {
                Text(detail.displayOverview)
                    .font(.system(size: XuvaScale.bodyFontSize(viewport)))
                    .foregroundStyle(XuvaTheme.secondaryText)
                    .lineLimit(6)
                    .frame(maxWidth: XuvaScale.heroContentMaxWidth(viewport), alignment: .leading)
                    .padding(.top, 22)
            }
            actionRow(viewport: viewport)
                .padding(.top, 28)
                .focusSection()
            creditsRow(viewport: viewport)
                .padding(.top, 22)
            studioChips(viewport: viewport)
                .padding(.top, 14)
            castStrip(viewport: viewport)
                .padding(.top, 36)
            sectionBody(viewport: viewport)
                .padding(.top, 32)
                .focusSection()
        }
        .frame(maxWidth: .infinity, alignment: .leading)
    }

    // MARK: – Meta strip (uppercase, 0.22em tracked)

    @ViewBuilder
    private func metaStrip(viewport: CGSize) -> some View {
        let parts = metaParts
        let size = XuvaScale.eyebrowFontSize(viewport) + 1
        HStack(alignment: .center, spacing: 10) {
            ForEach(Array(parts.enumerated()), id: \.offset) { idx, part in
                Text(part.text)
                    .font(.system(size: size, weight: .semibold))
                    .tracking(size * 0.20)
                    .foregroundStyle(part.tint)
                    .padding(.horizontal, part.boxed ? 8 : 0)
                    .padding(.vertical, part.boxed ? 4 : 0)
                    .overlay {
                        if part.boxed {
                            RoundedRectangle(cornerRadius: 4, style: .continuous)
                                .stroke(XuvaTheme.mutedText.opacity(0.55), lineWidth: 1)
                        }
                    }
                    .background(
                        part.background ?
                            RoundedRectangle(cornerRadius: 4, style: .continuous)
                                .fill(XuvaTheme.action.opacity(0.22))
                            : nil
                    )
                if idx < parts.count - 1, !part.suppressDivider, !parts[idx + 1].suppressDivider {
                    Circle()
                        .fill(XuvaTheme.mutedText.opacity(0.32))
                        .frame(width: 3, height: 3)
                }
            }
        }
        .textCase(.uppercase)
    }

    private struct MetaPart {
        let text: String
        var tint: Color = XuvaTheme.mutedText
        var boxed: Bool = false
        var background: Bool = false
        var suppressDivider: Bool = false
    }

    private var metaParts: [MetaPart] {
        var parts: [MetaPart] = []
        if let year = detail.displayYear { parts.append(MetaPart(text: String(year))) }
        for genre in detail.displayGenres.prefix(2) {
            parts.append(MetaPart(text: genre))
        }
        if let runtime = detail.displayRuntime {
            parts.append(MetaPart(text: runtime))
        }
        if let rating = detail.displayRating, rating > 0 {
            parts.append(MetaPart(text: String(format: "★ %.1f", rating), tint: Color(red: 1.0, green: 0.85, blue: 0.45)))
        }
        if let cr = detail.displayContentRating, !cr.isEmpty {
            parts.append(MetaPart(text: cr, boxed: true))
        }
        if let quality = detail.versions?.first?.qualityLabel, !quality.isEmpty {
            // Shorten "1080p H264 MOV,MP4,..." → "1080p"
            let short = quality.split(separator: " ").first.map(String.init) ?? quality
            parts.append(MetaPart(text: short, tint: XuvaTheme.primaryGlow, background: true))
        }
        return parts
    }

    // MARK: – Title (logo OR serif-display equivalent)

    @ViewBuilder
    private func heroTitle(viewport: CGSize) -> some View {
        if let logo = detail.displayLogoURL, !logo.isEmpty {
            RemoteLogo(
                urlString: logo,
                fallbackTitle: detail.displayTitle,
                maxWidth: XuvaScale.heroLogoMaxWidth(viewport),
                maxHeight: XuvaScale.heroLogoMaxHeight(viewport)
            )
        } else {
            // Web uses `font-serif-display` (Geist/Aptos/Inter Display) at
            // clamp(2rem, 5vw, 4rem) with leading-[0.95] tracking-tight.
            let titleSize = XuvaScale.clamped(40, viewport.width * 0.055, 96)
            Text(detail.displayTitle)
                .font(.system(size: titleSize, weight: .semibold, design: .default))
                .tracking(titleSize * -0.045)
                .lineSpacing(titleSize * -0.05)
                .foregroundStyle(XuvaTheme.text)
                .lineLimit(2)
                .minimumScaleFactor(0.6)
                .frame(maxWidth: XuvaScale.heroContentMaxWidth(viewport), alignment: .leading)
        }
    }

    // MARK: – Action row

    @ViewBuilder
    private func actionRow(viewport: CGSize) -> some View {
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

            if let trailer = detail.displayTrailerPath, !trailer.isEmpty {
                Button {
                    Task { await playTrailer() }
                } label: {
                    Label("Trailer", systemImage: "film")
                }
                .buttonStyle(XuvaSecondaryButtonStyle(viewport: viewport))
            }

            Button {
                guard let id = detail.item?.id, let kind = detail.kind else { return }
                _ = watchlist.toggle(
                    id: id, kind: kind,
                    title: detail.displayTitle,
                    year: detail.displayYear,
                    posterUrl: detail.displayPosterURL,
                    backdropUrl: detail.displayBackdropURL,
                    genres: detail.displayGenres
                )
            } label: {
                Image(systemName: inWatchlist ? "checkmark" : "plus")
            }
            .buttonStyle(XuvaIconButtonStyle(viewport: viewport))

            RouteBadge(decision: routeDecision, viewport: viewport)
        }
    }

    // MARK: – Credits row (DIRECTOR · WRITERS · VERSIONS)

    @ViewBuilder
    private func creditsRow(viewport: CGSize) -> some View {
        let directors = detail.displayDirectors
        let writers = detail.displayWriters
        let versionCount = detail.versions?.count ?? 0
        HStack(alignment: .top, spacing: 40) {
            if !directors.isEmpty {
                creditGroup(label: directors.count == 1 ? "Director" : "Directors",
                            values: directors,
                            viewport: viewport)
            }
            if !writers.isEmpty {
                creditGroup(label: writers.count == 1 ? "Writer" : "Writers",
                            values: Array(writers.prefix(3)),
                            viewport: viewport)
            }
            if versionCount > 0 {
                creditGroup(label: "Versions",
                            values: ["\(versionCount)"],
                            viewport: viewport)
            }
        }
    }

    // MARK: – Studio chips

    @ViewBuilder
    private func studioChips(viewport: CGSize) -> some View {
        let studios = detail.displayStudios
        if !studios.isEmpty {
            FlowLayout(spacing: 8) {
                ForEach(studios.prefix(4), id: \.self) { studio in
                    Text(studio)
                        .font(.system(size: XuvaScale.metaFontSize(viewport) - 2))
                        .foregroundStyle(XuvaTheme.mutedText)
                        .padding(.horizontal, 12)
                        .padding(.vertical, 5)
                        .background(XuvaTheme.surface.opacity(0.40), in: Capsule())
                        .overlay(Capsule().stroke(XuvaTheme.hairline))
                }
            }
        }
    }

    // MARK: – Cast strip

    @ViewBuilder
    private func castStrip(viewport: CGSize) -> some View {
        let cast = detail.displayCast
        if !cast.isEmpty {
            VStack(alignment: .leading, spacing: 18) {
                sectionTitle("Cast", viewport: viewport)
                ScrollView(.horizontal, showsIndicators: false) {
                    HStack(spacing: 14) {
                        ForEach(cast.prefix(16), id: \.id) { person in
                            CastCard(person: person, viewport: viewport)
                        }
                    }
                    .padding(.vertical, 8)
                    .focusSection()
                }
            }
        }
    }

    @ViewBuilder
    private func creditGroup(label: String, values: [String], viewport: CGSize) -> some View {
        let labelSize = XuvaScale.eyebrowFontSize(viewport)
        let valueSize = XuvaScale.metaFontSize(viewport) - 2
        VStack(alignment: .leading, spacing: 4) {
            Text(label)
                .font(.system(size: labelSize, weight: .semibold))
                .tracking(labelSize * 0.20)
                .textCase(.uppercase)
                .foregroundStyle(XuvaTheme.mutedText)
            ForEach(values, id: \.self) { v in
                Text(v)
                    .font(.system(size: valueSize, weight: .medium))
                    .foregroundStyle(XuvaTheme.text.opacity(0.82))
            }
        }
    }

    // MARK: – Sections (Versions / Audio & Subs)

    @ViewBuilder
    private func sectionBody(viewport: CGSize) -> some View {
        VStack(alignment: .leading, spacing: 36) {
            if detail.isSeries, let seasons = detail.seasons, !seasons.isEmpty {
                episodesSection(seasons: seasons, viewport: viewport)
            } else {
                if let versions = detail.versions, versions.count > 1 {
                    sectionTitle("Versions", viewport: viewport)
                    ScrollView(.horizontal, showsIndicators: false) {
                        HStack(spacing: 18) {
                            ForEach(versions, id: \.stableID) { v in
                                VersionCard(version: v, viewport: viewport, isSelected: v.stableID == selectedVersion?.stableID) {
                                    selectedVersionID = v.stableID
                                }
                            }
                        }
                        .padding(.vertical, 12)
                    }
                }
                if !detail.audioTracks.isEmpty || !detail.subtitleTracks.isEmpty {
                    sectionTitle("Audio & Subtitles", viewport: viewport)
                    HStack(alignment: .top, spacing: 18) {
                        TrackStack(title: "Audio", systemImage: "speaker.wave.2", tracks: detail.audioTracks, selectedTrackID: $selectedAudioID, allowsNone: false, viewport: viewport)
                        TrackStack(title: "Subtitles", systemImage: "captions.bubble", tracks: detail.subtitleTracks, selectedTrackID: $selectedSubtitleID, allowsNone: true, viewport: viewport)
                    }
                }
            }
        }
    }

    @ViewBuilder
    private func episodesSection(seasons: [SeasonItem], viewport: CGSize) -> some View {
        VStack(alignment: .leading, spacing: 18) {
            HStack(spacing: 16) {
                sectionTitle(currentSeason(seasons)?.displayTitle ?? "Episodes", viewport: viewport)
                Spacer()
                if seasons.count > 1 {
                    ScrollView(.horizontal, showsIndicators: false) {
                        HStack(spacing: 8) {
                            ForEach(seasons) { season in
                                SeasonChip(
                                    season: season,
                                    viewport: viewport,
                                    isSelected: season.seasonNumber == effectiveSeasonNumber(seasons)
                                ) {
                                    selectedSeasonNumber = season.seasonNumber
                                }
                            }
                        }
                    }
                    .frame(maxWidth: viewport.width * 0.45)
                }
            }
            let episodes = currentSeason(seasons)?.episodes ?? []
            if episodes.isEmpty {
                Text("No episodes available yet.")
                    .font(.system(size: XuvaScale.bodyFontSize(viewport)))
                    .foregroundStyle(XuvaTheme.muted)
            } else {
                LazyVStack(alignment: .leading, spacing: 12) {
                    ForEach(episodes) { episode in
                        EpisodeRow(episode: episode, viewport: viewport) {
                            Task { await store.playEpisode(episode) }
                        }
                    }
                }
                .focusSection()
            }
        }
    }

    private func effectiveSeasonNumber(_ seasons: [SeasonItem]) -> Int? {
        if let s = selectedSeasonNumber { return s }
        return seasons.first?.seasonNumber
    }

    private func currentSeason(_ seasons: [SeasonItem]) -> SeasonItem? {
        guard let n = effectiveSeasonNumber(seasons) else { return seasons.first }
        return seasons.first { $0.seasonNumber == n } ?? seasons.first
    }

    @ViewBuilder
    private func sectionTitle(_ title: String, viewport: CGSize) -> some View {
        let size = XuvaScale.eyebrowFontSize(viewport) + 1
        Text(title)
            .font(.system(size: size, weight: .semibold))
            .tracking(size * 0.22)
            .textCase(.uppercase)
            .foregroundStyle(XuvaTheme.mutedText)
    }

    // MARK: – Helpers

    private var inWatchlist: Bool {
        guard let id = detail.item?.id, let kind = detail.kind else { return false }
        return watchlist.isIn(id: id, kind: kind)
    }

    private var routeDecision: PlaybackDecision? {
        selectedVersion?.decision ?? detail.versions?.first?.decision
    }

    private var selectedVersion: MediaVersion? {
        let versions = detail.versions ?? []
        if let selectedVersionID, let v = versions.first(where: { $0.stableID == selectedVersionID }) {
            return v
        }
        return versions.first
    }

    private var selectedAudioTrack: MediaTrack? {
        let tracks = detail.audioTracks
        if let selectedAudioID, let t = tracks.first(where: { $0.stableID == selectedAudioID }) {
            return t
        }
        return tracks.first(where: { $0.default == true }) ?? tracks.first
    }

    private var selectedSubtitleTrack: MediaTrack? {
        guard let selectedSubtitleID else { return nil }
        return detail.subtitleTracks.first(where: { $0.stableID == selectedSubtitleID })
    }

    private func playTrailer() async {
        guard let trailer = detail.displayTrailerPath, !trailer.isEmpty,
              let api = store.api,
              let url = api.resolvedURL(trailer) else { return }
        store.playback = PlaybackStartResponse(
            sessionId: nil, heartbeatUrl: nil, stopUrl: nil, playbackStateUrl: nil,
            heartbeatIntervalMs: nil, decision: nil,
            route: PlaybackRoute(url: url.absoluteString, manifestUrl: nil, protocolValue: "direct", route: "trailer", status: "ready", decision: nil),
            mediaSourceId: nil, deviceId: nil, defaultSubtitlesEnabled: false
        )
        store.screen = .player
    }
}

private struct SeasonChip: View {
    let season: SeasonItem
    let viewport: CGSize
    let isSelected: Bool
    let action: () -> Void

    var body: some View {
        Button(action: action) {
            Text(season.displayTitle)
                .font(.system(size: XuvaScale.metaFontSize(viewport), weight: .semibold))
                .foregroundStyle(isSelected ? XuvaTheme.background : XuvaTheme.text)
                .padding(.horizontal, 16)
                .padding(.vertical, 8)
                .background(isSelected ? XuvaTheme.text : XuvaTheme.surface.opacity(0.6), in: Capsule())
                .overlay(Capsule().stroke(isSelected ? Color.clear : XuvaTheme.hairline))
        }
        .buttonStyle(.plain)
        .xuvaFocused(radius: 22)
    }
}

private struct EpisodeRow: View {
    let episode: EpisodeItem
    let viewport: CGSize
    let action: () -> Void

    var body: some View {
        let thumbW: CGFloat = XuvaScale.clamped(160, viewport.width * 0.14, 260)
        let thumbH = thumbW * 9 / 16
        Button(action: action) {
            HStack(alignment: .top, spacing: 18) {
                ZStack {
                    LinearGradient(
                        colors: [XuvaTheme.surface, XuvaTheme.elevated],
                        startPoint: .topLeading, endPoint: .bottomTrailing
                    )
                    if let url = episode.thumbnailUrl, !url.isEmpty {
                        RemoteImage(urlString: url, aspectRatio: 16 / 9)
                            .frame(width: thumbW, height: thumbH)
                            .clipped()
                    } else {
                        Image(systemName: "tv")
                            .font(.system(size: thumbW * 0.30))
                            .foregroundStyle(XuvaTheme.mutedText.opacity(0.40))
                    }
                }
                .frame(width: thumbW, height: thumbH)
                .clipShape(RoundedRectangle(cornerRadius: 12, style: .continuous))

                VStack(alignment: .leading, spacing: 6) {
                    HStack(spacing: 10) {
                        Text(episode.displayTitle)
                            .font(.system(size: XuvaScale.bodyFontSize(viewport), weight: .semibold))
                            .foregroundStyle(XuvaTheme.text)
                            .lineLimit(1)
                        if let runtime = episode.displayRuntime {
                            Text("· \(runtime)")
                                .font(.system(size: XuvaScale.metaFontSize(viewport) - 2))
                                .foregroundStyle(XuvaTheme.mutedText)
                        }
                        Spacer()
                    }
                    if let overview = episode.overview, !overview.isEmpty {
                        Text(overview)
                            .font(.system(size: XuvaScale.metaFontSize(viewport)))
                            .foregroundStyle(XuvaTheme.secondaryText)
                            .lineLimit(3)
                    }
                }
            }
            .padding(14)
            .background(XuvaTheme.surface.opacity(0.45), in: RoundedRectangle(cornerRadius: 16, style: .continuous))
            .overlay(RoundedRectangle(cornerRadius: 16, style: .continuous).stroke(XuvaTheme.hairline))
        }
        .buttonStyle(.plain)
        .xuvaFocused(radius: 16)
    }
}

private struct CastCard: View {
    let person: MetadataCredit
    let viewport: CGSize

    var body: some View {
        let cardW: CGFloat = XuvaScale.clamped(96, viewport.width * 0.075, 156)
        let cardH = cardW * 1.5
        VStack(alignment: .leading, spacing: 8) {
            ZStack {
                LinearGradient(
                    colors: [XuvaTheme.surface, XuvaTheme.elevated],
                    startPoint: .topLeading,
                    endPoint: .bottomTrailing
                )
                if let url = person.profileUrl, !url.isEmpty {
                    RemoteImage(urlString: url, aspectRatio: 2 / 3)
                        .frame(width: cardW, height: cardH)
                        .clipped()
                } else {
                    Image(systemName: "person.fill")
                        .font(.system(size: cardW * 0.40, weight: .semibold))
                        .foregroundStyle(XuvaTheme.mutedText.opacity(0.40))
                }
            }
            .frame(width: cardW, height: cardH)
            .clipShape(RoundedRectangle(cornerRadius: 12, style: .continuous))
            .overlay(RoundedRectangle(cornerRadius: 12, style: .continuous).stroke(XuvaTheme.hairline))

            VStack(alignment: .leading, spacing: 2) {
                Text(person.name ?? "")
                    .font(.system(size: XuvaScale.metaFontSize(viewport) - 3, weight: .semibold))
                    .foregroundStyle(XuvaTheme.text)
                    .lineLimit(1)
                if let character = person.character ?? person.role, !character.isEmpty {
                    Text(character)
                        .font(.system(size: XuvaScale.metaFontSize(viewport) - 4))
                        .foregroundStyle(XuvaTheme.mutedText)
                        .lineLimit(1)
                }
            }
            .frame(width: cardW, alignment: .leading)
        }
        .xuvaFocused(radius: 12)
    }
}

/// Simple flow layout for wrapping pill chips. Reproduces what Tailwind's
/// `flex flex-wrap gap-2` does on the web.
private struct FlowLayout: Layout {
    var spacing: CGFloat = 8

    func sizeThatFits(proposal: ProposedViewSize, subviews: Subviews, cache: inout ()) -> CGSize {
        let containerWidth = proposal.width ?? .infinity
        var x: CGFloat = 0
        var y: CGFloat = 0
        var rowHeight: CGFloat = 0
        var maxWidth: CGFloat = 0
        for sv in subviews {
            let size = sv.sizeThatFits(.unspecified)
            if x + size.width > containerWidth && x > 0 {
                y += rowHeight + spacing
                x = 0
                rowHeight = 0
            }
            x += size.width + spacing
            rowHeight = max(rowHeight, size.height)
            maxWidth = max(maxWidth, x)
        }
        return CGSize(width: maxWidth, height: y + rowHeight)
    }

    func placeSubviews(in bounds: CGRect, proposal: ProposedViewSize, subviews: Subviews, cache: inout ()) {
        var x = bounds.minX
        var y = bounds.minY
        var rowHeight: CGFloat = 0
        for sv in subviews {
            let size = sv.sizeThatFits(.unspecified)
            if x + size.width > bounds.maxX && x > bounds.minX {
                y += rowHeight + spacing
                x = bounds.minX
                rowHeight = 0
            }
            sv.place(at: CGPoint(x: x, y: y), anchor: .topLeading, proposal: ProposedViewSize(size))
            x += size.width + spacing
            rowHeight = max(rowHeight, size.height)
        }
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
        if let q = version.qualityLabel, !q.isEmpty { return q.split(separator: " ").first.map(String.init) ?? q }
        if let res = version.displayResolution { return res }
        return version.name ?? "Original"
    }

    private var subtitleParts: [String] {
        [version.displayResolution, version.displayVideoCodec, version.displayAudioSummary, version.displayDuration]
            .compactMap { $0 }
            .filter { !$0.isEmpty }
    }

    private var cardWidth: CGFloat { XuvaScale.clamped(280, viewport.width * 0.24, 460) }
    private var cardHeight: CGFloat { XuvaScale.clamped(168, viewport.height * 0.20, 220) }
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
