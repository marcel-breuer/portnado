import Foundation
import Testing
@testable import PortnadoCore

@Test
func bundledDaemonPathResolvesPackagedDaemon() throws {
    let temporaryRoot = FileManager.default.temporaryDirectory
        .appendingPathComponent("portnado-launcher-\(UUID().uuidString)")
    let daemonURL = temporaryRoot
        .appendingPathComponent("Portnado.app")
        .appendingPathComponent("Contents")
        .appendingPathComponent("Resources")
        .appendingPathComponent("bin")
        .appendingPathComponent("portnado-daemon")
    try FileManager.default.createDirectory(
        at: daemonURL.deletingLastPathComponent(),
        withIntermediateDirectories: true
    )
    FileManager.default.createFile(atPath: daemonURL.path, contents: Data())
    defer {
        try? FileManager.default.removeItem(at: temporaryRoot)
    }

    let path = PortnadoDaemonLauncher.defaultBundledDaemonPath(
        bundleURL: temporaryRoot.appendingPathComponent("Portnado.app")
    )

    #expect(path == daemonURL.path)
}

@Test
func ensureRunningSkipsLaunchWhenSocketIsReachable() async throws {
    let launcher = PortnadoDaemonLauncher(
        socketPath: "/tmp/portnado.sock",
        daemonPath: nil,
        isReachable: { true }
    )

    try await launcher.ensureRunning()
}

@Test
func ensureRunningReportsMissingBundledDaemon() async {
    let launcher = PortnadoDaemonLauncher(
        socketPath: "/tmp/portnado.sock",
        daemonPath: nil,
        isReachable: { false }
    )

    do {
        try await launcher.ensureRunning()
        Issue.record("Expected bundled daemon lookup to fail")
    } catch let error as PortnadoDaemonLauncher.LauncherError {
        #expect(error == .bundledDaemonMissing)
    } catch {
        Issue.record("Unexpected error: \(error)")
    }
}
