import SwiftUI

enum LorivoTheme {
    static let cinema = Color(red: 0.027, green: 0.027, blue: 0.027)
    static let carbon = Color(red: 0.051, green: 0.051, blue: 0.047)
    static let graphite = Color(red: 0.086, green: 0.086, blue: 0.078)
    static let text = Color(red: 0.957, green: 0.937, blue: 0.894)
    static let soft = Color(red: 0.824, green: 0.792, blue: 0.741)
    static let quiet = Color(red: 0.667, green: 0.627, blue: 0.573)
    static let amber = Color(red: 0.812, green: 0.682, blue: 0.451)
    static let focus = Color(red: 0.604, green: 0.800, blue: 0.776)
    static let red = Color(red: 0.733, green: 0.400, blue: 0.365)
    static let green = Color(red: 0.541, green: 0.749, blue: 0.561)

    static let cardRadius: CGFloat = 8
    static let panelRadius: CGFloat = 10
    static let horizontalMargin: CGFloat = 84
    static let rowSpacing: CGFloat = 42
}

struct LorivoFocusStyle: ButtonStyle {
    func makeBody(configuration: Configuration) -> some View {
        configuration.label
            .scaleEffect(configuration.isPressed ? 0.98 : 1)
            .animation(.easeOut(duration: 0.12), value: configuration.isPressed)
    }
}

struct RouteBadge: View {
    let text: String
    var tone: Color = LorivoTheme.focus

    var body: some View {
        Text(text)
            .font(.system(size: 22, weight: .semibold))
            .foregroundStyle(LorivoTheme.text)
            .padding(.horizontal, 16)
            .padding(.vertical, 8)
            .background(tone.opacity(0.16), in: Capsule())
            .overlay(Capsule().stroke(tone.opacity(0.35), lineWidth: 1))
    }
}
