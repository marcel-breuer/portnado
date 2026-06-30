import Foundation
import Testing
@testable import PortnadoCore

@Test
func encodesRouteActionRequest() throws {
    let data = try PortnadoIPCClient.encodeRequest(
        method: "route.approve",
        requestID: "test",
        params: ["id": "suggestion_app"]
    )
    let text = String(decoding: data, as: UTF8.self)

    #expect(text.hasSuffix("\n"))
    #expect(text.contains("\"method\":\"route.approve\""))
    #expect(text.contains("\"requestId\":\"test\""))
    #expect(text.contains("\"id\":\"suggestion_app\""))
}

@Test
func displayModelsDecodeDaemonPayloads() throws {
    let data = """
    {
      "daemonState": "running",
      "protocolVersion": 1,
      "version": "0.1.0-dev",
      "socketPath": "/tmp/portnado.sock",
      "startedAt": "2026-06-30T10:00:00Z"
    }
    """.data(using: .utf8)!

    let status = try JSONDecoder().decode(DaemonStatus.self, from: data)

    #expect(status.daemonState == "running")
    #expect(status.socketPath == "/tmp/portnado.sock")
}
