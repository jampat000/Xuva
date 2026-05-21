// swift-tools-version: 5.9
import PackageDescription

let package = Package(
    name: "XuvaClientCore",
    platforms: [
        .macOS(.v13),
        .iOS(.v17),
        .tvOS(.v17)
    ],
    products: [
        .library(name: "XuvaClientCore", targets: ["XuvaClientCore"])
    ],
    targets: [
        .target(name: "XuvaClientCore")
    ]
)
