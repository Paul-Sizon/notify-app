import SwiftUI

@main
struct iOSApp: App {
    init() {
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
