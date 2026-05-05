import Foundation
import Observation
import SwiftUI

/// Root observable state for the entire app.
///
/// `@Observable` (Swift 5.9+) re-renders only views that read each property —
/// no manual `@Published` boilerplate. `Set<String>` of seen signal IDs lets
/// us detect newly-arrived signals on each run and trigger mock pushes.
/// Backend URL store. Persists to UserDefaults so a physical device can be
/// pointed at a tunnel (e.g. https://*.trycloudflare.com) and remember it
/// across launches. Sim defaults to localhost.
enum BackendURL {
    static let key = "notify.baseURL"
    #if targetEnvironment(simulator)
    static let defaultURL = "http://localhost:8080"
    #else
    /// Cloudflared quick tunnel — rotates each time `cloudflared tunnel --url`
    /// restarts. Refresh this constant + reinstall when the URL changes,
    /// or just edit it via Account → Backend URL on-device.
    static let defaultURL = "https://courier-organizational-instrumental-monroe.trycloudflare.com"
    #endif

    static var current: String {
        UserDefaults.standard.string(forKey: key) ?? defaultURL
    }
    static func set(_ url: String) {
        let trimmed = url.trimmingCharacters(in: .whitespacesAndNewlines)
        if trimmed.isEmpty { UserDefaults.standard.removeObject(forKey: key) }
        else { UserDefaults.standard.set(trimmed, forKey: key) }
    }
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

    init(api: ApiService = ApiService(baseURL: BackendURL.current)) {
        self.api = api
    }

    /// Swap backend URL at runtime — clears local state so the new server's
    /// device id is fetched cleanly. Caller should follow up with `bootstrap()`.
    func switchBackend(to url: String) async {
        BackendURL.set(url)
        api.deviceId = nil
        subscriptions = []
        signalsBySub = [:]
        seenSignalIds = []
        api.rebuild(baseURL: BackendURL.current)
        toast = "Backend → \(url)"
        await bootstrap()
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
            // Authoritative resync — guards against local parse drift.
            await refresh()
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
