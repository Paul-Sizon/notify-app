import SwiftUI

@main
struct iOSApp: App {
    @State private var showSplash = true

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
