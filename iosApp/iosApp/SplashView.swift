import SwiftUI

/// Launch splash — black canvas, signal-green ECG waveform draws left→right
/// then dot-pulses on the R-peak before handing off to `ContentView`.
///
/// The system launch screen (`UILaunchScreen_Generation = YES`) is unavoidably
/// static, so this view is the *first* SwiftUI scene and runs the animation
/// before the real UI appears.
struct SplashView: View {
    let onFinish: () -> Void

    @State private var trim: CGFloat = 0

    private let drawDuration: Double = 1.4
    private let holdDuration: Double = 0.55

    var body: some View {
        ZStack {
            Color.black.ignoresSafeArea()

            GeometryReader { geo in
                let waveHeight: CGFloat = 220
                let waveRect = CGRect(
                    x: 0,
                    y: geo.size.height * 0.42 - waveHeight / 2,
                    width: geo.size.width,
                    height: waveHeight
                )

                ZStack {
                    // Faint baseline so screen never looks empty mid-draw.
                    Path { p in
                        p.move(to: CGPoint(x: 0, y: waveRect.midY))
                        p.addLine(to: CGPoint(x: waveRect.width, y: waveRect.midY))
                    }
                    .stroke(Theme.accent.opacity(0.08), lineWidth: 1)

                    // Animated ECG trace.
                    ECGWave()
                        .trim(from: 0, to: trim)
                        .stroke(
                            Theme.accent,
                            style: StrokeStyle(lineWidth: 3.5, lineCap: .round, lineJoin: .round)
                        )
                        .shadow(color: Theme.accentGlow, radius: 12)
                        .shadow(color: Theme.accent.opacity(0.35), radius: 28)
                        .frame(width: waveRect.width, height: waveRect.height)
                        .offset(y: waveRect.minY)
                }
            }
        }
        .onAppear { run() }
    }

    private func run() {
        withAnimation(.easeInOut(duration: drawDuration)) {
            trim = 1
        }
        DispatchQueue.main.asyncAfter(deadline: .now() + drawDuration + holdDuration) {
            onFinish()
        }
    }
}

/// Single-cycle ECG: baseline → P → QRS spike → T → baseline.
/// Coordinates are normalised to the supplied rect so it scales cleanly.
private struct ECGWave: Shape {
    func path(in rect: CGRect) -> Path {
        var p = Path()
        let mid = rect.midY
        let w = rect.width
        let h = rect.height

        p.move(to: CGPoint(x: 0, y: mid))
        p.addLine(to: CGPoint(x: w * 0.30, y: mid))

        // P-wave (small atrial bump)
        p.addLine(to: CGPoint(x: w * 0.33, y: mid - h * 0.06))
        p.addLine(to: CGPoint(x: w * 0.36, y: mid))

        // QRS complex (the spike)
        p.addLine(to: CGPoint(x: w * 0.42, y: mid + h * 0.10))   // Q dip
        p.addLine(to: CGPoint(x: w * 0.46, y: mid - h * 0.45))   // R peak
        p.addLine(to: CGPoint(x: w * 0.50, y: mid + h * 0.42))   // S trough
        p.addLine(to: CGPoint(x: w * 0.54, y: mid))

        // T-wave (repolarisation hump)
        p.addLine(to: CGPoint(x: w * 0.62, y: mid - h * 0.12))
        p.addLine(to: CGPoint(x: w * 0.70, y: mid))

        p.addLine(to: CGPoint(x: w, y: mid))
        return p
    }
}

#Preview {
    SplashView(onFinish: {})
}
