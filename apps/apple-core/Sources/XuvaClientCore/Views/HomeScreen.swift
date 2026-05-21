import SwiftUI

public struct HomeScreen: View {
    @EnvironmentObject private var store: XuvaClientStore
    @FocusState private var rootFocus: RootFocus?
    @Namespace private var rootNamespace

    public init() {}

    public var body: some View {
        GeometryReader { geometry in
            let viewport = geometry.size
            ZStack(alignment: .top) {
                heroBackdrop(viewport: viewport)
                ScrollView {
                    VStack(alignment: .leading, spacing: XuvaScale.sectionSpacing(viewport)) {
                        MediaTopBar(viewport: viewport)
                            .padding(.top, XuvaScale.safeTop(viewport))
                        HeroView(item: hero, heroes: heroes, viewport: viewport, parentFocus: $rootFocus)
                        rowsSection(viewport: viewport)
                    }
                    .padding(.bottom, viewport.height * 0.12)
                }
            }
            .background(XuvaTheme.background)
            .modifier(FocusScopeModifier(namespace: rootNamespace))
        }
        .defaultFocus($rootFocus, .heroPlay)
        .onAppear {
            // tvOS focus engine sometimes lands on the first focusable view (nav pill)
            // before defaultFocus resolves. Re-assert focus after layout settles.
            DispatchQueue.main.asyncAfter(deadline: .now() + 0.05) {
                rootFocus = .heroPlay
            }
        }
    }

    enum RootFocus: Hashable {
        case heroPlay
        case heroInfo
    }

    @ViewBuilder
    private func heroBackdrop(viewport: CGSize) -> some View {
        let height = viewport.height * XuvaScale.heroVerticalFraction(viewport) + 80
        ZStack {
            RemoteImage(urlString: hero.backdropUrl ?? hero.imageUrl, aspectRatio: 16 / 9)
                .frame(width: viewport.width, height: height)
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
        .frame(width: viewport.width, height: height)
        .ignoresSafeArea()
    }

    @ViewBuilder
    private func rowsSection(viewport: CGSize) -> some View {
        if visibleRows.isEmpty {
            EmptyLibraryView(section: store.activeSection, viewport: viewport)
                .padding(.horizontal, XuvaScale.safeHorizontal(viewport))
        } else {
            VStack(alignment: .leading, spacing: XuvaScale.sectionSpacing(viewport)) {
                ForEach(visibleRows) { row in
                    MediaRowView(row: row, viewport: viewport) { item in
                        Task { await store.open(item: item) }
                    }
                }
            }
        }
    }

    private var hero: HomeItem {
        if store.activeSection != "Home", let first = visibleRows.flatMap({ $0.items ?? [] }).first {
            return first
        }
        if store.heroIndex < heroes.count { return heroes[store.heroIndex] }
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
        switch store.activeSection {
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
    let viewport: CGSize
    var parentFocus: FocusState<HomeScreen.RootFocus?>.Binding
    @EnvironmentObject private var store: XuvaClientStore

    var body: some View {
        let isCompact = viewport.width < 600
        VStack(alignment: .leading, spacing: isCompact ? 14 : 22) {
            HStack(spacing: 14) {
                Rectangle()
                    .fill(XuvaTheme.text.opacity(0.40))
                    .frame(width: 36, height: 1)
                Text(heroEyebrow)
                    .font(.system(size: XuvaScale.eyebrowFontSize(viewport), weight: .semibold))
                    .tracking(5.6)
                    .foregroundStyle(XuvaTheme.mutedText)
            }
            HeroTitle(item: item, viewport: viewport)
            HomeMetaLine(item: item, viewport: viewport)
            Text(item.overview ?? item.subtitle ?? "")
                .font(.system(size: XuvaScale.bodyFontSize(viewport)))
                .foregroundStyle(XuvaTheme.secondaryText)
                .lineLimit(isCompact ? 4 : 3)
                .frame(maxWidth: XuvaScale.heroContentMaxWidth(viewport), alignment: .leading)
            HStack(spacing: 14) {
                Button {
                    Task { await store.open(item: item) }
                } label: {
                    Label(primaryActionTitle, systemImage: "play.fill")
                }
                .buttonStyle(XuvaPrimaryButtonStyle(viewport: viewport))
                .focused(parentFocus, equals: .heroPlay)

                Button {
                    Task { await store.open(item: item) }
                } label: {
                    Label("More info", systemImage: "info.circle")
                }
                .buttonStyle(XuvaSecondaryButtonStyle(viewport: viewport))
                .focused(parentFocus, equals: .heroInfo)

                #if !os(tvOS)
                Button {
                    Task { await store.open(item: item) }
                } label: {
                    Image(systemName: "plus")
                }
                .buttonStyle(XuvaIconButtonStyle(viewport: viewport))
                #endif
            }
            heroDots
                .padding(.top, 8)
        }
        .padding(.horizontal, XuvaScale.safeHorizontal(viewport))
        .padding(.top, viewport.height * XuvaScale.heroContentTopFraction(viewport))
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
                    .font(.system(size: XuvaScale.eyebrowFontSize(viewport) - 2, weight: .semibold))
                    .tracking(2.8)
                    .foregroundStyle(XuvaTheme.mutedText)
                ForEach(Array(heroes.enumerated()), id: \.element.id) { index, hero in
                    Capsule()
                        .fill(index == store.heroIndex ? XuvaTheme.text : XuvaTheme.text.opacity(0.24))
                        .frame(width: index == store.heroIndex ? 56 : 26, height: 3)
                        .accessibilityLabel("Showing \(hero.title ?? "featured title")")
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
                maxWidth: XuvaScale.heroLogoMaxWidth(viewport),
                maxHeight: XuvaScale.heroLogoMaxHeight(viewport)
            )
        } else {
            heroTextTitle
        }
    }

    private var heroTextTitle: Text {
        let title = item.title ?? "Your cinema awaits"
        let parts = title.split(separator: " ", omittingEmptySubsequences: true).map(String.init)
        let size = XuvaScale.heroTitleSize(viewport)
        guard parts.count > 1, let last = parts.last else {
            return Text(title)
                .font(.system(size: size, weight: .semibold, design: .default))
                .tracking(size * -0.045)
                .foregroundColor(XuvaTheme.text)
        }
        let leading = parts.dropLast().joined(separator: " ") + " "
        return Text(leading)
            .font(.system(size: size, weight: .semibold, design: .default))
            .tracking(size * -0.045)
            .foregroundColor(XuvaTheme.text) +
            Text(last)
                .font(.system(size: size, weight: .semibold, design: .default).italic())
                .tracking(size * -0.045)
                .foregroundColor(XuvaTheme.text)
    }
}

private struct FocusScopeModifier: ViewModifier {
    let namespace: Namespace.ID
    func body(content: Content) -> some View {
        #if os(tvOS)
        content.focusScope(namespace)
        #else
        content
        #endif
    }
}


struct MediaRowView: View {
    let row: HomeRow
    let viewport: CGSize
    let action: (HomeItem) -> Void

    var body: some View {
        VStack(alignment: .leading, spacing: XuvaScale.rowSpacing(viewport)) {
            VStack(alignment: .leading, spacing: 6) {
                if let eyebrow = rowEyebrow {
                    Text(eyebrow)
                        .font(.system(size: XuvaScale.eyebrowFontSize(viewport), weight: .semibold))
                        .tracking(3.6)
                        .foregroundStyle(XuvaTheme.mutedText)
                }
                Text(row.title ?? "Library")
                    .font(.system(size: XuvaScale.sectionTitleSize(viewport), weight: .bold))
                    .foregroundStyle(XuvaTheme.text)
            }
            .padding(.horizontal, XuvaScale.safeHorizontal(viewport))

            ScrollView(.horizontal, showsIndicators: false) {
                HStack(spacing: XuvaScale.posterRowSpacing(viewport)) {
                    ForEach(Array((row.items ?? []).enumerated()), id: \.element.id) { index, item in
                        PosterTile(item: item, viewport: viewport, ranked: isRanked, rank: index + 1, wide: isWideRow) {
                            action(item)
                        }
                    }
                }
                .padding(.horizontal, XuvaScale.safeHorizontal(viewport))
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
    let viewport: CGSize

    var body: some View {
        VStack(alignment: .leading, spacing: 12) {
            Label(title, systemImage: "film.stack")
                .font(.system(size: XuvaScale.bodyFontSize(viewport) + 4, weight: .bold))
                .foregroundStyle(XuvaTheme.text)
            Text(message)
                .font(.system(size: XuvaScale.bodyFontSize(viewport)))
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
    let viewport: CGSize

    private var sections: [String] {
        ["Home", "Movies", "TV", "Watchlist"]
    }

    var body: some View {
        let showInlineNav = viewport.width >= 700
        HStack(spacing: 0) {
            XuvaLogo(viewport: viewport)
                .padding(.trailing, showInlineNav ? viewport.width * 0.04 : 12)
            if showInlineNav {
                HStack(spacing: 6) {
                    ForEach(sections, id: \.self) { section in
                        TopNavPill(title: section, viewport: viewport, isActive: store.activeSection == section) {
                            store.setSection(section)
                        }
                    }
                }
            }
            Spacer()
            Button {
                Task { await store.loadHome() }
            } label: {
                Image(systemName: "arrow.clockwise")
            }
            .buttonStyle(XuvaIconButtonStyle(viewport: viewport))
        }
        .padding(.horizontal, XuvaScale.safeHorizontal(viewport))
        .frame(height: XuvaScale.navBarHeight(viewport))
    }
}

struct TopNavPill: View {
    let title: String
    let viewport: CGSize
    var isActive = false
    var action: (() -> Void)?

    var body: some View {
        Button {
            action?()
        } label: {
            Text(title)
                .font(.system(size: XuvaScale.metaFontSize(viewport), weight: .medium))
                .foregroundStyle(isActive ? XuvaTheme.text : XuvaTheme.mutedText)
                .padding(.horizontal, XuvaScale.buttonHorizontalPadding(viewport) * 0.75)
                .frame(height: XuvaScale.buttonHeight(viewport) * 0.7)
                .background(isActive ? XuvaTheme.focus.opacity(0.10) : Color.clear, in: Capsule())
                .overlay(Capsule().stroke(isActive ? XuvaTheme.focus.opacity(0.30) : Color.clear))
        }
        .buttonStyle(.plain)
    }
}

private struct HomeMetaLine: View {
    let item: HomeItem
    let viewport: CGSize

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
        .font(.system(size: XuvaScale.metaFontSize(viewport), weight: .semibold))
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
