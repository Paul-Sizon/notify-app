import SwiftUI

/// Full-bleed overlay shown while the agent is running.
/// "Phases" treatment: 4 stages step through with a pulsing ring + waveform
/// underline, plus a fake terminal log so the user feels the work happening.
struct LiveAgentView: View {
    let subscription: Subscription
    @State private var phase: Int = 0
    @State private var ringScale: CGFloat = 0.8

    private let phases = [
        ("Searching", "magnifyingglass"),
        ("Reading",   "doc.text"),
        ("Extracting", "sparkles"),
        ("Wrapping up", "checkmark.circle"),
    ]

    var body: some View {
        ZStack {
            Theme.bg.opacity(0.92).ignoresSafeArea()
            VStack(spacing: 32) {
                Spacer()
                pulseRing
                VStack(spacing: 8) {
                    Text("LIVE AGENT").font(Theme.eyebrow()).foregroundStyle(Theme.accent)
                        .tracking(2)
                    Text(subscription.query)
                        .font(.system(size: 22, weight: .semibold))
                        .foregroundStyle(Theme.label1)
                        .multilineTextAlignment(.center)
                        .frame(maxWidth: 320)
                }
                phaseList
                Spacer()
                Waveform(active: true).frame(height: 36)
                    .padding(.horizontal, 60)
                Spacer().frame(height: 40)
            }
            .padding(.horizontal, 22)
        }
        .onAppear {
            withAnimation(.easeInOut(duration: 1.6).repeatForever(autoreverses: true)) {
                ringScale = 1.15
            }
            startPhases()
        }
    }

    private var pulseRing: some View {
        ZStack {
            ForEach(0..<3, id: \.self) { i in
                Circle().stroke(Theme.accent.opacity(0.5 - Double(i) * 0.15), lineWidth: 1.5)
                    .frame(width: 110, height: 110)
                    .scaleEffect(ringScale + CGFloat(i) * 0.18)
                    .opacity(2.0 - ringScale - CGFloat(i) * 0.4)
            }
            Circle().fill(Theme.accentSoft).frame(width: 90, height: 90)
            Image(systemName: phases[phase].1)
                .font(.system(size: 32, weight: .light))
                .foregroundStyle(Theme.accent)
                .contentTransition(.symbolEffect(.replace))
        }
        .frame(width: 180, height: 180)
    }

    private var phaseList: some View {
        VStack(alignment: .leading, spacing: 10) {
            ForEach(0..<phases.count, id: \.self) { i in
                HStack(spacing: 10) {
                    ZStack {
                        Circle()
                            .stroke(i < phase ? Theme.accent : Theme.label4, lineWidth: 1.5)
                            .frame(width: 16, height: 16)
                        if i < phase {
                            Image(systemName: "checkmark")
                                .font(.system(size: 9, weight: .bold))
                                .foregroundStyle(Theme.accent)
                        } else if i == phase {
                            Circle().fill(Theme.accent).frame(width: 6, height: 6)
                        }
                    }
                    Text(phases[i].0)
                        .font(Theme.body())
                        .foregroundStyle(i <= phase ? Theme.label1 : Theme.label3)
                }
                .opacity(i <= phase ? 1 : 0.5)
            }
        }
        .frame(maxWidth: .infinity, alignment: .leading)
        .padding(.horizontal, 40)
    }

    private func startPhases() {
        for (i, _) in phases.enumerated().dropFirst() {
            DispatchQueue.main.asyncAfter(deadline: .now() + Double(i) * 0.6) {
                withAnimation(.spring(response: 0.4, dampingFraction: 0.8)) {
                    phase = i
                    Haptics.soft()
                }
            }
        }
    }
}

/// Animated waveform — three rows of bars whose scaleY oscillates on
/// staggered phases. Conveys "the agent is alive" without literal data.
struct Waveform: View {
    let active: Bool
    private let bars = 26
    @State private var phase: Double = 0

    var body: some View {
        TimelineView(.animation(minimumInterval: 1.0 / 30.0, paused: !active)) { ctx in
            let t = ctx.date.timeIntervalSince1970
            HStack(alignment: .center, spacing: 4) {
                ForEach(0..<bars, id: \.self) { i in
                    let x = Double(i) / Double(bars)
                    let h = abs(sin(t * 2 + x * 6)) * 0.7 + 0.3
                    Capsule().fill(Theme.accent.opacity(0.7))
                        .frame(width: 3, height: max(4, 32 * h))
                }
            }
            .frame(maxWidth: .infinity)
        }
    }
}
