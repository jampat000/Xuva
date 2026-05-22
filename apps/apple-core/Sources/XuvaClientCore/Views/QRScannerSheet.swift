import SwiftUI
#if !os(tvOS)
import AVFoundation

public struct QRScannerSheet: View {
    let onResult: (String) -> Void
    @Environment(\.dismiss) private var dismiss
    @State private var permissionDenied = false

    public init(onResult: @escaping (String) -> Void) {
        self.onResult = onResult
    }

    public var body: some View {
        NavigationStack {
            Group {
                if permissionDenied {
                    VStack(spacing: 16) {
                        Image(systemName: "camera.slash")
                            .font(.system(size: 48))
                            .foregroundStyle(XuvaTheme.muted)
                        Text("Camera access is required to scan a QR code.")
                            .font(.system(size: 16))
                            .foregroundStyle(XuvaTheme.muted)
                            .multilineTextAlignment(.center)
                        Button("Open Settings") {
                            if let url = URL(string: UIApplication.openSettingsURLString) {
                                UIApplication.shared.open(url)
                            }
                        }
                        .buttonStyle(.borderedProminent)
                    }
                    .padding(32)
                    .frame(maxWidth: .infinity, maxHeight: .infinity)
                    .background(XuvaTheme.background)
                } else {
                    QRCaptureView(onResult: { value in
                        onResult(value)
                    }, onPermissionDenied: {
                        permissionDenied = true
                    })
                    .ignoresSafeArea()
                    .overlay(alignment: .center) {
                        RoundedRectangle(cornerRadius: 20, style: .continuous)
                            .strokeBorder(Color.white.opacity(0.6), lineWidth: 3)
                            .frame(width: 220, height: 220)
                    }
                }
            }
            .navigationTitle("Scan QR Code")
            .navigationBarTitleDisplayMode(.inline)
            .toolbar {
                ToolbarItem(placement: .cancellationAction) {
                    Button("Cancel") { dismiss() }
                }
            }
        }
        .preferredColorScheme(.dark)
    }
}

private struct QRCaptureView: UIViewRepresentable {
    let onResult: (String) -> Void
    let onPermissionDenied: () -> Void

    func makeCoordinator() -> Coordinator {
        Coordinator(onResult: onResult, onPermissionDenied: onPermissionDenied)
    }

    func makeUIView(context: Context) -> UIView {
        let view = UIView(frame: .zero)
        view.backgroundColor = .black

        let status = AVCaptureDevice.authorizationStatus(for: .video)
        if status == .denied || status == .restricted {
            onPermissionDenied()
            return view
        }

        AVCaptureDevice.requestAccess(for: .video) { granted in
            DispatchQueue.main.async {
                if granted {
                    context.coordinator.startSession(in: view)
                } else {
                    onPermissionDenied()
                }
            }
        }
        if status == .authorized {
            context.coordinator.startSession(in: view)
        }
        return view
    }

    func updateUIView(_ uiView: UIView, context: Context) {
        context.coordinator.previewLayer?.frame = uiView.bounds
    }

    final class Coordinator: NSObject, AVCaptureMetadataOutputObjectsDelegate {
        private let onResult: (String) -> Void
        private let onPermissionDenied: () -> Void
        private let session = AVCaptureSession()
        private var fired = false
        var previewLayer: AVCaptureVideoPreviewLayer?

        init(onResult: @escaping (String) -> Void, onPermissionDenied: @escaping () -> Void) {
            self.onResult = onResult
            self.onPermissionDenied = onPermissionDenied
        }

        func startSession(in view: UIView) {
            guard let device = AVCaptureDevice.default(for: .video),
                  let input = try? AVCaptureDeviceInput(device: device),
                  session.canAddInput(input) else { return }
            session.addInput(input)

            let output = AVCaptureMetadataOutput()
            guard session.canAddOutput(output) else { return }
            session.addOutput(output)
            output.setMetadataObjectsDelegate(self, queue: .main)
            output.metadataObjectTypes = [.qr]

            let preview = AVCaptureVideoPreviewLayer(session: session)
            preview.videoGravity = .resizeAspectFill
            preview.frame = view.bounds
            view.layer.addSublayer(preview)
            previewLayer = preview

            DispatchQueue.global(qos: .userInitiated).async { [weak self] in
                self?.session.startRunning()
            }
        }

        func metadataOutput(_ output: AVCaptureMetadataOutput, didOutput metadataObjects: [AVMetadataObject], from connection: AVCaptureConnection) {
            guard !fired,
                  let obj = metadataObjects.first as? AVMetadataMachineReadableCodeObject,
                  let value = obj.stringValue else { return }
            fired = true
            session.stopRunning()
            onResult(value)
        }
    }
}
#endif
