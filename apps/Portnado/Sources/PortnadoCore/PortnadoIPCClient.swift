import Darwin
import Foundation

public struct DaemonStatus: Decodable, Equatable, Sendable {
    public let daemonState: String
    public let protocolVersion: Int
    public let version: String
    public let socketPath: String
    public let startedAt: String
}

public struct ServiceSummary: Decodable, Equatable, Identifiable, Sendable {
    public let routeId: String
    public let projectName: String
    public let serviceName: String
    public let `protocol`: String
    public let routeHost: String
    public let frontendPort: Int?
    public let backendHost: String?
    public let backendPort: Int?
    public let state: String
    public let confidence: String
    public let source: String

    public var id: String { routeId }
}

public struct ConfirmedRoute: Decodable, Equatable, Identifiable, Sendable {
    public let id: String
    public let serviceId: String
    public let serviceName: String?
    public let projectName: String?
    public let `protocol`: String
    public let routeHost: String
    public let frontendPort: Int?
    public let backendHost: String
    public let backendPort: Int
    public let state: String
    public let createdAt: String
    public let updatedAt: String
}

public struct PortnadoIPCClient: Sendable {
    public enum ClientError: Error, LocalizedError, Equatable {
        case socketPathTooLong
        case openSocketFailed(Int32)
        case connectFailed(Int32)
        case writeFailed(Int32)
        case readFailed(Int32)
        case emptyResponse
        case daemonError(String, String)
        case missingResult

        public var errorDescription: String? {
            switch self {
            case .socketPathTooLong:
                return "Control socket path is too long."
            case let .openSocketFailed(code):
                return "Could not open control socket: errno \(code)."
            case let .connectFailed(code):
                return "Could not connect to daemon: errno \(code)."
            case let .writeFailed(code):
                return "Could not write daemon request: errno \(code)."
            case let .readFailed(code):
                return "Could not read daemon response: errno \(code)."
            case .emptyResponse:
                return "Daemon returned an empty response."
            case let .daemonError(code, message):
                return "\(code): \(message)"
            case .missingResult:
                return "Daemon response did not include a result."
            }
        }
    }

    public let socketPath: String
    public let requestID: String

    public init(socketPath: String = DaemonReachability.defaultSocketPath(), requestID: String = "menubar") {
        self.socketPath = socketPath
        self.requestID = requestID
    }

    public func status() throws -> DaemonStatus {
        try call("daemon.status", as: DaemonStatus.self)
    }

    public func suggestions() throws -> [ServiceSummary] {
        try call("routes.list", as: [ServiceSummary].self)
    }

    public func routes() throws -> [ConfirmedRoute] {
        try call("route.list", as: [ConfirmedRoute].self)
    }

    public func approveRoute(id: String) throws -> ConfirmedRoute {
        try call("route.approve", params: ["id": id], as: ConfirmedRoute.self)
    }

    public func enableRoute(id: String) throws -> ConfirmedRoute {
        try call("route.enable", params: ["id": id], as: ConfirmedRoute.self)
    }

    public func disableRoute(id: String) throws -> ConfirmedRoute {
        try call("route.disable", params: ["id": id], as: ConfirmedRoute.self)
    }

    public static func encodeRequest(method: String, requestID: String, params: [String: String]? = nil) throws -> Data {
        var object: [String: Any] = [
            "protocolVersion": 1,
            "requestId": requestID,
            "method": method
        ]
        if let params {
            object["params"] = params
        }
        let data = try JSONSerialization.data(withJSONObject: object, options: [.sortedKeys])
        return data + Data([0x0A])
    }

    private func call<T: Decodable>(_ method: String, params: [String: String]? = nil, as type: T.Type) throws -> T {
        let request = try Self.encodeRequest(method: method, requestID: requestID, params: params)
        let responseData = try send(request)
        let response = try JSONDecoder().decode(IPCResponse<T>.self, from: responseData)
        if !response.ok {
            let error = response.error
            throw ClientError.daemonError(error?.code ?? "daemon_error", error?.message ?? "Daemon returned an error.")
        }
        guard let result = response.result else {
            throw ClientError.missingResult
        }
        return result
    }

    private func send(_ request: Data) throws -> Data {
        guard socketPath.utf8.count < MemoryLayout.size(ofValue: sockaddr_un().sun_path) else {
            throw ClientError.socketPathTooLong
        }

        let fileDescriptor = socket(AF_UNIX, SOCK_STREAM, 0)
        guard fileDescriptor >= 0 else {
            throw ClientError.openSocketFailed(errno)
        }
        defer {
            close(fileDescriptor)
        }

        var address = sockaddr_un()
        address.sun_family = sa_family_t(AF_UNIX)
        socketPath.withCString { source in
            withUnsafeMutableBytes(of: &address.sun_path) { destination in
                if let baseAddress = destination.baseAddress {
                    strncpy(baseAddress.assumingMemoryBound(to: CChar.self), source, destination.count - 1)
                }
            }
        }

        let addressLength = socklen_t(MemoryLayout<sa_family_t>.size + socketPath.utf8.count + 1)
        let connected = withUnsafePointer(to: &address) { pointer in
            pointer.withMemoryRebound(to: sockaddr.self, capacity: 1) { socketAddress in
                connect(fileDescriptor, socketAddress, addressLength)
            }
        }
        guard connected == 0 else {
            throw ClientError.connectFailed(errno)
        }

        try request.withUnsafeBytes { buffer in
            guard let baseAddress = buffer.baseAddress else {
                throw ClientError.writeFailed(EINVAL)
            }
            var written = 0
            while written < buffer.count {
                let count = Darwin.write(fileDescriptor, baseAddress.advanced(by: written), buffer.count - written)
                if count <= 0 {
                    throw ClientError.writeFailed(errno)
                }
                written += count
            }
        }

        var bytes = [UInt8]()
        bytes.reserveCapacity(4096)
        var buffer = [UInt8](repeating: 0, count: 4096)
        while bytes.count < 1_048_576 {
            let count = Darwin.read(fileDescriptor, &buffer, buffer.count)
            if count < 0 {
                throw ClientError.readFailed(errno)
            }
            if count == 0 {
                break
            }
            if let newline = buffer[..<count].firstIndex(of: 0x0A) {
                bytes.append(contentsOf: buffer[..<newline])
                break
            }
            bytes.append(contentsOf: buffer[..<count])
        }
        guard !bytes.isEmpty else {
            throw ClientError.emptyResponse
        }
        return Data(bytes)
    }
}

private struct IPCResponse<T: Decodable>: Decodable {
    let ok: Bool
    let result: T?
    let error: IPCError?
}

private struct IPCError: Decodable {
    let code: String
    let message: String
}
