import Foundation

public struct PortnadoDaemonLauncher: Sendable {
    public enum LauncherError: Error, LocalizedError, Equatable {
        case bundledDaemonMissing
        case bundledDaemonNotExecutable(String)
        case daemonExited(Int32)
        case daemonDidNotBecomeReachable(String)

        public var errorDescription: String? {
            switch self {
            case .bundledDaemonMissing:
                return "Could not find the bundled Portnado daemon."
            case let .bundledDaemonNotExecutable(path):
                return "Bundled Portnado daemon is not executable at \(path)."
            case let .daemonExited(status):
                return "Portnado daemon exited before it became reachable with status \(status)."
            case let .daemonDidNotBecomeReachable(path):
                return "Portnado daemon started but did not become reachable at \(path)."
            }
        }
    }

    private let daemonPath: String?
    private let socketPath: String
    private let maxAttempts: Int
    private let retryDelayNanoseconds: UInt64
    private let isReachable: @Sendable () -> Bool

    public init(
        reachability: DaemonReachability = DaemonReachability(),
        daemonPath: String? = PortnadoDaemonLauncher.defaultBundledDaemonPath(),
        maxAttempts: Int = 25,
        retryDelayNanoseconds: UInt64 = 100_000_000
    ) {
        self.init(
            socketPath: reachability.socketPath,
            daemonPath: daemonPath,
            maxAttempts: maxAttempts,
            retryDelayNanoseconds: retryDelayNanoseconds,
            isReachable: { reachability.isReachable() }
        )
    }

    public init(
        socketPath: String,
        daemonPath: String?,
        maxAttempts: Int = 25,
        retryDelayNanoseconds: UInt64 = 100_000_000,
        isReachable: @escaping @Sendable () -> Bool
    ) {
        self.socketPath = socketPath
        self.daemonPath = daemonPath
        self.maxAttempts = max(1, maxAttempts)
        self.retryDelayNanoseconds = retryDelayNanoseconds
        self.isReachable = isReachable
    }

    public static func defaultBundledDaemonPath(
        bundleURL: URL = Bundle.main.bundleURL,
        fileManager: FileManager = .default
    ) -> String? {
        let candidate = bundleURL
            .appendingPathComponent("Contents")
            .appendingPathComponent("Resources")
            .appendingPathComponent("bin")
            .appendingPathComponent("portnado-daemon")
            .path
        return fileManager.fileExists(atPath: candidate) ? candidate : nil
    }

    public func ensureRunning() async throws {
        if isReachable() {
            return
        }
        guard let daemonPath else {
            throw LauncherError.bundledDaemonMissing
        }
        guard FileManager.default.isExecutableFile(atPath: daemonPath) else {
            throw LauncherError.bundledDaemonNotExecutable(daemonPath)
        }

        let process = Process()
        process.executableURL = URL(fileURLWithPath: daemonPath)
        process.standardOutput = FileHandle(forWritingAtPath: "/dev/null")
        process.standardError = FileHandle(forWritingAtPath: "/dev/null")
        try process.run()

        for _ in 0..<maxAttempts {
            if isReachable() {
                return
            }
            if !process.isRunning {
                throw LauncherError.daemonExited(process.terminationStatus)
            }
            try await Task.sleep(nanoseconds: retryDelayNanoseconds)
        }

        throw LauncherError.daemonDidNotBecomeReachable(socketPath)
    }
}
