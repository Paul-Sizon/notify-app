import Foundation
import UserNotifications
import UIKit

/// Local-notification push mock.
///
/// Why this exists: APNs needs an Apple Developer account, an .p8 key, and a
/// device token round-trip. For the hackathon we sidestep all of it and
/// schedule a `UNNotificationRequest` with a 0.5s trigger — the OS delivers a
/// real banner, lockscreen entry, and sound. Tap-to-deep-link works the same
/// as remote APNs.
///
/// Swap for remote APNs later by replacing `deliver(...)` with a real APNs
/// device-token registration; the rest of the app stays unchanged.
@MainActor
final class MockPush: NSObject, UNUserNotificationCenterDelegate {
    static let shared = MockPush()

    /// Last notification tapped — observed by RootView to deep-link into the
    /// signal that prompted the alert.
    @Published var lastTappedSignalId: String?

    private override init() { super.init() }

    func bootstrap() {
        UNUserNotificationCenter.current().delegate = self
    }

    func requestAuthorization() async -> Bool {
        do {
            return try await UNUserNotificationCenter.current()
                .requestAuthorization(options: [.alert, .sound, .badge])
        } catch {
            return false
        }
    }

    /// Schedule a local push for a freshly-arrived signal.
    /// `subscriptionId` doubles as the deep-link target.
    func deliver(
        title: String,
        body: String,
        subscriptionId: String,
        signalId: String,
        delay: TimeInterval = 0.5
    ) {
        let content = UNMutableNotificationContent()
        content.title = title
        content.body = body
        content.sound = .default
        content.userInfo = [
            "subscription_id": subscriptionId,
            "signal_id": signalId,
        ]
        let trigger = UNTimeIntervalNotificationTrigger(timeInterval: delay, repeats: false)
        let req = UNNotificationRequest(identifier: "signal-\(signalId)", content: content, trigger: trigger)
        UNUserNotificationCenter.current().add(req)
    }

    // Foreground: still show the banner so users see it during dev/demo.
    nonisolated func userNotificationCenter(
        _ center: UNUserNotificationCenter,
        willPresent notification: UNNotification
    ) async -> UNNotificationPresentationOptions {
        [.banner, .sound, .list]
    }

    // Tap → set lastTappedSignalId for RootView to handle.
    nonisolated func userNotificationCenter(
        _ center: UNUserNotificationCenter,
        didReceive response: UNNotificationResponse
    ) async {
        let info = response.notification.request.content.userInfo
        if let sigId = info["signal_id"] as? String {
            await MainActor.run { self.lastTappedSignalId = sigId }
        }
    }
}
