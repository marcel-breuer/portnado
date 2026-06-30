import Darwin
import Foundation

public struct DaemonReachability: Sendable {
    public let socketPath: String

    public init(socketPath: String = DaemonReachability.defaultSocketPath()) {
        self.socketPath = socketPath
    }

    public static func defaultSocketPath(
        homeDirectory: URL = FileManager.default.homeDirectoryForCurrentUser
    ) -> String {
        homeDirectory
            .appendingPathComponent("Library")
            .appendingPathComponent("Application Support")
            .appendingPathComponent("Portnado")
            .appendingPathComponent("run")
            .appendingPathComponent("portnado.sock")
            .path
    }

    public func isReachable() -> Bool {
        guard socketPath.utf8.count < MemoryLayout.size(ofValue: sockaddr_un().sun_path) else {
            return false
        }

        let fileDescriptor = socket(AF_UNIX, SOCK_STREAM, 0)
        guard fileDescriptor >= 0 else {
            return false
        }
        defer {
            close(fileDescriptor)
        }

        var address = sockaddr_un()
        address.sun_family = sa_family_t(AF_UNIX)

        let copied = socketPath.withCString { source in
            withUnsafeMutableBytes(of: &address.sun_path) { destination in
                guard let baseAddress = destination.baseAddress else {
                    return false
                }
                strncpy(baseAddress.assumingMemoryBound(to: CChar.self), source, destination.count - 1)
                return true
            }
        }
        guard copied else {
            return false
        }

        let addressLength = socklen_t(MemoryLayout<sa_family_t>.size + socketPath.utf8.count + 1)
        return withUnsafePointer(to: &address) { pointer in
            pointer.withMemoryRebound(to: sockaddr.self, capacity: 1) { socketAddress in
                connect(fileDescriptor, socketAddress, addressLength) == 0
            }
        }
    }
}
