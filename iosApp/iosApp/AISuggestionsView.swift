import SwiftUI

/// AI-powered watcher suggestions from a free-text context description.
///
/// Three phases in one view:
///   1. `.input` — multi-line context field, char counter, "Find Signals" CTA
///   2. `.loading` — sparkle pulse while server runs the OpenAI call
///   3. `.reveal` — toggleable suggestion cards (mirrors onboarding reveal UX)
///
/// Activation reuses the same `POST /v1/subscriptions` path as onboarding so
/// the home screen picks up new rows on its next refresh.
struct AISuggestionsView: View {
    @Environment(\.dismiss) private var dismiss
    @Environment(AppState.self) private var state

    @State private var phase: Phase = .input
    @State private var contextText: String = Self.prefillContext
    @State private var suggestions: [OnboardingSuggestion] = []
    @State private var fallback: Bool = false
    @State private var error: String?
    @State private var activatedSubIds: [UUID: String] = [:]
    @State private var inFlight: Set<UUID> = []
    @FocusState private var focused: Bool

    private let charLimit = 2000
    private let minChars = 10

    private static let prefillContext = """
    San Francisco, Senior Software Engineer at a Series B startup (~150 people). Five years in, three at current company. Backend-leaning full-stack — Python and Go daily, currently leading a Postgres-to-distributed-storage migration. Reads Hacker News in the morning, Pragmatic Engineer on weekends. Vaguely thinking about leaving for a smaller team or starting something herself in 2 years.

    Lives in the Mission. Runs 4x/week, training for the SF Marathon. Member of a local run club. Hot yoga twice a week at a studio nearby. Climbs at Mission Cliffs on weekends. Mostly cooks — Whole Foods plus farmers' market on Saturdays. Eats out 2x/week at places with real vegetable programs (Souvla, Reem's, Nopa). Doesn't drink much; will go to a natural wine bar with friends. Sober-curious adjacent, interested in NA cocktails.

    Cultural taste: indie and electronic — Bon Iver, Caribou, Floating Points, Mitski. Catches small shows at The Independent, The Chapel, Great American. Avoids arena tours. Watches A24 movies. Genuinely interested in AI/ML developments but skeptical of hype cycles, occasionally reads papers. Goes to maybe one tech meetup a month — picks them carefully, hates recruiting-bait events.

    Civic life: watches SF politics closely — housing policy, public transit, Prop measures. Concerned about Mission gentrification, BART funding, street safety. She votes and reads the voter guide.
    """

    enum Phase { case input, loading, reveal }

    var body: some View {
        ZStack {
            Theme.bg.ignoresSafeArea()

            switch phase {
            case .input:
                inputContent
            case .loading:
                loadingContent
            case .reveal:
                revealContent
            }
        }
        .preferredColorScheme(.dark)
        .toolbar {
            ToolbarItem(placement: .principal) {
                Text("AI Suggestions")
                    .font(.system(size: 16, weight: .semibold))
                    .foregroundStyle(Theme.label1)
            }
            ToolbarItem(placement: .topBarLeading) {
                Button {
                    Haptics.tap()
                    dismiss()
                } label: {
                    Image(systemName: "chevron.left")
                        .font(.system(size: 16, weight: .semibold))
                        .foregroundStyle(Theme.label1)
                        .frame(width: 36, height: 36)
                        .background(Circle().fill(Theme.surface))
                }
            }
        }
        .toolbarBackground(Theme.bg, for: .navigationBar)
        .toolbarBackground(.visible, for: .navigationBar)
        .navigationBarTitleDisplayMode(.inline)
    }

    // ─── input phase ───
    private var inputContent: some View {
        ScrollView {
            VStack(spacing: 22) {
                Spacer().frame(height: 12)
                sparkleBadge
                VStack(spacing: 12) {
                    Text("Tell us what to look for.")
                        .font(.system(size: 30, weight: .bold))
                        .foregroundStyle(Theme.label1)
                        .multilineTextAlignment(.center)
                    Text("Describe your current interests, projects, or specific signals you want the AI to track.")
                        .font(Theme.body())
                        .foregroundStyle(Theme.label2)
                        .multilineTextAlignment(.center)
                        .lineSpacing(2)
                }
                .padding(.horizontal, 22)

                VStack(alignment: .leading, spacing: 10) {
                    Text("CONTEXT & PREFERENCES")
                        .font(Theme.eyebrow())
                        .foregroundStyle(Theme.label3)
                    contextField
                    HStack {
                        suggestionPill("Add Tech Focus", append: " I'm focused on the technical and engineering side.")
                        suggestionPill("Add Market Focus", append: " Track market moves and competitor activity.")
                    }
                    HStack {
                        Spacer()
                        Text("\(contextText.count) / \(charLimit)")
                            .font(.system(size: 12, weight: .regular).monospacedDigit())
                            .foregroundStyle(contextText.count > charLimit ? Theme.danger : Theme.label3)
                    }
                }
                .padding(.horizontal, 22)
                .padding(.top, 6)

                if let error {
                    Text(error)
                        .font(Theme.body())
                        .foregroundStyle(Theme.danger)
                        .padding(.horizontal, 22)
                }
            }
            .padding(.bottom, 120)
        }
        .scrollDismissesKeyboard(.interactively)
        .safeAreaInset(edge: .bottom, spacing: 0) { findButton }
        .onAppear {
            DispatchQueue.main.asyncAfter(deadline: .now() + 0.25) { focused = true }
        }
    }

    private var sparkleBadge: some View {
        ZStack {
            RoundedRectangle(cornerRadius: 14)
                .fill(Theme.surface)
                .overlay(RoundedRectangle(cornerRadius: 14).stroke(Theme.stroke, lineWidth: 0.5))
                .frame(width: 64, height: 64)
            Image(systemName: "sparkles")
                .font(.system(size: 26, weight: .semibold))
                .foregroundStyle(Theme.accent)
        }
    }

    private var contextField: some View {
        ZStack(alignment: .topLeading) {
            RoundedRectangle(cornerRadius: 16)
                .fill(Theme.surface)
                .overlay(RoundedRectangle(cornerRadius: 16).stroke(focused ? Theme.accent.opacity(0.5) : Theme.stroke, lineWidth: focused ? 1.2 : 0.5))
            if contextText.isEmpty {
                Text("e.g., I'm tracking advancements in solid-state batteries and emerging EV startups in Europe. Ignore consumer reviews, focus on patents, technical papers, and venture capital movements.")
                    .font(Theme.body())
                    .foregroundStyle(Theme.label3)
                    .padding(16)
                    .allowsHitTesting(false)
            }
            TextEditor(text: $contextText)
                .font(Theme.body())
                .foregroundStyle(Theme.label1)
                .tint(Theme.accent)
                .focused($focused)
                .scrollContentBackground(.hidden)
                .padding(10)
                .onChange(of: contextText) { _, newValue in
                    if newValue.count > charLimit {
                        contextText = String(newValue.prefix(charLimit))
                    }
                }
        }
        .frame(minHeight: 200)
        .animation(.easeOut(duration: 0.2), value: focused)
    }

    private func suggestionPill(_ label: String, append: String) -> some View {
        Button {
            Haptics.selection()
            let next = (contextText + append).trimmingCharacters(in: .whitespacesAndNewlines)
            contextText = String(next.prefix(charLimit))
        } label: {
            Text(label)
                .font(.system(size: 13, weight: .semibold))
                .foregroundStyle(Theme.label1)
                .padding(.horizontal, 14).padding(.vertical, 9)
                .background(Capsule().fill(Theme.surfaceHi))
                .overlay(Capsule().stroke(Theme.stroke, lineWidth: 0.5))
        }
        .buttonStyle(PressableStyle())
    }

    private var findButton: some View {
        let trimmed = contextText.trimmingCharacters(in: .whitespacesAndNewlines)
        let valid = trimmed.count >= minChars
        return Button {
            guard valid else { Haptics.error(); return }
            Haptics.tapMedium()
            focused = false
            Task { await fetch(trimmed) }
        } label: {
            HStack(spacing: 10) {
                Image(systemName: "scope").font(.system(size: 15, weight: .bold))
                Text("Find Signals").font(.system(size: 16, weight: .semibold))
            }
            .foregroundStyle(Theme.accentInk)
            .frame(maxWidth: .infinity, minHeight: 54)
            .background(Capsule().fill(Theme.accent))
            .opacity(valid ? 1 : 0.4)
        }
        .buttonStyle(PressableStyle())
        .disabled(!valid)
        .padding(.horizontal, 22)
        .padding(.top, 14)
        .padding(.bottom, 22)
        .background(
            Theme.bg
                .overlay(Rectangle().fill(Theme.stroke).frame(height: 0.5), alignment: .top)
                .ignoresSafeArea(edges: .bottom)
        )
    }

    // ─── loading phase ───
    private var loadingContent: some View {
        VStack(spacing: 20) {
            Spacer()
            PulsingSparkle()
            Text("Scanning the web for signals…")
                .font(Theme.body())
                .foregroundStyle(Theme.label2)
            Spacer()
        }
    }

    // ─── reveal phase ───
    private var revealContent: some View {
        ScrollView {
            VStack(alignment: .leading, spacing: 18) {
                VStack(alignment: .leading, spacing: 6) {
                    Text(fallback ? "Popular watchers for you" : "Suggested watchers")
                        .font(.system(size: 22, weight: .bold))
                        .foregroundStyle(Theme.label1)
                    Text("Tap to add. You can edit cadence later.")
                        .font(Theme.body())
                        .foregroundStyle(Theme.label2)
                }
                .padding(.horizontal, 22)
                .padding(.top, 8)

                VStack(spacing: 10) {
                    ForEach(suggestions) { s in
                        suggestionCard(s)
                    }
                }
                .padding(.horizontal, 22)

                if let error {
                    Text(error)
                        .font(.system(size: 13))
                        .foregroundStyle(Theme.danger)
                        .padding(.horizontal, 22)
                }
            }
            .padding(.bottom, 120)
        }
        .safeAreaInset(edge: .bottom, spacing: 0) { doneButton }
    }

    private func suggestionCard(_ s: OnboardingSuggestion) -> some View {
        let active = activatedSubIds[s.id] != nil
        return Button {
            Task {
                if active { await deactivate(s) } else { await activate(s) }
            }
        } label: {
            HStack(alignment: .top, spacing: 12) {
                ZStack {
                    Circle().fill(active ? Theme.accent : Theme.accentSoft).frame(width: 36, height: 36)
                    Image(systemName: active ? "checkmark" : (s.type == .event ? "calendar" : "newspaper"))
                        .font(.system(size: 14, weight: .bold))
                        .foregroundStyle(active ? Theme.accentInk : Theme.accent)
                }
                VStack(alignment: .leading, spacing: 4) {
                    Text(s.query)
                        .font(.system(size: 16, weight: .semibold))
                        .foregroundStyle(Theme.label1)
                        .lineLimit(2)
                        .multilineTextAlignment(.leading)
                    Text(s.reason)
                        .font(.system(size: 13))
                        .foregroundStyle(Theme.label3)
                        .lineLimit(2)
                        .multilineTextAlignment(.leading)
                    HStack(spacing: 6) {
                        Text(s.type == .event ? "Event" : "News")
                        Text("·")
                        Text(cadenceLabel(s.cadenceSeconds))
                    }
                    .font(.system(size: 12, weight: .medium))
                    .foregroundStyle(Theme.label3)
                }
                Spacer(minLength: 0)
            }
            .padding(14)
            .background(RoundedRectangle(cornerRadius: 16).fill(Theme.surface))
            .overlay(RoundedRectangle(cornerRadius: 16).stroke(active ? Theme.accent.opacity(0.6) : Theme.stroke, lineWidth: active ? 1.2 : 0.5))
        }
        .buttonStyle(PressableStyle())
        .disabled(inFlight.contains(s.id))
    }

    private var doneButton: some View {
        Button {
            Haptics.tapMedium()
            dismiss()
        } label: {
            Text(activatedSubIds.isEmpty ? "Done" : "Done · \(activatedSubIds.count) added")
                .font(.system(size: 16, weight: .semibold))
                .foregroundStyle(Theme.accentInk)
                .frame(maxWidth: .infinity, minHeight: 54)
                .background(Capsule().fill(Theme.accent))
        }
        .buttonStyle(PressableStyle())
        .padding(.horizontal, 22)
        .padding(.top, 14)
        .padding(.bottom, 22)
        .background(
            Theme.bg
                .overlay(Rectangle().fill(Theme.stroke).frame(height: 0.5), alignment: .top)
                .ignoresSafeArea(edges: .bottom)
        )
    }

    // ─── network ───
    private func fetch(_ text: String) async {
        phase = .loading
        error = nil
        let start = Date()
        let minLoad: TimeInterval = 1.0
        do {
            let result = try await state.api.suggestFromContext(text)
            await holdMin(start: start, minimum: minLoad)
            suggestions = result.suggestions
            fallback = result.fallback
            phase = .reveal
        } catch {
            await holdMin(start: start, minimum: minLoad)
            self.error = "Couldn't reach AI — \(error.localizedDescription)"
            phase = .input
            print("AI suggest failed: \(error)")
        }
    }

    private func holdMin(start: Date, minimum: TimeInterval) async {
        let elapsed = Date().timeIntervalSince(start)
        if elapsed < minimum {
            try? await Task.sleep(nanoseconds: UInt64((minimum - elapsed) * 1_000_000_000))
        }
    }

    private func activate(_ s: OnboardingSuggestion) async {
        guard activatedSubIds[s.id] == nil, !inFlight.contains(s.id) else { return }
        inFlight.insert(s.id)
        defer { inFlight.remove(s.id) }
        do {
            try await state.api.bootstrapDevice()
            let sub = try await state.api.createSubscription(
                query: s.query, type: s.type, cadenceSeconds: s.cadenceSeconds
            )
            activatedSubIds[s.id] = sub.id
            state.injectSubscription(sub)
            Haptics.success()
        } catch {
            Haptics.error()
            self.error = "Couldn't add — \(error.localizedDescription)"
        }
    }

    private func deactivate(_ s: OnboardingSuggestion) async {
        guard let subId = activatedSubIds[s.id], !inFlight.contains(s.id) else { return }
        inFlight.insert(s.id)
        defer { inFlight.remove(s.id) }
        do {
            try await state.api.deleteSubscription(subId)
            activatedSubIds[s.id] = nil
            state.removeSubscription(id: subId)
            Haptics.tap()
        } catch {
            Haptics.error()
            self.error = "Couldn't remove — \(error.localizedDescription)"
        }
    }

    private func cadenceLabel(_ s: Int) -> String {
        switch s {
        case ..<3600: return "\(s/60)m"
        case 3600: return "Hourly"
        case 21600: return "Every 6h"
        case 86400: return "Daily"
        default: return "\(s/3600)h"
        }
    }
}

private struct PulsingSparkle: View {
    @State private var scale: CGFloat = 0.85
    @State private var glow: Double = 0.3

    var body: some View {
        ZStack {
            Circle()
                .fill(Theme.accent.opacity(glow))
                .frame(width: 110, height: 110)
                .blur(radius: 16)
            ZStack {
                RoundedRectangle(cornerRadius: 18)
                    .fill(Theme.surface)
                    .overlay(RoundedRectangle(cornerRadius: 18).stroke(Theme.stroke, lineWidth: 0.5))
                    .frame(width: 80, height: 80)
                Image(systemName: "sparkles")
                    .font(.system(size: 32, weight: .semibold))
                    .foregroundStyle(Theme.accent)
            }
            .scaleEffect(scale)
        }
        .onAppear {
            withAnimation(.easeInOut(duration: 1.0).repeatForever(autoreverses: true)) {
                scale = 1.05
                glow = 0.6
            }
        }
    }
}
