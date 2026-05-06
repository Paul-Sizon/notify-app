package com.notify.anything.notify

import com.notify.anything.notify.ui.NotifyIcons

import androidx.compose.animation.AnimatedContent
import androidx.compose.animation.AnimatedVisibility
import androidx.compose.animation.core.Spring
import androidx.compose.animation.core.animateDpAsState
import androidx.compose.animation.core.animateFloatAsState
import androidx.compose.animation.core.spring
import androidx.compose.animation.core.tween
import androidx.compose.animation.fadeIn
import androidx.compose.animation.fadeOut
import androidx.compose.animation.scaleIn
import androidx.compose.animation.scaleOut
import androidx.compose.animation.slideInHorizontally
import androidx.compose.animation.slideOutHorizontally
import androidx.compose.animation.togetherWith
import androidx.compose.foundation.background
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.BoxWithConstraints
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.Spacer
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.navigationBarsPadding
import androidx.compose.foundation.layout.offset
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.layout.statusBarsPadding
import androidx.compose.foundation.layout.width
import androidx.compose.foundation.shape.CircleShape
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material3.Icon
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.getValue
import androidx.compose.runtime.remember
import androidx.compose.runtime.rememberCoroutineScope
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.draw.scale
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.graphics.vector.ImageVector
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import com.notify.anything.notify.platform.Haptics
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
import com.notify.anything.notify.ui.ProvideHaptics
import com.notify.anything.notify.ui.SignalDetailSheet
import com.notify.anything.notify.ui.SignalsScreen
import com.notify.anything.notify.ui.Tab
import com.notify.anything.notify.ui.Toast
import com.notify.anything.notify.ui.WatchersScreen
import com.notify.anything.notify.ui.softClick

@Composable
fun App(
    notifier: Notifier,
    prefs: Prefs,
    urlOpener: UrlOpener,
    haptics: Haptics,
    initialBaseUrl: String,
) {
    NotifyTheme {
        ProvideHaptics(haptics) {
            val scope = rememberCoroutineScope()
            val state = remember { AppState(scope, prefs, notifier, initialBaseUrl) }

            LaunchedEffect(Unit) { state.bootstrap() }

            Box(modifier = Modifier.fillMaxSize().background(NotifyColors.bg)) {
                // Animated screen swap. Direction follows tab order so movement
                // feels like sliding along a horizontal rail.
                AnimatedContent(
                    targetState = state.selectedTab,
                    transitionSpec = {
                        val forward = targetState.ordinal > initialState.ordinal
                        val dir = if (forward) 1 else -1
                        (slideInHorizontally(tween(280)) { (it * 0.18f * dir).toInt() } +
                            fadeIn(tween(280))) togetherWith
                            (slideOutHorizontally(tween(220)) { -(it * 0.12f * dir).toInt() } +
                                fadeOut(tween(180)))
                    },
                    label = "tab-screen",
                    modifier = Modifier.fillMaxSize().statusBarsPadding(),
                ) { tab ->
                    when (tab) {
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

                // FAB — fades + scales + spring on enter/exit
                AnimatedVisibility(
                    visible = state.selectedTab == Tab.WATCHERS && state.subscriptions.isNotEmpty(),
                    enter = scaleIn(spring(Spring.DampingRatioMediumBouncy, Spring.StiffnessMedium)) +
                        fadeIn(tween(200)),
                    exit = scaleOut(tween(180)) + fadeOut(tween(140)),
                    modifier = Modifier
                        .align(Alignment.BottomEnd)
                        .padding(end = 22.dp, bottom = 102.dp)
                        .navigationBarsPadding(),
                ) {
                    NotifyFab(
                        bumpKey = state.subscriptions.size,
                        onClick = { state.showAddSheet = true },
                    )
                }

                // Bottom tab bar
                Box(modifier = Modifier.align(Alignment.BottomCenter)) {
                    BottomTabs(state.selectedTab) { tab -> state.selectedTab = tab }
                }

                // Live agent overlay — fades in/out smoothly
                AnimatedVisibility(
                    visible = state.liveAgentSubId != null,
                    enter = fadeIn(tween(220)),
                    exit = fadeOut(tween(220)),
                ) {
                    val sub = state.subscriptions.firstOrNull { it.id == state.liveAgentSubId }
                    if (sub != null) LiveAgentOverlay(sub.query)
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
}

/**
 * Bottom nav with a single sliding pill indicator that animates between
 * the four tab cells. Avoids the visual stutter of "highlight off / new
 * highlight on" you get with per-cell background swaps.
 */
@Composable
private fun BottomTabs(active: Tab, onChange: (Tab) -> Unit) {
    val tabs = listOf(
        Tab.WATCHERS to ("Watchers" to NotifyIcons.Visibility),
        Tab.ALERTS to ("Alerts" to NotifyIcons.Notifications),
        Tab.SIGNALS to ("Signals" to NotifyIcons.GraphicEq),
        Tab.ACCOUNT to ("Account" to NotifyIcons.AccountCircle),
    )
    BoxWithConstraints(
        modifier = Modifier
            .fillMaxWidth()
            .background(NotifyColors.surface.copy(alpha = 0.96f))
            .padding(top = 10.dp)
            .navigationBarsPadding()
            .padding(bottom = 8.dp)
            .height(74.dp),
    ) {
        val cellWidth = maxWidth / tabs.size
        val activeIndex = tabs.indexOfFirst { it.first == active }.coerceAtLeast(0)
        val pillWidth = 52.dp
        val indicatorOffset by animateDpAsState(
            targetValue = cellWidth * activeIndex + (cellWidth - pillWidth) / 2,
            animationSpec = spring(Spring.DampingRatioMediumBouncy, Spring.StiffnessMediumLow),
            label = "tab-indicator",
        )
        Box(
            modifier = Modifier
                .offset(x = indicatorOffset, y = 0.dp)
                .padding(top = 4.dp)
                .size(width = pillWidth, height = 32.dp)
                .clip(CircleShape)
                .background(NotifyColors.accentSoft),
        )
        Row(
            modifier = Modifier.fillMaxSize(),
            horizontalArrangement = Arrangement.SpaceEvenly,
            verticalAlignment = Alignment.CenterVertically,
        ) {
            for ((tab, label) in tabs) {
                TabCell(active == tab, label.first, label.second) { onChange(tab) }
            }
        }
    }
}

@Composable
private fun TabCell(
    isActive: Boolean,
    label: String,
    icon: ImageVector,
    onTap: () -> Unit,
) {
    val iconScale by animateFloatAsState(
        targetValue = if (isActive) 1.12f else 1f,
        animationSpec = spring(Spring.DampingRatioMediumBouncy, Spring.StiffnessMedium),
        label = "tab-icon",
    )
    val tint by animateColor(
        if (isActive) NotifyColors.accent else NotifyColors.label2,
    )
    val textColor by animateColor(
        if (isActive) NotifyColors.accent else NotifyColors.label3,
    )
    Column(
        horizontalAlignment = Alignment.CenterHorizontally,
        verticalArrangement = Arrangement.spacedBy(4.dp),
        modifier = Modifier
            .clip(RoundedCornerShape(12.dp))
            .softClick(pressScale = 0.9f, haptic = { selection() }) { onTap() }
            .padding(horizontal = 14.dp, vertical = 4.dp),
    ) {
        Box(
            modifier = Modifier.size(width = 52.dp, height = 32.dp),
            contentAlignment = Alignment.Center,
        ) {
            Icon(
                icon, null,
                tint = tint,
                modifier = Modifier.size(19.dp).scale(iconScale),
            )
        }
        Text(
            label,
            style = NotifyType.eyebrow.copy(fontSize = 10.sp, letterSpacing = 0.2.sp),
            color = textColor,
        )
    }
}

@Composable
private fun animateColor(target: Color) =
    androidx.compose.animation.animateColorAsState(
        targetValue = target,
        animationSpec = tween(durationMillis = 220),
        label = "color",
    )
