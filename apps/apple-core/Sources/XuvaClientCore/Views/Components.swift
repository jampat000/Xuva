import SwiftUI

public struct XuvaLogo: View {
    let viewport: CGSize

    public init(viewport: CGSize = XuvaScale.screenSize) {
        self.viewport = viewport
    }

    public var body: some View {
        let mark = XuvaScale.clamped(28, viewport.width * 0.022, 56)
        let textSize = XuvaScale.clamped(16, viewport.width * 0.014, 32)
        HStack(spacing: mark * 0.30) {
            ZStack {
                RoundedRectangle(cornerRadius: mark * 0.28, style: .continuous)
                    .fill(XuvaTheme.action.opacity(0.18))
                RoundedRectangle(cornerRadius: mark * 0.31, style: .continuous)
                    .stroke(XuvaTheme.action.opacity(0.55), lineWidth: mark * 0.04)
                XuvaMark()
                    .stroke(
                        LinearGradient(
                            colors: [XuvaTheme.focus, XuvaTheme.text],
                            startPoint: .leading,
                            endPoint: .trailing
                        ),
                        style: StrokeStyle(lineWidth: max(2, mark * 0.085), lineCap: .round, lineJoin: .round)
                    )
                    .padding(mark * 0.25)
            }
            .frame(width: mark, height: mark)
            Text("Xuva")
                .font(.system(size: textSize, weight: .semibold))
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
        XuvaScale.clamped(24, maxHeight * 0.55, 86)
    }
}

struct PosterTile: View {
    let item: HomeItem
    let viewport: CGSize
    let ranked: Bool
    let rank: Int?
    let wide: Bool
    let action: () -> Void

    init(item: HomeItem, viewport: CGSize, ranked: Bool = false, rank: Int? = nil, wide: Bool = false, action: @escaping () -> Void) {
        self.item = item
        self.viewport = viewport
        self.ranked = ranked
        self.rank = rank
        self.wide = wide
        self.action = action
    }

    var body: some View {
        Button(action: action) {
            VStack(alignment: .leading, spacing: wide ? 10 : 0) {
                ZStack(alignment: .bottomLeading) {
                    RemoteImage(urlString: artworkURL, aspectRatio: wide ? 16 / 9 : 2 / 3)
                        .frame(width: posterWidth, height: posterHeight)
                        .clipShape(RoundedRectangle(cornerRadius: 12, style: .continuous))
                    LinearGradient(colors: [.clear, XuvaTheme.background.opacity(0.90)], startPoint: .center, endPoint: .bottom)
                        .clipShape(RoundedRectangle(cornerRadius: 12, style: .continuous))
                    VStack(alignment: .leading, spacing: 6) {
                        if let route = item.routeLabel, !route.isEmpty {
                            Text(route)
                                .font(.system(size: XuvaScale.eyebrowFontSize(viewport), weight: .bold))
                                .padding(.horizontal, 8)
                                .padding(.vertical, 4)
                                .background(XuvaTheme.focus.opacity(0.16), in: Capsule())
                                .foregroundStyle(XuvaTheme.focus)
                        }
                        if wide, let logoURL = item.logoUrl, !logoURL.isEmpty {
                            RemoteLogo(urlString: logoURL, fallbackTitle: item.title ?? "Untitled", maxWidth: posterWidth * 0.62, maxHeight: posterHeight * 0.35)
                        } else {
                            Text(item.title ?? "Untitled")
                                .font(.system(size: XuvaScale.metaFontSize(viewport) + 2, weight: .semibold))
                                .foregroundStyle(.white)
                                .lineLimit(2)
                        }
                        if !wide {
                            Text([item.year.map(String.init), item.subtitle].compactMap { $0 }.joined(separator: " · "))
                                .font(.system(size: XuvaScale.metaFontSize(viewport) - 2))
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
                            .font(.system(size: XuvaScale.metaFontSize(viewport), weight: .bold))
                            .padding(.horizontal, 9)
                            .padding(.vertical, 5)
                            .background(.black.opacity(0.58), in: Capsule())
                            .padding(12)
                            .frame(maxWidth: .infinity, maxHeight: .infinity, alignment: .topLeading)
                    }
                }
                if wide {
                    Text(item.title ?? "Untitled")
                        .font(.system(size: XuvaScale.metaFontSize(viewport) + 1, weight: .bold))
                        .foregroundStyle(XuvaTheme.text)
                        .lineLimit(1)
                    Text(wideSubtitle)
                        .font(.system(size: XuvaScale.metaFontSize(viewport) - 2))
                        .foregroundStyle(XuvaTheme.muted)
                        .lineLimit(1)
                }
            }
        }
        .buttonStyle(.plain)
        .xuvaFocused(radius: 12)
    }

    private var posterWidth: CGFloat {
        wide ? XuvaScale.widePosterWidth(viewport) : XuvaScale.posterWidth(viewport)
    }

    private var posterHeight: CGFloat {
        wide ? XuvaScale.widePosterHeight(viewport) : XuvaScale.posterHeight(viewport)
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
}

public struct RouteBadge: View {
    let decision: PlaybackDecision?
    let viewport: CGSize

    public init(decision: PlaybackDecision?, viewport: CGSize = XuvaScale.screenSize) {
        self.decision = decision
        self.viewport = viewport
    }

    public var body: some View {
        HStack(spacing: 7) {
            Circle().fill(color).frame(width: 7, height: 7)
            Text(decision?.badgeLabel ?? "Route")
                .font(.system(size: XuvaScale.eyebrowFontSize(viewport), weight: .bold))
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
        case "Pending": return XuvaTheme.warn
        default: return XuvaTheme.secondaryText
        }
    }
}

public struct XuvaPrimaryButtonStyle: ButtonStyle {
    let viewport: CGSize
    public init(viewport: CGSize = XuvaScale.screenSize) {
        self.viewport = viewport
    }

    public func makeBody(configuration: Configuration) -> some View {
        let h = XuvaScale.buttonHeight(viewport)
        configuration.label
            .font(.system(size: XuvaScale.buttonFontSize(viewport), weight: .semibold))
            .foregroundStyle(XuvaTheme.background)
            .padding(.horizontal, XuvaScale.buttonHorizontalPadding(viewport))
            .frame(height: h)
            .background(XuvaTheme.text, in: Capsule(style: .continuous))
            .scaleEffect(configuration.isPressed ? 0.97 : 1)
            .xuvaFocused(radius: h / 2)
    }
}

public struct XuvaSecondaryButtonStyle: ButtonStyle {
    let viewport: CGSize
    public init(viewport: CGSize = XuvaScale.screenSize) {
        self.viewport = viewport
    }

    public func makeBody(configuration: Configuration) -> some View {
        let h = XuvaScale.buttonHeight(viewport)
        configuration.label
            .font(.system(size: XuvaScale.buttonFontSize(viewport), weight: .medium))
            .foregroundStyle(XuvaTheme.text)
            .padding(.horizontal, XuvaScale.buttonHorizontalPadding(viewport))
            .frame(height: h)
            .background(Color.white.opacity(configuration.isPressed ? 0.12 : 0.06), in: Capsule(style: .continuous))
            .overlay(Capsule(style: .continuous).stroke(Color.white.opacity(0.15)))
            .xuvaFocused(radius: h / 2)
    }
}

public struct XuvaIconButtonStyle: ButtonStyle {
    let viewport: CGSize
    public init(viewport: CGSize = XuvaScale.screenSize) {
        self.viewport = viewport
    }

    public func makeBody(configuration: Configuration) -> some View {
        let size = XuvaScale.iconButtonSize(viewport)
        configuration.label
            .font(.system(size: XuvaScale.buttonFontSize(viewport) * 0.95))
            .foregroundStyle(XuvaTheme.text)
            .frame(width: size, height: size)
            .background(Color.white.opacity(configuration.isPressed ? 0.12 : 0.06), in: Circle())
            .overlay(Circle().stroke(Color.white.opacity(0.10)))
            .xuvaFocused(radius: size / 2)
    }
}

public struct MediaPill: View {
    let text: String
    let systemImage: String?
    let tint: Color
    let viewport: CGSize

    public init(text: String, systemImage: String? = nil, tint: Color = XuvaTheme.primaryGlow, viewport: CGSize = XuvaScale.screenSize) {
        self.text = text
        self.systemImage = systemImage
        self.tint = tint
        self.viewport = viewport
    }

    public var body: some View {
        HStack(spacing: 7) {
            if let systemImage {
                Image(systemName: systemImage)
                    .font(.system(size: XuvaScale.eyebrowFontSize(viewport), weight: .bold))
            }
            Text(text)
                .font(.system(size: XuvaScale.eyebrowFontSize(viewport), weight: .bold))
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
