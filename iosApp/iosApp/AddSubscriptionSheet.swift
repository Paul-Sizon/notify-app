import SwiftUI

/// Modal sheet for creating a watcher.
/// Three steps in one screen: query → type segment → cadence.
/// Animated cadence pill with haptic on each tick.
struct AddSubscriptionSheet: View {
    @Environment(\.dismiss) private var dismiss
    @State private var query = ""
    @State private var type: SubscriptionKind = .event
    @State private var cadenceIndex = 2 // default: 1 hour
    @FocusState private var queryFocused: Bool

    let onCreate: (String, SubscriptionKind, Int) -> Void

    private static let cadences: [(label: String, seconds: Int)] = [
        ("15 min", 15 * 60),
        ("30 min", 30 * 60),
        ("1 hour", 3600),
        ("3 hours", 3 * 3600),
        ("6 hours", 6 * 3600),
        ("Daily", 86400),
    ]

    var body: some View {
        VStack(alignment: .leading, spacing: 0) {
            grabber
            ScrollView {
                VStack(alignment: .leading, spacing: 26) {
                    title
                    queryField
                    typeSelector
                    cadenceSelector
                }
                .padding(.horizontal, 22)
                .padding(.top, 8)
            }
            footer
        }
        .background(Theme.bgElevated.ignoresSafeArea())
        .preferredColorScheme(.dark)
        .onAppear { DispatchQueue.main.asyncAfter(deadline: .now() + 0.3) { queryFocused = true } }
    }

    private var grabber: some View {
        RoundedRectangle(cornerRadius: 3)
            .fill(Theme.label4)
            .frame(width: 38, height: 5)
            .frame(maxWidth: .infinity)
            .padding(.top, 8)
    }

    private var title: some View {
        VStack(alignment: .leading, spacing: 4) {
            Text("New watcher").font(Theme.title2()).foregroundStyle(Theme.label1)
            Text("The agent will check the web on your cadence.")
                .font(Theme.body()).foregroundStyle(Theme.label2)
        }
    }

    private var queryField: some View {
        VStack(alignment: .leading, spacing: 8) {
            Text("WATCH FOR").font(Theme.eyebrow()).foregroundStyle(Theme.label3)
            TextField("Coldplay tour São Paulo 2026", text: $query, axis: .vertical)
                .font(Theme.title3())
                .foregroundStyle(Theme.label1)
                .tint(Theme.accent)
                .focused($queryFocused)
                .lineLimit(1...3)
                .padding(14)
                .background(RoundedRectangle(cornerRadius: Theme.rMed).fill(Theme.surface))
                .overlay {
                    RoundedRectangle(cornerRadius: Theme.rMed)
                        .stroke(queryFocused ? Theme.accent : Theme.stroke,
                                lineWidth: queryFocused ? 1.5 : 0.5)
                }
                .animation(.easeOut(duration: 0.2), value: queryFocused)
        }
    }

    private var typeSelector: some View {
        VStack(alignment: .leading, spacing: 8) {
            Text("KIND").font(Theme.eyebrow()).foregroundStyle(Theme.label3)
            HStack(spacing: 8) {
                typeChip(.event, label: "Event", icon: "calendar")
                typeChip(.news, label: "News", icon: "newspaper")
            }
        }
    }
    private func typeChip(_ k: SubscriptionKind, label: String, icon: String) -> some View {
        Button {
            Haptics.selection()
            withAnimation(.spring(response: 0.3, dampingFraction: 0.8)) { type = k }
        } label: {
            HStack(spacing: 8) {
                Image(systemName: icon).font(.system(size: 14, weight: .semibold))
                Text(label).font(Theme.bodyMed())
            }
            .padding(.horizontal, 16).padding(.vertical, 11)
            .frame(maxWidth: .infinity)
            .foregroundStyle(type == k ? Theme.accentInk : Theme.label1)
            .background {
                Capsule().fill(type == k ? Theme.accent : Theme.surface)
            }
            .overlay { Capsule().stroke(Theme.stroke, lineWidth: 0.5) }
        }
        .buttonStyle(PressableStyle())
    }

    private var cadenceSelector: some View {
        VStack(alignment: .leading, spacing: 12) {
            Text("CADENCE").font(Theme.eyebrow()).foregroundStyle(Theme.label3)
            HStack {
                Text("Every \(Self.cadences[cadenceIndex].label)")
                    .font(Theme.title3()).foregroundStyle(Theme.label1)
                    .contentTransition(.numericText())
                Spacer()
                ZStack {
                    Circle().fill(Theme.accentSoft).frame(width: 38, height: 38)
                    Image(systemName: "clock").foregroundStyle(Theme.accent)
                        .font(.system(size: 14, weight: .semibold))
                }
            }
            cadenceSlider
        }
    }
    private var cadenceSlider: some View {
        let count = Self.cadences.count
        return GeometryReader { geo in
            let w = geo.size.width
            let stepW = w / CGFloat(count - 1)
            ZStack(alignment: .leading) {
                Capsule().fill(Theme.surface).frame(height: 6)
                Capsule().fill(Theme.accent)
                    .frame(width: max(8, stepW * CGFloat(cadenceIndex)), height: 6)
                Circle().fill(.white).frame(width: 22, height: 22)
                    .overlay(Circle().stroke(Theme.accent, lineWidth: 3))
                    .offset(x: stepW * CGFloat(cadenceIndex) - 11)
                    .shadow(color: .black.opacity(0.3), radius: 6, y: 2)
            }
            .frame(height: 22)
            .contentShape(Rectangle())
            .gesture(
                DragGesture(minimumDistance: 0)
                    .onChanged { v in
                        let raw = max(0, min(w, v.location.x))
                        let idx = Int((raw / stepW).rounded())
                        if idx != cadenceIndex {
                            cadenceIndex = idx
                            Haptics.selection()
                        }
                    }
            )
        }
        .frame(height: 22)
    }

    private var footer: some View {
        HStack(spacing: 10) {
            Button {
                Haptics.tap()
                dismiss()
            } label: {
                Text("Cancel")
                    .font(Theme.bodyMed()).foregroundStyle(Theme.label1)
                    .frame(maxWidth: .infinity, minHeight: 50)
                    .background(Capsule().fill(Theme.surface))
                    .overlay(Capsule().stroke(Theme.stroke, lineWidth: 0.5))
            }
            .buttonStyle(PressableStyle())

            Button {
                let trimmed = query.trimmingCharacters(in: .whitespacesAndNewlines)
                guard trimmed.count >= 3 else { Haptics.error(); return }
                onCreate(trimmed, type, Self.cadences[cadenceIndex].seconds)
                dismiss()
            } label: {
                HStack(spacing: 8) {
                    Text("Start watching").font(.system(size: 15, weight: .semibold))
                    Image(systemName: "arrow.right").font(.system(size: 13, weight: .bold))
                }
                .foregroundStyle(Theme.accentInk)
                .frame(maxWidth: .infinity, minHeight: 50)
                .background(Capsule().fill(Theme.accent))
                .opacity(query.trimmingCharacters(in: .whitespacesAndNewlines).count >= 3 ? 1 : 0.4)
            }
            .buttonStyle(PressableStyle())
            .disabled(query.trimmingCharacters(in: .whitespacesAndNewlines).count < 3)
        }
        .padding(.horizontal, 22).padding(.top, 14).padding(.bottom, 26)
        .background(Theme.bgElevated)
    }
}
