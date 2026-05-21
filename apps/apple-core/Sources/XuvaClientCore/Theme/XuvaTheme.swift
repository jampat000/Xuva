import SwiftUI

public enum XuvaTheme {
    public static let background = Color(red: 0.043, green: 0.067, blue: 0.125)
    public static let ink = Color(red: 0.051, green: 0.078, blue: 0.141)
    public static let surface = Color(red: 0.067, green: 0.094, blue: 0.153)
    public static let elevated = Color(red: 0.122, green: 0.161, blue: 0.216)
    public static let soft = Color(red: 0.149, green: 0.196, blue: 0.267)
    public static let text = Color(red: 0.973, green: 0.980, blue: 0.988)
    public static let secondaryText = Color(red: 0.796, green: 0.835, blue: 0.882)
    public static let mutedText = Color(red: 0.580, green: 0.639, blue: 0.722)
    public static let action = Color(red: 0.486, green: 0.361, blue: 1.000)
    public static let actionStrong = Color(red: 0.416, green: 0.290, blue: 0.941)
    public static let focus = Color(red: 0.608, green: 0.486, blue: 1.000)
    public static let primary = action
    public static let primaryGlow = focus
    public static let accent = action
    public static let good = Color(red: 0.482, green: 0.784, blue: 0.573)
    public static let warn = Color(red: 0.886, green: 0.749, blue: 0.451)
    public static let danger = Color(red: 0.863, green: 0.545, blue: 0.514)
    public static let muted = secondaryText
    public static let hairline = Color.white.opacity(0.10)
    public static let hairlineStrong = action.opacity(0.36)

    public static var backgroundWash: some View {
        ZStack {
            background
            RadialGradient(
                colors: [action.opacity(0.14), Color.clear],
                center: UnitPoint(x: 0.18, y: 0.0),
                startRadius: 60,
                endRadius: 900
            )
            RadialGradient(
                colors: [Color(red: 0.435, green: 0.560, blue: 1.0).opacity(0.06), Color.clear],
                center: UnitPoint(x: 0.92, y: 0.18),
                startRadius: 140,
                endRadius: 980
            )
        }
        .ignoresSafeArea()
    }
}

public struct XuvaFocusModifier: ViewModifier {
    @Environment(\.isFocused) private var isFocused
    let radius: CGFloat

    public func body(content: Content) -> some View {
        content
            .scaleEffect(isFocused ? 1.035 : 1)
            .shadow(color: XuvaTheme.focus.opacity(isFocused ? 0.42 : 0), radius: isFocused ? 26 : 0, x: 0, y: 14)
            .overlay(
                RoundedRectangle(cornerRadius: radius, style: .continuous)
                    .stroke(isFocused ? XuvaTheme.focus.opacity(0.95) : Color.clear, lineWidth: 3)
            )
            .animation(.spring(response: 0.25, dampingFraction: 0.78), value: isFocused)
    }
}

public extension View {
    func xuvaFocused(radius: CGFloat = 18) -> some View {
        modifier(XuvaFocusModifier(radius: radius))
    }
}
