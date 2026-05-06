import SwiftUI

@main
struct iOSApp: App {
    @State private var showSplash = true

    init() {
        // One-shot migration: drop deviceId minted against the old localhost
        // /tunnel backend. The Pi-backed build needs a fresh device registered
        // against the new server's DB; the old UUID would 401 every request.
        let migrated = "notify.deviceId.migrated_to_pi"
        if !UserDefaults.standard.bool(forKey: migrated) {
            UserDefaults.standard.removeObject(forKey: "notify.deviceId")
            UserDefaults.standard.removeObject(forKey: "onboarding_completed_v1")
            UserDefaults.standard.set(true, forKey: migrated)
        }

        UINavigationBar.appearance().barTintColor = UIColor(Theme.bg)
    }

    var body: some Scene {
        WindowGroup {
            ZStack {
                ContentView()
                    .preferredColorScheme(.dark)
                    .tint(Theme.accent)

                if showSplash {
                    SplashView {
                        withAnimation(.easeOut(duration: 0.35)) {
                            showSplash = false
                        }
                    }
                    .transition(.opacity)
                    .zIndex(1)
                }
            }
        }
    }
}
