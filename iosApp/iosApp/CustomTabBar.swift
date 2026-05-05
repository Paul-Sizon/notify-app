import SwiftUI

/// Bottom tab bar — outline icons → fill on active, accent label.
/// Built custom (not TabView) so we can do precise spring animation on the
/// pill highlight + matched-geometry between active states without fighting
/// SwiftUI's defaults.
struct CustomTabBar: View {
    let active: AppState.Tab
    let onChange: (AppState.Tab) -> Void

    @Namespace private var ns

    var body: some View {
        HStack(spacing: 0) {
            ForEach(AppState.Tab.allCases, id: \.self) { tab in
                cell(tab)
                    .frame(maxWidth: .infinity)
                    .contentShape(Rectangle())
                    .onTapGesture { onChange(tab) }
            }
        }
        .padding(.top, 10)
        .padding(.bottom, 18)
        .background {
            // Glass: blur + tinted overlay + top hairline
            ZStack {
                Rectangle().fill(.ultraThinMaterial)
                Rectangle().fill(Theme.surface.opacity(0.55))
                VStack(spacing: 0) {
                    Rectangle().fill(Theme.strokeHi).frame(height: 0.5)
                    Spacer()
                }
            }
            .ignoresSafeArea(edges: .bottom)
        }
        .frame(height: 78)
    }

    @ViewBuilder
    private func cell(_ tab: AppState.Tab) -> some View {
        let isActive = active == tab
        VStack(spacing: 4) {
            ZStack {
                if isActive {
                    Capsule()
                        .fill(Theme.accentSoft)
                        .frame(width: 52, height: 32)
                        .matchedGeometryEffect(id: "tab.bg", in: ns)
                }
                Image(systemName: tab.icon(active: isActive))
                    .font(.system(size: 19, weight: isActive ? .semibold : .regular))
                    .foregroundStyle(isActive ? Theme.accent : Theme.label2)
                    .symbolEffect(.bounce, value: isActive)
            }
            .frame(height: 32)

            Text(tab.title)
                .font(.system(size: 10, weight: isActive ? .semibold : .medium))
                .foregroundStyle(isActive ? Theme.accent : Theme.label3)
                .tracking(0.2)
        }
    }
}

extension AppState.Tab {
    var title: String {
        switch self {
        case .watchers: "Watchers"
        case .alerts:   "Alerts"
        case .signals:  "Signals"
        case .account:  "Account"
        }
    }
    func icon(active: Bool) -> String {
        switch self {
        case .watchers: active ? "eye.fill" : "eye"
        case .alerts:   active ? "bell.fill" : "bell"
        case .signals:  active ? "waveform.path" : "waveform"
        case .account:  active ? "person.crop.circle.fill" : "person.crop.circle"
        }
    }
}
