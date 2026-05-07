import Foundation
import Shared

/// Thin Swift wrapper around the KMP `ApiClient`.
///
/// Two responsibilities:
/// 1. Persist `deviceId` across launches (so subscriptions outlive process
///    restarts — server uses it to scope every request).
/// 2. Translate KMP DTOs → Swift value types so the rest of the app never
///    imports `Shared` directly.
final class ApiService {
    private var client: ApiClient
    private(set) var baseURL: String
    private let store = UserDefaults.standard
    private let deviceIdKey = "notify.deviceId"

    var deviceId: String? {
        get { client.deviceId }
        set {
            client.deviceId = newValue
            if let v = newValue { store.set(v, forKey: deviceIdKey) }
            else { store.removeObject(forKey: deviceIdKey) }
        }
    }

    init(baseURL: String) {
        self.baseURL = baseURL
        self.client = ApiClient(baseUrl: baseURL, engine: nil, enableLogging: false)
        if let saved = store.string(forKey: deviceIdKey) {
            self.client.deviceId = saved
        }
    }

    /// Ensures `deviceId` is set — registers a fresh device if cold start.
    /// Pass the hex APNs device token from `PushService.awaitToken()`. If nil
    /// (sim, perms denied, no network), falls back to a placeholder so the
    /// app remains usable; pushes won't deliver until the user re-launches
    /// with a real token available.
    func bootstrapDevice(apnsToken: String?) async throws {
        if deviceId != nil { return }
        let token = apnsToken ?? "ios-no-apns-\(UUID().uuidString)"
        let id = try await client.registerDevice(apnsToken: token)
        self.deviceId = id
    }

    func listSubscriptions() async throws -> [Subscription] {
        let dtos = try await client.listSubscriptions()
        return dtos.map(Subscription.init)
    }

    func createSubscription(query: String, type: SubscriptionKind, cadenceSeconds: Int) async throws -> Subscription {
        let kind: SubscriptionType = (type == .event) ? .event : .news
        let dto = try await client.createSubscription(
            query: query,
            type: kind,
            cadenceSeconds: Int32(cadenceSeconds)
        )
        return Subscription(dto)
    }

    func deleteSubscription(_ id: String) async throws {
        try await client.deleteSubscription(id: id)
    }

    func runSubscription(_ id: String) async throws -> Int {
        let resp = try await client.runSubscription(id: id)
        return Int(resp.newSignals)
    }

    /// Free-text AI suggester. Stateless, no auth.
    func suggestFromContext(_ context: String) async throws -> OnboardingResult {
        let resp = try await client.suggestFromContext(context: context)
        let sugs = resp.suggestions.map { dto -> OnboardingSuggestion in
            OnboardingSuggestion(
                query: dto.query,
                type: SubscriptionKind(wire: dto.type),
                cadenceSeconds: Int(dto.cadenceSeconds),
                reason: dto.reason
            )
        }
        return OnboardingResult(suggestions: sugs, fallback: resp.fallback)
    }

    /// Onboarding suggest call. Stateless on the server; safe to retry.
    /// Does NOT require `bootstrapDevice()` first — the endpoint has no auth.
    func suggestOnboarding(
        city: String,
        country: String,
        role: OnboardingRole,
        roleOther: String?,
        interests: [OnboardingInterest]
    ) async throws -> OnboardingResult {
        let req = OnboardingRequest(
            city: city,
            country: country,
            role: role.wire,
            roleOther: roleOther,
            interests: interests.map { $0.wire }
        )
        let resp = try await client.suggestOnboarding(req: req)
        let sugs = resp.suggestions.map { dto -> OnboardingSuggestion in
            OnboardingSuggestion(
                query: dto.query,
                type: SubscriptionKind(wire: dto.type),
                cadenceSeconds: Int(dto.cadenceSeconds),
                reason: dto.reason
            )
        }
        return OnboardingResult(suggestions: sugs, fallback: resp.fallback)
    }

    func listSignals(subscriptionId: String, limit: Int = 50) async throws -> [Signal] {
        let dtos = try await client.listSignals(
            subscriptionId: subscriptionId,
            limit: Int32(limit),
            before: nil
        )
        return dtos.map(Signal.init)
    }
}
