import SwiftUI

/// Primary tab — list of every subscription.
/// Splits into `Watching` (active) and `Resolved · paused` (events whose
/// occursAt has been confirmed in the future). Pull-to-refresh re-fetches.
struct WatchersView: View {
    @Environment(AppState.self) private var state
    @Environment(\.openSubscription) private var openSubscription
    let onAdd: () -> Void
    let onAI: () -> Void

    var body: some View {
        ScrollView {
            VStack(alignment: .leading, spacing: 22) {
                header
                    .padding(.horizontal, 22)
                    .padding(.top, 8)

                if state.subscriptions.isEmpty {
                    EmptyStateView(
                        title: "Nothing to watch yet",
                        subtitle: "Add a topic and the agent quietly checks the web on your cadence — pinging only when something changes.",
                        cta: "Add your first watcher",
                        action: onAdd
                    )
                    .frame(maxWidth: .infinity)
                    .padding(.top, 48)
                } else {
                    activeList
                    if !state.resolvedSubscriptions.isEmpty {
                        resolvedList
                    }
                }
            }
            .padding(.bottom, 24)
        }
        .background(Theme.bg.ignoresSafeArea())
        .refreshable {
            Haptics.soft()
            await state.refresh()
        }
        .scrollIndicators(.hidden)
    }

    private var header: some View {
        VStack(alignment: .leading, spacing: 6) {
            HStack(alignment: .lastTextBaseline) {
                Text("Watching")
                    .font(Theme.title1())
                    .tracking(-0.4)
                    .foregroundStyle(Theme.label1)
                Text("\(state.activeSubscriptions.count)")
                    .font(.system(size: 22, weight: .regular).monospacedDigit())
                    .foregroundStyle(Theme.label3)
                Spacer()
                aiButton
                statusOrb
            }
            Text("Quiet by default. Loud when it matters.")
                .font(Theme.body())
                .foregroundStyle(Theme.label2)
        }
    }

    private var aiButton: some View {
        Button {
            Haptics.tap()
            onAI()
        } label: {
            HStack(spacing: 6) {
                Image(systemName: "sparkles").font(.system(size: 12, weight: .bold))
                Text("AI").font(.system(size: 13, weight: .semibold))
            }
            .foregroundStyle(Theme.accent)
            .padding(.horizontal, 12).padding(.vertical, 7)
            .background(Capsule().fill(Theme.accentSoft))
            .overlay(Capsule().stroke(Theme.accent.opacity(0.3), lineWidth: 0.5))
        }
        .buttonStyle(PressableStyle())
    }

    private var statusOrb: some View {
        ZStack {
            Circle().fill(Theme.accent).frame(width: 8, height: 8)
            Circle().stroke(Theme.accent.opacity(0.5), lineWidth: 1)
                .frame(width: 18, height: 18)
                .scaleEffect(state.loading ? 1.4 : 1)
                .opacity(state.loading ? 0 : 1)
                .animation(.easeOut(duration: 1.2).repeatForever(autoreverses: false),
                           value: state.loading)
        }
    }

    private var activeList: some View {
        VStack(spacing: 12) {
            ForEach(state.activeSubscriptions) { sub in
                SubscriptionCard(
                    subscription: sub,
                    signals: state.signals(for: sub.id),
                    confirmedDate: nil,
                    onTap: { openSubscription(sub) },
                    onRun: { Task { await state.run(sub) } },
                    onDelete: { Task { await state.delete(sub) } }
                )
            }
        }
        .padding(.horizontal, 22)
    }

    private var resolvedList: some View {
        VStack(alignment: .leading, spacing: 14) {
            HStack(spacing: 8) {
                Text("RESOLVED · PAUSED").font(Theme.eyebrow()).foregroundStyle(Theme.label3)
                Rectangle().fill(Theme.stroke).frame(height: 1)
            }
            VStack(spacing: 12) {
                ForEach(state.resolvedSubscriptions) { sub in
                    SubscriptionCard(
                        subscription: sub,
                        signals: state.signals(for: sub.id),
                        confirmedDate: state.confirmedDate(for: sub),
                        onTap: { openSubscription(sub) },
                        onRun: { Task { await state.run(sub) } },
                        onDelete: { Task { await state.delete(sub) } }
                    )
                }
            }
        }
        .padding(.horizontal, 22)
        .padding(.top, 4)
    }
}
