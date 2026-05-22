import SwiftUI

/// Full-screen log viewer. Accessible via long-press on the version label
/// in PairingScreen or any future entry point.
public struct DiagnosticLogView: View {
    @Environment(\.dismiss) private var dismiss
    @State private var lines: [String] = []
    @State private var copied = false

    public init() {}

    public var body: some View {
        NavigationStack {
            ScrollViewReader { proxy in
                ScrollView {
                    LazyVStack(alignment: .leading, spacing: 2) {
                        ForEach(Array(lines.enumerated()), id: \.offset) { _, line in
                            Text(line)
                                .font(.system(.caption2, design: .monospaced))
                                .foregroundStyle(.green.opacity(0.85))
                                .frame(maxWidth: .infinity, alignment: .leading)
                                .padding(.horizontal, 12)
                        }
                    }
                    .padding(.vertical, 8)
                    Color.clear.frame(height: 1).id("bottom")
                }
                .background(Color.black)
                .onAppear {
                    lines = XuvaLogBuffer.shared.lines
                    proxy.scrollTo("bottom", anchor: .bottom)
                }
            }
            .navigationTitle("Diagnostic Log (\(lines.count) lines)")
            #if os(tvOS)
            .navigationBarTitleDisplayMode(.inline)
            #else
            .navigationBarTitleDisplayMode(.inline)
            #endif
            .toolbar {
                ToolbarItem(placement: .navigationBarTrailing) {
                    Button(copied ? "Copied!" : "Copy") {
                        let text = XuvaLogBuffer.shared.text
                        #if os(tvOS)
                        // tvOS has no UIPasteboard — show an alert or do nothing.
                        #else
                        UIPasteboard.general.string = text
                        #endif
                        copied = true
                        Task {
                            try? await Task.sleep(nanoseconds: 2_000_000_000)
                            copied = false
                        }
                    }
                }
                ToolbarItem(placement: .navigationBarLeading) {
                    Button("Clear") {
                        XuvaLogBuffer.shared.clear()
                        lines = []
                    }
                }
                ToolbarItem(placement: .cancellationAction) {
                    Button("Done") { dismiss() }
                }
            }
        }
    }
}
