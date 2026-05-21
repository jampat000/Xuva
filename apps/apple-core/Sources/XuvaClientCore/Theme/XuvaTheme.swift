import SwiftUI

public enum XuvaTheme {
    public static let background = Color(red: 0.032, green: 0.028, blue: 0.054)
    public static let ink = Color(red: 0.050, green: 0.047, blue: 0.082)
    public static let surface = Color(red: 0.078, green: 0.073, blue: 0.122)
    public static let elevated = Color(red: 0.114, green: 0.106, blue: 0.176)
    public static let text = Color(red: 0.980, green: 0.976, blue: 1.000)
    public static let secondaryText = Color(red: 0.702, green: 0.690, blue: 0.760)
    public static let mutedText = Color(red: 0.455, green: 0.435, blue: 0.520)
    public static let focus = Color(red: 0.650, green: 0.545, blue: 1.000)
    public static let action = Color(red: 0.486, green: 0.361, blue: 1.000)
    public static let primary = action
    public static let primaryGlow = focus
    public static let accent = focus
    public static let good = Color(red: 0.396, green: 0.847, blue: 0.529)
    public static let warn = Color(red: 0.886, green: 0.749, blue: 0.451)
    public static let danger = Color(red: 0.862, green: 0.545, blue: 0.514)
    public static let muted = secondaryText
    public static let hairline = Color(red: 0.196, green: 0.169, blue: 0.137)

    public static var backgroundWash: some View {
        ZStack {
            background
            RadialGradient(
                colors: [action.opacity(0.34), Color.clear],
                center: UnitPoint(x: 0.28, y: 0.0),
                startRadius: 40,
                endRadius: 780
            )
            RadialGradient(
                colors: [Color(red: 0.435, green: 0.650, blue: 1.0).opacity(0.12), Color.clear],
                center: UnitPoint(x: 0.88, y: 0.18),
                startRadius: 120,
                endRadius: 900
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
