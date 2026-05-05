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
    private let client: ApiClient
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
        self.client = ApiClient(baseUrl: baseURL, engine: nil, enableLogging: false)
        if let saved = store.string(forKey: deviceIdKey) {
            self.client.deviceId = saved
        }
    }

    /// Ensures `deviceId` is set — registers a fresh device if cold start.
    /// `apnsToken` is a placeholder for the mocked-push build; pass any
    /// stable string. The Go backend validates it's non-empty.
    func bootstrapDevice(mockToken: String = "ios-mock-\(UUID().uuidString)") async throws {
        if deviceId != nil { return }
        let id = try await client.registerDevice(apnsToken: mockToken)
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

    func listSignals(subscriptionId: String, limit: Int = 50) async throws -> [Signal] {
        let dtos = try await client.listSignals(
            subscriptionId: subscriptionId,
            limit: Int32(limit),
            before: nil
        )
        return dtos.map(Signal.init)
    }
}
