import SwiftUI

/// Root container — owns AppState, mounts CustomTabBar + active screen.
/// Bootstrap (device registration + initial fetch) runs in `.task`.
struct ContentView: View {
    @State private var state = AppState()
    @State private var showAdd = false
    @State private var showAI = false
    @State private var detailSub: Subscription?
    @State private var showOnboarding: Bool = !OnboardingFlags.completed

    var body: some View {
        ZStack(alignment: .bottom) {
            Theme.bg.ignoresSafeArea()

            // Active tab content
            currentTab
                .environment(state)
                .padding(.bottom, 92) // leave room for tab bar

            // Toast banners (top of stack)
            VStack(spacing: 8) {
                ToastBanner(message: Bindable(state).lastError, tone: .error)
                ToastBanner(message: Bindable(state).toast, tone: .info)
                Spacer()
            }
            .padding(.top, 50)
            .allowsHitTesting(state.lastError != nil || state.toast != nil)

            // FAB (only on watchers tab, never empty)
            if state.selectedTab == .watchers && !state.subscriptions.isEmpty {
                GlassFAB { showAdd = true }
                    .padding(.bottom, 102)
                    .padding(.trailing, 22)
                    .frame(maxWidth: .infinity, alignment: .trailing)
                    .transition(.scale.combined(with: .opacity))
            }

            // Custom tab bar
            CustomTabBar(active: state.selectedTab) { tab in
                Haptics.selection()
                withAnimation(.spring(response: 0.32, dampingFraction: 0.85)) {
                    state.selectedTab = tab
                }
            }
        }
        .preferredColorScheme(.dark)
        // Live-agent overlay — declared before .sheet so sheet presentations
        // render on top. Otherwise a stuck liveAgentSubId blocks all touches
        // on any modal the user opens.
        .overlay {
            if let runningId = state.liveAgentSubId,
               let sub = state.subscriptions.first(where: { $0.id == runningId }) {
                LiveAgentView(subscription: sub)
                    .transition(.opacity.combined(with: .move(edge: .bottom)))
                    .zIndex(100)
            }
        }
        .animation(.spring(response: 0.4, dampingFraction: 0.85), value: state.liveAgentSubId)
        .task {
            PushService.shared.bootstrap()
            // Kick the permission prompt + APNs registration concurrently.
            // `state.bootstrap()` will await the token (with timeout) so the
            // first /v1/devices call ships a real APNs token.
            Task { _ = await PushService.shared.requestAuthorization() }
            await state.bootstrap()
            // Existing-user hack (spec §5): if the user already has watchers,
            // they pre-date the onboarding feature — don't drop them into the
            // flow. Mark complete and dismiss.
            if !state.subscriptions.isEmpty && !OnboardingFlags.completed {
                OnboardingFlags.completed = true
                showOnboarding = false
            }
        }
        .fullScreenCover(isPresented: $showOnboarding) {
            OnboardingFlowView(appState: state) {
                withAnimation(.spring(response: 0.4, dampingFraction: 0.85)) {
                    showOnboarding = false
                }
            }
            .interactiveDismissDisabled()
        }
        .sheet(isPresented: $showAdd) {
            AddSubscriptionSheet(
                onCreate: { query, type, cadence in
                    Task { await state.create(query: query, type: type, cadenceSeconds: cadence) }
                },
                onAI: {
                    showAdd = false
                    DispatchQueue.main.asyncAfter(deadline: .now() + 0.4) { showAI = true }
                }
            )
            .environment(state)
            .presentationDetents([.large])
            .presentationDragIndicator(.visible)
            .presentationBackground(Theme.bgElevated)
        }
        .sheet(isPresented: $showAI) {
            NavigationStack {
                AISuggestionsView()
            }
            .environment(state)
            .presentationDetents([.large])
            .presentationDragIndicator(.hidden)
            .presentationBackground(Theme.bg)
        }
        .sheet(item: $detailSub) { sub in
            SignalDetailView(subscription: sub)
                .environment(state)
                .presentationDetents([.large])
                .presentationDragIndicator(.visible)
                .presentationBackground(Theme.bg)
        }
        .environment(\.openSubscription, { sub in detailSub = sub })
    }

    @ViewBuilder
    private var currentTab: some View {
        switch state.selectedTab {
        case .watchers: WatchersView(onAdd: { showAdd = true }, onAI: { showAI = true })
        case .alerts:   AlertsView()
        case .signals:  SignalsView()
        case .account:  AccountView()
        }
    }
}

// Environment plumbing so any card can request a detail open without
// drilling closures down through three components.
private struct OpenSubscriptionKey: EnvironmentKey {
    static let defaultValue: (Subscription) -> Void = { _ in }
}
extension EnvironmentValues {
    var openSubscription: (Subscription) -> Void {
        get { self[OpenSubscriptionKey.self] }
        set { self[OpenSubscriptionKey.self] = newValue }
    }
}
