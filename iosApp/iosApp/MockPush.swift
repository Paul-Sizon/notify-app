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
    /// Returns a status string: nil on success, error message otherwise — so
    /// callers can surface "permission denied" etc. as a toast instead of
    /// silently failing.
    @discardableResult
    func deliver(
        title: String,
        body: String,
        subscriptionId: String,
        signalId: String,
        delay: TimeInterval = 0.5
    ) async -> String? {
        let center = UNUserNotificationCenter.current()
        let settings = await center.notificationSettings()
        switch settings.authorizationStatus {
        case .denied:
            return "Notifications denied — enable in Settings → notify."
        case .notDetermined:
            let granted = (try? await center.requestAuthorization(options: [.alert, .sound, .badge])) ?? false
            if !granted { return "Permission not granted." }
        case .authorized, .provisional, .ephemeral:
            break
        @unknown default:
            return "Unknown permission state."
        }
        let content = UNMutableNotificationContent()
        content.title = title
        content.body = body
        content.sound = .default
        content.userInfo = [
            "subscription_id": subscriptionId,
            "signal_id": signalId,
        ]
        if let attachment = Self.iconAttachment() {
            content.attachments = [attachment]
        }
        let trigger = UNTimeIntervalNotificationTrigger(timeInterval: max(0.1, delay), repeats: false)
        let req = UNNotificationRequest(identifier: "signal-\(signalId)", content: content, trigger: trigger)
        do {
            try await center.add(req)
            return nil
        } catch {
            return "Schedule failed: \(error.localizedDescription)"
        }
    }

    /// Build a `UNNotificationAttachment` from the bundled `NotificationIcon`
    /// asset. Notifications render the first attachment as the large
    /// right-side thumbnail in the banner and on the lockscreen — without
    /// this the slot stays blank. Attachments require an on-disk file URL,
    /// so we PNG-encode the asset once into `tmp/` and reuse it.
    private static func iconAttachment() -> UNNotificationAttachment? {
        let url = FileManager.default.temporaryDirectory
            .appendingPathComponent("notify-notification-icon.png")
        if !FileManager.default.fileExists(atPath: url.path) {
            guard let img = UIImage(named: "NotificationIcon"),
                  let data = img.pngData() else { return nil }
            do { try data.write(to: url, options: .atomic) }
            catch { return nil }
        }
        return try? UNNotificationAttachment(
            identifier: "notify-icon",
            url: url,
            options: [UNNotificationAttachmentOptionsTypeHintKey: "public.png"]
        )
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
