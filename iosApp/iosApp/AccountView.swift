import SwiftUI
import UserNotifications

/// Settings tab — device identity, agent stats, debug actions.
/// "Send test notification" schedules a local `UNNotificationRequest` to
/// verify the banner UI and permission state without an APNs round-trip.
struct AccountView: View {
    @Environment(AppState.self) private var state
    @State private var showingResetConfirm = false

    var body: some View {
        ScrollView {
            VStack(alignment: .leading, spacing: 22) {
                header.padding(.horizontal, 22).padding(.top, 8)
                deviceCard.padding(.horizontal, 22)
                statsRow.padding(.horizontal, 22)
                groups.padding(.horizontal, 22)
            }
            .padding(.bottom, 24)
        }
        .background(Theme.bg.ignoresSafeArea())
        .scrollIndicators(.hidden)
        .alert("Reset device?", isPresented: $showingResetConfirm) {
            Button("Reset", role: .destructive) {
                state.api.deviceId = nil
                Task {
                    await state.bootstrap()
                    Haptics.warning()
                }
            }
            Button("Cancel", role: .cancel) {}
        } message: {
            Text("Forgets the local device id. Subscriptions on the server stay attached to it but become inaccessible from this device.")
        }
    }

    private var header: some View {
        VStack(alignment: .leading, spacing: 6) {
            Text("Account").font(Theme.title1()).tracking(-0.4).foregroundStyle(Theme.label1)
        }
    }

    private var deviceCard: some View {
        VStack(alignment: .leading, spacing: 12) {
            HStack(spacing: 12) {
                ZStack {
                    Circle().fill(Theme.accentSoft).frame(width: 44, height: 44)
                    Image(systemName: "iphone.gen3.radiowaves.left.and.right")
                        .foregroundStyle(Theme.accent).font(.system(size: 18, weight: .semibold))
                }
                VStack(alignment: .leading, spacing: 2) {
                    Text("This device").font(Theme.bodyMed()).foregroundStyle(Theme.label1)
                    Text(state.api.deviceId ?? "—")
                        .font(.system(size: 12, weight: .regular).monospaced())
                        .foregroundStyle(Theme.label3).lineLimit(1)
                }
                Spacer()
            }
        }
        .padding(16)
        .background(RoundedRectangle(cornerRadius: Theme.rLg).fill(Theme.surface))
        .overlay(RoundedRectangle(cornerRadius: Theme.rLg).stroke(Theme.stroke, lineWidth: 0.5))
    }

    private var statsRow: some View {
        HStack(spacing: 10) {
            statTile("Watching",  "\(state.activeSubscriptions.count)",  icon: "eye.fill")
            statTile("Resolved",  "\(state.resolvedSubscriptions.count)", icon: "checkmark.seal.fill")
            statTile("Signals",   "\(state.allSignalsRecent.count)",     icon: "waveform.path")
        }
    }
    private func statTile(_ label: String, _ value: String, icon: String) -> some View {
        VStack(alignment: .leading, spacing: 8) {
            Image(systemName: icon).foregroundStyle(Theme.accent)
                .font(.system(size: 14, weight: .semibold))
            Text(value).font(.system(size: 22, weight: .semibold).monospacedDigit())
                .foregroundStyle(Theme.label1)
            Text(label.uppercased()).font(Theme.eyebrow()).foregroundStyle(Theme.label3)
        }
        .padding(14).frame(maxWidth: .infinity, alignment: .leading)
        .background(RoundedRectangle(cornerRadius: Theme.rMed).fill(Theme.surface))
        .overlay(RoundedRectangle(cornerRadius: Theme.rMed).stroke(Theme.stroke, lineWidth: 0.5))
    }

    private var groups: some View {
        VStack(spacing: 22) {
            settingsGroup("Notifications") {
                row("Send test notification", icon: "bell.badge", accent: true) {
                    Haptics.tapMedium()
                    Task { await scheduleLocalTestNotification(state: state) }
                }
                row("Request permission again", icon: "lock.shield") {
                    Task {
                        let granted = await PushService.shared.requestAuthorization()
                        state.toast = granted ? "Notifications authorized." : "Permission denied — open Settings."
                    }
                }
            }

            settingsGroup("Agent") {
                row("Run all watchers now", icon: "arrow.triangle.2.circlepath") {
                    Haptics.tapMedium()
                    Task {
                        for sub in state.activeSubscriptions { await state.run(sub) }
                    }
                }
                row("Refresh from server", icon: "arrow.clockwise") {
                    Task { await state.refresh() }
                }
            }

            settingsGroup("Danger zone") {
                row("Forget this device", icon: "trash", danger: true) {
                    showingResetConfirm = true
                }
            }
        }
    }

    private func settingsGroup<Content: View>(_ title: String, @ViewBuilder _ content: () -> Content) -> some View {
        VStack(alignment: .leading, spacing: 8) {
            Text(title.uppercased()).font(Theme.eyebrow()).foregroundStyle(Theme.label3)
                .padding(.leading, 4)
            VStack(spacing: 1) { content() }
                .clipShape(RoundedRectangle(cornerRadius: Theme.rMed))
                .overlay(RoundedRectangle(cornerRadius: Theme.rMed).stroke(Theme.stroke, lineWidth: 0.5))
        }
    }

    private func row(_ title: String, icon: String, accent: Bool = false, danger: Bool = false, action: @escaping () -> Void) -> some View {
        Button(action: action) {
            HStack(spacing: 12) {
                Image(systemName: icon)
                    .frame(width: 22)
                    .foregroundStyle(danger ? Theme.danger : (accent ? Theme.accent : Theme.label2))
                Text(title)
                    .foregroundStyle(danger ? Theme.danger : Theme.label1)
                    .font(Theme.body())
                Spacer()
                Image(systemName: "chevron.right")
                    .font(.system(size: 12, weight: .semibold))
                    .foregroundStyle(Theme.label4)
            }
            .padding(.horizontal, 16).padding(.vertical, 14)
            .background(Theme.surface)
        }
        .buttonStyle(.plain)
    }
}

@MainActor
private func scheduleLocalTestNotification(state: AppState) async {
    let center = UNUserNotificationCenter.current()
    let settings = await center.notificationSettings()
    switch settings.authorizationStatus {
    case .denied:
        state.lastError = "Notifications denied — enable in Settings → notify."
        return
    case .notDetermined:
        let granted = (try? await center.requestAuthorization(options: [.alert, .sound, .badge])) ?? false
        if !granted { state.lastError = "Permission not granted."; return }
    case .authorized, .provisional, .ephemeral:
        break
    @unknown default:
        state.lastError = "Unknown permission state."
        return
    }
    let content = UNMutableNotificationContent()
    content.title = "Test alert"
    content.body = "If you can see this, local notifications work."
    content.sound = .default
    let trigger = UNTimeIntervalNotificationTrigger(timeInterval: 0.6, repeats: false)
    let req = UNNotificationRequest(identifier: "test-\(UUID().uuidString)", content: content, trigger: trigger)
    do {
        try await center.add(req)
        state.toast = "Test notification scheduled — leave foreground or wait."
    } catch {
        state.lastError = "Schedule failed: \(error.localizedDescription)"
    }
}
