import Foundation
import Observation

/// Owns onboarding state across the four screens. Pure VM — no SwiftUI here.
///
/// Activation calls the existing `POST /v1/subscriptions` path via `AppState`
/// so the home screen automatically picks up the new rows on its next refresh
/// (no separate "onboarding subscription" type, per spec §1).
@Observable
@MainActor
final class OnboardingViewModel {
    enum Step: Int, CaseIterable { case city, role, interests, reveal }

    // ─── inputs ───
    var step: Step = .city
    var city: String?
    var country: String?
    var role: OnboardingRole?
    var roleOther: String = ""
    var interests: Set<OnboardingInterest> = []

    // ─── reveal screen ───
    var suggestions: [OnboardingSuggestion] = []
    var fallback: Bool = false
    var loading: Bool = false
    var error: String?

    /// local-suggestion-id → server-subscription-id (so deactivation can DELETE).
    var activatedSubIds: [UUID: String] = [:]
    var inFlight: Set<UUID> = []

    private let appState: AppState

    init(appState: AppState) {
        self.appState = appState
    }

    // ─── flow control ───

    func advance(to s: Step) {
        step = s
    }

    var canSubmitInterests: Bool { !interests.isEmpty }

    func toggleInterest(_ i: OnboardingInterest) {
        if interests.contains(i) {
            interests.remove(i)
        } else if interests.count < 8 {
            interests.insert(i)
        }
        // 9th tap intentionally rejected silently — UI shakes the chip.
    }

    // ─── network ───

    /// Calls the suggest endpoint and enforces a minimum 1.2s loading window
    /// (per spec §4.6): instant reveal "feels cheaper than a brief moment of
    /// anticipation."
    func fetchSuggestions() async {
        guard let role else { return }
        loading = true
        error = nil

        let start = Date()
        let minLoad: TimeInterval = 1.2
        do {
            let result = try await appState.api.suggestOnboarding(
                city: city ?? "Worldwide",
                country: country ?? "",
                role: role,
                roleOther: role == .other ? roleOther.trimmingCharacters(in: .whitespacesAndNewlines) : nil,
                interests: Array(interests)
            )
            await holdMinimum(start: start, minimum: minLoad)
            suggestions = result.suggestions
            fallback = result.fallback
        } catch {
            await holdMinimum(start: start, minimum: minLoad)
            // Soft failure — show a local fallback so the user is never stuck
            // staring at an error during onboarding.
            suggestions = Self.localFallback(city: city, country: country)
            fallback = true
            self.error = "Couldn't personalize — showing popular watchers."
        }
        loading = false
    }

    private func holdMinimum(start: Date, minimum: TimeInterval) async {
        let elapsed = Date().timeIntervalSince(start)
        if elapsed < minimum {
            try? await Task.sleep(nanoseconds: UInt64((minimum - elapsed) * 1_000_000_000))
        }
    }

    /// Optimistic activate — flip card immediately; revert on failure.
    func activate(_ s: OnboardingSuggestion) async {
        guard activatedSubIds[s.id] == nil, !inFlight.contains(s.id) else { return }
        inFlight.insert(s.id)
        defer { inFlight.remove(s.id) }
        do {
            try await appState.api.bootstrapDevice()
            let sub = try await appState.api.createSubscription(
                query: s.query,
                type: s.type,
                cadenceSeconds: s.cadenceSeconds
            )
            activatedSubIds[s.id] = sub.id
            // Mirror into AppState so the home screen shows it after dismiss
            // without waiting for a refresh round-trip.
            appState.injectSubscription(sub)
            Haptics.success()
        } catch {
            Haptics.error()
            self.error = "Couldn't start that one — try again."
        }
    }

    func deactivate(_ s: OnboardingSuggestion) async {
        guard let subId = activatedSubIds[s.id], !inFlight.contains(s.id) else { return }
        inFlight.insert(s.id)
        defer { inFlight.remove(s.id) }
        do {
            try await appState.api.bootstrapDevice()
            try await appState.api.deleteSubscription(subId)
            activatedSubIds[s.id] = nil
            appState.removeSubscription(id: subId)
            Haptics.tap()
        } catch {
            Haptics.error()
            self.error = "Couldn't stop that one — try again."
        }
    }

    func isActivated(_ s: OnboardingSuggestion) -> Bool {
        activatedSubIds[s.id] != nil
    }

    var activatedCount: Int { activatedSubIds.count }

    func complete() {
        OnboardingFlags.completed = true
    }

    // ─── helpers ───

    private static func localFallback(city: String?, country: String?) -> [OnboardingSuggestion] {
        let c = (city?.lowercased()) ?? "your city"
        let cn = (country?.lowercased()) ?? "your country"
        return [
            OnboardingSuggestion(query: "tech events \(c)", type: .event, cadenceSeconds: 21600,
                                 reason: "Local meetups and conferences."),
            OnboardingSuggestion(query: "concerts \(c)", type: .event, cadenceSeconds: 21600,
                                 reason: "Live music near you."),
            OnboardingSuggestion(query: "major concert announcements \(cn)", type: .event, cadenceSeconds: 86400,
                                 reason: "Big tours coming through."),
            OnboardingSuggestion(query: "cryptocurrency regulation \(cn)", type: .news, cadenceSeconds: 86400,
                                 reason: "Regulatory shifts in your country."),
            OnboardingSuggestion(query: "ai industry news", type: .news, cadenceSeconds: 86400,
                                 reason: "Major AI announcements."),
        ]
    }
}
