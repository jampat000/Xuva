import SwiftUI

// MARK: – tvOS-only focus API shims
// focusSection() and prefersDefaultFocus(_:in:) are tvOS-only. These no-op
// extensions let the code compile on iOS/macOS without guarding every call site.

#if !os(tvOS)
extension View {
    func focusSection() -> some View { self }
    func prefersDefaultFocus(_ condition: Bool = true, in namespace: Namespace.ID) -> some View { self }
    func focusScope(_ namespace: Namespace.ID) -> some View { self }
}
#endif
