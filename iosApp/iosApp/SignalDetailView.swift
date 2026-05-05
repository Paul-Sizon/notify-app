import SwiftUI

/// One subscription's full detail: hero query, cadence/source stats,
/// "Run agent now" CTA, then a chronological signal stream.
struct SignalDetailView: View {
    let subscription: Subscription
    @Environment(AppState.self) private var state
    @Environment(\.dismiss) private var dismiss

    private var signals: [Signal] { state.signals(for: subscription.id) }
    private var confirmed: Date? { state.confirmedDate(for: subscription) }

    var body: some View {
        ScrollView {
            VStack(alignment: .leading, spacing: 22) {
                hero.padding(.horizontal, 22).padding(.top, 14)
                runRow.padding(.horizontal, 22)
                if !signals.isEmpty {
                    signalsList
                } else {
                    emptyState
                }
            }
            .padding(.bottom, 40)
        }
        .background(Theme.bg.ignoresSafeArea())
        .scrollIndicators(.hidden)
        .toolbar {
            ToolbarItem(placement: .topBarTrailing) {
                Menu {
                    Button(role: .destructive) {
                        Task {
                            await state.delete(subscription)
                            dismiss()
                        }
                    } label: { Label("Delete watcher", systemImage: "trash") }
                } label: {
                    Image(systemName: "ellipsis.circle").foregroundStyle(Theme.label1)
                }
            }
        }
    }

    private var hero: some View {
        VStack(alignment: .leading, spacing: 14) {
            HStack(spacing: 8) {
                TypePill(type: subscription.type)
                if let c = confirmed {
                    HStack(spacing: 6) {
                        Image(systemName: "calendar.badge.checkmark").font(.system(size: 11, weight: .semibold))
                        Text(c.formatted(.dateTime.month(.abbreviated).day().year()))
                            .font(.system(size: 12, weight: .semibold))
                    }
                    .padding(.horizontal, 8).padding(.vertical, 5)
                    .background(Capsule().fill(Theme.accentSoft))
                    .foregroundStyle(Theme.accent)
                }
            }
            Text(subscription.query)
                .font(.system(size: 28, weight: .semibold))
                .tracking(-0.4)
                .foregroundStyle(Theme.label1)
            CadenceChip(cadenceSeconds: subscription.cadenceSeconds, lastRunAt: subscription.lastRunAt)
        }
    }

    private var runRow: some View {
        Button {
            Haptics.tapMedium()
            Task { await state.run(subscription) }
        } label: {
            HStack(spacing: 10) {
                ZStack {
                    Circle().fill(Theme.accentInk.opacity(0.2)).frame(width: 28, height: 28)
                    Image(systemName: "sparkles").foregroundStyle(Theme.accentInk)
                        .font(.system(size: 13, weight: .bold))
                }
                Text("Run agent now").font(.system(size: 15, weight: .semibold))
                    .foregroundStyle(Theme.accentInk)
                Spacer()
                Image(systemName: "arrow.right").font(.system(size: 13, weight: .bold))
                    .foregroundStyle(Theme.accentInk)
            }
            .padding(.horizontal, 16).padding(.vertical, 14)
            .background(Capsule().fill(Theme.accent))
        }
        .buttonStyle(PressableStyle())
    }

    private var signalsList: some View {
        VStack(alignment: .leading, spacing: 10) {
            HStack {
                Text("\(signals.count) SIGNAL\(signals.count == 1 ? "" : "S")")
                    .font(Theme.eyebrow()).foregroundStyle(Theme.label3)
                Spacer()
            }
            .padding(.horizontal, 22)
            VStack(spacing: 10) {
                ForEach(signals) { sig in
                    SignalRow(signal: sig)
                }
            }
            .padding(.horizontal, 22)
        }
    }

    private var emptyState: some View {
        VStack(spacing: 12) {
            ZStack {
                Circle().fill(Theme.accentSoft).frame(width: 76, height: 76)
                Image(systemName: "magnifyingglass").font(.system(size: 26, weight: .light))
                    .foregroundStyle(Theme.accent)
            }
            Text("No signals yet")
                .font(Theme.title3()).foregroundStyle(Theme.label1)
            Text("Pull the trigger above to ask the agent to look right now.")
                .font(Theme.body()).foregroundStyle(Theme.label2)
                .multilineTextAlignment(.center).frame(maxWidth: 260)
        }
        .frame(maxWidth: .infinity).padding(.top, 30)
    }
}

struct SignalRow: View {
    let signal: Signal

    var body: some View {
        VStack(alignment: .leading, spacing: 8) {
            Text(signal.title)
                .font(Theme.bodyMed()).foregroundStyle(Theme.label1)
                .multilineTextAlignment(.leading)
                .lineLimit(3)
            if let body = signal.body, !body.isEmpty {
                Text(body).font(Theme.body()).foregroundStyle(Theme.label2)
                    .lineLimit(3)
            }
            HStack(spacing: 8) {
                if let dom = signal.sourceDomains.first {
                    HStack(spacing: 4) {
                        Image(systemName: "link").font(.system(size: 10, weight: .semibold))
                        Text(dom).font(.system(size: 11, weight: .medium))
                    }
                    .foregroundStyle(Theme.label3)
                }
                Text(signal.firstSeenAt.formatted(.relative(presentation: .named)))
                    .font(.system(size: 11, weight: .regular))
                    .foregroundStyle(Theme.label3)
                Spacer()
                ConfidenceBar(value: signal.confidence)
            }
        }
        .padding(14)
        .frame(maxWidth: .infinity, alignment: .leading)
        .background(RoundedRectangle(cornerRadius: Theme.rMed).fill(Theme.surface))
        .overlay(RoundedRectangle(cornerRadius: Theme.rMed).stroke(Theme.stroke, lineWidth: 0.5))
        .onTapGesture {
            Haptics.tap()
            if let urlStr = signal.url, let url = URL(string: urlStr) {
                UIApplication.shared.open(url)
            }
        }
    }
}

struct ConfidenceBar: View {
    let value: Float
    var body: some View {
        HStack(spacing: 3) {
            ForEach(0..<3) { i in
                Capsule()
                    .fill(Float(i + 1) * 0.34 <= value ? Theme.accent : Theme.surfaceHi)
                    .frame(width: 6, height: 4)
            }
        }
    }
}
