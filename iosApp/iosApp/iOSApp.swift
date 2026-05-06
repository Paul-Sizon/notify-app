import SwiftUI

@main
struct iOSApp: App {
    @State private var showSplash = true

    init() {
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
