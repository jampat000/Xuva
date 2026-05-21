import SwiftUI

public struct XuvaLogo: View {
    public init() {}

    public var body: some View {
        HStack(spacing: 12) {
            ZStack {
                RoundedRectangle(cornerRadius: 9, style: .continuous)
                    .fill(XuvaTheme.action.opacity(0.18))
                RoundedRectangle(cornerRadius: 10, style: .continuous)
                    .stroke(
                        LinearGradient(
                            colors: [
                                Color(red: 0.550, green: 0.380, blue: 1.000),
                                Color(red: 0.720, green: 0.560, blue: 1.000),
                                Color(red: 0.560, green: 0.700, blue: 1.000)
                            ],
                            startPoint: .topLeading,
                            endPoint: .bottomTrailing
                        ),
                        lineWidth: 1.4
                    )
                    .opacity(0.70)
                XuvaMark()
                    .stroke(
                        LinearGradient(
                            colors: [
                                Color(red: 0.550, green: 0.380, blue: 1.000),
                                Color(red: 0.980, green: 0.976, blue: 1.000)
                            ],
                            startPoint: .leading,
                            endPoint: .trailing
                        ),
                        style: StrokeStyle(lineWidth: 3.2, lineCap: .round, lineJoin: .round)
                    )
                    .padding(8)
            }
            .frame(width: 36, height: 36)
            Text("Xuva")
                .font(.title3.weight(.semibold))
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
                    LinearGradient(colors: [XuvaTheme.surface, XuvaTheme.background], startPoint: .topLeading, endPoint: .bottomTrailing)
                    Image(systemName: "film")
                        .font(.system(size: 42, weight: .semibold))
                        .foregroundStyle(.white.opacity(0.20))
                }
            }
        }
        .aspectRatio(aspectRatio, contentMode: .fill)
    }
}

struct PosterTile: View {
    let item: HomeItem
    let ranked: Bool
    let action: () -> Void

    var body: some View {
        Button(action: action) {
            ZStack(alignment: .bottomLeading) {
                RemoteImage(urlString: item.posterUrl ?? item.imageUrl, aspectRatio: 2 / 3)
                    .frame(width: posterWidth, height: posterHeight)
                    .clipShape(RoundedRectangle(cornerRadius: 8, style: .continuous))
                LinearGradient(colors: [.clear, .black.opacity(0.82)], startPoint: .center, endPoint: .bottom)
                    .clipShape(RoundedRectangle(cornerRadius: 8, style: .continuous))
                VStack(alignment: .leading, spacing: 5) {
                    if let route = item.routeLabel {
                        Text(route)
                            .font(.caption2.bold())
                            .padding(.horizontal, 8)
                            .padding(.vertical, 4)
                            .background(XuvaTheme.focus.opacity(0.16), in: Capsule())
                            .foregroundStyle(XuvaTheme.focus)
                    }
                    Text(item.title ?? "Untitled")
                        .font(.headline)
                        .lineLimit(2)
                    Text([item.year.map(String.init), item.subtitle].compactMap { $0 }.joined(separator: " · "))
                        .font(.caption)
                        .foregroundStyle(.white.opacity(0.68))
                        .lineLimit(1)
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
                    Text("#")
                        .font(.caption.bold())
                        .padding(.horizontal, 9)
                        .padding(.vertical, 5)
                        .background(.black.opacity(0.58), in: Capsule())
                        .padding(12)
                        .frame(maxWidth: .infinity, maxHeight: .infinity, alignment: .topLeading)
                }
            }
        }
        .buttonStyle(.plain)
        .xuvaFocused(radius: 8)
    }

    private var posterWidth: CGFloat {
        #if os(tvOS)
        return 190
        #else
        return 132
        #endif
    }

    private var posterHeight: CGFloat {
        #if os(tvOS)
        return 285
        #else
        return 198
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
        .background(color.opacity(0.16), in: Capsule())
        .overlay(Capsule().stroke(color.opacity(0.30)))
    }

    private var color: Color {
        switch decision?.badgeLabel {
        case "Direct Play": return XuvaTheme.good
        case "Remux": return XuvaTheme.accent
        case "Adaptive": return XuvaTheme.primaryGlow
        case "Transcoding": return XuvaTheme.danger
        case "Audio Tx": return XuvaTheme.warn
        default: return .white.opacity(0.46)
        }
    }
}

public struct XuvaPrimaryButtonStyle: ButtonStyle {
    public init() {}

    public func makeBody(configuration: Configuration) -> some View {
        configuration.label
            .font(.headline)
            .foregroundStyle(.white)
            .padding(.horizontal, 22)
            .frame(height: 54)
            .background(
                LinearGradient(
                    colors: [
                        Color(red: 0.486, green: 0.361, blue: 1.000),
                        Color(red: 0.612, green: 0.486, blue: 1.000)
                    ],
                    startPoint: .leading,
                    endPoint: .trailing
                ),
                in: RoundedRectangle(cornerRadius: 12, style: .continuous)
            )
            .scaleEffect(configuration.isPressed ? 0.97 : 1)
            .xuvaFocused(radius: 12)
    }
}

public struct XuvaSecondaryButtonStyle: ButtonStyle {
    public init() {}

    public func makeBody(configuration: Configuration) -> some View {
        configuration.label
            .font(.headline)
            .foregroundStyle(XuvaTheme.text)
            .padding(.horizontal, 22)
            .frame(height: 54)
            .background(XuvaTheme.elevated.opacity(configuration.isPressed ? 1 : 0.78), in: RoundedRectangle(cornerRadius: 12, style: .continuous))
            .overlay(RoundedRectangle(cornerRadius: 12, style: .continuous).stroke(XuvaTheme.hairline))
            .xuvaFocused(radius: 12)
    }
}

public struct XuvaIconButtonStyle: ButtonStyle {
    public init() {}

    public func makeBody(configuration: Configuration) -> some View {
        configuration.label
            .font(.headline)
            .foregroundStyle(XuvaTheme.text)
            .frame(width: 48, height: 48)
            .background(XuvaTheme.elevated.opacity(configuration.isPressed ? 1 : 0.78), in: RoundedRectangle(cornerRadius: 12, style: .continuous))
            .overlay(RoundedRectangle(cornerRadius: 12, style: .continuous).stroke(XuvaTheme.hairline))
            .xuvaFocused(radius: 12)
    }
}
