import SwiftUI

public enum XuvaBrand {
    /// Brand gradient — matches `--gradient-primary` in apps/web/svelte
    /// (135° from oklch(0.55 0.22 285) → oklch(0.65 0.20 280) → oklch(0.72 0.18 265)).
    public static let gradient = LinearGradient(
        colors: [
            Color(red: 0.435, green: 0.259, blue: 0.878),
            Color(red: 0.529, green: 0.388, blue: 0.941),
            Color(red: 0.604, green: 0.490, blue: 1.000)
        ],
        startPoint: .topLeading,
        endPoint: .bottomTrailing
    )

    public static let wordmarkGradient = LinearGradient(
        colors: [
            Color(red: 0.486, green: 0.361, blue: 1.000),
            Color(red: 0.616, green: 0.478, blue: 1.000)
        ],
        startPoint: .topLeading,
        endPoint: .bottomTrailing
    )

    public static let chevronHighlight = LinearGradient(
        colors: [
            Color.white.opacity(0.95),
            Color.white.opacity(0.55)
        ],
        startPoint: .topLeading,
        endPoint: .bottomTrailing
    )
}

public struct XuvaLogo: View {
    let viewport: CGSize

    public init(viewport: CGSize = XuvaScale.screenSize) {
        self.viewport = viewport
    }

    public var body: some View {
        let mark = XuvaScale.clamped(28, viewport.width * 0.024, 64)
        let textSize = XuvaScale.clamped(17, viewport.width * 0.015, 34)
        HStack(spacing: mark * 0.32) {
            XuvaIconMark(size: mark)
            (
                Text("X").foregroundStyle(XuvaTheme.text)
                + Text("uva").foregroundStyle(XuvaBrand.wordmarkGradient)
            )
            .font(.system(size: textSize, weight: .semibold).width(.condensed))
            .tracking(-0.4)
        }
    }
}

/// The Xuva mark — rounded square with two chevrons (left in brand gradient,
/// right in white) — matching `apps/web/svelte/src/lib/components/Logo.svelte`.
public struct XuvaIconMark: View {
    let size: CGFloat

    public init(size: CGFloat) {
        self.size = size
    }

    public var body: some View {
        let stroke = max(2.5, size * 0.105)
        ZStack {
            RoundedRectangle(cornerRadius: size * 0.25, style: .continuous)
                .fill(XuvaBrand.gradient.opacity(0.18))
            RoundedRectangle(cornerRadius: size * 0.25, style: .continuous)
                .strokeBorder(XuvaBrand.gradient.opacity(0.50), lineWidth: max(1, size * 0.022))

            ChevronShape(offsetXFraction: 0.292)
                .stroke(XuvaBrand.gradient, style: StrokeStyle(lineWidth: stroke, lineCap: .round, lineJoin: .round))
                .blur(radius: size * 0.035)
                .opacity(0.65)

            ChevronShape(offsetXFraction: 0.292)
                .stroke(XuvaBrand.gradient, style: StrokeStyle(lineWidth: stroke, lineCap: .round, lineJoin: .round))

            ChevronShape(offsetXFraction: 0.458)
                .stroke(XuvaBrand.chevronHighlight, style: StrokeStyle(lineWidth: stroke, lineCap: .round, lineJoin: .round))
        }
        .frame(width: size, height: size)
    }
}

/// A single right-pointing chevron. `offsetXFraction` is the horizontal start
/// position in the icon's 48-unit grid (matching the SVG `M14 …` / `M22 …`).
private struct ChevronShape: Shape {
    /// Where the chevron's leading vertex starts horizontally, as a fraction of the icon.
    let offsetXFraction: CGFloat

    func path(in rect: CGRect) -> Path {
        // Web SVG uses a 48-unit viewBox with chevrons spanning 14→28 (x) and 12→36 (y).
        // We rescale into the available rect's interior.
        let unit = rect.width
        let topY = rect.minY + unit * 0.250
        let midY = rect.midY
        let bottomY = rect.minY + unit * 0.750
        let startX = rect.minX + unit * offsetXFraction
        let tipX = startX + unit * 0.292

        var path = Path()
        path.move(to: CGPoint(x: startX, y: topY))
        path.addLine(to: CGPoint(x: tipX, y: midY))
        path.addLine(to: CGPoint(x: startX, y: bottomY))
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
