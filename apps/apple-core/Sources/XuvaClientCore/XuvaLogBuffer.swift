import Foundation

/// Thread-safe ring buffer that retains the last `capacity` log lines.
/// Lines are added via `xuvaLog(_:)`. The DiagnosticLogView reads `lines`.
public final class XuvaLogBuffer: @unchecked Sendable {
    public static let shared = XuvaLogBuffer()

    private let capacity: Int
    private var buffer: [String] = []
    private let lock = NSLock()

    public init(capacity: Int = 500) {
        self.capacity = capacity
    }

    public func append(_ line: String) {
        let stamped = "[\(timestamp())] \(line)"
        lock.withLock {
            buffer.append(stamped)
            if buffer.count > capacity {
                buffer.removeFirst(buffer.count - capacity)
            }
        }
    }

    public var lines: [String] {
        lock.withLock { buffer }
    }

    public func clear() {
        lock.withLock { buffer.removeAll() }
    }

    public var text: String {
        lines.joined(separator: "\n")
    }

    private func timestamp() -> String {
        let f = DateFormatter()
        f.dateFormat = "HH:mm:ss.SSS"
        return f.string(from: Date())
    }
}

/// Print and buffer a log line. Prefer this over bare `print` for [XUVA] messages.
public func xuvaLog(_ message: String) {
    let line = "[XUVA] \(message)"
    print(line)
    XuvaLogBuffer.shared.append(line)
}
