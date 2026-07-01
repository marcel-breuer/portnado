import PortnadoCore
import AppKit
import SwiftUI

@main
struct PortnadoApp: App {
    @StateObject private var model = MenuBarModel()

    var body: some Scene {
        MenuBarExtra("Portnado", systemImage: model.systemImage) {
            VStack(alignment: .leading, spacing: 12) {
                BrandHeader(model: model)
                StatusPanel(model: model)
                Divider()

                if !model.suggestions.isEmpty {
                    SectionHeader(title: "Suggested Routes", systemImage: "sparkles")
                    ForEach(model.suggestions) { suggestion in
                        RouteSuggestionRow(suggestion: suggestion, model: model)
                    }
                    Divider()
                }

                if !model.routes.isEmpty {
                    SectionHeader(title: "Confirmed Routes", systemImage: "checkmark.seal")
                    ForEach(model.routes) { route in
                        ConfirmedRouteRow(route: route, model: model)
                    }
                    Divider()
                }

                Toggle("Show inactive routes", isOn: $model.showInactiveRoutes)
                    .tint(BrandPalette.primary)
                Toggle("Copy addresses with http://", isOn: $model.copyWithScheme)
                    .tint(BrandPalette.primary)
                Divider()

                HStack(spacing: 8) {
                    Button {
                        model.refresh()
                    } label: {
                        Label("Refresh", systemImage: "arrow.clockwise")
                    }
                    .keyboardShortcut("r")
                    .tint(BrandPalette.primary)

                    Button {
                        NSApplication.shared.terminate(nil)
                    } label: {
                        Label("Quit", systemImage: "power")
                    }
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

private enum BrandPalette {
    static let primary = Color(hex: 0x9333EA)
    static let primaryDeep = Color(hex: 0x7C3AED)
    static let indigo = Color(hex: 0x6366F1)
    static let purple100 = Color(hex: 0xF3E8FF)
    static let slate500 = Color(hex: 0x64748B)
    static let slate950 = Color(hex: 0x020617)
    static let success = Color(hex: 0x10B981)
    static let warning = Color(hex: 0xF59E0B)
}

private extension Color {
    init(hex: UInt32, opacity: Double = 1) {
        let red = Double((hex >> 16) & 0xff) / 255
        let green = Double((hex >> 8) & 0xff) / 255
        let blue = Double(hex & 0xff) / 255
        self.init(.sRGB, red: red, green: green, blue: blue, opacity: opacity)
    }
}

private struct BrandHeader: View {
    @ObservedObject var model: MenuBarModel

    var body: some View {
        HStack(spacing: 10) {
            ZStack {
                RoundedRectangle(cornerRadius: 8, style: .continuous)
                    .fill(
                        LinearGradient(
                            colors: [BrandPalette.primaryDeep, BrandPalette.indigo],
                            startPoint: .topLeading,
                            endPoint: .bottomTrailing
                        )
                    )
                Image(systemName: model.systemImage)
                    .font(.system(size: 15, weight: .semibold))
                    .foregroundStyle(.white)
            }
            .frame(width: 34, height: 34)

            VStack(alignment: .leading, spacing: 2) {
                Text("Portnado")
                    .font(.headline)
                    .foregroundStyle(.primary)
                Text("Local routing control")
                    .font(.caption2)
                    .foregroundStyle(.secondary)
            }
        }
    }
}

private struct StatusPanel: View {
    @ObservedObject var model: MenuBarModel
    @Environment(\.colorScheme) private var colorScheme

    var body: some View {
        VStack(alignment: .leading, spacing: 6) {
            HStack(spacing: 6) {
                Circle()
                    .fill(model.reachable ? BrandPalette.success : BrandPalette.warning)
                    .frame(width: 8, height: 8)
                Text(model.statusText)
                    .font(.subheadline.weight(.semibold))
                    .foregroundStyle(.primary)
            }

            Text(model.detailText)
                .font(.caption)
                .foregroundStyle(.secondary)
                .textSelection(.enabled)
        }
        .padding(10)
        .frame(maxWidth: .infinity, alignment: .leading)
        .background(
            RoundedRectangle(cornerRadius: 8, style: .continuous)
                .fill(statusBackground)
                .overlay(
                    RoundedRectangle(cornerRadius: 8, style: .continuous)
                        .stroke(BrandPalette.primary.opacity(0.18), lineWidth: 1)
            )
        )
    }

    private var statusBackground: Color {
        colorScheme == .dark ? BrandPalette.primaryDeep.opacity(0.18) : BrandPalette.purple100.opacity(0.65)
    }
}

private struct SectionHeader: View {
    let title: String
    let systemImage: String

    var body: some View {
        Label(title, systemImage: systemImage)
            .font(.caption.weight(.semibold))
            .foregroundStyle(BrandPalette.primaryDeep)
    }
}

private struct RouteSuggestionRow: View {
    let suggestion: ServiceSummary
    @ObservedObject var model: MenuBarModel

    var body: some View {
        VStack(alignment: .leading, spacing: 6) {
            Text("\(suggestion.projectName) / \(suggestion.serviceName)")
                .font(.caption.weight(.semibold))
                .foregroundStyle(.primary)
            Text(model.displayAddress(host: suggestion.routeHost, port: suggestion.frontendPort))
                .font(.caption2)
                .foregroundStyle(.secondary)
                .textSelection(.enabled)
            HStack {
                Button {
                    model.approve(suggestion)
                } label: {
                    Label("Approve", systemImage: "checkmark.circle")
                }
                .tint(BrandPalette.primary)
                Button {
                    model.copyAddress(host: suggestion.routeHost, port: suggestion.frontendPort)
                } label: {
                    Label("Copy", systemImage: "doc.on.doc")
                }
            }
        }
        .routeSurface(accent: BrandPalette.primary)
        .accessibilityElement(children: .combine)
        .accessibilityLabel("Suggested route \(suggestion.routeHost)")
    }
}

private struct ConfirmedRouteRow: View {
    let route: ConfirmedRoute
    @ObservedObject var model: MenuBarModel

    var body: some View {
        VStack(alignment: .leading, spacing: 6) {
            Text("\(route.projectName ?? "Project") / \(route.serviceName ?? "Service")")
                .font(.caption.weight(.semibold))
                .foregroundStyle(.primary)
            Text(model.displayAddress(host: route.routeHost, port: route.frontendPort))
                .font(.caption2)
                .foregroundStyle(.secondary)
                .textSelection(.enabled)
            HStack {
                Button {
                    model.toggle(route)
                } label: {
                    Label(route.state == "active" ? "Disable" : "Enable", systemImage: route.state == "active" ? "pause.circle" : "play.circle")
                }
                .tint(route.state == "active" ? BrandPalette.warning : BrandPalette.primary)
                Button {
                    model.copyAddress(host: route.routeHost, port: route.frontendPort)
                } label: {
                    Label("Copy", systemImage: "doc.on.doc")
                }
            }
        }
        .routeSurface(accent: route.state == "active" ? BrandPalette.success : BrandPalette.slate500)
        .accessibilityElement(children: .combine)
        .accessibilityLabel("Confirmed route \(route.routeHost)")
    }
}

private extension View {
    func routeSurface(accent: Color) -> some View {
        modifier(RouteSurfaceModifier(accent: accent))
    }
}

private struct RouteSurfaceModifier: ViewModifier {
    let accent: Color
    @Environment(\.colorScheme) private var colorScheme

    func body(content: Content) -> some View {
        content
            .padding(8)
            .frame(maxWidth: .infinity, alignment: .leading)
            .background(
                HStack(spacing: 0) {
                    RoundedRectangle(cornerRadius: 3, style: .continuous)
                        .fill(accent)
                        .frame(width: 3)
                    RoundedRectangle(cornerRadius: 8, style: .continuous)
                        .fill(surfaceColor)
                }
            )
            .overlay(
                RoundedRectangle(cornerRadius: 8, style: .continuous)
                    .stroke(BrandPalette.primary.opacity(colorScheme == .dark ? 0.22 : 0.10), lineWidth: 1)
            )
    }

    private var surfaceColor: Color {
        colorScheme == .dark ? BrandPalette.slate950.opacity(0.55) : Color.white.opacity(0.72)
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
