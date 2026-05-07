import Foundation
import UserNotifications
import UIKit

/// Real APNs glue.
///
/// Lifecycle:
/// 1. App launch → `requestAuthorization()` prompts user.
/// 2. On grant → `UIApplication.registerForRemoteNotifications()` triggers
///    APNs token round-trip handled by `AppDelegate`.
/// 3. `AppDelegate` forwards the hex token via `setToken(_:)`.
/// 4. `awaitToken(timeout:)` lets `AppState.bootstrap()` block briefly so the
///    first `registerDevice` call ships a real token instead of a placeholder.
@MainActor
final class PushService: NSObject, UNUserNotificationCenterDelegate {
    static let shared = PushService()

    @Published var lastTappedSignalId: String?

    private var token: String?
    private var waiters: [CheckedContinuation<String?, Never>] = []

    private override init() { super.init() }

    func bootstrap() {
        UNUserNotificationCenter.current().delegate = self
    }

    @discardableResult
    func requestAuthorization() async -> Bool {
        let center = UNUserNotificationCenter.current()
        let granted = (try? await center.requestAuthorization(options: [.alert, .sound, .badge])) ?? false
        if granted {
            UIApplication.shared.registerForRemoteNotifications()
        }
        return granted
    }

    /// Called from `AppDelegate.didRegisterForRemoteNotificationsWithDeviceToken`.
    func setToken(_ hex: String) {
        token = hex
        let pending = waiters
        waiters.removeAll()
        for w in pending { w.resume(returning: hex) }
    }

    func setRegistrationFailed() {
        let pending = waiters
        waiters.removeAll()
        for w in pending { w.resume(returning: nil) }
    }

    /// Wait up to `timeout` seconds for APNs to deliver a token. Returns nil
    /// on timeout (sim, perms denied, no network) so callers can fall back.
    func awaitToken(timeout: TimeInterval = 5) async -> String? {
        if let token { return token }
        return await withCheckedContinuation { (cont: CheckedContinuation<String?, Never>) in
            waiters.append(cont)
            Task { [weak self] in
                try? await Task.sleep(nanoseconds: UInt64(timeout * 1_000_000_000))
                guard let self else { return }
                let pending = self.waiters
                self.waiters.removeAll()
                for w in pending { w.resume(returning: self.token) }
            }
        }
    }

    // Foreground: still show banner so dev/demo see it.
    nonisolated func userNotificationCenter(
        _ center: UNUserNotificationCenter,
        willPresent notification: UNNotification
    ) async -> UNNotificationPresentationOptions {
        [.banner, .sound, .list]
    }

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

/// `UIApplicationDelegate` shim — SwiftUI app needs this for APNs callbacks.
final class AppDelegate: NSObject, UIApplicationDelegate {
    func application(
        _ application: UIApplication,
        didFinishLaunchingWithOptions launchOptions: [UIApplication.LaunchOptionsKey: Any]? = nil
    ) -> Bool {
        return true
    }

    func application(
        _ application: UIApplication,
        didRegisterForRemoteNotificationsWithDeviceToken deviceToken: Data
    ) {
        let hex = deviceToken.map { String(format: "%02x", $0) }.joined()
        Task { @MainActor in PushService.shared.setToken(hex) }
        #if DEBUG
        print("APNs token:", hex)
        #endif
    }

    func application(
        _ application: UIApplication,
        didFailToRegisterForRemoteNotificationsWithError error: Error
    ) {
        Task { @MainActor in PushService.shared.setRegistrationFailed() }
        print("APNs register failed:", error.localizedDescription)
    }
}
