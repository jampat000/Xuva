import SwiftUI

struct HomeView: View {
    @EnvironmentObject private var appState: LorivoAppState

    var body: some View {
        ZStack(alignment: .bottomLeading) {
            HeroBackdrop(title: appState.focusedPoster.title)

            VStack(alignment: .leading, spacing: 40) {
                HeaderBar()
                HeroCopy(poster: appState.focusedPoster)
                ForEach(homeRows) { row in
                    PosterRow(title: row.title, posters: row.items.map { $0.posterModel() })
                }
            }
            .padding(.horizontal, LorivoTheme.horizontalMargin)
            .padding(.top, 54)
            .padding(.bottom, 62)
        }
        .ignoresSafeArea()
    }

    private var homeRows: [TVHomeRow] {
        let rows = appState.home?.rows.filter { !$0.items.isEmpty } ?? []
        if rows.isEmpty {
            return [TVHomeRow(id: "samples", title: "Ready for your library", items: MediaPoster.samples.map {
                TVHomeItem(id: $0.id, kind: "sample", title: $0.title, subtitle: $0.subtitle, posterUrl: nil, backdropUrl: nil, mediaSourceId: nil, route: $0.route)
            })]
        }
        return rows
    }
}

private struct HeaderBar: View {
    @EnvironmentObject private var appState: LorivoAppState

    var body: some View {
        HStack {
            Text("Lorivo")
                .font(.system(size: 34, weight: .bold))
                .foregroundStyle(LorivoTheme.text)
            Text("Movies")
                .font(.system(size: 24, weight: .semibold))
                .foregroundStyle(LorivoTheme.soft)
                .padding(.leading, 30)
            Text("TV Shows")
                .font(.system(size: 24, weight: .semibold))
                .foregroundStyle(LorivoTheme.quiet)
                .padding(.leading, 22)
            Spacer()
            RouteBadge(text: appState.bootstrap?.server.name ?? "Local Server")
        }
    }
}

private struct HeroBackdrop: View {
    let title: String

    var body: some View {
        ZStack {
            LinearGradient(colors: [LorivoTheme.graphite, LorivoTheme.carbon, LorivoTheme.cinema], startPoint: .topTrailing, endPoint: .bottomLeading)
            Text(title.prefix(1))
                .font(.system(size: 360, weight: .black))
                .foregroundStyle(LorivoTheme.text.opacity(0.035))
                .offset(x: 420, y: -120)
            LinearGradient(colors: [.black.opacity(0.12), .black.opacity(0.86)], startPoint: .top, endPoint: .bottom)
        }
    }
}

private struct HeroCopy: View {
    let poster: MediaPoster

    var body: some View {
        VStack(alignment: .leading, spacing: 16) {
            RouteBadge(text: poster.route)
            Text(poster.title)
                .font(.system(size: 68, weight: .bold))
                .foregroundStyle(LorivoTheme.text)
            Text(poster.subtitle)
                .font(.system(size: 26, weight: .medium))
                .foregroundStyle(LorivoTheme.soft)
            HStack(spacing: 18) {
                Button("Play") {}
                    .buttonStyle(.borderedProminent)
                    .tint(LorivoTheme.amber)
                Button("Details") {}
                    .buttonStyle(.bordered)
            }
            .font(.system(size: 26, weight: .bold))
            .padding(.top, 8)
        }
        .frame(maxWidth: 820, alignment: .leading)
    }
}

private struct PosterRow: View {
    let title: String
    let posters: [MediaPoster]

    var body: some View {
        VStack(alignment: .leading, spacing: 18) {
            Text(title)
                .font(.system(size: 30, weight: .semibold))
                .foregroundStyle(LorivoTheme.text)
            ScrollView(.horizontal, showsIndicators: false) {
                HStack(spacing: 22) {
                    ForEach(posters) { poster in
                        PosterCard(poster: poster)
                    }
                }
                .padding(.vertical, 10)
            }
        }
    }
}
