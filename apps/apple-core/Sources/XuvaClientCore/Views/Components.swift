import SwiftUI

public struct XuvaLogo: View {
    public init() {}

    public var body: some View {
        HStack(spacing: 10) {
            ZStack {
                RoundedRectangle(cornerRadius: 9, style: .continuous)
                    .fill(XuvaTheme.action.opacity(0.18))
                RoundedRectangle(cornerRadius: 10, style: .continuous)
                    .stroke(XuvaTheme.action.opacity(0.55), lineWidth: 1.2)
                XuvaMark()
                    .stroke(
                        LinearGradient(
                            colors: [XuvaTheme.focus, XuvaTheme.text],
                            startPoint: .leading,
                            endPoint: .trailing
                        ),
                        style: StrokeStyle(lineWidth: 2.6, lineCap: .round, lineJoin: .round)
                    )
                    .padding(8)
            }
            .frame(width: 32, height: 32)
            Text("Xuva")
                .font(.title3.weight(.semibold))
                .tracking(0.5)
        }
        .foregroundStyle(XuvaTheme.text)
    }
}

private struct XuvaMark: Shape {
    func path(in rect: CGRect) -> Path {
        var path = Path()
        let leftX = rect.minX + rect.width * 0.16
        let midX = rect.minX + rect.width * 0.50
        let rightX = rect.minX + rect.width * 0.84
        let topY = rect.minY + rect.height * 0.16
        let centerY = rect.midY
        let bottomY = rect.minY + rect.height * 0.84

        path.move(to: CGPoint(x: leftX, y: topY))
        path.addLine(to: CGPoint(x: midX, y: centerY))
        path.addLine(to: CGPoint(x: leftX, y: bottomY))
        path.move(to: CGPoint(x: midX, y: topY))
        path.addLine(to: CGPoint(x: rightX, y: centerY))
        path.addLine(to: CGPoint(x: midX, y: bottomY))
        return path
    }
}

struct RemoteImage: View {
    let urlString: String?
    let aspectRatio: CGFloat

    var body: some View {
        AsyncImage(url: URL(string: urlString ?? "")) { phase in
            switch phase {
            case let .success(image):
                image.resizable().scaledToFill()
            default:
                ZStack {
                    LinearGradient(colors: [XuvaTheme.elevated, XuvaTheme.background], startPoint: .topLeading, endPoint: .bottomTrailing)
                    Image(systemName: "film")
                        .font(.system(size: 42, weight: .semibold))
                        .foregroundStyle(.white.opacity(0.20))
                }
            }
        }
        .aspectRatio(aspectRatio, contentMode: .fill)
    }
}

struct RemoteLogo: View {
    let urlString: String?
    let fallbackTitle: String
    let maxWidth: CGFloat
    let maxHeight: CGFloat

    var body: some View {
        AsyncImage(url: URL(string: urlString ?? "")) { phase in
            switch phase {
            case let .success(image):
                image
                    .resizable()
                    .scaledToFit()
                    .shadow(color: .black.opacity(0.72), radius: 18, y: 8)
            default:
                Text(fallbackTitle)
                    .font(.system(size: fallbackSize, weight: .black, design: .rounded))
                    .foregroundStyle(XuvaTheme.text)
                    .lineLimit(3)
                    .minimumScaleFactor(0.54)
                    .multilineTextAlignment(.leading)
            }
        }
        .frame(maxWidth: maxWidth, maxHeight: maxHeight, alignment: .leading)
    }

    private var fallbackSize: CGFloat {
        #if os(tvOS)
        return maxHeight > 120 ? 72 : 24
        #else
        return maxHeight > 100 ? 44 : 18
        #endif
    }
}

struct PosterTile: View {
    let item: HomeItem
    let ranked: Bool
    let rank: Int?
    let wide: Bool
    let action: () -> Void

    init(item: HomeItem, ranked: Bool = false, rank: Int? = nil, wide: Bool = false, action: @escaping () -> Void) {
        self.item = item
        self.ranked = ranked
        self.rank = rank
        self.wide = wide
        self.action = action
    }

    var body: some View {
        Button(action: action) {
            VStack(alignment: .leading, spacing: wide ? 12 : 0) {
                ZStack(alignment: .bottomLeading) {
                    RemoteImage(urlString: artworkURL, aspectRatio: wide ? 16 / 9 : 2 / 3)
                        .frame(width: posterWidth, height: posterHeight)
                        .clipShape(RoundedRectangle(cornerRadius: 10, style: .continuous))
                    LinearGradient(colors: [.clear, XuvaTheme.background.opacity(0.90)], startPoint: .center, endPoint: .bottom)
                        .clipShape(RoundedRectangle(cornerRadius: 10, style: .continuous))
                    VStack(alignment: .leading, spacing: 6) {
                        if let route = item.routeLabel, !route.isEmpty {
                            Text(route)
                                .font(.caption2.bold())
                                .padding(.horizontal, 8)
                                .padding(.vertical, 4)
                                .background(XuvaTheme.focus.opacity(0.16), in: Capsule())
                                .foregroundStyle(XuvaTheme.focus)
                        }
                        if wide, let logoURL = item.logoUrl, !logoURL.isEmpty {
                            RemoteLogo(urlString: logoURL, fallbackTitle: item.title ?? "Untitled", maxWidth: posterWidth * 0.62, maxHeight: 58)
                        } else {
                            Text(item.title ?? "Untitled")
                                .font(tileTitleFont)
                                .foregroundStyle(.white)
                                .lineLimit(2)
                        }
                        if !wide {
                            Text([item.year.map(String.init), item.subtitle].compactMap { $0 }.joined(separator: " · "))
                                .font(.caption)
                                .foregroundStyle(.white.opacity(0.68))
                                .lineLimit(1)
                        }
                        if let progress = item.progress, progress > 0 {
                            GeometryReader { proxy in
                                ZStack(alignment: .leading) {
                                    Capsule().fill(.white.opacity(0.18))
                                    Capsule().fill(XuvaTheme.action)
                                        .frame(width: proxy.size.width * min(max(progress, 0), 1))
                                }
                            }
                            .frame(height: 4)
                        }
                    }
                    .padding(14)
                    if ranked {
                        Text(rank.map { "#\($0)" } ?? "#")
                            .font(.caption.bold())
                            .padding(.horizontal, 9)
                            .padding(.vertical, 5)
                            .background(.black.opacity(0.58), in: Capsule())
                            .padding(12)
                            .frame(maxWidth: .infinity, maxHeight: .infinity, alignment: .topLeading)
                    }
                }
                if wide {
                    Text(item.title ?? "Untitled")
                        .font(.callout.weight(.bold))
                        .foregroundStyle(XuvaTheme.text)
                        .lineLimit(1)
                    Text(wideSubtitle)
                        .font(.caption)
                        .foregroundStyle(XuvaTheme.muted)
                        .lineLimit(1)
                }
            }
        }
        .buttonStyle(.plain)
        .xuvaFocused(radius: 10)
    }

    private var posterWidth: CGFloat {
        wide ? XuvaScale.widePosterWidth() : XuvaScale.posterWidth()
    }

    private var posterHeight: CGFloat {
        wide ? XuvaScale.widePosterHeight() : XuvaScale.posterHeight()
    }

    private var artworkURL: String? {
        wide ? (item.backdropUrl ?? item.imageUrl ?? item.posterUrl) : (item.posterUrl ?? item.imageUrl ?? item.backdropUrl)
    }

    private var wideSubtitle: String {
        var parts: [String] = []
        if let year = item.year { parts.append(String(year)) }
        if let kind = item.kind, !kind.isEmpty { parts.append(kind.capitalized) }
        if let progress = item.progress, progress > 0 {
            parts.append("Resume from \(Int((progress * 100).rounded()))%")
        }
        return parts.joined(separator: " · ")
    }

    private var tileTitleFont: Font {
        #if os(tvOS)
        return .system(size: 22, weight: .semibold)
        #else
        return .headline
        #endif
    }
}

public struct RouteBadge: View {
    let decision: PlaybackDecision?

    public init(decision: PlaybackDecision?) {
        self.decision = decision
    }

    public var body: some View {
        HStack(spacing: 7) {
            Circle().fill(color).frame(width: 7, height: 7)
            Text(decision?.badgeLabel ?? "Route")
                .font(.caption2.weight(.bold))
                .tracking(1.2)
                .textCase(.uppercase)
        }
        .foregroundStyle(color)
        .padding(.horizontal, 12)
        .padding(.vertical, 8)
        .background(color.opacity(0.14), in: Capsule())
        .overlay(Capsule().stroke(color.opacity(0.32)))
    }

    private var color: Color {
        switch decision?.badgeLabel {
        case "Direct Play": return XuvaTheme.focus
        case "Remux": return XuvaTheme.action
        case "Adaptive": return XuvaTheme.primaryGlow
        case "Transcoding": return XuvaTheme.danger
        case "Audio Tx": return XuvaTheme.warn
        default: return XuvaTheme.secondaryText
        }
    }
}

private enum XuvaButtonMetrics {
    static var height: CGFloat {
        #if os(tvOS)
        return 72
        #else
        return 52
        #endif
    }

    static var font: Font {
        #if os(tvOS)
        return .system(size: 26, weight: .semibold)
        #else
        return .system(size: 16, weight: .semibold)
        #endif
    }

    static var horizontalPadding: CGFloat {
        #if os(tvOS)
        return 34
        #else
        return 24
        #endif
    }

    static var iconSize: CGFloat {
        #if os(tvOS)
        return 60
        #else
        return 44
        #endif
    }
}

public struct XuvaPrimaryButtonStyle: ButtonStyle {
    public init() {}

    public func makeBody(configuration: Configuration) -> some View {
        configuration.label
            .font(XuvaButtonMetrics.font)
            .foregroundStyle(XuvaTheme.background)
            .padding(.horizontal, XuvaButtonMetrics.horizontalPadding)
            .frame(height: XuvaButtonMetrics.height)
            .background(XuvaTheme.text, in: Capsule(style: .continuous))
            .scaleEffect(configuration.isPressed ? 0.97 : 1)
            .xuvaFocused(radius: XuvaButtonMetrics.height / 2)
    }
}

public struct XuvaSecondaryButtonStyle: ButtonStyle {
    public init() {}

    public func makeBody(configuration: Configuration) -> some View {
        configuration.label
            .font(XuvaButtonMetrics.font)
            .foregroundStyle(XuvaTheme.text)
            .padding(.horizontal, XuvaButtonMetrics.horizontalPadding)
            .frame(height: XuvaButtonMetrics.height)
            .background(Color.white.opacity(configuration.isPressed ? 0.12 : 0.06), in: Capsule(style: .continuous))
            .overlay(Capsule(style: .continuous).stroke(Color.white.opacity(0.15)))
            .xuvaFocused(radius: XuvaButtonMetrics.height / 2)
    }
}

public struct XuvaIconButtonStyle: ButtonStyle {
    public init() {}

    public func makeBody(configuration: Configuration) -> some View {
        configuration.label
            .font(XuvaButtonMetrics.font)
            .foregroundStyle(XuvaTheme.text)
            .frame(width: XuvaButtonMetrics.iconSize, height: XuvaButtonMetrics.iconSize)
            .background(Color.white.opacity(configuration.isPressed ? 0.12 : 0.06), in: Circle())
            .overlay(Circle().stroke(Color.white.opacity(0.10)))
            .xuvaFocused(radius: XuvaButtonMetrics.iconSize / 2)
    }
}

public struct MediaPill: View {
    let text: String
    let systemImage: String?
    let tint: Color

    public init(text: String, systemImage: String? = nil, tint: Color = XuvaTheme.primaryGlow) {
        self.text = text
        self.systemImage = systemImage
        self.tint = tint
    }

    public var body: some View {
        HStack(spacing: 7) {
            if let systemImage {
                Image(systemName: systemImage)
                    .font(.caption.weight(.bold))
            }
            Text(text)
                .font(.caption.weight(.bold))
                .tracking(1.4)
                .textCase(.uppercase)
                .lineLimit(1)
        }
        .foregroundStyle(tint)
        .padding(.horizontal, 12)
        .padding(.vertical, 8)
        .background(tint.opacity(0.14), in: Capsule())
        .overlay(Capsule().stroke(tint.opacity(0.28)))
    }
}
