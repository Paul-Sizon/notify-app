import SwiftUI

// ─── GlassFAB ──────────────────────────────────────────────────────
/// Floating "+" button. Spring-scales on press, glows accent.
struct GlassFAB: View {
    let action: () -> Void
    @State private var pressed = false

    var body: some View {
        Button {
            Haptics.tapMedium()
            action()
        } label: {
            ZStack {
                Circle().fill(Theme.accent)
                Image(systemName: "plus")
                    .font(.system(size: 22, weight: .bold))
                    .foregroundStyle(Theme.accentInk)
            }
            .frame(width: 56, height: 56)
            .shadow(color: Theme.accentGlow, radius: 16, y: 6)
            .shadow(color: .black.opacity(0.3), radius: 8, y: 3)
        }
        .buttonStyle(PressableStyle())
    }
}

/// Press-style: gentle scale-down on press for any tappable surface.
struct PressableStyle: ButtonStyle {
    var scale: CGFloat = 0.94
    func makeBody(configuration: Configuration) -> some View {
        configuration.label
            .scaleEffect(configuration.isPressed ? scale : 1)
            .animation(.spring(response: 0.28, dampingFraction: 0.7), value: configuration.isPressed)
    }
}

// ─── CadenceChip ────────────────────────────────────────────────────
/// Small uppercase pill displaying poll cadence + relative last-run.
struct CadenceChip: View {
    let cadenceSeconds: Int
    let lastRunAt: Date?

    var body: some View {
        HStack(spacing: 6) {
            Circle().fill(Theme.accent).frame(width: 5, height: 5)
                .overlay(Circle().fill(Theme.accent).blur(radius: 4).opacity(0.8))
            Text(cadenceLabel)
                .font(Theme.eyebrow())
                .foregroundStyle(Theme.label2)
            if let r = relative {
                Text("·").foregroundStyle(Theme.label4)
                Text(r).font(Theme.eyebrow()).foregroundStyle(Theme.label3)
            }
        }
        .padding(.horizontal, 10).padding(.vertical, 6)
        .background(Capsule().fill(Theme.surfaceMute))
        .overlay(Capsule().stroke(Theme.stroke, lineWidth: 0.5))
    }

    private var cadenceLabel: String {
        let h = cadenceSeconds / 3600
        let m = (cadenceSeconds % 3600) / 60
        if h >= 24 { return "EVERY \(h/24)D" }
        if h >= 1  { return m == 0 ? "EVERY \(h)H" : "EVERY \(h)H \(m)M" }
        return "EVERY \(m)M"
    }
    private var relative: String? {
        guard let lastRunAt else { return "NEVER" }
        let secs = Int(Date().timeIntervalSince(lastRunAt))
        if secs < 60 { return "JUST NOW" }
        if secs < 3600 { return "\(secs/60)M AGO" }
        if secs < 86400 { return "\(secs/3600)H AGO" }
        return "\(secs/86400)D AGO"
    }
}

// ─── SubscriptionCard ───────────────────────────────────────────────
/// One row in the Watchers list. Two visual states:
/// - active: query bold, cadence chip live, optional unread accent dot
/// - resolved: muted bg, struck query, accent date stamp + RESOLVED tag
struct SubscriptionCard: View {
    let subscription: Subscription
    let signals: [Signal]
    let confirmedDate: Date?
    let onTap: () -> Void
    let onRun: () -> Void

    private var isResolved: Bool { confirmedDate != nil }
    private var unread: Bool { (signals.first?.firstSeenAt ?? .distantPast) > Date().addingTimeInterval(-3600) }

    var body: some View {
        Button(action: { Haptics.tap(); onTap() }) {
            VStack(alignment: .leading, spacing: 12) {
                // Top row: type pill + unread dot or RESOLVED tag
                HStack {
                    TypePill(type: subscription.type)
                    Spacer()
                    if isResolved {
                        Text("RESOLVED")
                            .font(Theme.eyebrow())
                            .foregroundStyle(Theme.accent)
                            .padding(.horizontal, 8).padding(.vertical, 4)
                            .background(Capsule().fill(Theme.accentSoft))
                    } else if unread {
                        ZStack {
                            Circle().fill(Theme.accent).frame(width: 8, height: 8)
                            Circle().stroke(Theme.accent.opacity(0.4), lineWidth: 6)
                                .frame(width: 8, height: 8).blur(radius: 3)
                        }
                    }
                }

                // Query text
                Text(subscription.query)
                    .font(Theme.title3())
                    .foregroundStyle(isResolved ? Theme.label2 : Theme.label1)
                    .strikethrough(isResolved, color: Theme.label3)
                    .lineLimit(2)
                    .multilineTextAlignment(.leading)

                // Footer row
                HStack(spacing: 10) {
                    if let date = confirmedDate {
                        HStack(spacing: 6) {
                            Image(systemName: "calendar.badge.checkmark")
                                .font(.system(size: 12, weight: .semibold))
                            Text(date.formatted(.dateTime.month(.abbreviated).day().year()))
                                .font(.system(size: 13, weight: .semibold))
                        }
                        .foregroundStyle(Theme.accent)
                    } else {
                        CadenceChip(cadenceSeconds: subscription.cadenceSeconds, lastRunAt: subscription.lastRunAt)
                    }
                    Spacer()
                    Text("\(signals.count)")
                        .font(.system(size: 13, weight: .semibold).monospacedDigit())
                        .foregroundStyle(Theme.label2)
                    Image(systemName: "waveform")
                        .font(.system(size: 12, weight: .medium))
                        .foregroundStyle(Theme.label3)
                }
            }
            .padding(18)
            .frame(maxWidth: .infinity, alignment: .leading)
            .background {
                RoundedRectangle(cornerRadius: Theme.rLg, style: .continuous)
                    .fill(isResolved ? Theme.surfaceMute : Theme.surface)
            }
            .overlay {
                RoundedRectangle(cornerRadius: Theme.rLg, style: .continuous)
                    .stroke(Theme.stroke, lineWidth: 0.5)
            }
        }
        .buttonStyle(PressableStyle(scale: 0.98))
        .contextMenu {
            Button {
                Haptics.tapMedium()
                onRun()
            } label: { Label("Run now", systemImage: "play.circle") }
        }
    }
}

struct TypePill: View {
    let type: SubscriptionKind
    var body: some View {
        let label = type == .event ? "EVENT" : "NEWS"
        let icon  = type == .event ? "calendar" : "newspaper"
        HStack(spacing: 5) {
            Image(systemName: icon).font(.system(size: 10, weight: .semibold))
            Text(label).font(Theme.eyebrow())
        }
        .foregroundStyle(Theme.label3)
        .padding(.horizontal, 8).padding(.vertical, 4)
        .background(Capsule().fill(Theme.surfaceHi.opacity(0.5)))
    }
}

// ─── Section Header ────────────────────────────────────────────────
struct SectionHeader: View {
    let title: String
    let count: Int?
    var subtitle: String?

    var body: some View {
        HStack(alignment: .lastTextBaseline) {
            Text(title)
                .font(Theme.title1())
                .foregroundStyle(Theme.label1)
                .tracking(-0.4)
            if let count {
                Text("\(count)")
                    .font(.system(size: 22, weight: .regular).monospacedDigit())
                    .foregroundStyle(Theme.label3)
            }
            Spacer()
        }
        if let s = subtitle {
            Text(s).font(Theme.body()).foregroundStyle(Theme.label2)
        }
    }
}

// ─── Toast ─────────────────────────────────────────────────────────
/// Auto-dismissing banner. Bound to a single `String?` — set it, banner
/// appears with spring; banner clears the binding on a 2.4s timer.
struct ToastBanner: View {
    @Binding var message: String?
    var tone: Tone = .info
    enum Tone { case info, error
        var bg: Color { self == .error ? Color(hex: 0x3A1418) : Theme.surfaceHi }
        var stroke: Color { self == .error ? Theme.danger.opacity(0.5) : Theme.stroke }
        var icon: String { self == .error ? "exclamationmark.triangle.fill" : "info.circle.fill" }
        var iconColor: Color { self == .error ? Theme.danger : Theme.accent }
    }

    var body: some View {
        Group {
            if let msg = message {
                HStack(spacing: 10) {
                    Image(systemName: tone.icon).foregroundStyle(tone.iconColor)
                        .font(.system(size: 15, weight: .semibold))
                    Text(msg).font(Theme.body()).foregroundStyle(Theme.label1)
                        .lineLimit(3).multilineTextAlignment(.leading)
                    Spacer(minLength: 0)
                    Button { withAnimation { message = nil } } label: {
                        Image(systemName: "xmark").font(.system(size: 11, weight: .bold))
                            .foregroundStyle(Theme.label3).padding(6)
                    }
                }
                .padding(.horizontal, 14).padding(.vertical, 12)
                .background(RoundedRectangle(cornerRadius: Theme.rMed).fill(tone.bg))
                .overlay(RoundedRectangle(cornerRadius: Theme.rMed).stroke(tone.stroke, lineWidth: 0.75))
                .shadow(color: .black.opacity(0.4), radius: 18, y: 8)
                .padding(.horizontal, 16)
                .transition(.move(edge: .top).combined(with: .opacity))
                .task(id: msg) {
                    try? await Task.sleep(nanoseconds: 2_400_000_000)
                    withAnimation { if message == msg { message = nil } }
                }
            }
        }
        .animation(.spring(response: 0.36, dampingFraction: 0.82), value: message)
    }
}

// ─── Empty State ───────────────────────────────────────────────────
struct EmptyStateView: View {
    let title: String
    let subtitle: String
    let cta: String?
    let action: (() -> Void)?

    var body: some View {
        VStack(spacing: 18) {
            ZStack {
                Circle().fill(Theme.accentSoft).frame(width: 92, height: 92)
                Image(systemName: "waveform")
                    .font(.system(size: 38, weight: .light))
                    .foregroundStyle(Theme.accent)
            }
            VStack(spacing: 6) {
                Text(title).font(Theme.title2()).foregroundStyle(Theme.label1)
                Text(subtitle).font(Theme.body()).foregroundStyle(Theme.label2)
                    .multilineTextAlignment(.center)
                    .frame(maxWidth: 260)
            }
            if let cta, let action {
                Button(action: { Haptics.tapMedium(); action() }) {
                    Text(cta)
                        .font(.system(size: 15, weight: .semibold))
                        .foregroundStyle(Theme.accentInk)
                        .padding(.horizontal, 24).padding(.vertical, 13)
                        .background(Capsule().fill(Theme.accent))
                }
                .buttonStyle(PressableStyle())
            }
        }
        .padding(40)
    }
}
