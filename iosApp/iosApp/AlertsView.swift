import SwiftUI

/// Push-history feed. Groups signals by Today / Yesterday / Earlier.
/// Each row is a one-line "what arrived" with the parent watcher's query
/// and the source domain.
struct AlertsView: View {
    @Environment(AppState.self) private var state
    @Environment(\.openSubscription) private var openSubscription

    private var grouped: [(String, [(Signal, Subscription)])] {
        let cal = Calendar.current
        let today = cal.startOfDay(for: Date())
        let yest  = cal.date(byAdding: .day, value: -1, to: today)!
        let week  = cal.date(byAdding: .day, value: -7, to: today)!
        var todayG: [(Signal, Subscription)] = []
        var yestG:  [(Signal, Subscription)] = []
        var weekG:  [(Signal, Subscription)] = []
        var olderG: [(Signal, Subscription)] = []
        for pair in state.allSignalsRecent {
            let d = pair.0.firstSeenAt
            if d >= today { todayG.append(pair) }
            else if d >= yest { yestG.append(pair) }
            else if d >= week { weekG.append(pair) }
            else { olderG.append(pair) }
        }
        var out: [(String, [(Signal, Subscription)])] = []
        if !todayG.isEmpty { out.append(("Today", todayG)) }
        if !yestG.isEmpty  { out.append(("Yesterday", yestG)) }
        if !weekG.isEmpty  { out.append(("This week", weekG)) }
        if !olderG.isEmpty { out.append(("Earlier", olderG)) }
        return out
    }

    var body: some View {
        ScrollView {
            VStack(alignment: .leading, spacing: 24) {
                header.padding(.horizontal, 22).padding(.top, 8)
                if state.allSignalsRecent.isEmpty {
                    EmptyStateView(
                        title: "No alerts yet",
                        subtitle: "Watchers will buzz your phone here when something new is found.",
                        cta: nil, action: nil
                    )
                    .frame(maxWidth: .infinity).padding(.top, 48)
                } else {
                    ForEach(grouped, id: \.0) { (label, items) in
                        VStack(alignment: .leading, spacing: 10) {
                            Text(label.uppercased())
                                .font(Theme.eyebrow())
                                .foregroundStyle(Theme.label3)
                                .padding(.horizontal, 22)
                            VStack(spacing: 10) {
                                ForEach(items, id: \.0.id) { (sig, sub) in
                                    AlertRow(signal: sig, subscription: sub) { openSubscription(sub) }
                                }
                            }
                            .padding(.horizontal, 22)
                        }
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
            HStack(alignment: .lastTextBaseline) {
                Text("Alerts").font(Theme.title1()).tracking(-0.4).foregroundStyle(Theme.label1)
                Text("\(state.allSignalsRecent.count)")
                    .font(.system(size: 22, weight: .regular).monospacedDigit())
                    .foregroundStyle(Theme.label3)
                Spacer()
            }
            Text("Every signal that buzzed your phone.")
                .font(Theme.body()).foregroundStyle(Theme.label2)
        }
    }
}

struct AlertRow: View {
    let signal: Signal
    let subscription: Subscription
    let onTap: () -> Void

    var body: some View {
        Button(action: { Haptics.tap(); onTap() }) {
            HStack(alignment: .top, spacing: 12) {
                ZStack {
                    Circle().fill(Theme.accentSoft).frame(width: 32, height: 32)
                    Image(systemName: subscription.type == .event ? "calendar" : "newspaper")
                        .font(.system(size: 13, weight: .semibold))
                        .foregroundStyle(Theme.accent)
                }
                VStack(alignment: .leading, spacing: 4) {
                    Text(subscription.query)
                        .font(.system(size: 12, weight: .semibold))
                        .foregroundStyle(Theme.label3)
                        .tracking(0.3)
                        .textCase(.uppercase)
                    Text(signal.title)
                        .font(.system(size: 15, weight: .medium))
                        .foregroundStyle(Theme.label1)
                        .lineLimit(2)
                        .multilineTextAlignment(.leading)
                    HStack(spacing: 6) {
                        if let dom = signal.sourceDomains.first {
                            Text(dom).font(Theme.eyebrow()).foregroundStyle(Theme.label3)
                            Text("·").foregroundStyle(Theme.label4)
                        }
                        Text(signal.firstSeenAt.formatted(.relative(presentation: .named)))
                            .font(Theme.eyebrow()).foregroundStyle(Theme.label3)
                    }
                }
                Spacer(minLength: 0)
            }
            .padding(14)
            .frame(maxWidth: .infinity, alignment: .leading)
            .background(RoundedRectangle(cornerRadius: Theme.rMed).fill(Theme.surface))
            .overlay(RoundedRectangle(cornerRadius: Theme.rMed).stroke(Theme.stroke, lineWidth: 0.5))
        }
        .buttonStyle(PressableStyle(scale: 0.98))
    }
}
