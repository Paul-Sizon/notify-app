import Foundation
import Observation
import SwiftUI

/// Root observable state for the entire app.
///
/// `@Observable` (Swift 5.9+) re-renders only views that read each property —
/// no manual `@Published` boilerplate. `Set<String>` of seen signal IDs lets
/// us detect newly-arrived signals on each run and trigger mock pushes.
/// Backend URL — hardcoded, served from the Pi via Tailscale Funnel.
enum BackendURL {
    static let url = "https://raspberrypi.taile76757.ts.net"
}

@Observable
@MainActor
final class AppState {
    private(set) var subscriptions: [Subscription] = []
    private(set) var signalsBySub: [String: [Signal]] = [:]
    var seenSignalIds: Set<String> = []

    var loading: Bool = false
    var lastError: String?
    var toast: String?

    /// Tab index — own state lives here, not in TabView, so we can deep-link
    /// from a notification tap.
    var selectedTab: Tab = .watchers

    /// Latest live-agent run for the watchers screen (visualizer overlay).
    var liveAgentSubId: String?

    let api: ApiService

    init(api: ApiService = ApiService(baseURL: BackendURL.url)) {
        self.api = api
    }

    enum Tab: Int, Hashable, CaseIterable {
        case watchers, alerts, signals, account
    }

    func bootstrap() async {
        do {
            try await api.bootstrapDevice()
            await refresh()
        } catch {
            lastError = "bootstrap: \(error.localizedDescription)"
        }
    }

    func refresh() async {
        loading = true
        defer { loading = false }
        do {
            let subs = try await api.listSubscriptions()
            self.subscriptions = subs.sorted { $0.createdAt > $1.createdAt }
            // Parallelise signal fetches — N subs = N concurrent requests, not N sequential.
            let pairs: [(String, [Signal])] = try await withThrowingTaskGroup(of: (String, [Signal]).self) { group in
                for sub in subs {
                    let id = sub.id
                    group.addTask { [api] in
                        let sigs = try await api.listSignals(subscriptionId: id, limit: 30)
                        return (id, sigs)
                    }
                }
                var out: [(String, [Signal])] = []
                for try await pair in group { out.append(pair) }
                return out
            }
            for (id, sigs) in pairs {
                signalsBySub[id] = sigs
                for s in sigs { seenSignalIds.insert(s.id) }
            }
        } catch {
            lastError = "refresh: \(error.localizedDescription)"
        }
    }

    func create(query: String, type: SubscriptionKind, cadenceSeconds: Int) async {
        do {
            let s = try await api.createSubscription(query: query, type: type, cadenceSeconds: cadenceSeconds)
            subscriptions.insert(s, at: 0)
            signalsBySub[s.id] = []
            Haptics.success()
            toast = "Watcher added: \(s.query)"
            // Authoritative resync runs in background — UI already optimistic,
            // no reason to make the caller (and the toast) wait on a full refetch.
            Task { await refresh() }
            // Kick the agent immediately so the user sees results without
            // waiting for the cadence boundary. `run` drives the LiveAgentView
            // overlay, fetches signals, and pushes for any new ones.
            Task { await run(s) }
        } catch {
            Haptics.error()
            lastError = "create failed: \(error.localizedDescription)"
        }
    }

    func delete(_ sub: Subscription) async {
        do {
            try await api.deleteSubscription(sub.id)
            subscriptions.removeAll { $0.id == sub.id }
            signalsBySub[sub.id] = nil
            Haptics.warning()
        } catch {
            lastError = "delete: \(error.localizedDescription)"
        }
    }

    /// Trigger an agent run — backend executes search → extract → insert,
    /// returns count of new signals. Newly arrived signals get mock pushes.
    func run(_ sub: Subscription) async {
        liveAgentSubId = sub.id
        do {
            _ = try await api.runSubscription(sub.id)
            let sigs = try await api.listSignals(subscriptionId: sub.id, limit: 30)
            let newOnes = sigs.filter { !seenSignalIds.contains($0.id) }
            signalsBySub[sub.id] = sigs
            for s in newOnes {
                seenSignalIds.insert(s.id)
                if let err = await MockPush.shared.deliver(
                    title: sub.query,
                    body: s.title,
                    subscriptionId: sub.id,
                    signalId: s.id
                ) {
                    lastError = err
                }
            }
            // give the visualizer a beat before dismissing
            try? await Task.sleep(nanoseconds: 600_000_000)
            liveAgentSubId = nil
            if !newOnes.isEmpty { Haptics.success() } else { Haptics.tap() }
        } catch {
            liveAgentSubId = nil
            Haptics.error()
            lastError = "run: \(error.localizedDescription)"
        }
    }

    /// Onboarding-side helper: drop a freshly-created subscription into local
    /// state so the home screen shows it immediately when onboarding dismisses,
    /// without waiting for a refresh round-trip.
    func injectSubscription(_ s: Subscription) {
        if !subscriptions.contains(where: { $0.id == s.id }) {
            subscriptions.insert(s, at: 0)
            signalsBySub[s.id] = []
        }
    }

    func removeSubscription(id: String) {
        subscriptions.removeAll { $0.id == id }
        signalsBySub[id] = nil
    }

    // ─── Derived ───
    func signals(for subId: String) -> [Signal] { signalsBySub[subId] ?? [] }

    var allSignalsRecent: [(Signal, Subscription)] {
        var out: [(Signal, Subscription)] = []
        for sub in subscriptions {
            for s in (signalsBySub[sub.id] ?? []) { out.append((s, sub)) }
        }
        return out.sorted { $0.0.firstSeenAt > $1.0.firstSeenAt }
    }

    var resolvedSubscriptions: [Subscription] {
        subscriptions.filter { sub in
            (signalsBySub[sub.id] ?? []).contains { $0.isResolved }
        }
    }

    var activeSubscriptions: [Subscription] {
        let resolvedIds = Set(resolvedSubscriptions.map(\.id))
        return subscriptions.filter { !resolvedIds.contains($0.id) }
    }

    func confirmedDate(for sub: Subscription) -> Date? {
        (signalsBySub[sub.id] ?? [])
            .compactMap(\.occursAt)
            .filter { $0 > Date() }
            .min()
    }
}
