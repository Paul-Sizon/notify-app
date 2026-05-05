import SwiftUI

@main
struct iOSApp: App {
    init() {
        // Drop any persisted backend URL override left over from earlier
        // builds — the build-time constant in AppState.swift is now the
        // single source of truth, and a stale persisted URL was making
        // physical devices hit dead tunnel hosts after rotation.
        BackendURL.clearLegacyOverride()

        // Force dark navigation bar tint to match theme.
        UINavigationBar.appearance().barTintColor = UIColor(Theme.bg)
    }
    var body: some Scene {
        WindowGroup {
            ContentView()
                .preferredColorScheme(.dark)
                .tint(Theme.accent)
        }
    }
}
