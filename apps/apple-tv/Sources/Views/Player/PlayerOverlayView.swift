import SwiftUI

struct PlayerOverlayView: View {
    let title: String
    let route: String
    let progress: Double

    var body: some View {
        VStack {
            Spacer()
            VStack(alignment: .leading, spacing: 20) {
                HStack {
                    VStack(alignment: .leading, spacing: 8) {
                        Text(title)
                            .font(.system(size: 36, weight: .bold))
                            .foregroundStyle(XuvaTheme.text)
                        RouteBadge(text: route)
                    }
                    Spacer()
                    Button("Audio") {}
                    Button("Subtitles") {}
                    Button("Quality") {}
                    Button("Inspector") {}
                }
                .font(.system(size: 24, weight: .semibold))

                ProgressView(value: progress)
                    .tint(XuvaTheme.amber)
                    .frame(height: 8)

                HStack(spacing: 18) {
                    Button("Skip Back") {}
                    Button("Play / Pause") {}
                    Button("Skip Forward") {}
                    Spacer()
                    Text(route)
                        .font(.system(size: 18, weight: .medium))
                        .foregroundStyle(XuvaTheme.quiet)
                }
            }
            .padding(42)
            .background(.black.opacity(0.76), in: RoundedRectangle(cornerRadius: XuvaTheme.panelRadius))
            .padding(.horizontal, 70)
            .padding(.bottom, 44)
        }
    }
}
