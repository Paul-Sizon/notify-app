package com.notify.anything.notify.ui

import com.notify.anything.notify.ui.NotifyIcons

import androidx.compose.foundation.background
import androidx.compose.foundation.border
import androidx.compose.foundation.clickable
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.PaddingValues
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.Spacer
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.shape.CircleShape
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material3.AlertDialog
import androidx.compose.material3.Icon
import androidx.compose.material3.Text
import androidx.compose.material3.TextButton
import androidx.compose.runtime.Composable
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.setValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.text.font.FontFamily
import androidx.compose.ui.unit.dp
import com.notify.anything.notify.platform.Notifier

@Composable
fun AccountScreen(state: AppState, notifier: Notifier) {
    var showResetConfirm by remember { mutableStateOf(false) }

    LazyColumn(
        modifier = Modifier.fillMaxSize().background(NotifyColors.bg),
        contentPadding = PaddingValues(top = 8.dp, bottom = 120.dp),
    ) {
        item {
            Text(
                "Account",
                style = NotifyType.largeTitle, color = NotifyColors.label1,
                modifier = Modifier.padding(horizontal = 20.dp).padding(top = 24.dp, bottom = 24.dp),
            )
        }
        item { DeviceCard(state, modifier = Modifier.padding(horizontal = 22.dp).padding(bottom = 16.dp)) }
        item { StatsRow(state, modifier = Modifier.padding(horizontal = 22.dp).padding(bottom = 16.dp)) }

        item {
            SettingsGroup("Notifications", modifier = Modifier.padding(horizontal = 22.dp).padding(bottom = 16.dp)) {
                SettingsRow("Send test notification", icon = NotifyIcons.Notifications, accent = true) {
                    notifier.deliver(
                        title = "Test alert",
                        body = "If you see this, notifications work.",
                        subscriptionId = "debug",
                        signalId = "debug-${kotlinx.datetime.Clock.System.now().toEpochMilliseconds()}",
                    )
                    state.toast = "Test notification scheduled."
                }
            }
        }

        item {
            SettingsGroup("Agent", modifier = Modifier.padding(horizontal = 22.dp).padding(bottom = 16.dp)) {
                SettingsRow("Run all watchers now", icon = NotifyIcons.Sync) {
                    for (sub in state.activeSubscriptions) state.run(sub)
                }
                SettingsRow("Refresh from server", icon = NotifyIcons.Refresh) {
                    state.refresh()
                }
            }
        }

        item {
            SettingsGroup("Danger zone", modifier = Modifier.padding(horizontal = 22.dp).padding(bottom = 16.dp)) {
                SettingsRow("Forget this device", icon = NotifyIcons.DeleteForever, danger = true) {
                    showResetConfirm = true
                }
            }
        }
    }

    if (showResetConfirm) {
        AlertDialog(
            onDismissRequest = { showResetConfirm = false },
            title = { Text("Reset device?") },
            text = {
                Text("Forgets the local device id. Subscriptions on the server stay attached but become inaccessible from this device.")
            },
            confirmButton = {
                TextButton(onClick = {
                    showResetConfirm = false
                    state.forgetDevice()
                }) { Text("Reset", color = NotifyColors.danger) }
            },
            dismissButton = { TextButton(onClick = { showResetConfirm = false }) { Text("Cancel") } },
            containerColor = NotifyColors.bgElevated,
            titleContentColor = NotifyColors.label1,
            textContentColor = NotifyColors.label2,
        )
    }

}

@Composable
private fun DeviceCard(state: AppState, modifier: Modifier = Modifier) {
    Row(
        modifier = modifier
            .fillMaxWidth()
            .clip(RoundedCornerShape(20.dp))
            .background(NotifyColors.surface)
            .border(0.5.dp, NotifyColors.stroke, RoundedCornerShape(20.dp))
            .padding(16.dp),
        verticalAlignment = Alignment.CenterVertically,
        horizontalArrangement = Arrangement.spacedBy(12.dp),
    ) {
        Box(
            modifier = Modifier.size(44.dp).clip(CircleShape).background(NotifyColors.accentSoft),
            contentAlignment = Alignment.Center,
        ) {
            Icon(NotifyIcons.PhoneAndroid, null, tint = NotifyColors.accent, modifier = Modifier.size(20.dp))
        }
        Column(modifier = Modifier.weight(1f)) {
            Text("This device", style = NotifyType.bodyMed, color = NotifyColors.label1)
            Text(
                state.deviceId ?: "—",
                style = NotifyType.caption.copy(fontFamily = FontFamily.Monospace),
                color = NotifyColors.label3,
                maxLines = 1,
            )
        }
    }
}

@Composable
private fun StatsRow(state: AppState, modifier: Modifier = Modifier) {
    Row(
        modifier = modifier.fillMaxWidth(),
        horizontalArrangement = Arrangement.spacedBy(10.dp),
    ) {
        StatTile("Watching", "${state.activeSubscriptions.size}", NotifyIcons.Visibility, Modifier.weight(1f))
        StatTile("Resolved", "${state.resolvedSubscriptions.size}", NotifyIcons.CheckCircle, Modifier.weight(1f))
        StatTile("Signals", "${state.allSignalsRecent.size}", NotifyIcons.GraphicEq, Modifier.weight(1f))
    }
}

@Composable
private fun StatTile(label: String, value: String, icon: androidx.compose.ui.graphics.vector.ImageVector, modifier: Modifier = Modifier) {
    Column(
        modifier = modifier
            .clip(RoundedCornerShape(14.dp))
            .background(NotifyColors.surface)
            .border(0.5.dp, NotifyColors.stroke, RoundedCornerShape(14.dp))
            .padding(14.dp),
        verticalArrangement = Arrangement.spacedBy(8.dp),
    ) {
        Icon(icon, null, tint = NotifyColors.accent, modifier = Modifier.size(16.dp))
        Text(
            value,
            style = NotifyType.title2.copy(fontFamily = FontFamily.Monospace),
            color = NotifyColors.label1,
        )
        Text(label.uppercase(), style = NotifyType.eyebrow, color = NotifyColors.label3)
    }
}

@Composable
private fun SettingsGroup(title: String, modifier: Modifier = Modifier, content: @Composable () -> Unit) {
    Column(modifier = modifier, verticalArrangement = Arrangement.spacedBy(8.dp)) {
        Text(
            title.uppercase(),
            style = NotifyType.eyebrow,
            color = NotifyColors.label3,
            modifier = Modifier.padding(start = 4.dp),
        )
        Column(
            modifier = Modifier
                .clip(RoundedCornerShape(14.dp))
                .border(0.5.dp, NotifyColors.stroke, RoundedCornerShape(14.dp)),
        ) {
            content()
        }
    }
}

@Composable
private fun SettingsRow(
    title: String,
    icon: androidx.compose.ui.graphics.vector.ImageVector,
    subtitle: String? = null,
    accent: Boolean = false,
    danger: Boolean = false,
    onClick: () -> Unit,
) {
    val tint = when {
        danger -> NotifyColors.danger
        accent -> NotifyColors.accent
        else -> NotifyColors.label2
    }
    val titleColor = when {
        danger -> NotifyColors.danger
        else -> NotifyColors.label1
    }
    Row(
        modifier = Modifier
            .fillMaxWidth()
            .background(NotifyColors.surface)
            .clickable { onClick() }
            .padding(horizontal = 16.dp, vertical = 14.dp),
        verticalAlignment = Alignment.CenterVertically,
        horizontalArrangement = Arrangement.spacedBy(12.dp),
    ) {
        Icon(icon, null, tint = tint, modifier = Modifier.size(22.dp))
        Column(modifier = Modifier.weight(1f)) {
            Text(title, style = NotifyType.body, color = titleColor)
            if (subtitle != null) {
                Text(
                    subtitle,
                    style = NotifyType.caption.copy(fontFamily = FontFamily.Monospace),
                    color = NotifyColors.label3,
                    maxLines = 1,
                )
            }
        }
        Icon(NotifyIcons.ChevronRight, null, tint = NotifyColors.label4, modifier = Modifier.size(14.dp))
    }
}
