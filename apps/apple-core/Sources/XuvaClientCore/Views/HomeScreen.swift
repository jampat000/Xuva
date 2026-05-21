import SwiftUI

public struct HomeScreen: View {
    @EnvironmentObject private var store: XuvaClientStore

    public init() {}

    public var body: some View {
        GeometryReader { geometry in
            ScrollView {
                VStack(alignment: .leading, spacing: geometry.size.width < 700 ? 28 : 42) {
                    HeroView(item: hero)
                    if rows.isEmpty {
                        EmptyLibraryView()
                            .padding(.horizontal, horizontalPadding(for: geometry.size))
                    } else {
                        ForEach(rows) { row in
                            MediaRowView(row: row, horizontalPadding: horizontalPadding(for: geometry.size)) { item in
                                Task { await store.open(item: item) }
                            }
                        }
                    }
                }
                .padding(.bottom, 80)
            }
        }
        .safeAreaInset(edge: .top) {
            HStack {
                XuvaLogo()
                Spacer()
                Button {
                    Task { await store.loadHome() }
                } label: {
                    Image(systemName: "arrow.clockwise")
                }
                .buttonStyle(XuvaIconButtonStyle())
                Button {
                    store.resetConnection()
                } label: {
                    Image(systemName: "rectangle.portrait.and.arrow.right")
                }
                .buttonStyle(XuvaIconButtonStyle())
            }
            .padding(.horizontal, topBarPadding)
            .padding(.vertical, 18)
            .background(XuvaTheme.background.opacity(0.78))
        }
    }

    private var hero: HomeItem {
        store.home?.hero ?? store.home?.heroes?.first ?? rows.flatMap { $0.items ?? [] }.first ?? HomeItem(id: "empty", kind: "movie", title: "Your cinema awaits", subtitle: "Add media to Xuva to fill this screen.", year: nil, rating: nil, runtime: nil, progress: nil, posterUrl: nil, backdropUrl: nil, imageUrl: nil, logoUrl: nil, mediaSourceId: nil, routeLabel: nil, genres: ["Local-first"], overview: "Xuva will show your own library here after the server returns a client home feed.")
    }

    private var rows: [HomeRow] {
        store.home?.rows ?? []
    }

    private var topBarPadding: CGFloat {
        #if os(tvOS)
        return 64
        #else
        return 22
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
    @EnvironmentObject private var store: XuvaClientStore

    var body: some View {
        GeometryReader { geometry in
            ZStack(alignment: .bottomLeading) {
                RemoteImage(urlString: item.backdropUrl ?? item.imageUrl, aspectRatio: 16 / 9)
                    .frame(maxWidth: .infinity)
                    .frame(height: heroHeight(for: geometry.size))
                    .clipped()
                LinearGradient(colors: [.clear, XuvaTheme.background], startPoint: .top, endPoint: .bottom)
                LinearGradient(colors: [XuvaTheme.background, .clear], startPoint: .leading, endPoint: .trailing)
                VStack(alignment: .leading, spacing: geometry.size.width < 700 ? 12 : 18) {
                    Text("XUVA PRESENTS")
                        .font(.caption.weight(.bold))
                        .tracking(3.2)
                        .foregroundStyle(XuvaTheme.primaryGlow)
                    Text(item.title ?? "Your cinema awaits")
                        .font(.system(size: titleSize(for: geometry.size), weight: .black, design: .rounded))
                        .foregroundStyle(XuvaTheme.text)
                        .lineLimit(3)
                        .minimumScaleFactor(0.6)
                        .frame(maxWidth: 820, alignment: .leading)
                    Text(item.overview ?? item.subtitle ?? "")
                        .font(geometry.size.width < 700 ? .body : .title3)
                        .foregroundStyle(XuvaTheme.muted)
                        .lineLimit(3)
                        .frame(maxWidth: 680, alignment: .leading)
                    HStack(spacing: 12) {
                        Button {
                            Task { await store.open(item: item) }
                        } label: {
                            Label("Open", systemImage: "play.fill")
                        }
                        .buttonStyle(XuvaPrimaryButtonStyle())

                        Button("Details") {
                            Task { await store.open(item: item) }
                        }
                        .buttonStyle(XuvaSecondaryButtonStyle())
                    }
                }
                .padding(.horizontal, horizontalPadding(for: geometry.size))
                .padding(.bottom, bottomPadding(for: geometry.size))
            }
        }
        .frame(height: preferredHeight)
    }

    private var preferredHeight: CGFloat {
        #if os(tvOS)
        return 620
        #else
        return 430
        #endif
    }

    private func heroHeight(for size: CGSize) -> CGFloat {
        #if os(tvOS)
        return 620
        #else
        return min(430, max(340, size.width * 0.92))
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

    private func bottomPadding(for size: CGSize) -> CGFloat {
        #if os(tvOS)
        return 72
        #else
        return 28
        #endif
    }
}

struct MediaRowView: View {
    let row: HomeRow
    let horizontalPadding: CGFloat
    let action: (HomeItem) -> Void

    var body: some View {
        VStack(alignment: .leading, spacing: 16) {
            HStack(alignment: .lastTextBaseline) {
                Text(row.title ?? "Library")
                    .font(.title2.weight(.bold))
                    .foregroundStyle(XuvaTheme.text)
                if let subtitle = row.subtitle {
                    Text(subtitle)
                        .font(.callout)
                        .foregroundStyle(XuvaTheme.muted)
                }
            }
            .padding(.horizontal, horizontalPadding)

            ScrollView(.horizontal, showsIndicators: false) {
                HStack(spacing: 22) {
                    ForEach(row.items ?? []) { item in
                        PosterTile(item: item, ranked: row.title?.lowercased().contains("top") == true) {
                            action(item)
                        }
                    }
                }
                .padding(.horizontal, horizontalPadding)
                .padding(.vertical, 12)
            }
        }
    }
}

struct EmptyLibraryView: View {
    var body: some View {
        VStack(alignment: .leading, spacing: 12) {
            Label("No library rows yet", systemImage: "film.stack")
                .font(.headline)
                .foregroundStyle(XuvaTheme.text)
            Text("Once the server returns a client home feed, Xuva will show Continue Watching, Movies, TV, and Recently Added here.")
                .foregroundStyle(XuvaTheme.muted)
        }
        .padding(22)
        .frame(maxWidth: 620, alignment: .leading)
        .background(XuvaTheme.surface.opacity(0.82), in: RoundedRectangle(cornerRadius: 12, style: .continuous))
        .overlay(RoundedRectangle(cornerRadius: 12, style: .continuous).stroke(XuvaTheme.hairline))
    }
}
