import Foundation
import Shared

/// Swift-native value types decoupled from KMP-generated classes.
///
/// Why this exists: SwiftUI views work best with `Hashable`/`Identifiable`
/// structs and `Date`. KMP gives us serialized strings + Kotlin classes whose
/// `Hashable` conformance comes from `KotlinObject`. Mapping at the boundary
/// lets every `View`, `@State`, and `ForEach` use familiar Swift idioms.
struct Subscription: Identifiable, Hashable {
    let id: String
    let query: String
    let type: SubscriptionKind
    let cadenceSeconds: Int
    let lastRunAt: Date?
    let nextRunAt: Date
    let createdAt: Date
}

enum SubscriptionKind: String, Hashable {
    case event, news
    init(wire: String) { self = SubscriptionKind(rawValue: wire) ?? .event }
}

struct Signal: Identifiable, Hashable {
    let id: String
    let subscriptionId: String
    let title: String
    let body: String?
    let url: String?
    let occursAt: Date?
    let sourceDomains: [String]
    let confidence: Float
    let firstSeenAt: Date

    /// Heuristic for the design's "resolved" state: an event with a confirmed
    /// future occurrence date — the agent has nothing more to discover.
    var isResolved: Bool {
        guard let occursAt else { return false }
        return occursAt > Date()
    }
}

extension Subscription {
    init(_ d: SubscriptionDTO) {
        self.id = d.id
        self.query = d.query
        self.type = SubscriptionKind(wire: d.type)
        self.cadenceSeconds = Int(d.cadenceSeconds)
        self.lastRunAt = d.lastRunAt.flatMap(ISO8601Date.parse)
        self.nextRunAt = ISO8601Date.parse(d.nextRunAt) ?? Date()
        self.createdAt = ISO8601Date.parse(d.createdAt) ?? Date()
    }
}

extension Signal {
    init(_ d: SignalDTO) {
        self.id = d.id
        self.subscriptionId = d.subscriptionId
        self.title = d.title
        self.body = d.body
        self.url = d.url
        self.occursAt = d.occursAt.flatMap(ISO8601Date.parse)
        self.sourceDomains = d.sourceDomains
        self.confidence = d.confidence
        self.firstSeenAt = ISO8601Date.parse(d.firstSeenAt) ?? Date()
    }
}

/// Server emits two date shapes: full RFC3339 for timestamps and `YYYY-MM-DD`
/// for event dates (occurs_at). One parser handles both.
enum ISO8601Date {
    private static let iso: ISO8601DateFormatter = {
        let f = ISO8601DateFormatter()
        f.formatOptions = [.withInternetDateTime, .withFractionalSeconds]
        return f
    }()
    private static let isoNoFrac: ISO8601DateFormatter = {
        let f = ISO8601DateFormatter()
        f.formatOptions = [.withInternetDateTime]
        return f
    }()
    private static let dayOnly: DateFormatter = {
        let f = DateFormatter()
        f.dateFormat = "yyyy-MM-dd"
        f.timeZone = TimeZone(identifier: "UTC")
        f.locale = Locale(identifier: "en_US_POSIX")
        return f
    }()

    static func parse(_ s: String) -> Date? {
        iso.date(from: s) ?? isoNoFrac.date(from: s) ?? dayOnly.date(from: s)
    }
}
