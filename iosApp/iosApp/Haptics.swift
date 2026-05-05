import UIKit

/// Centralized haptics so each kind has consistent voice across screens.
/// Generators are short-lived per call — Apple recommends `prepare()` only
/// when the trigger time is predictable and ≤100ms away. We don't have that.
enum Haptics {
    static func selection() {
        UISelectionFeedbackGenerator().selectionChanged()
    }
    static func tap() {
        UIImpactFeedbackGenerator(style: .light).impactOccurred()
    }
    static func tapMedium() {
        UIImpactFeedbackGenerator(style: .medium).impactOccurred()
    }
    static func tapHeavy() {
        UIImpactFeedbackGenerator(style: .heavy).impactOccurred()
    }
    static func soft() {
        UIImpactFeedbackGenerator(style: .soft).impactOccurred(intensity: 0.7)
    }
    static func success() {
        UINotificationFeedbackGenerator().notificationOccurred(.success)
    }
    static func warning() {
        UINotificationFeedbackGenerator().notificationOccurred(.warning)
    }
    static func error() {
        UINotificationFeedbackGenerator().notificationOccurred(.error)
    }
}
