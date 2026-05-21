# Xuva tvOS

Native Apple TV media player for Xuva.

The tvOS app uses shared code from `../apple-core` and focuses on couch playback: pairing, home rows, detail screens, route forecast, and AVPlayer playback.

## Build

```sh
xcodebuild -project apps/tvos/XuvaTV.xcodeproj -scheme "Xuva TV" -destination 'generic/platform=tvOS Simulator' build
```

## Run

Open `apps/tvos/XuvaTV.xcodeproj` in Xcode, select the `Xuva TV` scheme, choose an Apple TV simulator or device, and run.
