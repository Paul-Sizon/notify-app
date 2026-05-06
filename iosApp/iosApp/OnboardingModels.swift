import Foundation

/// Closed enum of onboarding role chips. `wire` is the canonical string the
/// server validates against (mirrored in `internal/api/onboarding.go`).
enum OnboardingRole: String, CaseIterable, Hashable {
    case developer, founder, designer, investor, student, other

    var wire: String { rawValue }

    var label: String {
        switch self {
        case .developer: "Developer"
        case .founder:   "Founder"
        case .designer:  "Designer"
        case .investor:  "Investor"
        case .student:   "Student"
        case .other:     "Other"
        }
    }

    var icon: String {
        switch self {
        case .developer: "chevron.left.forwardslash.chevron.right"
        case .founder:   "flag.fill"
        case .designer:  "paintbrush.fill"
        case .investor:  "chart.line.uptrend.xyaxis"
        case .student:   "graduationcap.fill"
        case .other:     "person.fill"
        }
    }
}

/// Closed enum of interest chips — order is the display order.
/// Mirrored server-side; adding a value requires a server change too.
enum OnboardingInterest: String, CaseIterable, Hashable {
    case concerts
    case tech_meetups
    case crypto_web3
    case fintech
    case startups_vc
    case ai_ml
    case sports
    case art_design
    case food_restaurants
    case politics_policy
    case gaming
    case film_tv

    var wire: String { rawValue }

    var label: String {
        switch self {
        case .concerts:         "Concerts"
        case .tech_meetups:     "Tech meetups"
        case .crypto_web3:      "Crypto & web3"
        case .fintech:          "Fintech"
        case .startups_vc:      "Startups & VC"
        case .ai_ml:            "AI & ML"
        case .sports:           "Sports"
        case .art_design:       "Art & design"
        case .food_restaurants: "Food & restaurants"
        case .politics_policy:  "Politics & policy"
        case .gaming:           "Gaming"
        case .film_tv:          "Film & TV"
        }
    }
}

/// One suggestion in Swift land. Local UUID is used to track activation
/// state in the view (server doesn't return one — server is stateless here).
struct OnboardingSuggestion: Identifiable, Hashable {
    let id = UUID()
    let query: String
    let type: SubscriptionKind
    let cadenceSeconds: Int
    let reason: String
}

struct OnboardingResult {
    let suggestions: [OnboardingSuggestion]
    let fallback: Bool
}

/// Persistence keys live here so the view + AppState can both read the flag
/// without stringly-typed drift.
enum OnboardingFlags {
    static let completedKey = "onboarding_completed_v1"

    static var completed: Bool {
        get { UserDefaults.standard.bool(forKey: completedKey) }
        set { UserDefaults.standard.set(newValue, forKey: completedKey) }
    }
}
