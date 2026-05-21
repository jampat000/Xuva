import SwiftUI

public struct HomeScreen: View {
    @EnvironmentObject private var store: XuvaClientStore
    @State private var heroIndex = 0
    @State private var activeSection = "Home"

    public init() {}

    public var body: some View {
        GeometryReader { geometry in
            ZStack(alignment: .top) {
                heroBackdrop(geometry: geometry)
                ScrollView {
                    VStack(alignment: .leading, spacing: XuvaScale.sectionSpacing) {
                        MediaTopBar(activeSection: $activeSection)
                            .padding(.top, XuvaScale.safeTop)
                        HeroView(item: hero, heroes: heroes, selectedIndex: $heroIndex, viewport: geometry.size)
                            .padding(.top, 8)
                        rowsSection(geometry: geometry)
                    }
                    .padding(.bottom, 120)
                }
            }
            .background(XuvaTheme.background)
        }
    }

    @ViewBuilder
    private func heroBackdrop(geometry: GeometryProxy) -> some View {
        let height = geometry.size.height * XuvaScale.heroVerticalFraction + 80
        ZStack {
            RemoteImage(urlString: hero.backdropUrl ?? hero.imageUrl, aspectRatio: 16 / 9)
                .frame(width: geometry.size.width, height: height)
                .clipped()
                .opacity(0.55)
            LinearGradient(
                colors: [.clear, XuvaTheme.background.opacity(0.70), XuvaTheme.background],
                startPoint: .top,
                endPoint: .bottom
            )
            LinearGradient(
                colors: [XuvaTheme.background.opacity(0.92), XuvaTheme.background.opacity(0.35), .clear],
                startPoint: .leading,
                endPoint: .trailing
            )
        }
        .frame(width: geometry.size.width, height: height)
        .ignoresSafeArea()
    }

    @ViewBuilder
    private func rowsSection(geometry: GeometryProxy) -> some View {
        if visibleRows.isEmpty {
            EmptyLibraryView(section: activeSection)
                .padding(.horizontal, XuvaScale.safeHorizontal)
        } else {
            VStack(alignment: .leading, spacing: XuvaScale.sectionSpacing) {
                ForEach(visibleRows) { row in
                    MediaRowView(row: row) { item in
                        Task { await store.open(item: item) }
                    }
                }
            }
        }
    }

    private var hero: HomeItem {
        if activeSection != "Home", let first = visibleRows.flatMap({ $0.items ?? [] }).first {
            return first
        }
        if heroIndex < heroes.count { return heroes[heroIndex] }
        return store.home?.hero ?? heroes.first ?? rows.flatMap { $0.items ?? [] }.first ?? HomeItem(
            id: "empty",
            kind: "movie",
            title: "Your cinema awaits",
            subtitle: "Library is ready for titles.",
            genres: ["Local-first"],
            overview: "Movies and shows will appear here as soon as they are available."
        )
    }

    private var heroes: [HomeItem] {
        let explicit = store.home?.heroes ?? []
        if !explicit.isEmpty { return explicit }
        if let hero = store.home?.hero { return [hero] }
        let pool = rows.flatMap { $0.items ?? [] }.filter { ($0.backdropUrl ?? "").isEmpty == false }
        return Array(pool.prefix(5))
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
}

struct HeroView: View {
    let item: HomeItem
    let heroes: [HomeItem]
    @Binding var selectedIndex: Int
    let viewport: CGSize
    @EnvironmentObject private var store: XuvaClientStore
    @FocusState private var heroFocus: HeroFocusItem?

    var body: some View {
        let isCompact = viewport.width < 700
        VStack(alignment: .leading, spacing: isCompact ? 16 : 22) {
            HStack(spacing: 14) {
                Rectangle()
                    .fill(XuvaTheme.text.opacity(0.40))
                    .frame(width: 36, height: 1)
                Text(heroEyebrow)
                    .font(.system(size: XuvaScale.eyebrowFontSize(), weight: .semibold))
                    .tracking(5.6)
                    .foregroundStyle(XuvaTheme.mutedText)
            }
            HeroTitle(item: item, viewport: viewport)
            HomeMetaLine(item: item)
            Text(item.overview ?? item.subtitle ?? "")
                .font(.system(size: XuvaScale.bodyFontSize()))
                .foregroundStyle(XuvaTheme.secondaryText)
                .lineLimit(isCompact ? 4 : 3)
                .frame(maxWidth: viewport.width * 0.55, alignment: .leading)
            HStack(spacing: 14) {
                Button {
                    Task { await store.open(item: item) }
                } label: {
                    Label(primaryActionTitle, systemImage: "play.fill")
                }
                .buttonStyle(XuvaPrimaryButtonStyle())
                .focused($heroFocus, equals: .play)

                Button {
                    Task { await store.open(item: item) }
                } label: {
                    Label("More info", systemImage: "info.circle")
                }
                .buttonStyle(XuvaSecondaryButtonStyle())
                .focused($heroFocus, equals: .info)

                #if !os(tvOS)
                Button {
                    Task { await store.open(item: item) }
                } label: {
                    Image(systemName: "plus")
                }
                .buttonStyle(XuvaIconButtonStyle())
                #endif
            }
            .onAppear { heroFocus = .play }
            heroDots
                .padding(.top, 8)
        }
        .padding(.horizontal, XuvaScale.safeHorizontal)
        .padding(.top, viewport.height * 0.18)
        .frame(maxWidth: .infinity, alignment: .leading)
    }

    private var heroEyebrow: String {
        if item.progress ?? 0 > 0 { return "CONTINUE WATCHING" }
        return "XUVA PRESENTS"
    }

    private var primaryActionTitle: String {
        if item.progress ?? 0 > 0 { return "Resume" }
        return "Play"
    }

    @ViewBuilder
    private var heroDots: some View {
        if heroes.count > 1 {
            HStack(spacing: 12) {
                Text("FEATURED")
                    .font(.system(size: XuvaScale.eyebrowFontSize() - 2, weight: .semibold))
                    .tracking(2.8)
                    .foregroundStyle(XuvaTheme.mutedText)
                ForEach(Array(heroes.enumerated()), id: \.element.id) { index, hero in
                    Button {
                        selectedIndex = index
                    } label: {
                        Capsule()
                            .fill(index == selectedIndex ? XuvaTheme.text : XuvaTheme.text.opacity(0.24))
                            .frame(width: index == selectedIndex ? 56 : 26, height: 3)
                    }
                    .buttonStyle(.plain)
                    .accessibilityLabel("Show \(hero.title ?? "featured title")")
                }
            }
        }
    }
}

private struct HeroTitle: View {
    let item: HomeItem
    let viewport: CGSize

    var body: some View {
        if let logo = item.logoUrl, !logo.isEmpty {
            RemoteLogo(
                urlString: logo,
                fallbackTitle: item.title ?? "Your cinema awaits",
                maxWidth: XuvaScale.heroLogoMaxWidth(viewportWidth: viewport.width),
                maxHeight: XuvaScale.heroLogoMaxHeight(viewportWidth: viewport.width)
            )
        } else {
            heroTextTitle
        }
    }

    private var heroTextTitle: Text {
        let title = item.title ?? "Your cinema awaits"
        let parts = title.split(separator: " ", omittingEmptySubsequences: true).map(String.init)
        let size = XuvaScale.heroTitleSize(viewportWidth: viewport.width)
        guard parts.count > 1, let last = parts.last else {
            return Text(title)
                .font(.system(size: size, weight: .bold))
                .foregroundColor(XuvaTheme.text)
        }
        let leading = parts.dropLast().joined(separator: " ") + " "
        return Text(leading)
            .font(.system(size: size, weight: .bold))
            .foregroundColor(XuvaTheme.text) +
            Text(last)
                .font(.system(size: size, weight: .bold).italic())
                .foregroundColor(XuvaTheme.text)
    }
}

private enum HeroFocusItem: Hashable {
    case play
    case info
}

struct MediaRowView: View {
    let row: HomeRow
    let action: (HomeItem) -> Void

    var body: some View {
        VStack(alignment: .leading, spacing: XuvaScale.rowSpacing) {
            VStack(alignment: .leading, spacing: 6) {
                if let eyebrow = rowEyebrow {
                    Text(eyebrow)
                        .font(.system(size: XuvaScale.eyebrowFontSize(), weight: .semibold))
                        .tracking(3.6)
                        .foregroundStyle(XuvaTheme.mutedText)
                }
                Text(row.title ?? "Library")
                    .font(.system(size: XuvaScale.sectionTitleSize(), weight: .bold))
                    .foregroundStyle(XuvaTheme.text)
            }
            .padding(.horizontal, XuvaScale.safeHorizontal)

            ScrollView(.horizontal, showsIndicators: false) {
                HStack(spacing: XuvaScale.posterRowSpacing()) {
                    ForEach(Array((row.items ?? []).enumerated()), id: \.element.id) { index, item in
                        PosterTile(item: item, ranked: isRanked, rank: index + 1, wide: isWideRow) {
                            action(item)
                        }
                    }
                }
                .padding(.horizontal, XuvaScale.safeHorizontal)
                .padding(.vertical, 16)
            }
        }
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
        if normalizedRowText.contains("recent") { return "RECENTLY ADDED" }
        if normalizedRowText.contains("trend") { return "TRENDING NOW" }
        if isRanked { return "TOP IN YOUR LIBRARY" }
        if normalizedRowText.contains("movie") { return "FROM YOUR LIBRARY" }
        if normalizedRowText.contains("tv") || normalizedRowText.contains("series") { return "FROM YOUR LIBRARY" }
        return nil
    }
}

struct EmptyLibraryView: View {
    let section: String

    var body: some View {
        VStack(alignment: .leading, spacing: 12) {
            Label(title, systemImage: "film.stack")
                .font(.system(size: XuvaScale.bodyFontSize() + 4, weight: .bold))
                .foregroundStyle(XuvaTheme.text)
            Text(message)
                .font(.system(size: XuvaScale.bodyFontSize()))
                .foregroundStyle(XuvaTheme.muted)
        }
        .padding(28)
        .frame(maxWidth: 620, alignment: .leading)
        .background(XuvaTheme.surface.opacity(0.82), in: RoundedRectangle(cornerRadius: 18, style: .continuous))
        .overlay(RoundedRectangle(cornerRadius: 18, style: .continuous).stroke(XuvaTheme.hairline))
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

struct MediaTopBar: View {
    @EnvironmentObject private var store: XuvaClientStore
    @Binding var activeSection: String

    private var sections: [String] {
        #if os(tvOS)
        return ["Home", "Movies", "TV", "Watchlist"]
        #else
        return ["Home", "Movies", "TV", "Watchlist"]
        #endif
    }

    var body: some View {
        HStack(spacing: 0) {
            XuvaLogo()
                .padding(.trailing, XuvaScale.platform == .tv ? 64 : 24)
            #if os(tvOS)
            HStack(spacing: 8) {
                ForEach(sections, id: \.self) { section in
                    TopNavPill(title: section, isActive: activeSection == section) {
                        activeSection = section
                    }
                }
            }
            #else
            if compactWidth {
                EmptyView()
            } else {
                HStack(spacing: 4) {
                    ForEach(sections, id: \.self) { section in
                        TopNavPill(title: section, isActive: activeSection == section) {
                            activeSection = section
                        }
                    }
                }
            }
            #endif
            Spacer()
            Button {
                Task { await store.loadHome() }
            } label: {
                Image(systemName: "arrow.clockwise")
            }
            .buttonStyle(XuvaIconButtonStyle())
        }
        .padding(.horizontal, XuvaScale.safeHorizontal)
        .frame(height: XuvaScale.navBarHeight())
    }

    private var compactWidth: Bool {
        #if os(tvOS)
        return false
        #else
        return true
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
                .font(.system(size: XuvaScale.platform == .tv ? 24 : 15, weight: .medium))
                .foregroundStyle(isActive ? XuvaTheme.text : XuvaTheme.mutedText)
                .padding(.horizontal, XuvaScale.platform == .tv ? 26 : 18)
                .frame(height: XuvaScale.platform == .tv ? 54 : 36)
                .background(isActive ? XuvaTheme.focus.opacity(0.10) : Color.clear, in: Capsule())
                .overlay(Capsule().stroke(isActive ? XuvaTheme.focus.opacity(0.30) : Color.clear))
        }
        .buttonStyle(.plain)
    }
}

private struct HomeMetaLine: View {
    let item: HomeItem

    var body: some View {
        HStack(spacing: 12) {
            ForEach(parts, id: \.self) { part in
                Text(part)
                if part != parts.last {
                    Circle()
                        .fill(XuvaTheme.muted.opacity(0.45))
                        .frame(width: 5, height: 5)
                }
            }
        }
        .font(.system(size: XuvaScale.metaFontSize(), weight: .semibold))
        .foregroundStyle(XuvaTheme.secondaryText)
    }

    private var parts: [String] {
        var parts: [String] = []
        if let year = item.year { parts.append(String(year)) }
        if let runtime = item.runtime { parts.append(runtime) }
        if let runtimeMin = item.runtimeMinutes, runtimeMin > 0 {
            let h = runtimeMin / 60
            let m = runtimeMin % 60
            parts.append(h > 0 ? "\(h)h \(m)m" : "\(m)m")
        }
        if let rating = item.rating, rating > 0 {
            parts.append(String(format: "★ %.1f", rating))
        }
        if let genres = item.genres, !genres.isEmpty {
            parts.append(genres.prefix(2).joined(separator: " / "))
        }
        return parts.filter { !$0.isEmpty }
    }
}
