import SwiftUI

public struct DetailScreen: View {
    @EnvironmentObject private var store: XuvaClientStore

    public init() {}

    public var body: some View {
        guard let detail = store.selectedDetail else {
            return AnyView(Text("No title selected").foregroundStyle(XuvaTheme.text))
        }
        return AnyView(
            GeometryReader { geometry in
                ScrollView {
                    ZStack(alignment: .bottomLeading) {
                        RemoteImage(urlString: detail.displayBackdropURL, aspectRatio: 16 / 9)
                            .frame(height: backdropHeight(for: geometry.size))
                            .clipped()
                        LinearGradient(colors: [.clear, XuvaTheme.background], startPoint: .top, endPoint: .bottom)
                        LinearGradient(colors: [XuvaTheme.background, .clear], startPoint: .leading, endPoint: .trailing)
                    }

                    detailContent(detail, geometry: geometry)
                        .padding(.horizontal, horizontalPadding(for: geometry.size))
                        .padding(.top, contentTopOffset(for: geometry.size))
                        .padding(.bottom, 80)
                }
            }
        )
    }

    @ViewBuilder
    private func detailContent(_ detail: DetailResponse, geometry: GeometryProxy) -> some View {
        let isCompact = geometry.size.width < 700
        let layout: AnyLayout = if isCompact {
            AnyLayout(VStackLayout(alignment: .leading, spacing: 22))
        } else {
            AnyLayout(HStackLayout(alignment: .top, spacing: 42))
        }

        layout {
            RemoteImage(urlString: detail.displayPosterURL, aspectRatio: 2 / 3)
                .frame(width: posterSize(for: geometry.size).width, height: posterSize(for: geometry.size).height)
                .clipShape(RoundedRectangle(cornerRadius: 8, style: .continuous))
                .shadow(color: .black.opacity(0.46), radius: 32, y: 24)

            VStack(alignment: .leading, spacing: isCompact ? 14 : 18) {
                Button {
                    store.backToHome()
                } label: {
                    Label("Back", systemImage: "chevron.left")
                }
                .buttonStyle(XuvaSecondaryButtonStyle())

                Text(metadataLine(detail))
                    .font(.caption.weight(.bold))
                    .tracking(2.4)
                    .foregroundStyle(XuvaTheme.muted)
                Text(detail.displayTitle)
                    .font(.system(size: titleSize(for: geometry.size), weight: .black, design: .rounded))
                    .foregroundStyle(XuvaTheme.text)
                    .lineLimit(3)
                    .minimumScaleFactor(0.6)
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
                HStack(spacing: 12) {
                    Button {
                        Task { await store.play() }
                    } label: {
                        Label("Play", systemImage: "play.fill")
                    }
                    .buttonStyle(XuvaPrimaryButtonStyle())

                    RouteBadge(decision: detail.playbackDecision ?? detail.versions?.first?.decision)
                }
                VersionDeck(versions: detail.versions ?? []) { version in
                    Task { await store.play(version: version) }
                }
            }
            if !isCompact {
                Spacer()
            }
        }
    }

    private func backdropHeight(for size: CGSize) -> CGFloat {
        #if os(tvOS)
        return 560
        #else
        return size.width > 700 ? 420 : 280
        #endif
    }

    private func contentTopOffset(for size: CGSize) -> CGFloat {
        #if os(tvOS)
        return -180
        #else
        return size.width > 700 ? -120 : -80
        #endif
    }

    private func posterSize(for size: CGSize) -> CGSize {
        #if os(tvOS)
        return CGSize(width: 250, height: 375)
        #else
        return size.width > 700 ? CGSize(width: 190, height: 285) : CGSize(width: 132, height: 198)
        #endif
    }

    private func titleSize(for size: CGSize) -> CGFloat {
        #if os(tvOS)
        return 60
        #else
        return size.width > 700 ? 44 : 34
        #endif
    }

    private func horizontalPadding(for size: CGSize) -> CGFloat {
        #if os(tvOS)
        return 64
        #else
        return size.width > 700 ? 36 : 20
        #endif
    }

    private func metadataLine(_ detail: DetailResponse) -> String {
        [detail.displayYear.map(String.init), detail.displayRuntime, detail.contentRating, detail.displayGenres.prefix(2).joined(separator: " / ")]
            .compactMap { $0 }
            .filter { !$0.isEmpty }
            .joined(separator: "  ·  ")
    }
}

struct VersionDeck: View {
    let versions: [MediaVersion]
    let play: (MediaVersion) -> Void

    var body: some View {
        if !versions.isEmpty {
            VStack(alignment: .leading, spacing: 12) {
                Text("Versions")
                    .font(.headline)
                ForEach(versions, id: \.stableID) { version in
                    Button {
                        play(version)
                    } label: {
                        HStack(spacing: 14) {
                            Image(systemName: "film.stack")
                            VStack(alignment: .leading, spacing: 4) {
                                Text(version.qualityLabel ?? version.name ?? "Original")
                                    .font(.headline)
                                Text([version.resolution, version.videoCodec, version.audioSummary].compactMap { $0 }.joined(separator: " · "))
                                    .font(.caption)
                                    .foregroundStyle(XuvaTheme.muted)
                            }
                            Spacer()
                            RouteBadge(decision: version.decision)
                        }
                        .padding(16)
                        .background(XuvaTheme.surface.opacity(0.80), in: RoundedRectangle(cornerRadius: 18, style: .continuous))
                    }
                    .buttonStyle(.plain)
                    .xuvaFocused(radius: 18)
                }
            }
            .frame(maxWidth: 720, alignment: .leading)
            .padding(.top, 18)
        }
    }
}
