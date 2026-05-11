import SwiftUI

struct PosterCard: View {
    @EnvironmentObject private var appState: LorivoAppState
    @FocusState private var focused: Bool
    let poster: MediaPoster

    var body: some View {
        Button {
            appState.focusedPoster = poster
        } label: {
            VStack(alignment: .leading, spacing: 12) {
                ZStack(alignment: .bottomLeading) {
                    RoundedRectangle(cornerRadius: LorivoTheme.cardRadius)
                        .fill(
                            LinearGradient(
                                colors: [LorivoTheme.graphite, LorivoTheme.carbon],
                                startPoint: .top,
                                endPoint: .bottom
                            )
                        )
                    Text(poster.title.prefix(1))
                        .font(.system(size: 92, weight: .black))
                        .foregroundStyle(LorivoTheme.text.opacity(0.12))
                    LinearGradient(colors: [.clear, .black.opacity(0.76)], startPoint: .center, endPoint: .bottom)
                        .clipShape(RoundedRectangle(cornerRadius: LorivoTheme.cardRadius))
                    Text(poster.route)
                        .font(.system(size: 16, weight: .bold))
                        .foregroundStyle(LorivoTheme.text)
                        .padding(.horizontal, 10)
                        .padding(.vertical, 6)
                        .background(LorivoTheme.focus.opacity(0.18), in: Capsule())
                        .padding(12)
                }
                .frame(width: 214, height: 322)
                .overlay(
                    RoundedRectangle(cornerRadius: LorivoTheme.cardRadius)
                        .stroke(focused ? LorivoTheme.focus : .clear, lineWidth: 4)
                )
                .shadow(color: focused ? LorivoTheme.focus.opacity(0.22) : .black.opacity(0.28), radius: focused ? 28 : 14, x: 0, y: focused ? 18 : 10)

                Text(poster.title)
                    .font(.system(size: 22, weight: .semibold))
                    .foregroundStyle(LorivoTheme.text)
                    .lineLimit(1)
                Text(poster.subtitle)
                    .font(.system(size: 17, weight: .medium))
                    .foregroundStyle(LorivoTheme.quiet)
                    .lineLimit(1)
            }
            .frame(width: 214, alignment: .leading)
            .scaleEffect(focused ? 1.045 : 1)
            .offset(y: focused ? -8 : 0)
            .animation(.easeOut(duration: 0.16), value: focused)
        }
        .buttonStyle(LorivoFocusStyle())
        .focused($focused)
        .onChange(of: focused) { value in
            if value {
                appState.focusedPoster = poster
            }
        }
    }
}
