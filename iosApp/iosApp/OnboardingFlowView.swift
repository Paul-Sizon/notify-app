import SwiftUI

/// Four-screen onboarding: city → role → interests → reveal.
/// Mounted as a fullscreenCover from `ContentView` while
/// `OnboardingFlags.completed == false`.
struct OnboardingFlowView: View {
    @State private var model: OnboardingViewModel
    let onFinish: () -> Void

    init(appState: AppState, onFinish: @escaping () -> Void) {
        _model = State(initialValue: OnboardingViewModel(appState: appState))
        self.onFinish = onFinish
    }

    var body: some View {
        ZStack(alignment: .top) {
            Theme.bg.ignoresSafeArea()

            VStack(spacing: 0) {
                ProgressBar(progress: progress)
                    .frame(height: 2)
                    .padding(.top, 6)
                    .padding(.horizontal, 22)

                ZStack {
                    switch model.step {
                    case .city:
                        CityScreen(model: model, onContinue: goNext, onSkip: skipCity)
                            .transition(.asymmetric(
                                insertion: .move(edge: .leading).combined(with: .opacity),
                                removal:   .move(edge: .leading).combined(with: .opacity)))
                    case .role:
                        RoleScreen(model: model, onAdvance: goNext, onBack: goBack, onSkip: skip)
                            .transition(slideTransition)
                    case .interests:
                        InterestsScreen(model: model, onContinue: goNext, onBack: goBack, onSkip: skip)
                            .transition(slideTransition)
                    case .reveal:
                        RevealScreen(model: model, onContinue: finish, onBack: goBack)
                            .transition(slideTransition)
                    }
                }
                .animation(.spring(response: 0.35, dampingFraction: 0.85), value: model.step)
            }
        }
        .preferredColorScheme(.dark)
        .toast(message: Bindable(model).error)
    }

    private var progress: Double {
        Double(model.step.rawValue + 1) / Double(OnboardingViewModel.Step.allCases.count)
    }

    private var slideTransition: AnyTransition {
        .asymmetric(
            insertion: .move(edge: .trailing).combined(with: .opacity),
            removal:   .move(edge: .leading).combined(with: .opacity))
    }

    // ─── nav ───
    private func goNext() {
        Haptics.tap()
        guard let next = OnboardingViewModel.Step(rawValue: model.step.rawValue + 1) else { return }
        if next == .reveal {
            Task { await model.fetchSuggestions() }
        }
        model.advance(to: next)
    }

    private func goBack() {
        Haptics.tap()
        guard let prev = OnboardingViewModel.Step(rawValue: model.step.rawValue - 1) else { return }
        model.advance(to: prev)
    }

    private func skip() {
        // Skip from screens 2/3 jumps to reveal with whatever's been picked
        // so far. City always falls back to "Worldwide" — see skipCity.
        Haptics.tap()
        if model.city == nil { model.city = "Worldwide"; model.country = "" }
        if model.role == nil { model.role = .other; model.roleOther = "Curious" }
        if model.interests.isEmpty { model.interests.insert(.tech_meetups) }
        Task { await model.fetchSuggestions() }
        model.advance(to: .reveal)
    }

    private func skipCity() {
        model.city = "Worldwide"
        model.country = ""
        goNext()
    }

    private func finish() {
        Haptics.success()
        model.complete()
        onFinish()
    }
}

// ─── thin progress bar ───

private struct ProgressBar: View {
    let progress: Double
    var body: some View {
        GeometryReader { geo in
            ZStack(alignment: .leading) {
                Capsule().fill(Theme.surface)
                Capsule().fill(Theme.accent)
                    .frame(width: max(2, geo.size.width * progress))
                    .animation(.spring(response: 0.4, dampingFraction: 0.85), value: progress)
            }
        }
    }
}

// ─── shared header (back + skip) ───

private struct OnboardingHeader: View {
    let onBack: (() -> Void)?
    let onSkip: (() -> Void)?
    var body: some View {
        HStack {
            if let onBack {
                Button(action: onBack) {
                    Image(systemName: "chevron.left")
                        .font(.system(size: 17, weight: .semibold))
                        .foregroundStyle(Theme.label1)
                        .frame(width: 36, height: 36)
                        .background(Circle().fill(Theme.surface))
                }
                .buttonStyle(PressableStyle())
            } else { Spacer().frame(width: 36, height: 36) }

            Spacer()

            if let onSkip {
                Button(action: onSkip) {
                    Text("Skip")
                        .font(Theme.bodyMed())
                        .foregroundStyle(Theme.label2)
                }
            }
        }
        .padding(.horizontal, 22)
        .padding(.top, 12)
    }
}

// ─── screen 1: city ───

private struct CityScreen: View {
    @Bindable var model: OnboardingViewModel
    let onContinue: () -> Void
    let onSkip: () -> Void

    @State private var completer = CityCompleter()
    @State private var locator = CurrentLocationFinder()
    @State private var typed: String = ""
    @State private var freeAcceptAt: Date? = nil
    @State private var locating: Bool = false
    @State private var locateError: String? = nil
    @FocusState private var focused: Bool

    var body: some View {
        VStack(alignment: .leading, spacing: 0) {
            OnboardingHeader(onBack: nil, onSkip: onSkip)

            VStack(alignment: .leading, spacing: 10) {
                Text("Where are you?").font(Theme.title1()).foregroundStyle(Theme.label1)
                Text("We use this to find local events.")
                    .font(Theme.body()).foregroundStyle(Theme.label2)
            }
            .padding(.horizontal, 22)
            .padding(.top, 32)

            content
                .padding(.horizontal, 22)
                .padding(.top, 22)

            Spacer()

            footer
        }
        .onAppear { focused = true }
    }

    @ViewBuilder
    private var content: some View {
        if let city = model.city {
            // Selected pill replaces the field.
            HStack(spacing: 10) {
                Image(systemName: "location.fill").foregroundStyle(Theme.accent)
                VStack(alignment: .leading, spacing: 2) {
                    Text(city).font(Theme.title3()).foregroundStyle(Theme.label1)
                    if let sub = selectedSubtitle, !sub.isEmpty {
                        Text(sub).font(Theme.caption()).foregroundStyle(Theme.label2)
                    }
                }
                Spacer()
                Button {
                    model.city = nil; model.region = nil; model.country = nil
                    typed = ""; focused = true
                } label: {
                    Image(systemName: "xmark.circle.fill")
                        .foregroundStyle(Theme.label3)
                }
            }
            .padding(14)
            .background(RoundedRectangle(cornerRadius: Theme.rMed).fill(Theme.surface))
            .overlay(RoundedRectangle(cornerRadius: Theme.rMed).stroke(Theme.accent, lineWidth: 1.5))
        } else {
            VStack(alignment: .leading, spacing: 10) {
                TextField("City", text: $typed)
                    .font(Theme.title3())
                    .foregroundStyle(Theme.label1)
                    .tint(Theme.accent)
                    .focused($focused)
                    .autocorrectionDisabled()
                    .textInputAutocapitalization(.words)
                    .padding(14)
                    .background(RoundedRectangle(cornerRadius: Theme.rMed).fill(Theme.surface))
                    .overlay(RoundedRectangle(cornerRadius: Theme.rMed)
                        .stroke(focused ? Theme.accent : Theme.stroke, lineWidth: focused ? 1.5 : 0.5))
                    .onChange(of: typed) { _, new in
                        completer.query = new
                        freeAcceptAt = Date().addingTimeInterval(1.5)
                    }

                // Use-my-location row — visible only when nothing typed yet.
                if typed.isEmpty {
                    Button(action: useCurrentLocation) {
                        HStack(spacing: 10) {
                            if locating {
                                ProgressView().tint(Theme.accent)
                            } else {
                                Image(systemName: "location.circle.fill")
                                    .foregroundStyle(Theme.accent)
                            }
                            Text(locating ? "Finding you…" : "Use my current location")
                                .font(Theme.body())
                                .foregroundStyle(Theme.label1)
                            Spacer()
                        }
                        .padding(.vertical, 10)
                    }
                    .disabled(locating)
                    if let err = locateError {
                        Text(err).font(Theme.caption()).foregroundStyle(Theme.label3)
                    }
                    Divider().background(Theme.stroke)
                }

                ForEach(completer.results.prefix(6)) { row in
                    Button {
                        Haptics.selection()
                        apply(row)
                    } label: {
                        HStack(spacing: 10) {
                            if !row.flag.isEmpty {
                                Text(row.flag).font(.system(size: 22))
                            }
                            VStack(alignment: .leading, spacing: 2) {
                                Text(row.city).font(Theme.body()).foregroundStyle(Theme.label1)
                                if !row.subtitle.isEmpty {
                                    Text(row.subtitle).font(Theme.caption()).foregroundStyle(Theme.label2)
                                }
                            }
                            Spacer()
                        }
                        .padding(.vertical, 10)
                    }
                    Divider().background(Theme.stroke)
                }
            }
        }
    }

    private var selectedSubtitle: String? {
        let r = model.region ?? ""
        let c = model.country ?? ""
        if r.isEmpty { return c.isEmpty ? nil : c }
        if c.isEmpty { return r }
        return "\(r), \(c)"
    }

    private func apply(_ s: CitySuggestion) {
        model.city = s.city
        model.region = s.region
        model.country = s.country
        typed = ""
        focused = false
    }

    private func useCurrentLocation() {
        locateError = nil
        locating = true
        Task {
            defer { locating = false }
            do {
                let s = try await locator.fetch()
                Haptics.selection()
                apply(s)
            } catch CurrentLocationFinder.FinderError.denied {
                locateError = "Location permission denied. Type your city instead."
            } catch {
                locateError = "Couldn't find your location."
            }
        }
    }

    private var canContinue: Bool {
        if model.city != nil { return true }
        // 1.5s of typing without picking → accept free text per spec §4.3.
        let trimmed = typed.trimmingCharacters(in: .whitespacesAndNewlines)
        guard trimmed.count >= 2 else { return false }
        return freeAcceptAt.map { Date() > $0 } ?? false
    }

    private var footer: some View {
        Button {
            if model.city == nil {
                let trimmed = typed.trimmingCharacters(in: .whitespacesAndNewlines)
                model.city = trimmed
                model.region = ""
                model.country = ""
            }
            onContinue()
        } label: {
            Text("Continue")
                .font(.system(size: 15, weight: .semibold))
                .foregroundStyle(Theme.accentInk)
                .frame(maxWidth: .infinity, minHeight: 50)
                .background(Capsule().fill(Theme.accent))
                .opacity(canContinue ? 1 : 0.4)
        }
        .buttonStyle(PressableStyle())
        .disabled(!canContinue)
        .padding(.horizontal, 22)
        .padding(.bottom, 26)
    }
}

// ─── screen 2: role ───

private struct RoleScreen: View {
    @Bindable var model: OnboardingViewModel
    let onAdvance: () -> Void
    let onBack: () -> Void
    let onSkip: () -> Void

    @FocusState private var otherFocused: Bool

    var body: some View {
        VStack(alignment: .leading, spacing: 0) {
            OnboardingHeader(onBack: onBack, onSkip: onSkip)

            Text("What do you do?")
                .font(Theme.title1()).foregroundStyle(Theme.label1)
                .padding(.horizontal, 22).padding(.top, 32)

            grid
                .padding(.horizontal, 22).padding(.top, 28)

            if model.role == .other {
                TextField("e.g. Architect", text: Bindable(model).roleOther)
                    .font(Theme.body())
                    .foregroundStyle(Theme.label1)
                    .tint(Theme.accent)
                    .submitLabel(.done)
                    .focused($otherFocused)
                    .padding(14)
                    .background(RoundedRectangle(cornerRadius: Theme.rMed).fill(Theme.surface))
                    .overlay(RoundedRectangle(cornerRadius: Theme.rMed)
                        .stroke(otherFocused ? Theme.accent : Theme.stroke, lineWidth: otherFocused ? 1.5 : 0.5))
                    .padding(.horizontal, 22).padding(.top, 16)
                    .onAppear { otherFocused = true }
                    .onSubmit {
                        if !model.roleOther.trimmingCharacters(in: .whitespacesAndNewlines).isEmpty {
                            onAdvance()
                        }
                    }
            }

            Spacer()
        }
    }

    private var grid: some View {
        let cols = [GridItem(.flexible(), spacing: 10), GridItem(.flexible(), spacing: 10)]
        return LazyVGrid(columns: cols, spacing: 10) {
            ForEach(OnboardingRole.allCases, id: \.self) { r in
                RoleChip(role: r, selected: model.role == r) {
                    Haptics.selection()
                    withAnimation(.spring(response: 0.2, dampingFraction: 0.7)) {
                        model.role = r
                    }
                    if r == .other {
                        // Wait for textfield. Don't auto-advance.
                        return
                    }
                    DispatchQueue.main.asyncAfter(deadline: .now() + 0.3) { onAdvance() }
                }
            }
        }
    }
}

private struct RoleChip: View {
    let role: OnboardingRole
    let selected: Bool
    let onTap: () -> Void

    @State private var pulse: CGFloat = 1.0

    var body: some View {
        Button {
            withAnimation(.easeOut(duration: 0.1)) { pulse = 0.96 }
            DispatchQueue.main.asyncAfter(deadline: .now() + 0.1) {
                withAnimation(.spring(response: 0.25, dampingFraction: 0.6)) { pulse = 1.0 }
            }
            onTap()
        } label: {
            VStack(spacing: 8) {
                Image(systemName: role.icon)
                    .font(.system(size: 22, weight: .medium))
                    .foregroundStyle(selected ? Theme.accentInk : Theme.label1)
                Text(role.label)
                    .font(Theme.bodyMed())
                    .foregroundStyle(selected ? Theme.accentInk : Theme.label1)
            }
            .frame(maxWidth: .infinity)
            .frame(height: 92)
            .background(
                RoundedRectangle(cornerRadius: Theme.rLg)
                    .fill(selected ? Theme.accent : Theme.surface)
            )
            .overlay(
                RoundedRectangle(cornerRadius: Theme.rLg)
                    .stroke(Theme.stroke, lineWidth: 0.5)
            )
            .scaleEffect(pulse)
        }
        .buttonStyle(.plain)
    }
}

// ─── screen 3: interests ───

private struct InterestsScreen: View {
    @Bindable var model: OnboardingViewModel
    let onContinue: () -> Void
    let onBack: () -> Void
    let onSkip: () -> Void

    @State private var shakeId: OnboardingInterest? = nil

    var body: some View {
        VStack(alignment: .leading, spacing: 0) {
            OnboardingHeader(onBack: onBack, onSkip: onSkip)

            VStack(alignment: .leading, spacing: 10) {
                Text("What are you into?")
                    .font(Theme.title1()).foregroundStyle(Theme.label1)
                Text("Pick a few. You can add more later.")
                    .font(Theme.body()).foregroundStyle(Theme.label2)
            }
            .padding(.horizontal, 22).padding(.top, 32)

            ScrollView {
                FlowLayout(spacing: 8) {
                    ForEach(OnboardingInterest.allCases, id: \.self) { i in
                        InterestChip(
                            label: i.label,
                            selected: model.interests.contains(i),
                            shake: shakeId == i
                        ) {
                            tap(i)
                        }
                    }
                }
                .padding(.horizontal, 22)
                .padding(.top, 28)
            }

            footer
        }
    }

    private func tap(_ i: OnboardingInterest) {
        if model.interests.contains(i) {
            model.toggleInterest(i)
            Haptics.selection()
        } else if model.interests.count >= 8 {
            // 9th tap shakes per spec §4.5
            Haptics.error()
            withAnimation(.default) { shakeId = i }
            DispatchQueue.main.asyncAfter(deadline: .now() + 0.4) { shakeId = nil }
        } else {
            model.toggleInterest(i)
            Haptics.selection()
        }
    }

    private var footer: some View {
        Button(action: onContinue) {
            Text("Continue")
                .font(.system(size: 15, weight: .semibold))
                .foregroundStyle(Theme.accentInk)
                .frame(maxWidth: .infinity, minHeight: 50)
                .background(Capsule().fill(Theme.accent))
                .opacity(model.canSubmitInterests ? 1 : 0.4)
        }
        .buttonStyle(PressableStyle())
        .disabled(!model.canSubmitInterests)
        .padding(.horizontal, 22).padding(.bottom, 26).padding(.top, 12)
    }
}

private struct InterestChip: View {
    let label: String
    let selected: Bool
    let shake: Bool
    let onTap: () -> Void

    @State private var offset: CGFloat = 0

    var body: some View {
        Button(action: onTap) {
            Text(label)
                .font(Theme.bodyMed())
                .foregroundStyle(selected ? Theme.accentInk : Theme.label1)
                .padding(.horizontal, 14).padding(.vertical, 10)
                .background(Capsule().fill(selected ? Theme.accent : Theme.surface))
                .overlay(Capsule().stroke(selected ? Theme.accent : Theme.stroke, lineWidth: 0.5))
        }
        .buttonStyle(.plain)
        .offset(x: offset)
        .onChange(of: shake) { _, new in
            guard new else { return }
            withAnimation(.default.repeatCount(3, autoreverses: true).speed(4)) { offset = -6 }
            DispatchQueue.main.asyncAfter(deadline: .now() + 0.4) { offset = 0 }
        }
    }
}

// ─── flow layout (chip wrapping) ───

/// Minimal SwiftUI flow layout. iOS 16+.
private struct FlowLayout: Layout {
    var spacing: CGFloat = 8

    func sizeThatFits(proposal: ProposedViewSize, subviews: Subviews, cache: inout ()) -> CGSize {
        let width = proposal.width ?? .infinity
        var x: CGFloat = 0, y: CGFloat = 0, lineHeight: CGFloat = 0
        for sub in subviews {
            let s = sub.sizeThatFits(.unspecified)
            if x + s.width > width { x = 0; y += lineHeight + spacing; lineHeight = 0 }
            x += s.width + spacing
            lineHeight = max(lineHeight, s.height)
        }
        return CGSize(width: width, height: y + lineHeight)
    }

    func placeSubviews(in bounds: CGRect, proposal: ProposedViewSize, subviews: Subviews, cache: inout ()) {
        var x: CGFloat = bounds.minX, y: CGFloat = bounds.minY, lineHeight: CGFloat = 0
        for sub in subviews {
            let s = sub.sizeThatFits(.unspecified)
            if x + s.width > bounds.maxX { x = bounds.minX; y += lineHeight + spacing; lineHeight = 0 }
            sub.place(at: CGPoint(x: x, y: y), proposal: ProposedViewSize(s))
            x += s.width + spacing
            lineHeight = max(lineHeight, s.height)
        }
    }
}

// ─── screen 4: reveal ───

private struct RevealScreen: View {
    @Bindable var model: OnboardingViewModel
    let onContinue: () -> Void
    let onBack: () -> Void

    var body: some View {
        VStack(alignment: .leading, spacing: 0) {
            OnboardingHeader(onBack: onBack, onSkip: nil)

            if model.loading {
                loadingState
            } else {
                resultsState
            }
        }
    }

    private var loadingState: some View {
        VStack(alignment: .leading, spacing: 22) {
            VStack(alignment: .leading, spacing: 10) {
                Text("Finding things you'll care about...")
                    .font(Theme.title2()).foregroundStyle(Theme.label1)
                ProgressView().progressViewStyle(.circular).tint(Theme.accent)
            }
            .padding(.horizontal, 22).padding(.top, 32)

            VStack(spacing: 12) {
                ForEach(0..<5, id: \.self) { _ in shimmerCard }
            }
            .padding(.horizontal, 22)

            Spacer()
        }
    }

    private var shimmerCard: some View {
        RoundedRectangle(cornerRadius: Theme.rMed)
            .fill(Theme.surface)
            .frame(height: 96)
            .overlay(
                RoundedRectangle(cornerRadius: Theme.rMed)
                    .stroke(Theme.stroke, lineWidth: 0.5))
            .shimmer()
    }

    private var resultsState: some View {
        VStack(alignment: .leading, spacing: 0) {
            VStack(alignment: .leading, spacing: 10) {
                Text("Here's what we'd watch for you")
                    .font(Theme.title2()).foregroundStyle(Theme.label1)
                Text("Tap any to start watching. You can always add more later.")
                    .font(Theme.body()).foregroundStyle(Theme.label2)
                if model.fallback {
                    Text("Showing popular watchers — couldn't personalize this time.")
                        .font(Theme.caption()).foregroundStyle(Theme.warn)
                        .padding(.top, 4)
                }
            }
            .padding(.horizontal, 22).padding(.top, 24)

            ScrollView {
                VStack(spacing: 12) {
                    ForEach(Array(model.suggestions.enumerated()), id: \.element.id) { idx, s in
                        SuggestionCard(
                            suggestion: s,
                            active: model.isActivated(s),
                            inFlight: model.inFlight.contains(s.id)
                        ) {
                            if model.isActivated(s) {
                                Task { await model.deactivate(s) }
                            } else {
                                Task { await model.activate(s) }
                            }
                        }
                        .transition(.asymmetric(
                            insertion: .opacity.combined(with: .move(edge: .bottom)),
                            removal: .opacity))
                        .animation(.spring(response: 0.4, dampingFraction: 0.85)
                                    .delay(Double(idx) * 0.06), value: model.suggestions.count)
                    }
                }
                .padding(.horizontal, 22).padding(.top, 18).padding(.bottom, 16)
            }

            Button(action: onContinue) {
                Text(model.activatedCount == 0
                     ? "Skip for now"
                     : "Continue with \(model.activatedCount) watcher\(model.activatedCount == 1 ? "" : "s")")
                    .font(.system(size: 15, weight: .semibold))
                    .foregroundStyle(model.activatedCount == 0 ? Theme.label1 : Theme.accentInk)
                    .frame(maxWidth: .infinity, minHeight: 50)
                    .background(Capsule().fill(model.activatedCount == 0 ? Theme.surface : Theme.accent))
                    .overlay(Capsule().stroke(Theme.stroke, lineWidth: 0.5))
            }
            .buttonStyle(PressableStyle())
            .padding(.horizontal, 22).padding(.bottom, 26).padding(.top, 12)
        }
    }
}

private struct SuggestionCard: View {
    let suggestion: OnboardingSuggestion
    let active: Bool
    let inFlight: Bool
    let onTap: () -> Void

    var body: some View {
        Button(action: onTap) {
            VStack(alignment: .leading, spacing: 10) {
                HStack(alignment: .top) {
                    Text(suggestion.query)
                        .font(Theme.title3())
                        .foregroundStyle(active ? Theme.accentInk : Theme.label1)
                        .multilineTextAlignment(.leading)
                    Spacer()
                    if inFlight {
                        ProgressView().tint(active ? Theme.accentInk : Theme.accent)
                    } else if active {
                        HStack(spacing: 4) {
                            Image(systemName: "antenna.radiowaves.left.and.right")
                                .font(.system(size: 11, weight: .bold))
                            Text("Watching").font(.system(size: 12, weight: .semibold))
                        }
                        .foregroundStyle(Theme.accentInk)
                    } else {
                        Text("Tap to watch")
                            .font(.system(size: 12, weight: .semibold))
                            .foregroundStyle(Theme.label3)
                    }
                }

                HStack(spacing: 6) {
                    chip(text: suggestion.type == .event ? "Event" : "News",
                         icon: suggestion.type == .event ? "calendar" : "newspaper")
                    chip(text: cadenceLabel(suggestion.cadenceSeconds), icon: "clock")
                }

                Text(suggestion.reason)
                    .font(Theme.caption())
                    .foregroundStyle(active ? Theme.accentInk.opacity(0.75) : Theme.label2)
                    .lineLimit(2)
            }
            .padding(14)
            .frame(maxWidth: .infinity, alignment: .leading)
            .background(
                RoundedRectangle(cornerRadius: Theme.rMed)
                    .fill(active ? Theme.accent : Theme.surface))
            .overlay(
                RoundedRectangle(cornerRadius: Theme.rMed)
                    .stroke(active ? Theme.accent : Theme.stroke, lineWidth: active ? 1 : 0.5))
            .scaleEffect(active ? 1.02 : 1.0)
            .animation(.spring(response: 0.25, dampingFraction: 0.7), value: active)
        }
        .buttonStyle(.plain)
    }

    private func chip(text: String, icon: String) -> some View {
        HStack(spacing: 4) {
            Image(systemName: icon).font(.system(size: 10, weight: .semibold))
            Text(text).font(.system(size: 11, weight: .semibold))
        }
        .padding(.horizontal, 8).padding(.vertical, 4)
        .background(Capsule().fill(active ? Color.white.opacity(0.18) : Theme.surfaceHi))
        .foregroundStyle(active ? Theme.accentInk : Theme.label2)
    }

    private func cadenceLabel(_ s: Int) -> String {
        switch s {
        case 3600:  "1h"
        case 21600: "6h"
        case 86400: "Daily"
        default:    "\(s/60)m"
        }
    }
}

// ─── shimmer + toast helpers (lightweight, local to this file) ───

private struct ShimmerModifier: ViewModifier {
    @State private var phase: CGFloat = -1
    func body(content: Content) -> some View {
        content.overlay(
            LinearGradient(
                colors: [.clear, Theme.label3.opacity(0.25), .clear],
                startPoint: .leading, endPoint: .trailing)
            .offset(x: phase * 200)
            .mask(content)
        )
        .onAppear {
            withAnimation(.linear(duration: 1.4).repeatForever(autoreverses: false)) {
                phase = 1
            }
        }
    }
}
private extension View {
    func shimmer() -> some View { modifier(ShimmerModifier()) }

    /// Inline toast for onboarding errors. Doesn't share the global ToastBanner
    /// because it's bound to model.error rather than AppState.lastError.
    func toast(message: Binding<String?>) -> some View {
        overlay(alignment: .top) {
            if let m = message.wrappedValue {
                Text(m)
                    .font(Theme.caption())
                    .foregroundStyle(Theme.label1)
                    .padding(.horizontal, 14).padding(.vertical, 10)
                    .background(Capsule().fill(Theme.danger.opacity(0.9)))
                    .padding(.top, 50)
                    .transition(.move(edge: .top).combined(with: .opacity))
                    .onAppear {
                        DispatchQueue.main.asyncAfter(deadline: .now() + 2.5) {
                            message.wrappedValue = nil
                        }
                    }
            }
        }
    }
}
