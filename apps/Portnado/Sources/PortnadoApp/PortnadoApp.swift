import PortnadoCore
import AppKit
import SwiftUI

@main
struct PortnadoApp: App {
    @StateObject private var model = MenuBarModel()

    var body: some Scene {
        MenuBarExtra("Portnado", systemImage: model.systemImage) {
            VStack(alignment: .leading, spacing: 10) {
                Text("Portnado")
                    .font(.headline)
                Text(model.statusText)
                Text(model.detailText)
                    .font(.caption)
                    .textSelection(.enabled)
                Divider()

                if !model.suggestions.isEmpty {
                    Text("Suggested Routes")
                        .font(.subheadline)
                    ForEach(model.suggestions) { suggestion in
                        RouteSuggestionRow(suggestion: suggestion, model: model)
                    }
                    Divider()
                }

                if !model.routes.isEmpty {
                    Text("Confirmed Routes")
                        .font(.subheadline)
                    ForEach(model.routes) { route in
                        ConfirmedRouteRow(route: route, model: model)
                    }
                    Divider()
                }

                Toggle("Show inactive routes", isOn: $model.showInactiveRoutes)
                Toggle("Copy addresses with http://", isOn: $model.copyWithScheme)
                Divider()

                Button {
                    model.refresh()
                } label: {
                    Label("Refresh", systemImage: "arrow.clockwise")
                }
                .keyboardShortcut("r")
                Button {
                    NSApplication.shared.terminate(nil)
                } label: {
                    Label("Quit", systemImage: "power")
                }
            }
            .frame(minWidth: 330, alignment: .leading)
            .padding(12)
            .onAppear {
                model.refresh()
            }
        }
        .menuBarExtraStyle(.menu)
    }
}

private struct RouteSuggestionRow: View {
    let suggestion: ServiceSummary
    @ObservedObject var model: MenuBarModel

    var body: some View {
        VStack(alignment: .leading, spacing: 4) {
            Text("\(suggestion.projectName) / \(suggestion.serviceName)")
                .font(.caption)
            Text(model.displayAddress(host: suggestion.routeHost, port: suggestion.frontendPort))
                .font(.caption2)
                .textSelection(.enabled)
            HStack {
                Button {
                    model.approve(suggestion)
                } label: {
                    Label("Approve", systemImage: "checkmark.circle")
                }
                Button {
                    model.copyAddress(host: suggestion.routeHost, port: suggestion.frontendPort)
                } label: {
                    Label("Copy", systemImage: "doc.on.doc")
                }
            }
        }
        .accessibilityElement(children: .combine)
        .accessibilityLabel("Suggested route \(suggestion.routeHost)")
    }
}

private struct ConfirmedRouteRow: View {
    let route: ConfirmedRoute
    @ObservedObject var model: MenuBarModel

    var body: some View {
        VStack(alignment: .leading, spacing: 4) {
            Text("\(route.projectName ?? "Project") / \(route.serviceName ?? "Service")")
                .font(.caption)
            Text(model.displayAddress(host: route.routeHost, port: route.frontendPort))
                .font(.caption2)
                .textSelection(.enabled)
            HStack {
                Button {
                    model.toggle(route)
                } label: {
                    Label(route.state == "active" ? "Disable" : "Enable", systemImage: route.state == "active" ? "pause.circle" : "play.circle")
                }
                Button {
                    model.copyAddress(host: route.routeHost, port: route.frontendPort)
                } label: {
                    Label("Copy", systemImage: "doc.on.doc")
                }
            }
        }
        .accessibilityElement(children: .combine)
        .accessibilityLabel("Confirmed route \(route.routeHost)")
    }
}

@MainActor
final class MenuBarModel: ObservableObject {
    @Published private(set) var reachable = false
    @Published private(set) var suggestions: [ServiceSummary] = []
    @Published private(set) var routes: [ConfirmedRoute] = []
    @Published private(set) var message = "Checking daemon..."
    @Published var showInactiveRoutes = true
    @Published var copyWithScheme = true

    private let reachability: DaemonReachability
    private let client: PortnadoIPCClient

    init(reachability: DaemonReachability = DaemonReachability(), client: PortnadoIPCClient = PortnadoIPCClient()) {
        self.reachability = reachability
        self.client = client
        self.reachable = reachability.isReachable()
    }

    var statusText: String {
        reachable ? "Daemon running" : "Daemon unavailable"
    }

    var detailText: String {
        message.isEmpty ? reachability.socketPath : message
    }

    var systemImage: String {
        reachable ? "point.3.connected.trianglepath.dotted" : "exclamationmark.triangle"
    }

    func refresh() {
        Task {
            do {
                let status = try client.status()
                let suggestions = try client.suggestions()
                let routes = try client.routes()
                await MainActor.run {
                    self.reachable = true
                    self.suggestions = suggestions
                    self.routes = showInactiveRoutes ? routes : routes.filter { $0.state == "active" }
                    self.message = "\(status.version) on \(status.socketPath)"
                }
            } catch {
                await MainActor.run {
                    self.reachable = false
                    self.suggestions = []
                    self.routes = []
                    self.message = error.localizedDescription
                }
            }
        }
    }

    func approve(_ suggestion: ServiceSummary) {
        Task {
            do {
                _ = try client.approveRoute(id: suggestion.routeId)
                await MainActor.run {
                    self.message = "Approved \(suggestion.routeHost)"
                    self.refresh()
                }
            } catch {
                await MainActor.run { self.message = error.localizedDescription }
            }
        }
    }

    func toggle(_ route: ConfirmedRoute) {
        Task {
            do {
                if route.state == "active" {
                    _ = try client.disableRoute(id: route.id)
                } else {
                    _ = try client.enableRoute(id: route.id)
                }
                await MainActor.run {
                    self.message = "Updated \(route.routeHost)"
                    self.refresh()
                }
            } catch {
                await MainActor.run { self.message = error.localizedDescription }
            }
        }
    }

    func copyAddress(host: String, port: Int?) {
        NSPasteboard.general.clearContents()
        NSPasteboard.general.setString(displayAddress(host: host, port: port), forType: .string)
        message = "Copied \(host)"
    }

    func displayAddress(host: String, port: Int?) -> String {
        let address: String
        if let port, port != 0 {
            address = "\(host):\(port)"
        } else {
            address = host
        }
        return copyWithScheme ? "http://\(address)" : address
    }
}
