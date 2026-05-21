import SwiftUI

public struct DetailScreen: View {
    @EnvironmentObject private var store: XuvaClientStore

    public init() {}

    public var body: some View {
        if let detail = store.selectedDetail {
            DetailContentView(detail: detail)
        } else {
            Text("No title selected")
                .foregroundStyle(XuvaTheme.text)
        }
    }
}

private struct DetailContentView: View {
    @EnvironmentObject private var store: XuvaClientStore
    let detail: DetailResponse
    @State private var selectedVersionID: String?
    @State private var selectedAudioID: String?
    @State private var selectedSubtitleID: String?

    var body: some View {
        GeometryReader { geometry in
            ZStack(alignment: .top) {
                RemoteImage(urlString: detail.displayBackdropURL, aspectRatio: 16 / 9)
                    .frame(width: geometry.size.width, height: backdropHeight(for: geometry.size))
                    .clipped()
                    .opacity(0.38)
                    .ignoresSafeArea()
                LinearGradient(
                    colors: [.clear, XuvaTheme.background.opacity(0.78), XuvaTheme.background],
                    startPoint: .top,
                    endPoint: .bottom
                )
                .ignoresSafeArea()
                LinearGradient(
                    colors: [XuvaTheme.background, XuvaTheme.background.opacity(0.24), .clear],
                    startPoint: .leading,
                    endPoint: .trailing
                )
                .ignoresSafeArea()

                ScrollView {
                    VStack(alignment: .leading, spacing: geometry.size.width < 700 ? 28 : 42) {
                        DetailTopBar(horizontalPadding: horizontalPadding(for: geometry.size))
                        detailHero(geometry: geometry)
                        versionSection
                        trackSection
                        creditSection
                        collectionSection
                    }
                    .padding(.bottom, 96)
                }
            }
        }
    }

    @ViewBuilder
    private func detailHero(geometry: GeometryProxy) -> some View {
        let isCompact = geometry.size.width < 700
        let layout: AnyLayout = if isCompact {
            AnyLayout(VStackLayout(alignment: .leading, spacing: 24))
        } else {
            AnyLayout(HStackLayout(alignment: .bottom, spacing: 46))
        }

        layout {
            RemoteImage(urlString: detail.displayPosterURL, aspectRatio: 2 / 3)
                .frame(width: posterSize(for: geometry.size).width, height: posterSize(for: geometry.size).height)
                .clipShape(RoundedRectangle(cornerRadius: 12, style: .continuous))
                .shadow(color: .black.opacity(0.46), radius: 32, y: 24)

            VStack(alignment: .leading, spacing: isCompact ? 14 : 18) {
                Text(metadataLine(detail))
                    .font(.caption.weight(.bold))
                    .tracking(2.4)
                    .foregroundStyle(XuvaTheme.muted)
                RemoteLogo(
                    urlString: detail.displayLogoURL,
                    fallbackTitle: detail.displayTitle,
                    maxWidth: isCompact ? 520 : 760,
                    maxHeight: isCompact ? 112 : 156
                )
                if let tagline = detail.tagline ?? detail.metadata?.tagline, !tagline.isEmpty {
                    Text(tagline)
                        .italic()
                        .foregroundStyle(XuvaTheme.muted)
                }
                Text(detail.displayOverview)
                    .font(isCompact ? .body : .title3)
                    .foregroundStyle(XuvaTheme.muted)
                    .lineLimit(isCompact ? 6 : 5)
                    .frame(maxWidth: 820, alignment: .leading)
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

                    Button {
                        store.backToHome()
                    } label: {
                        Label("Library", systemImage: "square.grid.2x2")
                    }
                    .buttonStyle(XuvaSecondaryButtonStyle())

                    RouteBadge(decision: routeDecision)
                }
                DetailFactStrip(detail: detail)
                PlaybackForecastCard(decision: routeDecision)
                    .padding(.top, 6)
            }
            if !isCompact {
                Spacer(minLength: 24)
            }
        }
        .padding(.horizontal, horizontalPadding(for: geometry.size))
        .padding(.top, topOffset(for: geometry.size))
    }

    @ViewBuilder
    private var versionSection: some View {
        if !(detail.versions ?? []).isEmpty {
            VStack(alignment: .leading, spacing: 16) {
                SectionHeading(title: "Versions", subtitle: "Source quality and route")
                ScrollView(.horizontal, showsIndicators: false) {
                    HStack(spacing: 16) {
                        ForEach(detail.versions ?? [], id: \.stableID) { version in
                            VersionCard(version: version, isSelected: version.stableID == selectedVersion?.stableID) {
                                selectedVersionID = version.stableID
                            }
                        }
                    }
                    .padding(.vertical, 12)
                }
            }
            .padding(.horizontal, sectionPadding)
        }
    }

    @ViewBuilder
    private var trackSection: some View {
        if !(detail.audioTracks ?? []).isEmpty || !(detail.subtitleTracks ?? []).isEmpty {
            VStack(alignment: .leading, spacing: 16) {
                SectionHeading(title: "Audio and subtitles", subtitle: "Language, channels, and captions")
                HStack(alignment: .top, spacing: 16) {
                    TrackStack(title: "Audio", systemImage: "speaker.wave.2", tracks: detail.audioTracks ?? [], selectedTrackID: $selectedAudioID, allowsNone: false)
                    TrackStack(title: "Subtitles", systemImage: "captions.bubble", tracks: detail.subtitleTracks ?? [], selectedTrackID: $selectedSubtitleID, allowsNone: true)
                }
            }
            .padding(.horizontal, sectionPadding)
        }
    }

    @ViewBuilder
    private var creditSection: some View {
        let cast = detail.displayCast
        let credits = detail.displayDirectors + detail.displayWriters
        let studios = detail.displayStudios
        if !cast.isEmpty || !credits.isEmpty || !studios.isEmpty {
            VStack(alignment: .leading, spacing: 18) {
                SectionHeading(title: "People and studio", subtitle: "Credits and production")
                if !credits.isEmpty {
                    HStack(spacing: 10) {
                        ForEach(credits.prefix(6), id: \.self) { person in
                            MediaPill(text: person, systemImage: "person.fill", tint: XuvaTheme.secondaryText)
                        }
                    }
                }
                if !studios.isEmpty {
                    HStack(spacing: 10) {
                        ForEach(studios.prefix(6), id: \.self) { studio in
                            MediaPill(text: studio, systemImage: "building.2", tint: XuvaTheme.secondaryText)
                        }
                    }
                }
                if !cast.isEmpty {
                    ScrollView(.horizontal, showsIndicators: false) {
                        HStack(spacing: 14) {
                            ForEach(cast.prefix(16), id: \.stableID) { person in
                                CastCard(person: person)
                            }
                        }
                        .padding(.vertical, 10)
                    }
                }
            }
            .padding(.horizontal, sectionPadding)
        }
    }

    @ViewBuilder
    private var collectionSection: some View {
        if let collection = detail.metadata?.collection, let name = collection.name, !name.isEmpty {
            VStack(alignment: .leading, spacing: 16) {
                SectionHeading(title: "Collection", subtitle: "Related titles")
                HStack(spacing: 18) {
                    RemoteImage(urlString: collection.backdropUrl ?? collection.posterUrl, aspectRatio: 16 / 9)
                        .frame(width: 360, height: 202)
                        .clipShape(RoundedRectangle(cornerRadius: 10, style: .continuous))
                    VStack(alignment: .leading, spacing: 10) {
                        RemoteLogo(urlString: collection.logoUrl, fallbackTitle: name, maxWidth: 380, maxHeight: 86)
                        Text("Titles from the same collection.")
                            .foregroundStyle(XuvaTheme.muted)
                            .lineLimit(2)
                    }
                }
                .padding(18)
                .background(XuvaTheme.elevated.opacity(0.58), in: RoundedRectangle(cornerRadius: 10, style: .continuous))
                .overlay(RoundedRectangle(cornerRadius: 10, style: .continuous).stroke(XuvaTheme.hairline))
            }
            .padding(.horizontal, sectionPadding)
        }
    }

    private func backdropHeight(for size: CGSize) -> CGFloat {
        #if os(tvOS)
        return 760
        #else
        return size.width > 700 ? 560 : 420
        #endif
    }

    private func topOffset(for size: CGSize) -> CGFloat {
        #if os(tvOS)
        return 36
        #else
        return size.width > 700 ? 26 : 18
        #endif
    }

    private func posterSize(for size: CGSize) -> CGSize {
        #if os(tvOS)
        return CGSize(width: 285, height: 428)
        #else
        return size.width > 700 ? CGSize(width: 190, height: 285) : CGSize(width: 132, height: 198)
        #endif
    }

    private func horizontalPadding(for size: CGSize) -> CGFloat {
        #if os(tvOS)
        return 64
        #else
        return size.width > 700 ? 36 : 20
        #endif
    }

    private var sectionPadding: CGFloat {
        #if os(tvOS)
        return 64
        #else
        return 28
        #endif
    }

    private var routeDecision: PlaybackDecision? {
        selectedVersion?.decision ?? detail.playbackDecision ?? detail.versions?.first?.decision
    }

    private var selectedVersion: MediaVersion? {
        let versions = detail.versions ?? []
        if let selectedVersionID, let version = versions.first(where: { $0.stableID == selectedVersionID }) {
            return version
        }
        return versions.first
    }

    private var selectedAudioTrack: MediaTrack? {
        let tracks = detail.audioTracks ?? []
        if let selectedAudioID, let track = tracks.first(where: { $0.stableID == selectedAudioID }) {
            return track
        }
        return tracks.first(where: { $0.default == true })
    }

    private var selectedSubtitleTrack: MediaTrack? {
        guard let selectedSubtitleID else { return nil }
        return (detail.subtitleTracks ?? []).first(where: { $0.stableID == selectedSubtitleID })
    }

    private func metadataLine(_ detail: DetailResponse) -> String {
        [detail.displayYear.map(String.init), detail.displayRuntime, detail.contentRating, detail.displayGenres.prefix(2).joined(separator: " / ")]
            .compactMap { $0 }
            .filter { !$0.isEmpty }
            .joined(separator: "  ·  ")
    }
}

private struct DetailTopBar: View {
    @EnvironmentObject private var store: XuvaClientStore
    let horizontalPadding: CGFloat

    var body: some View {
        HStack(spacing: 16) {
            Button {
                store.backToHome()
            } label: {
                Image(systemName: "chevron.left")
            }
            .buttonStyle(XuvaIconButtonStyle())
            XuvaLogo()
            HStack(spacing: 6) {
                TopNavPill(title: "Home")
                TopNavPill(title: "Movies", isActive: store.selectedDetail?.kind?.lowercased() != "series")
                TopNavPill(title: "TV", isActive: store.selectedDetail?.kind?.lowercased() == "series")
            }
            Spacer()
        }
        .padding(.horizontal, horizontalPadding)
        .padding(.top, topPadding)
    }

    private var topPadding: CGFloat {
        #if os(tvOS)
        return 46
        #else
        return 18
        #endif
    }
}

private struct DetailFactStrip: View {
    let detail: DetailResponse

    var body: some View {
        HStack(spacing: 10) {
            if let quality = detail.versions?.first?.qualityLabel, !quality.isEmpty {
                MediaPill(text: quality, systemImage: "sparkles.tv", tint: XuvaTheme.action)
            }
            if let versionCount = detail.versions?.count, versionCount > 1 {
                MediaPill(text: "\(versionCount) Versions", systemImage: "film.stack", tint: XuvaTheme.secondaryText)
            }
            if let audioCount = detail.audioTracks?.count, audioCount > 0 {
                MediaPill(text: "\(audioCount) Audio", systemImage: "speaker.wave.2", tint: XuvaTheme.secondaryText)
            }
            if let subtitleCount = detail.subtitleTracks?.count, subtitleCount > 0 {
                MediaPill(text: "\(subtitleCount) Subs", systemImage: "captions.bubble", tint: XuvaTheme.secondaryText)
            }
        }
    }
}

private struct SectionHeading: View {
    let title: String
    let subtitle: String

    var body: some View {
        VStack(alignment: .leading, spacing: 5) {
            Text(title)
                .font(.system(size: sectionTitleSize, weight: .bold, design: .rounded))
                .foregroundStyle(XuvaTheme.text)
            Text(subtitle)
                .font(.callout.weight(.medium))
                .foregroundStyle(XuvaTheme.muted)
        }
    }

    private var sectionTitleSize: CGFloat {
        #if os(tvOS)
        return 34
        #else
        return 22
        #endif
    }
}

private struct PlaybackForecastCard: View {
    let decision: PlaybackDecision?

    var body: some View {
        VStack(alignment: .leading, spacing: 12) {
            HStack(spacing: 10) {
                Image(systemName: "waveform.path.ecg")
                Text("Playback path")
                    .font(.headline.weight(.bold))
                Spacer()
                RouteBadge(decision: decision)
            }
            Text(decision?.reasonText ?? decision?.serverImpact ?? "Best available playback route.")
                .font(.callout)
                .foregroundStyle(XuvaTheme.secondaryText)
                .lineLimit(3)
            HStack(spacing: 8) {
                ForecastStep(label: "Video", value: decision?.videoAction)
                ForecastStep(label: "Audio", value: decision?.audioAction)
                ForecastStep(label: "Subs", value: decision?.subtitleAction)
                ForecastStep(label: "Container", value: decision?.containerAction)
            }
        }
        .padding(18)
        .frame(maxWidth: 820, alignment: .leading)
        .background(XuvaTheme.ink.opacity(0.72), in: RoundedRectangle(cornerRadius: 10, style: .continuous))
        .overlay(RoundedRectangle(cornerRadius: 10, style: .continuous).stroke(XuvaTheme.hairline))
    }
}

private struct ForecastStep: View {
    let label: String
    let value: String?

    var body: some View {
        VStack(alignment: .leading, spacing: 4) {
            Text(label)
                .font(.caption2.weight(.bold))
                .tracking(1.1)
                .foregroundStyle(XuvaTheme.muted)
            Text(displayValue)
                .font(.caption.weight(.semibold))
                .foregroundStyle(XuvaTheme.text)
                .lineLimit(1)
        }
        .frame(width: 104, alignment: .leading)
    }

    private var displayValue: String {
        guard let value, !value.isEmpty else { return "Auto" }
        return value.replacingOccurrences(of: "_", with: " ").capitalized
    }
}

private struct VersionCard: View {
    let version: MediaVersion
    let isSelected: Bool
    let select: () -> Void

    var body: some View {
        Button(action: select) {
            VStack(alignment: .leading, spacing: 14) {
                HStack {
                    Image(systemName: "film.stack")
                        .foregroundStyle(XuvaTheme.action)
                    Spacer()
                    if isSelected {
                        MediaPill(text: "Selected", systemImage: "checkmark.circle.fill", tint: XuvaTheme.focus)
                    } else {
                        RouteBadge(decision: version.decision)
                    }
                }
                Text(version.qualityLabel ?? version.name ?? "Original")
                    .font(.headline.weight(.bold))
                    .foregroundStyle(XuvaTheme.text)
                    .lineLimit(2)
                Text([version.resolution, version.videoCodec, version.audioSummary].compactMap { $0 }.joined(separator: " · "))
                    .font(.caption)
                    .foregroundStyle(XuvaTheme.muted)
                    .lineLimit(3)
                Spacer()
                Label(isSelected ? "Ready" : "Available", systemImage: isSelected ? "checkmark" : "play.fill")
                    .font(.caption.weight(.bold))
                    .foregroundStyle(XuvaTheme.text)
            }
            .padding(18)
            .frame(width: cardWidth, height: 178, alignment: .leading)
            .background(isSelected ? XuvaTheme.focus.opacity(0.12) : XuvaTheme.elevated.opacity(0.72), in: RoundedRectangle(cornerRadius: 10, style: .continuous))
            .overlay(RoundedRectangle(cornerRadius: 10, style: .continuous).stroke(isSelected ? XuvaTheme.focus.opacity(0.42) : XuvaTheme.hairline))
        }
        .buttonStyle(.plain)
        .xuvaFocused(radius: 10)
    }

    private var cardWidth: CGFloat {
        #if os(tvOS)
        return 360
        #else
        return 300
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
                .font(.headline.weight(.bold))
                .foregroundStyle(XuvaTheme.text)
            if allowsNone {
                TrackButton(title: "Off", subtitle: "No subtitle track", isSelected: selectedTrackID == nil) {
                    selectedTrackID = nil
                }
            }
            if tracks.isEmpty {
                Text("No tracks available")
                    .foregroundStyle(XuvaTheme.muted)
            } else {
                ForEach(tracks.prefix(4), id: \.stableID) { track in
                    TrackButton(
                        title: track.title ?? track.language ?? "Track \(track.index ?? 0)",
                        subtitle: [track.codec, track.channels.map { "\($0)ch" }, track.external == true ? "External" : nil, track.default == true ? "Default" : nil, track.forced == true ? "Forced" : nil].compactMap { $0 }.joined(separator: " · "),
                        isSelected: selectedTrackID == track.stableID || (!allowsNone && selectedTrackID == nil && track.default == true)
                    ) {
                        selectedTrackID = track.stableID
                    }
                }
            }
        }
        .padding(18)
        .frame(maxWidth: 420, minHeight: 150, alignment: .topLeading)
        .background(XuvaTheme.elevated.opacity(0.68), in: RoundedRectangle(cornerRadius: 10, style: .continuous))
        .overlay(RoundedRectangle(cornerRadius: 10, style: .continuous).stroke(XuvaTheme.hairline))
    }
}

private struct TrackButton: View {
    let title: String
    let subtitle: String
    let isSelected: Bool
    let select: () -> Void

    var body: some View {
        Button(action: select) {
            HStack(spacing: 10) {
                Circle()
                    .fill(isSelected ? XuvaTheme.focus : XuvaTheme.muted.opacity(0.42))
                    .frame(width: 8, height: 8)
                VStack(alignment: .leading, spacing: 3) {
                    Text(title)
                        .font(.callout.weight(.semibold))
                        .foregroundStyle(XuvaTheme.text)
                        .lineLimit(1)
                    Text(subtitle)
                        .font(.caption)
                        .foregroundStyle(XuvaTheme.muted)
                        .lineLimit(1)
                }
                Spacer()
                if isSelected {
                    Image(systemName: "checkmark")
                        .font(.caption.weight(.bold))
                        .foregroundStyle(XuvaTheme.focus)
                }
            }
            .padding(.horizontal, 12)
            .padding(.vertical, 10)
            .background(isSelected ? XuvaTheme.focus.opacity(0.10) : Color.white.opacity(0.035), in: RoundedRectangle(cornerRadius: 8, style: .continuous))
            .overlay(RoundedRectangle(cornerRadius: 8, style: .continuous).stroke(isSelected ? XuvaTheme.focus.opacity(0.34) : Color.clear))
        }
        .buttonStyle(.plain)
        .xuvaFocused(radius: 8)
    }
}

private struct CastCard: View {
    let person: MetadataCredit

    var body: some View {
        VStack(alignment: .leading, spacing: 9) {
            RemoteImage(urlString: person.profileUrl, aspectRatio: 2 / 3)
                .frame(width: cardWidth, height: cardHeight)
                .clipShape(RoundedRectangle(cornerRadius: 10, style: .continuous))
            Text(person.name ?? "Unknown")
                .font(.caption.weight(.bold))
                .foregroundStyle(XuvaTheme.text)
                .lineLimit(1)
            if let character = person.character, !character.isEmpty {
                Text(character)
                    .font(.caption2)
                    .foregroundStyle(XuvaTheme.muted)
                    .lineLimit(1)
            }
        }
        .frame(width: cardWidth, alignment: .leading)
    }

    private var cardWidth: CGFloat {
        #if os(tvOS)
        return 132
        #else
        return 92
        #endif
    }

    private var cardHeight: CGFloat {
        #if os(tvOS)
        return 198
        #else
        return 138
        #endif
    }
}
