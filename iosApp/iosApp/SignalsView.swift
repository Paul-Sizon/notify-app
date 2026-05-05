import SwiftUI

/// Firehose tab — every signal across all watchers, plus a 14-day activity
/// strip. The strip uses tiny bars whose heights map to per-day signal
/// counts; today's bar is rendered in accent so the eye lands on "right now."
struct SignalsView: View {
    @Environment(AppState.self) private var state
    @Environment(\.openSubscription) private var openSubscription

    private var dayBuckets: [(Date, Int)] {
        let cal = Calendar.current
        let today = cal.startOfDay(for: Date())
        let days = (0..<14).map { cal.date(byAdding: .day, value: -$0, to: today)! }.reversed()
        var counts: [Date: Int] = [:]
        for (sig, _) in state.allSignalsRecent {
            let d = cal.startOfDay(for: sig.firstSeenAt)
            counts[d, default: 0] += 1
        }
        return days.map { ($0, counts[$0] ?? 0) }
    }

    var body: some View {
        ScrollView {
            VStack(alignment: .leading, spacing: 22) {
                header.padding(.horizontal, 22).padding(.top, 8)
                statsCard.padding(.horizontal, 22)
                activityBars.padding(.horizontal, 22)
                if !state.allSignalsRecent.isEmpty {
                    VStack(alignment: .leading, spacing: 10) {
                        Text("LATEST")
                            .font(Theme.eyebrow()).foregroundStyle(Theme.label3)
                            .padding(.horizontal, 22)
                        VStack(spacing: 10) {
                            ForEach(state.allSignalsRecent.prefix(40), id: \.0.id) { (sig, sub) in
                                AlertRow(signal: sig, subscription: sub) { openSubscription(sub) }
                            }
                        }
                        .padding(.horizontal, 22)
                    }
                }
            }
            .padding(.bottom, 24)
        }
        .background(Theme.bg.ignoresSafeArea())
        .scrollIndicators(.hidden)
        .refreshable { await state.refresh() }
    }

    private var header: some View {
        VStack(alignment: .leading, spacing: 6) {
            Text("Signals").font(Theme.title1()).tracking(-0.4).foregroundStyle(Theme.label1)
            Text("Everything the agents have surfaced.")
                .font(Theme.body()).foregroundStyle(Theme.label2)
        }
    }

    private var statsCard: some View {
        HStack(spacing: 0) {
            stat("WATCHING", "\(state.activeSubscriptions.count)")
            divider
            stat("RESOLVED", "\(state.resolvedSubscriptions.count)")
            divider
            stat("SIGNALS",  "\(state.allSignalsRecent.count)")
        }
        .padding(.vertical, 18).padding(.horizontal, 4)
        .background(RoundedRectangle(cornerRadius: Theme.rLg).fill(Theme.surface))
        .overlay(RoundedRectangle(cornerRadius: Theme.rLg).stroke(Theme.stroke, lineWidth: 0.5))
    }
    private func stat(_ label: String, _ value: String) -> some View {
        VStack(spacing: 6) {
            Text(value).font(.system(size: 28, weight: .semibold).monospacedDigit())
                .foregroundStyle(Theme.label1)
            Text(label).font(Theme.eyebrow()).foregroundStyle(Theme.label3)
        }
        .frame(maxWidth: .infinity)
    }
    private var divider: some View {
        Rectangle().fill(Theme.stroke).frame(width: 0.5, height: 36)
    }

    private var activityBars: some View {
        let buckets = dayBuckets
        let maxV = max(1, buckets.map(\.1).max() ?? 1)
        let cal = Calendar.current
        let today = cal.startOfDay(for: Date())
        return VStack(alignment: .leading, spacing: 10) {
            HStack {
                Text("LAST 14 DAYS").font(Theme.eyebrow()).foregroundStyle(Theme.label3)
                Spacer()
                Text("\(buckets.last?.1 ?? 0) today")
                    .font(Theme.eyebrow()).foregroundStyle(Theme.accent)
            }
            HStack(alignment: .bottom, spacing: 6) {
                ForEach(Array(buckets.enumerated()), id: \.offset) { (i, pair) in
                    let isToday = cal.isDate(pair.0, inSameDayAs: today)
                    Capsule()
                        .fill(isToday ? Theme.accent : Theme.surfaceHi)
                        .frame(height: max(4, CGFloat(pair.1) / CGFloat(maxV) * 56))
                        .frame(maxWidth: .infinity)
                }
            }
            .frame(height: 60)
        }
    }
}
