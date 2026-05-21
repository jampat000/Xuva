# Xuva iOS

Universal iPhone and iPad media player for Xuva.

The iOS app uses shared code from `../apple-core` and mirrors the tvOS client without admin or settings screens.

## Build

```sh
xcodebuild -project apps/ios/XuvaIOS.xcodeproj -scheme "Xuva iOS" -destination 'generic/platform=iOS Simulator' build
```

## Run

Open `apps/ios/XuvaIOS.xcodeproj` in Xcode, select the `Xuva iOS` scheme, choose an iPhone or iPad simulator/device, and run.
