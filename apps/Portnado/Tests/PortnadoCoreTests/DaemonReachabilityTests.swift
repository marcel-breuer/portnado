import Foundation
import Testing
@testable import PortnadoCore

@Test
func defaultSocketPathUsesApplicationSupportRunDirectory() {
    let home = URL(fileURLWithPath: "/Users/example")
    let path = DaemonReachability.defaultSocketPath(homeDirectory: home)

    #expect(path == "/Users/example/Library/Application Support/Portnado/run/portnado.sock")
}

@Test
func missingSocketIsNotReachable() {
    let reachability = DaemonReachability(socketPath: "/tmp/portnado-missing-\(UUID().uuidString).sock")

    #expect(reachability.isReachable() == false)
}
