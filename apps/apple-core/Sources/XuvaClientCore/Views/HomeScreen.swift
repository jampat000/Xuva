import SwiftUI

public struct HomeScreen: View {
    @EnvironmentObject private var store: XuvaClientStore
    @State private var heroIndex = 0
    @State private var activeSection = "Home"

    public init() {}

    public var body: some View {
        GeometryReader { geometry in
            ZStack(alignment: .top) {
                RemoteImage(urlString: hero.backdropUrl ?? hero.imageUrl, aspectRatio: 16 / 9)
                    .frame(width: geometry.size.width, height: heroBackdropHeight(for: geometry.size))
                    .clipped()
                    .opacity(0.34)
                    .blur(radius: 0.4)
                    .ignoresSafeArea()
                LinearGradient(
                    colors: [
                        XuvaTheme.background.opacity(0.08),
                        XuvaTheme.background.opacity(0.68),
                        XuvaTheme.background
                    ],
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
                    VStack(alignment: .leading, spacing: geometry.size.width < 700 ? 30 : 46) {
                        MediaTopBar(activeSection: $activeSection, horizontalPadding: horizontalPadding(for: geometry.size))
                        HeroView(item: hero, heroes: heroes, selectedIndex: $heroIndex)
                            .padding(.top, geometry.size.width < 700 ? 18 : 26)

                        if visibleRows.isEmpty {
                            EmptyLibraryView(section: activeSection)
                                .padding(.horizontal, horizontalPadding(for: geometry.size))
                        } else {
                            ForEach(visibleRows) { row in
                                MediaRowView(row: row, horizontalPadding: horizontalPadding(for: geometry.size)) { item in
                                    Task { await store.open(item: item) }
                                }
                            }
                        }
                    }
                    .padding(.bottom, 110)
                }
            }
        }
    }

    private var hero: HomeItem {
        if activeSection != "Home", let first = visibleRows.flatMap({ $0.items ?? [] }).first {
            return first
        }
        if heroIndex < heroes.count { return heroes[heroIndex] }
        return store.home?.hero ?? heroes.first ?? rows.flatMap { $0.items ?? [] }.first ?? HomeItem(id: "empty", kind: "movie", title: "Your cinema awaits", subtitle: "Library is ready for titles.", year: nil, rating: nil, runtime: nil, progress: nil, posterUrl: nil, backdropUrl: nil, imageUrl: nil, logoUrl: nil, mediaSourceId: nil, routeLabel: nil, genres: ["Local-first"], overview: "Movies and shows will appear here as soon as they are available.")
    }

    private var heroes: [HomeItem] {
        let explicit = store.home?.heroes ?? []
        if !explicit.isEmpty { return explicit }
        if let hero = store.home?.hero { return [hero] }
        return []
    }

    private var rows: [HomeRow] {
        store.home?.rows ?? []
    }

    private var visibleRows: [HomeRow] {
        let populatedRows = rows.filter { !($0.items ?? []).isEmpty }
        switch activeSection {
        case "Movies":
            return populatedRows.filter { rowMatches($0, terms: ["movie"]) }
        case "TV":
            return populatedRows.filter { rowMatches($0, terms: ["tv", "series", "show", "episode"]) }
        case "Watchlist":
            return populatedRows.filter { rowMatches($0, terms: ["watchlist", "watch list", "saved"]) }
        default:
            return populatedRows
        }
    }

    private func rowMatches(_ row: HomeRow, terms: [String]) -> Bool {
        let haystack = "\(row.id) \(row.title ?? "") \(row.subtitle ?? "") \(row.kind ?? "")".lowercased()
        return terms.contains { haystack.contains($0) }
    }

    private func heroBackdropHeight(for size: CGSize) -> CGFloat {
        #if os(tvOS)
        return 840
        #else
        return size.width > 700 ? 620 : 520
        #endif
    }

    private func horizontalPadding(for size: CGSize) -> CGFloat {
        #if os(tvOS)
        return 64
        #else
        return size.width > 700 ? 36 : 20
        #endif
    }
}

struct HeroView: View {
    let item: HomeItem
    let heroes: [HomeItem]
    @Binding var selectedIndex: Int
    @EnvironmentObject private var store: XuvaClientStore

    var body: some View {
        GeometryReader { geometry in
            let isCompact = geometry.size.width < 700
            let layout: AnyLayout = isCompact
                ? AnyLayout(VStackLayout(alignment: .leading, spacing: 28))
                : AnyLayout(HStackLayout(alignment: .bottom, spacing: 44))

            layout {
                VStack(alignment: .leading, spacing: isCompact ? 14 : 20) {
                    HStack(spacing: 14) {
                        Rectangle()
                            .fill(XuvaTheme.text.opacity(0.40))
                            .frame(width: 32, height: 1)
                        Text(heroEyebrow)
                            .font(.caption2.weight(.semibold))
                            .tracking(5.6)
                            .foregroundStyle(XuvaTheme.mutedText)
                    }
                    HeroTitle(item: item, isCompact: isCompact, maxWidth: isCompact ? 520 : 760, maxHeight: isCompact ? 120 : 172)
                    HomeMetaLine(item: item)
                    Text(item.overview ?? item.subtitle ?? "")
                        .font(isCompact ? .body : .title3)
                        .foregroundStyle(XuvaTheme.muted)
                        .lineLimit(isCompact ? 4 : 3)
                        .frame(maxWidth: 760, alignment: .leading)
                    HStack(spacing: 12) {
                        Button {
                            Task { await store.open(item: item) }
                        } label: {
                            Label(primaryActionTitle, systemImage: "play.fill")
                        }
                        .buttonStyle(XuvaPrimaryButtonStyle())

                        Button {
                            Task { await store.open(item: item) }
                        } label: {
                            Label("More info", systemImage: "info.circle")
                        }
                        .buttonStyle(XuvaSecondaryButtonStyle())

                        Button {
                            Task { await store.open(item: item) }
                        } label: {
                            Image(systemName: "plus")
                        }
                        .buttonStyle(XuvaIconButtonStyle())
                    }
                    if let route = item.routeLabel, !route.isEmpty {
                        MediaPill(text: route, systemImage: "waveform.path.ecg", tint: XuvaTheme.focus)
                            .padding(.top, 2)
                    }
                    heroDots
                        .padding(.top, 10)
                }

                if !isCompact {
                    FeaturedPosterCard(item: item)
                }
            }
            .padding(.horizontal, horizontalPadding(for: geometry.size))
            .frame(maxWidth: .infinity, maxHeight: .infinity, alignment: .bottomLeading)
        }
        .frame(height: preferredHeight)
    }

    private var heroEyebrow: String {
        if item.progress ?? 0 > 0 { return "CONTINUE WATCHING" }
        return "XUVA PRESENTS"
    }

    private var primaryActionTitle: String {
        if item.progress ?? 0 > 0 { return "Resume" }
        return "Play"
    }

    private var preferredHeight: CGFloat {
        #if os(tvOS)
        return 560
        #else
        return 460
        #endif
    }

    private func titleSize(for size: CGSize) -> CGFloat {
        #if os(tvOS)
        return 72
        #else
        return size.width > 700 ? 48 : 38
        #endif
    }

    private func horizontalPadding(for size: CGSize) -> CGFloat {
        #if os(tvOS)
        return 64
        #else
        return size.width > 700 ? 36 : 20
        #endif
    }

    @ViewBuilder
    private var heroDots: some View {
        if heroes.count > 1 {
            HStack(spacing: 12) {
                Text("FEATURED")
                    .font(.caption2.weight(.bold))
                    .tracking(2.6)
                    .foregroundStyle(XuvaTheme.mutedText)
                ForEach(Array(heroes.enumerated()), id: \.element.id) { index, hero in
                    Button {
                        selectedIndex = index
                    } label: {
                        Capsule()
                            .fill(index == selectedIndex ? XuvaTheme.text : XuvaTheme.text.opacity(0.24))
                            .frame(width: index == selectedIndex ? 48 : 24, height: 3)
                    }
                    .buttonStyle(.plain)
                    .accessibilityLabel("Show \(hero.title ?? "featured title")")
                }
            }
        }
    }
}

struct MediaRowView: View {
    let row: HomeRow
    let horizontalPadding: CGFloat
    let action: (HomeItem) -> Void

    var body: some View {
        VStack(alignment: .leading, spacing: 18) {
            VStack(alignment: .leading, spacing: 6) {
                if let eyebrow = rowEyebrow {
                    Text(eyebrow)
                        .font(.caption2.weight(.semibold))
                        .tracking(3.6)
                        .foregroundStyle(XuvaTheme.mutedText)
                }
                Text(row.title ?? "Library")
                    .font(rowTitleFont)
                    .foregroundStyle(XuvaTheme.text)
                Text(row.subtitle ?? rowFallbackSubtitle)
                    .font(.callout.weight(.medium))
                    .foregroundStyle(XuvaTheme.muted)
            }
            .padding(.horizontal, horizontalPadding)

            ScrollView(.horizontal, showsIndicators: false) {
                HStack(spacing: 22) {
                    ForEach(Array((row.items ?? []).enumerated()), id: \.element.id) { index, item in
                        PosterTile(item: item, ranked: isRanked, rank: index + 1, wide: isWideRow) {
                            action(item)
                        }
                    }
                }
                .padding(.horizontal, horizontalPadding)
                .padding(.vertical, 12)
            }
        }
    }

    private var rowTitleFont: Font {
        #if os(tvOS)
        return .system(size: 34, weight: .bold, design: .rounded)
        #else
        return .title2.weight(.bold)
        #endif
    }

    private var rowFallbackSubtitle: String {
        let count = row.items?.count ?? 0
        return count == 1 ? "1 title" : "\(count) titles"
    }

    private var normalizedRowText: String {
        "\(row.id) \(row.title ?? "") \(row.kind ?? "")".lowercased()
    }

    private var isWideRow: Bool {
        normalizedRowText.contains("continue") || normalizedRowText.contains("watching")
    }

    private var isRanked: Bool {
        normalizedRowText.contains("top") || normalizedRowText.contains("rated") || normalizedRowText.contains("trend")
    }

    private var rowEyebrow: String? {
        if normalizedRowText.contains("continue") { return "PICK UP WHERE YOU LEFT OFF" }
        if normalizedRowText.contains("movie") { return "FRESH IN YOUR LIBRARY" }
        if normalizedRowText.contains("tv") || normalizedRowText.contains("series") { return "NEW EPISODES DROPPED" }
        if normalizedRowText.contains("recent") { return "RECENTLY ADDED" }
        if isRanked { return "FROM YOUR LIBRARY" }
        return row.subtitle == nil ? nil : row.subtitle?.uppercased()
    }
}

struct EmptyLibraryView: View {
    let section: String

    var body: some View {
        VStack(alignment: .leading, spacing: 12) {
            Label(title, systemImage: "film.stack")
                .font(.headline)
                .foregroundStyle(XuvaTheme.text)
            Text(message)
                .foregroundStyle(XuvaTheme.muted)
        }
        .padding(22)
        .frame(maxWidth: 620, alignment: .leading)
        .background(XuvaTheme.surface.opacity(0.82), in: RoundedRectangle(cornerRadius: 12, style: .continuous))
        .overlay(RoundedRectangle(cornerRadius: 12, style: .continuous).stroke(XuvaTheme.hairline))
    }

    private var title: String {
        section == "Home" ? "Your library is empty" : "No \(section.lowercased()) here yet"
    }

    private var message: String {
        switch section {
        case "Watchlist":
            return "Saved titles will appear here."
        case "Movies", "TV":
            return "\(section) titles will appear here as your library updates."
        default:
            return "Movies and shows will appear here as your library updates."
        }
    }
}

private struct MediaTopBar: View {
    @EnvironmentObject private var store: XuvaClientStore
    @Binding var activeSection: String
    let horizontalPadding: CGFloat

    var body: some View {
        HStack(spacing: 18) {
            XuvaLogo()
            HStack(spacing: 6) {
                ForEach(["Home", "Movies", "TV", "Watchlist"], id: \.self) { section in
                    TopNavPill(title: section, isActive: activeSection == section) {
                        activeSection = section
                    }
                }
            }
            Spacer()
            Button {
                Task { await store.loadHome() }
            } label: {
                Image(systemName: "arrow.clockwise")
            }
            .buttonStyle(XuvaIconButtonStyle())
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

struct TopNavPill: View {
    let title: String
    var isActive = false
    var action: (() -> Void)?

    var body: some View {
        Button {
            action?()
        } label: {
            Text(title)
                .font(.callout.weight(.medium))
                .foregroundStyle(isActive ? XuvaTheme.text : XuvaTheme.mutedText)
                .padding(.horizontal, 18)
                .frame(height: 36)
                .background(isActive ? XuvaTheme.focus.opacity(0.10) : Color.clear, in: Capsule())
                .overlay(Capsule().stroke(isActive ? XuvaTheme.focus.opacity(0.30) : Color.clear))
        }
        .buttonStyle(.plain)
    }
}

private struct HomeMetaLine: View {
    let item: HomeItem

    var body: some View {
        HStack(spacing: 10) {
            ForEach(parts, id: \.self) { part in
                Text(part)
                if part != parts.last {
                    Circle()
                        .fill(XuvaTheme.muted.opacity(0.45))
                        .frame(width: 4, height: 4)
                }
            }
        }
        .font(.callout.weight(.semibold))
        .foregroundStyle(XuvaTheme.secondaryText)
    }

    private var parts: [String] {
        [
            item.year.map(String.init),
            item.runtime,
            item.rating.map { String(format: "%.1f", $0) },
            item.genres?.prefix(2).joined(separator: " / ")
        ]
        .compactMap { $0 }
        .filter { !$0.isEmpty }
    }
}

private struct HeroTitle: View {
    let item: HomeItem
    let isCompact: Bool
    let maxWidth: CGFloat
    let maxHeight: CGFloat

    var body: some View {
        if let logo = item.logoUrl, !logo.isEmpty {
            RemoteLogo(
                urlString: logo,
                fallbackTitle: item.title ?? "Your cinema awaits",
                maxWidth: maxWidth,
                maxHeight: maxHeight
            )
        } else {
            heroTextTitle
        }
    }

    private var heroTextTitle: Text {
        let title = item.title ?? "Your cinema awaits"
        let parts = title.split(separator: " ", omittingEmptySubsequences: true).map(String.init)
        guard parts.count > 1, let last = parts.last else {
            return Text(title)
                .font(.system(size: titleSize, weight: .bold))
                .foregroundColor(XuvaTheme.text) +
                Text("")
        }
        let leading = parts.dropLast().joined(separator: " ") + " "
        return Text(leading)
            .font(.system(size: titleSize, weight: .bold))
            .foregroundColor(XuvaTheme.text) +
            Text(last)
                .font(.system(size: titleSize, weight: .bold).italic())
                .foregroundColor(XuvaTheme.text)
    }

    private var titleSize: CGFloat {
        #if os(tvOS)
        return 88
        #else
        return isCompact ? 44 : 64
        #endif
    }
}

private struct FeaturedPosterCard: View {
    let item: HomeItem

    var body: some View {
        ZStack(alignment: .bottomLeading) {
            RemoteImage(urlString: item.posterUrl ?? item.imageUrl, aspectRatio: 2 / 3)
                .frame(width: 250, height: 375)
                .clipShape(RoundedRectangle(cornerRadius: 14, style: .continuous))
            LinearGradient(colors: [.clear, .black.opacity(0.84)], startPoint: .center, endPoint: .bottom)
                .clipShape(RoundedRectangle(cornerRadius: 14, style: .continuous))
            VStack(alignment: .leading, spacing: 8) {
                Text("Featured")
                    .font(.caption2.weight(.bold))
                    .tracking(2.4)
                    .foregroundStyle(.white.opacity(0.72))
                RemoteLogo(urlString: item.logoUrl, fallbackTitle: item.title ?? "Xuva", maxWidth: 170, maxHeight: 64)
            }
            .padding(18)
        }
        .shadow(color: .black.opacity(0.46), radius: 28, y: 22)
    }
}
