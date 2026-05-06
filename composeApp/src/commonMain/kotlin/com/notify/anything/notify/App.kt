package com.notify.anything.notify

import com.notify.anything.notify.ui.NotifyIcons

import androidx.compose.animation.AnimatedVisibility
import androidx.compose.animation.fadeIn
import androidx.compose.animation.fadeOut
import androidx.compose.foundation.background
import androidx.compose.foundation.border
import androidx.compose.foundation.clickable
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.Spacer
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.navigationBarsPadding
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.layout.statusBarsPadding
import androidx.compose.foundation.shape.CircleShape
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material3.Icon
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.remember
import androidx.compose.runtime.rememberCoroutineScope
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.graphics.vector.ImageVector
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import com.notify.anything.notify.platform.Notifier
import com.notify.anything.notify.platform.Prefs
import com.notify.anything.notify.platform.UrlOpener
import com.notify.anything.notify.ui.AccountScreen
import com.notify.anything.notify.ui.AddSubscriptionSheet
import com.notify.anything.notify.ui.AlertsScreen
import com.notify.anything.notify.ui.AppState
import com.notify.anything.notify.ui.LiveAgentOverlay
import com.notify.anything.notify.ui.NotifyColors
import com.notify.anything.notify.ui.NotifyFab
import com.notify.anything.notify.ui.NotifyTheme
import com.notify.anything.notify.ui.NotifyType
import com.notify.anything.notify.ui.SignalDetailSheet
import com.notify.anything.notify.ui.SignalsScreen
import com.notify.anything.notify.ui.Tab
import com.notify.anything.notify.ui.Toast
import com.notify.anything.notify.ui.WatchersScreen

@Composable
fun App(
    notifier: Notifier,
    prefs: Prefs,
    urlOpener: UrlOpener,
    initialBaseUrl: String,
) {
    NotifyTheme {
        val scope = rememberCoroutineScope()
        val state = remember { AppState(scope, prefs, notifier, initialBaseUrl) }

        LaunchedEffect(Unit) { state.bootstrap() }

        Box(modifier = Modifier.fillMaxSize().background(NotifyColors.bg)) {
            Box(
                modifier = Modifier
                    .fillMaxSize()
                    .statusBarsPadding(),
            ) {
                when (state.selectedTab) {
                    Tab.WATCHERS -> WatchersScreen(state)
                    Tab.ALERTS -> AlertsScreen(state)
                    Tab.SIGNALS -> SignalsScreen(state)
                    Tab.ACCOUNT -> AccountScreen(state, notifier)
                }
            }

            // Toasts (top)
            Column(
                modifier = Modifier.fillMaxWidth().statusBarsPadding().padding(top = 12.dp),
                verticalArrangement = Arrangement.spacedBy(8.dp),
            ) {
                Toast(state.lastError, onClear = { state.lastError = null }, isError = true)
                Toast(state.toast, onClear = { state.toast = null }, isError = false)
            }

            // FAB (Watchers tab, when not empty)
            AnimatedVisibility(
                visible = state.selectedTab == Tab.WATCHERS && state.subscriptions.isNotEmpty(),
                enter = fadeIn(),
                exit = fadeOut(),
                modifier = Modifier
                    .align(Alignment.BottomEnd)
                    .padding(end = 22.dp, bottom = 102.dp)
                    .navigationBarsPadding(),
            ) {
                NotifyFab(onClick = { state.showAddSheet = true })
            }

            // Bottom tab bar
            Box(modifier = Modifier.align(Alignment.BottomCenter)) {
                BottomTabs(state.selectedTab) { tab -> state.selectedTab = tab }
            }

            // Live agent overlay
            if (state.liveAgentSubId != null) {
                val sub = state.subscriptions.firstOrNull { it.id == state.liveAgentSubId }
                if (sub != null) {
                    LiveAgentOverlay(sub.query)
                }
            }
        }

        // Sheets
        if (state.showAddSheet) {
            AddSubscriptionSheet(
                onDismiss = { state.showAddSheet = false },
                onCreate = { q, t, c -> state.create(q, t, c) },
            )
        }
        val detailSub = state.subscriptions.firstOrNull { it.id == state.detailSubscriptionId }
        if (detailSub != null) {
            SignalDetailSheet(
                state = state,
                subscription = detailSub,
                urlOpener = urlOpener,
                onDismiss = { state.detailSubscriptionId = null },
            )
        }
    }
}

@Composable
private fun BottomTabs(active: Tab, onChange: (Tab) -> Unit) {
    Row(
        modifier = Modifier
            .fillMaxWidth()
            .background(NotifyColors.surface.copy(alpha = 0.96f))
            .border(0.5.dp, NotifyColors.strokeHi, RoundedCornerShape(0.dp))
            .padding(top = 10.dp)
            .navigationBarsPadding()
            .padding(bottom = 8.dp)
            .height(74.dp),
        horizontalArrangement = Arrangement.SpaceEvenly,
        verticalAlignment = Alignment.CenterVertically,
    ) {
        TabCell(active == Tab.WATCHERS, "Watchers", NotifyIcons.Visibility, NotifyIcons.Visibility) { onChange(Tab.WATCHERS) }
        TabCell(active == Tab.ALERTS, "Alerts", NotifyIcons.Notifications, NotifyIcons.Notifications) { onChange(Tab.ALERTS) }
        TabCell(active == Tab.SIGNALS, "Signals", NotifyIcons.GraphicEq, NotifyIcons.GraphicEq) { onChange(Tab.SIGNALS) }
        TabCell(active == Tab.ACCOUNT, "Account", NotifyIcons.AccountCircle, NotifyIcons.AccountCircle) { onChange(Tab.ACCOUNT) }
    }
}

@Composable
private fun TabCell(
    isActive: Boolean,
    label: String,
    iconOutlined: ImageVector,
    iconFilled: ImageVector,
    onTap: () -> Unit,
) {
    Column(
        horizontalAlignment = Alignment.CenterHorizontally,
        verticalArrangement = Arrangement.spacedBy(4.dp),
        modifier = Modifier
            .clip(RoundedCornerShape(12.dp))
            .clickable { onTap() }
            .padding(horizontal = 14.dp, vertical = 4.dp),
    ) {
        Box(
            modifier = Modifier
                .size(width = 52.dp, height = 32.dp)
                .clip(CircleShape)
                .background(if (isActive) NotifyColors.accentSoft else androidx.compose.ui.graphics.Color.Transparent),
            contentAlignment = Alignment.Center,
        ) {
            Icon(
                if (isActive) iconFilled else iconOutlined, null,
                tint = if (isActive) NotifyColors.accent else NotifyColors.label2,
                modifier = Modifier.size(19.dp),
            )
        }
        Text(
            label,
            style = NotifyType.eyebrow.copy(fontSize = 10.sp, letterSpacing = 0.2.sp),
            color = if (isActive) NotifyColors.accent else NotifyColors.label3,
        )
    }
}

