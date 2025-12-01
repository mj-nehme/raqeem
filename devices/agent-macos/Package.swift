// swift-tools-version:5.9
import PackageDescription

let package = Package(
    name: "DeviceAgent",
    platforms: [
        .macOS(.v13)
    ],
    products: [
        .executable(name: "device-agent", targets: ["DeviceAgent"]),
    ],
    dependencies: [],
    targets: [
        .executableTarget(
            name: "DeviceAgent",
            path: "Sources/DeviceAgent"
        )
    ]
)
