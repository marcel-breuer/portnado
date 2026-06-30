// swift-tools-version: 6.0

import PackageDescription

let package = Package(
    name: "Portnado",
    platforms: [
        .macOS(.v14)
    ],
    products: [
        .executable(name: "Portnado", targets: ["PortnadoApp"])
    ],
    targets: [
        .target(name: "PortnadoCore"),
        .executableTarget(
            name: "PortnadoApp",
            dependencies: ["PortnadoCore"]
        ),
        .testTarget(
            name: "PortnadoCoreTests",
            dependencies: ["PortnadoCore"]
        )
    ]
)
