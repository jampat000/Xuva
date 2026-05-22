import SwiftUI

public struct HomeScreen: View {
    @EnvironmentObject private var store: XuvaClientStore
    @EnvironmentObject private var watchlist: XuvaWatchlist
    @State private var showSettings = false

    public init() {}

    public var body: some View {
        GeometryReader { geometry in
            let viewport = geometry.size
            ZStack(alignment: .top) {
                heroBackdrop(viewport: viewport)
                // MediaTopBar lives OUTSIDE the ScrollView so it stays on-screen
                // as the user scrolls down into content rows. The tvOS focus engine
                // can only navigate to views that are currently rendered — if the
                // nav bar scrolls off-screen, "swipe up" from the rows has nowhere
                // to land and the user is stuck.
                VStack(spacing: 0) {
                    MediaTopBar(viewport: viewport, onSettings: { showSettings = true })
                        .padding(.top, XuvaScale.safeTop(viewport))
                    ScrollView {
                        VStack(alignment: .leading, spacing: XuvaScale.sectionSpacing(viewport)) {
                            HeroView(item: hero, heroes: heroes, viewport: viewport)
                            rowsSection(viewport: viewport)
                        }
                        .padding(.bottom, 60)
                    }
                    // focusSection on the ScrollView lets tvOS navigate vertically
                    // out of it — without this, UP from inside the scroll view has
                    // nowhere to land and the nav bar above is unreachable.
                    .focusSection()
                    .frame(maxWidth: .infinity, maxHeight: .infinity)
                }
                .frame(maxWidth: .infinity, maxHeight: .infinity)
            }
            .frame(maxWidth: .infinity, maxHeight: .infinity)
            .background(XuvaTheme.background)
        }
        .ignoresSafeArea()
        .fullScreenCover(isPresented: $showSettings) {
            SettingsScreen { showSettings = false }
        }
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
        // "recently-added" is excluded — movies and series already appear in
        // their own dedicated rows (matches web behaviour exactly).
        let populatedRows = rows
            .filter { !($0.items ?? []).isEmpty }
            .filter { !$0.id.lowercased().contains("recent") }
        switch store.activeSection {
        case "Movies":
            return populatedRows.filter { rowMatches($0, terms: ["movie"]) }
        case "TV":
            return populatedRows.filter { rowMatches($0, terms: ["tv", "series", "show", "episode"]) }
        case "Watchlist":
            let saved = watchlist.asHomeItems()
            guard !saved.isEmpty else { return [] }
            return [HomeRow(id: "watchlist-local", title: "Your Watchlist", subtitle: "\(saved.count) saved", kind: "watchlist", items: saved)]
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
    @EnvironmentObject private var store: XuvaClientStore
    @EnvironmentObject private var watchlist: XuvaWatchlist
    // @FocusState gives reliable programmatic routing to Play on every entry.
    // prefersDefaultFocus(in:) only applies on the FIRST scope entry; after that
    // the engine restores the last-focused item (often More Info if the user
    // tapped it), so DOWN from nav pills lands on More Info instead of Play.
    // A FocusState bool + task-delayed activation picks Play consistently.
    #if os(tvOS)
    @FocusState private var playFocused: Bool
    #endif

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
                // Play — navigates to detail screen where playback begins.
                Button {
                    Task { await store.open(item: item) }
                } label: {
                    Label(primaryActionTitle, systemImage: "play.fill")
                }
                .buttonStyle(XuvaPrimaryButtonStyle(viewport: viewport))
                #if os(tvOS)
                .focused($playFocused)
                #endif

                // More Info — opens the detail screen for this title.
                Button {
                    Task { await store.open(item: item) }
                } label: {
                    Label("More Info", systemImage: "info.circle")
                }
                .buttonStyle(XuvaSecondaryButtonStyle(viewport: viewport))

                // Watchlist toggle — adds or removes this title.
                Button {
                    guard let kind = item.kind else { return }
                    _ = watchlist.toggle(
                        id: item.id,
                        kind: kind,
                        title: item.title ?? "",
                        year: item.year,
                        posterUrl: item.posterUrl,
                        backdropUrl: item.backdropUrl,
                        genres: item.genres
                    )
                } label: {
                    Image(systemName: inWatchlist ? "checkmark" : "plus")
                }
                .buttonStyle(XuvaIconButtonStyle(viewport: viewport))
            }
            .frame(maxWidth: .infinity, alignment: .leading)
            // focusSection declares the button row as a discrete region so UP
            // from Play exits cleanly toward the nav bar section above.
            .focusSection()
            heroDots
                .padding(.top, 8)
        }
        .padding(.horizontal, XuvaScale.safeHorizontal(viewport))
        .padding(.top, viewport.height * XuvaScale.heroContentTopFraction(viewport))
        .frame(maxWidth: .infinity, alignment: .leading)
        .focusSection()
        #if os(tvOS)
        // Route focus to Play on initial appearance and whenever the featured
        // hero title changes (auto-advance or manual dot tap).  The 150 ms
        // delay gives the focus engine time to settle after the view tree
        // re-renders before we override it.
        .task(id: item.id) {
            try? await Task.sleep(nanoseconds: 150_000_000)
            playFocused = true
        }
        #endif
    }

    private var heroEyebrow: String {
        if item.progress ?? 0 > 0 { return "CONTINUE WATCHING" }
        return "XUVA PRESENTS"
    }

    private var primaryActionTitle: String {
        if item.progress ?? 0 > 0 { return "Resume" }
        return "Play"
    }

    private var inWatchlist: Bool {
        guard let kind = item.kind else { return false }
        return watchlist.isIn(id: item.id, kind: kind)
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
            // focusSection on the scroll view lets the tvOS focus engine exit
            // it vertically — without this, up/down swipes are absorbed and
            // you can only navigate left/right within the row.
            .focusSection()
        }
        .focusSection()
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
        // Server-provided eyebrow wins (e.g. "From your library", "Your region · AU").
        if let serverEyebrow = row.eyebrow, !serverEyebrow.isEmpty { return serverEyebrow }
        // Fallback labels match the web page's hardcoded eyebrows exactly.
        if normalizedRowText.contains("continue") { return "Pick up where you left off" }
        if normalizedRowText.contains("trend") { return "Trending now" }
        if isRanked { return "From your library" }
        if normalizedRowText.contains("movie") { return "Fresh in your library" }
        if normalizedRowText.contains("tv") || normalizedRowText.contains("series") { return "New episodes dropped" }
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
    var onSettings: (() -> Void)? = nil

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
                // Declare the nav pills as a distinct focus sub-section so
                // that UP from the left-aligned hero Play button enters the
                // pills cluster rather than the right-side refresh/settings
                // icons.  The focus engine resolves within a section by
                // geometric proximity — without this sub-section the whole
                // top bar is one flat region and refresh can win.
                .focusSection()
            }
            Spacer()
            // Wrap refresh + settings in their own sub-section so they are
            // treated as a discrete focus cluster on the RIGHT side of the
            // bar.  The nav pills already have their own .focusSection() on
            // the LEFT.  When focus travels UP from the hero Play button
            // (which is left-aligned), the engine resolves the nearest
            // sub-section by geometric proximity and lands on the pills
            // cluster rather than the refresh icon on the far right.
            HStack(spacing: 8) {
                Button {
                    Task { await store.loadHome() }
                } label: {
                    Image(systemName: "arrow.clockwise")
                }
                .buttonStyle(XuvaIconButtonStyle(viewport: viewport))
                if let onSettings {
                    Button {
                        onSettings()
                    } label: {
                        Image(systemName: "gearshape")
                    }
                    .buttonStyle(XuvaIconButtonStyle(viewport: viewport))
                }
            }
            .focusSection()
        }
        .padding(.horizontal, XuvaScale.safeHorizontal(viewport))
        .frame(height: XuvaScale.navBarHeight(viewport))
        .focusSection()
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
                .padding(.horizontal, XuvaScale.buttonHorizontalPadding(viewport) * 0.75)
                .frame(height: XuvaScale.buttonHeight(viewport) * 0.7)
        }
        .buttonStyle(NavPillButtonStyle(isActive: isActive))
    }
}

/// Custom nav pill style — uses our own subtle focus ring instead of the
/// tvOS default focus card which inflates the pill into a giant white
/// halo that overlaps neighbouring pills on a real TV.
private struct NavPillButtonStyle: ButtonStyle {
    let isActive: Bool
    @Environment(\.isFocused) private var isFocused

    func makeBody(configuration: Configuration) -> some View {
        NavPillBody(configuration: configuration, isActive: isActive)
    }

    private struct NavPillBody: View {
        let configuration: Configuration
        let isActive: Bool
        @Environment(\.isFocused) private var isFocused

        var body: some View {
            configuration.label
                .foregroundStyle(isActive || isFocused ? XuvaTheme.text : XuvaTheme.mutedText)
                .background(
                    isFocused
                        ? XuvaTheme.text.opacity(0.14)
                        : (isActive ? XuvaTheme.focus.opacity(0.12) : Color.clear),
                    in: Capsule()
                )
                .overlay(
                    Capsule().stroke(
                        isFocused
                            ? XuvaTheme.text.opacity(0.55)
                            : (isActive ? XuvaTheme.focus.opacity(0.30) : Color.clear),
                        lineWidth: isFocused ? 2 : 1
                    )
                )
                .scaleEffect(isFocused ? 1.04 : 1)
                .animation(.easeOut(duration: 0.15), value: isFocused)
                // Suppress the tvOS system blue halo — the custom ring
                // drawn above via isFocused / overlay is our focus indicator.
                .focusEffectDisabled()
        }
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
