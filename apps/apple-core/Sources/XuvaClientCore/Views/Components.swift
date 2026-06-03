import SwiftUI

public enum XuvaBrand {
    /// Brand gradient — matches `--gradient-primary` in apps/web/svelte
    /// (135° from oklch(0.55 0.22 285) → oklch(0.65 0.20 280) → oklch(0.72 0.18 265)).
    ///
    /// Colors are expressed in Display P3 for accurate hue on wide-gamut Apple TV
    /// displays. The original sRGB conversions were computed with a broken tool and
    /// had incorrect green values; stops 1 and 2 also exceed the sRGB gamut (P3
    /// stop 2 is still boundary-clipped but is far closer to the intended hue).
    /// Middle stop is explicit at 0.55 to match the web favicon/Logo.svelte source.
    ///
    /// Reference conversions (python oklch → linear-sRGB → XYZ-D65 → linear-P3 → γ):
    ///   oklch(0.55 0.22 285) → P3 (0.406, 0.316, 0.882)   [sRGB in-gamut]
    ///   oklch(0.65 0.20 280) → P3 (0.483, 0.476, 0.987)   [sRGB slightly OOG]
    ///   oklch(0.72 0.18 265) → P3 (0.467, 0.615, 1.000)   [P3 boundary-clipped]
    public static let gradient = LinearGradient(
        gradient: Gradient(stops: [
            .init(color: Color(.displayP3, red: 0.406, green: 0.316, blue: 0.882), location: 0.00),
            .init(color: Color(.displayP3, red: 0.483, green: 0.476, blue: 0.987), location: 0.55),
            .init(color: Color(.displayP3, red: 0.467, green: 0.615, blue: 1.000), location: 1.00),
        ]),
        startPoint: .topLeading,
        endPoint: .bottomTrailing
    )

    /// Wordmark gradient — matches `--gradient-primary` (2-stop variant used for
    /// the "uva" text in the logo lockup, 135° diagonal).
    ///   oklch(0.58 0.22 285) → P3 (0.438, 0.356, 0.922)   [sRGB in-gamut]
    ///   oklch(0.70 0.18 265) → P3 (0.444, 0.590, 1.000)   [P3 boundary-clipped]
    public static let wordmarkGradient = LinearGradient(
        gradient: Gradient(stops: [
            .init(color: Color(.displayP3, red: 0.438, green: 0.356, blue: 0.922), location: 0.00),
            .init(color: Color(.displayP3, red: 0.444, green: 0.590, blue: 1.000), location: 1.00),
        ]),
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
            .font(.system(size: textSize, weight: .semibold, design: .default))
            .tracking(textSize * -0.025)
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

            // Glow pass — both chevrons are wrapped in a single filter group in
            // the web favicon, so both get the blur. Replicate that here.
            ChevronShape(offsetXFraction: 0.292)
                .stroke(XuvaBrand.gradient, style: StrokeStyle(lineWidth: stroke, lineCap: .round, lineJoin: .round))
                .blur(radius: size * 0.040)
                .opacity(0.65)

            ChevronShape(offsetXFraction: 0.458)
                .stroke(XuvaBrand.chevronHighlight, style: StrokeStyle(lineWidth: stroke, lineCap: .round, lineJoin: .round))
                .blur(radius: size * 0.040)
                .opacity(0.55)

            // Solid pass on top of the glow
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
            VStack(alignment: .leading, spacing: wide ? 10 : 6) {
                ZStack(alignment: .bottomLeading) {
                    RemoteImage(urlString: artworkURL, aspectRatio: wide ? 16 / 9 : 2 / 3)
                        .frame(width: posterWidth, height: posterHeight)
                        .clipShape(RoundedRectangle(cornerRadius: 12, style: .continuous))
                    LinearGradient(colors: [.clear, XuvaTheme.background.opacity(0.90)], startPoint: .center, endPoint: .bottom)
                        .clipShape(RoundedRectangle(cornerRadius: 12, style: .continuous))
                    VStack(alignment: .leading, spacing: 4) {
                        if let route = item.routeLabel, !route.isEmpty {
                            Text(route)
                                .font(.system(size: overlayEyebrowSize, weight: .bold))
                                .padding(.horizontal, 6)
                                .padding(.vertical, 3)
                                .background(XuvaTheme.focus.opacity(0.16), in: Capsule())
                                .foregroundStyle(XuvaTheme.focus)
                        }
                        if wide, let logoURL = item.logoUrl, !logoURL.isEmpty {
                            RemoteLogo(urlString: logoURL, fallbackTitle: item.title ?? "Untitled", maxWidth: posterWidth * 0.62, maxHeight: posterHeight * 0.35)
                        } else {
                            // Title lives on the poster only — not repeated below.
                            Text(item.title ?? "Untitled")
                                .font(.system(size: overlayTitleSize, weight: .semibold))
                                .foregroundStyle(.white)
                                .lineLimit(2)
                        }
                        // Year + rating on the poster (matching web/desktop layout).
                        let hasYear = item.year != nil
                        let hasRating = (item.rating ?? 0) > 0
                        if hasYear || hasRating {
                            HStack(spacing: 4) {
                                if let year = item.year {
                                    Text(String(year))
                                }
                                if hasYear && hasRating {
                                    Circle().fill(.white.opacity(0.45)).frame(width: 3, height: 3)
                                }
                                if let rating = item.rating, rating > 0 {
                                    Image(systemName: "star.fill")
                                        .foregroundStyle(Color(red: 1.0, green: 0.82, blue: 0.35))
                                    Text(String(format: "%.1f", rating))
                                }
                            }
                            .font(.system(size: overlaySubSize * 0.85, weight: .medium))
                            .foregroundStyle(.white.opacity(0.78))
                        }
                        if let progress = item.progress, progress > 0 {
                            GeometryReader { proxy in
                                ZStack(alignment: .leading) {
                                    Capsule().fill(.white.opacity(0.18))
                                    Capsule().fill(XuvaTheme.action)
                                        .frame(width: proxy.size.width * min(max(progress, 0), 1))
                                }
                            }
                            .frame(height: 3)
                        }
                    }
                    .padding(overlayPad)
                    if ranked {
                        Text(rank.map { "#\($0)" } ?? "#")
                            .font(.system(size: overlaySubSize, weight: .bold))
                            .padding(.horizontal, 7)
                            .padding(.vertical, 4)
                            .background(.black.opacity(0.58), in: Capsule())
                            .padding(10)
                            .frame(maxWidth: .infinity, maxHeight: .infinity, alignment: .topLeading)
                    }
                }
                // Below the card: supplementary metadata only — title is already
                // on the poster so we never repeat it here.
                // Always reserve at least one line height so all tiles in a row
                // bottom-align consistently, even when a particular card has no
                // runtime / subtitle to display.
                let meta = wide ? wideSubtitle : posterMeta
                Text(meta)
                    .font(.system(size: overlaySubSize))
                    .foregroundStyle(XuvaTheme.muted)
                    .lineLimit(1)
                    // Fixed height (exactly one line) keeps every tile in the
                    // row the same total height regardless of whether the card
                    // has a runtime string or not — eliminates the uneven
                    // vertical spacing the user sees when tiles vary in height.
                    .frame(minWidth: posterWidth, maxWidth: posterWidth,
                           minHeight: overlaySubSize * 1.4, maxHeight: overlaySubSize * 1.4,
                           alignment: .topLeading)
            }
        }
        // PosterTileButtonStyle reads isFocused from inside makeBody — the only
        // correct way to detect button focus from a modifier. .xuvaFocused()
        // applied outside a Button reads the parent's isFocused, always false.
        .buttonStyle(PosterTileButtonStyle(cardWidth: posterWidth, cardHeight: posterHeight))
    }

    private var posterWidth: CGFloat {
        wide ? XuvaScale.widePosterWidth(viewport) : XuvaScale.posterWidth(viewport)
    }

    // Overlay sizes derived from card width so they scale with the card,
    // not with the full viewport (which produces 28pt text in a 220pt card on tvOS).
    private var overlayPad: CGFloat       { max(8,  posterWidth * 0.058) }
    private var overlayTitleSize: CGFloat { max(13, posterWidth * 0.083) }
    private var overlaySubSize: CGFloat   { max(10, posterWidth * 0.063) }
    private var overlayEyebrowSize: CGFloat { max(8, posterWidth * 0.050) }

    private var posterHeight: CGFloat {
        wide ? XuvaScale.widePosterHeight(viewport) : XuvaScale.posterHeight(viewport)
    }

    private var artworkURL: String? {
        wide ? (item.backdropUrl ?? item.imageUrl ?? item.posterUrl) : (item.posterUrl ?? item.imageUrl ?? item.backdropUrl)
    }

    // Runtime below the portrait poster — year/rating are shown on the poster itself.
    private var posterMeta: String {
        if let minutes = item.runtimeMinutes, minutes > 0 {
            let h = minutes / 60; let m = minutes % 60
            return h > 0 ? "\(h)h \(m)m" : "\(m)m"
        }
        return item.runtime ?? ""
    }

    // Metadata line below wide (Continue Watching) cards: year · kind · progress
    private var wideSubtitle: String {
        var parts: [String] = []
        if let year = item.year { parts.append(String(year)) }
        if let kind = item.kind, !kind.isEmpty { parts.append(kind.capitalized) }
        if let progress = item.progress, progress > 0 {
            parts.append("Resume \(Int((progress * 100).rounded()))%")
        }
        return parts.joined(separator: "  ·  ")
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
        PrimaryBody(configuration: configuration, viewport: viewport)
    }

    // Inner View struct so @Environment(\.isFocused) reads the button's own
    // focus state, not the parent container's (ButtonStyle.makeBody has no
    // direct environment access — it must delegate to a proper View type).
    private struct PrimaryBody: View {
        let configuration: Configuration
        let viewport: CGSize
        @Environment(\.isFocused) private var isFocused

        var body: some View {
            let h = XuvaScale.buttonHeight(viewport)
            configuration.label
                .font(.system(size: XuvaScale.buttonFontSize(viewport), weight: .semibold))
                .foregroundStyle(XuvaTheme.background)
                .padding(.horizontal, XuvaScale.buttonHorizontalPadding(viewport))
                .frame(height: h)
                .background(XuvaTheme.text, in: Capsule(style: .continuous))
                .scaleEffect(configuration.isPressed ? 0.97 : (isFocused ? 1.035 : 1))
                .shadow(color: XuvaTheme.focus.opacity(isFocused ? 0.42 : 0), radius: isFocused ? 26 : 0, x: 0, y: 14)
                .overlay(Capsule(style: .continuous).stroke(XuvaTheme.focus.opacity(isFocused ? 0.95 : 0), lineWidth: 3))
                .focusEffectDisabled()
                .animation(.spring(response: 0.25, dampingFraction: 0.78), value: isFocused)
        }
    }
}

public struct XuvaSecondaryButtonStyle: ButtonStyle {
    let viewport: CGSize
    public init(viewport: CGSize = XuvaScale.screenSize) {
        self.viewport = viewport
    }

    public func makeBody(configuration: Configuration) -> some View {
        SecondaryBody(configuration: configuration, viewport: viewport)
    }

    private struct SecondaryBody: View {
        let configuration: Configuration
        let viewport: CGSize
        @Environment(\.isFocused) private var isFocused

        var body: some View {
            let h = XuvaScale.buttonHeight(viewport)
            configuration.label
                .font(.system(size: XuvaScale.buttonFontSize(viewport), weight: .medium))
                .foregroundStyle(XuvaTheme.text)
                .padding(.horizontal, XuvaScale.buttonHorizontalPadding(viewport))
                .frame(height: h)
                .background(Color.white.opacity(configuration.isPressed ? 0.12 : 0.06), in: Capsule(style: .continuous))
                .overlay(Capsule(style: .continuous).stroke(
                    isFocused ? XuvaTheme.focus.opacity(0.95) : Color.white.opacity(0.15),
                    lineWidth: isFocused ? 3 : 1))
                .scaleEffect(isFocused ? 1.035 : 1)
                .shadow(color: XuvaTheme.focus.opacity(isFocused ? 0.42 : 0), radius: isFocused ? 26 : 0, x: 0, y: 14)
                .focusEffectDisabled()
                .animation(.spring(response: 0.25, dampingFraction: 0.78), value: isFocused)
        }
    }
}

/// A button style for destructive actions (delete, remove). Uses a red tint
/// with the same shape and focus behaviour as XuvaSecondaryButtonStyle.
public struct XuvaDestructiveButtonStyle: ButtonStyle {
    let viewport: CGSize
    public init(viewport: CGSize = XuvaScale.screenSize) {
        self.viewport = viewport
    }

    public func makeBody(configuration: Configuration) -> some View {
        DestructiveBody(configuration: configuration, viewport: viewport)
    }

    private struct DestructiveBody: View {
        let configuration: Configuration
        let viewport: CGSize
        @Environment(\.isFocused) private var isFocused
        private let tint = Color(red: 0.95, green: 0.28, blue: 0.28)

        var body: some View {
            let h = XuvaScale.buttonHeight(viewport)
            configuration.label
                .font(.system(size: XuvaScale.buttonFontSize(viewport), weight: .medium))
                .foregroundStyle(isFocused ? tint : XuvaTheme.mutedText)
                .padding(.horizontal, XuvaScale.buttonHorizontalPadding(viewport))
                .frame(height: h)
                .background(tint.opacity(configuration.isPressed ? 0.15 : isFocused ? 0.10 : 0.05), in: Capsule(style: .continuous))
                .overlay(Capsule(style: .continuous).stroke(
                    isFocused ? tint.opacity(0.80) : tint.opacity(0.20),
                    lineWidth: isFocused ? 2.5 : 1))
                .scaleEffect(isFocused ? 1.035 : 1)
                .shadow(color: tint.opacity(isFocused ? 0.35 : 0), radius: isFocused ? 22 : 0, x: 0, y: 10)
                .focusEffectDisabled()
                .animation(.spring(response: 0.25, dampingFraction: 0.78), value: isFocused)
        }
    }
}

public struct XuvaIconButtonStyle: ButtonStyle {
    let viewport: CGSize
    public init(viewport: CGSize = XuvaScale.screenSize) {
        self.viewport = viewport
    }

    public func makeBody(configuration: Configuration) -> some View {
        IconBody(configuration: configuration, viewport: viewport)
    }

    private struct IconBody: View {
        let configuration: Configuration
        let viewport: CGSize
        @Environment(\.isFocused) private var isFocused

        var body: some View {
            let size = XuvaScale.iconButtonSize(viewport)
            configuration.label
                .font(.system(size: XuvaScale.buttonFontSize(viewport) * 0.95))
                .foregroundStyle(XuvaTheme.text)
                .frame(width: size, height: size)
                .background(Color.white.opacity(configuration.isPressed ? 0.12 : 0.06), in: Circle())
                .overlay(Circle().stroke(
                    isFocused ? XuvaTheme.focus.opacity(0.95) : Color.white.opacity(0.10),
                    lineWidth: isFocused ? 3 : 1))
                .scaleEffect(isFocused ? 1.035 : 1)
                .shadow(color: XuvaTheme.focus.opacity(isFocused ? 0.42 : 0), radius: isFocused ? 26 : 0, x: 0, y: 14)
                .focusEffectDisabled()
                .animation(.spring(response: 0.25, dampingFraction: 0.78), value: isFocused)
        }
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

/// A completely transparent button style. tvOS only applies its card-lift
/// focus halo to its own built-in styles (.plain, .card); a custom type gets
/// no platform focus decoration, so xuvaFocused() is the sole visual.
/// A pass-through button style that strips all system decoration.
/// `.focusEffectDisabled()` suppresses the tvOS default blue halo so that
/// each call site's `.xuvaFocused(radius:)` modifier is the sole focus
/// indicator — consistent with every other Xuva button style.
struct XuvaNakedButtonStyle: ButtonStyle {
    func makeBody(configuration: Configuration) -> some View {
        configuration.label
            .focusEffectDisabled()
    }
}

/// Focus-aware "card" button style — replaces the broken
/// `.buttonStyle(.plain).xuvaFocused(radius:)` pattern that silently produced
/// an invisible focus ring (XuvaFocusModifier reads `\.isFocused` from outside
/// the button, where it's always false on tvOS).
///
/// Renders the same scale + glow + ring that `xuvaFocused()` was meant to
/// render, but reads `@Environment(\.isFocused)` from inside `makeBody` where
/// it reflects the button's own focus state.
public struct XuvaCardButtonStyle: ButtonStyle {
    let radius: CGFloat

    public init(radius: CGFloat = 18) {
        self.radius = radius
    }

    public func makeBody(configuration: Configuration) -> some View {
        CardBody(configuration: configuration, radius: radius)
    }

    private struct CardBody: View {
        let configuration: Configuration
        let radius: CGFloat
        @Environment(\.isFocused) private var isFocused

        var body: some View {
            configuration.label
                .scaleEffect(configuration.isPressed ? 0.97 : (isFocused ? 1.03 : 1))
                // Symmetric soft halo. Earlier version used a y:16 drop shadow
                // with radius 30 — that biased the glow downward (looked
                // "misaligned" against the card edge) and the radius was huge
                // against slim cards like the discovery rows. Centered, smaller.
                .shadow(
                    color: XuvaTheme.focus.opacity(isFocused ? 0.38 : 0),
                    radius: isFocused ? 14 : 0, x: 0, y: 0
                )
                // Inset ring so the stroke sits flush on the card edge rather
                // than half-on / half-off (default .stroke straddles the path).
                // Lower opacity + thinner line matches Apple TV's native halo
                // weight — visible without screaming "FOCUSED!" at the user.
                .overlay(
                    RoundedRectangle(cornerRadius: radius, style: .continuous)
                        .inset(by: 1)
                        .stroke(isFocused ? XuvaTheme.focus.opacity(0.70) : Color.clear, lineWidth: 2)
                )
                .focusEffectDisabled()
                .animation(.spring(response: 0.25, dampingFraction: 0.78), value: isFocused)
        }
    }
}

/// Focus-aware button style for poster tile cards. Applies scale, glow, and
/// a ring *inside* makeBody so @Environment(\.isFocused) correctly reads the
/// button's own focus state rather than the parent container's.
///
/// cardWidth/cardHeight define the image card rectangle so the ring is
/// positioned over just the artwork, not the subtitle text below it.
struct PosterTileButtonStyle: ButtonStyle {
    let cardWidth: CGFloat
    let cardHeight: CGFloat

    func makeBody(configuration: Configuration) -> some View {
        PosterTileBody(configuration: configuration, cardWidth: cardWidth, cardHeight: cardHeight)
    }

    private struct PosterTileBody: View {
        let configuration: Configuration
        let cardWidth: CGFloat
        let cardHeight: CGFloat
        @Environment(\.isFocused) private var isFocused

        var body: some View {
            configuration.label
                .scaleEffect(configuration.isPressed ? 0.97 : (isFocused ? 1.05 : 1))
                // Brand-purple symmetric glow. Replaces the old y:22 drop
                // shadow + harsh white ring combo that clashed with the rest
                // of the UI (everything else uses XuvaTheme.focus). Same
                // shape as XuvaCardButtonStyle, just sized for poster cards.
                .shadow(
                    color: XuvaTheme.focus.opacity(isFocused ? 0.45 : 0),
                    radius: isFocused ? 22 : 0, x: 0, y: 0
                )
                // Inset 2pt brand-purple ring on the artwork rect, not on the
                // outer label (which includes the title/subtitle below the
                // card). Matches XuvaCardButtonStyle weight + opacity.
                .overlay(alignment: .top) {
                    RoundedRectangle(cornerRadius: 12, style: .continuous)
                        .inset(by: 1)
                        .stroke(XuvaTheme.focus.opacity(isFocused ? 0.70 : 0), lineWidth: 2)
                        .frame(width: cardWidth, height: cardHeight)
                }
                .animation(.spring(response: 0.22, dampingFraction: 0.76), value: isFocused)
        }
    }
}
