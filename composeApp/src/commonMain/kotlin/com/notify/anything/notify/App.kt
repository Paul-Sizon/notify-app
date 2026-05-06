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
 * Bottom nav per Claude Design handoff: flat, no pill, no border. The
 * surface is a vertical gradient that fades from transparent at the top
 * down to the page background — content underneath the bar scrolls
 * softly out of view rather than getting cropped by a hard divider.
 */
@Composable
private fun BottomTabs(active: Tab, onChange: (Tab) -> Unit) {
    val tabs = listOf(
        Tab.WATCHERS to ("Watchers" to TabKind.WATCHERS),
        Tab.ALERTS to ("Alerts" to TabKind.ALERTS),
        Tab.SIGNALS to ("Signals" to TabKind.SIGNALS),
        Tab.ACCOUNT to ("Account" to TabKind.ACCOUNT),
    )
    Box(
        modifier = Modifier
            .fillMaxWidth()
            .background(
                androidx.compose.ui.graphics.Brush.verticalGradient(
                    0.0f to Color.Transparent,
                    0.30f to NotifyColors.bg.copy(alpha = 0.85f),
                    0.70f to NotifyColors.bg,
                    1.0f to NotifyColors.bg,
                ),
            )
            .padding(top = 18.dp)
            .navigationBarsPadding()
            .padding(bottom = 14.dp),
    ) {
        Row(
            modifier = Modifier.fillMaxWidth().padding(horizontal = 8.dp),
            horizontalArrangement = Arrangement.SpaceBetween,
            verticalAlignment = Alignment.CenterVertically,
        ) {
            for ((tab, label) in tabs) {
                Box(modifier = Modifier.weight(1f), contentAlignment = Alignment.Center) {
                    TabCell(active == tab, label.first, label.second) { onChange(tab) }
                }
            }
        }
    }
}

private enum class TabKind { WATCHERS, ALERTS, SIGNALS, ACCOUNT }

@Composable
private fun TabCell(
    isActive: Boolean,
    label: String,
    kind: TabKind,
    onTap: () -> Unit,
) {
    // Inactive tabs sit at 0.98x to give the active one a subtle pop without
    // any backdrop chip — keeps the design's flat look.
    val cellScale by animateFloatAsState(
        targetValue = if (isActive) 1f else 0.98f,
        animationSpec = spring(Spring.DampingRatioMediumBouncy, Spring.StiffnessMedium),
        label = "tab-cell",
    )
    val tint by animateColor(
        if (isActive) NotifyColors.accent else NotifyColors.label3,
    )
    Column(
        horizontalAlignment = Alignment.CenterHorizontally,
        verticalArrangement = Arrangement.spacedBy(4.dp),
        modifier = Modifier
            .scale(cellScale)
            .clip(RoundedCornerShape(12.dp))
            .softClick(pressScale = 0.92f, haptic = { selection() }) { onTap() }
            .padding(horizontal = 4.dp, vertical = 8.dp),
    ) {
        TabGlyph(kind, active = isActive, color = tint)
        Text(
            label,
            style = androidx.compose.ui.text.TextStyle(
                fontSize = 11.sp,
                fontWeight = androidx.compose.ui.text.font.FontWeight.SemiBold,
                letterSpacing = 0.1.sp,
            ),
            color = tint,
        )
    }
}

/**
 * Hand-drawn line/fill tab icons matching the design's SVG paths in
 * `tab-bar.jsx`. Compose-Multiplatform 1.10 doesn't ship outlined icon
 * variants, so we render the four glyphs ourselves with `Canvas` —
 * cheaper than dragging in a 3000-icon extended set just for these.
 */
@Composable
private fun TabGlyph(kind: TabKind, active: Boolean, color: Color) {
    val size = 22.dp
    androidx.compose.foundation.Canvas(modifier = Modifier.size(size)) {
        val w = this.size.width
        val h = this.size.height
        val sx = w / 24f
        val sy = h / 24f
        val stroke = androidx.compose.ui.graphics.drawscope.Stroke(
            width = 1.7f * sx,
            cap = androidx.compose.ui.graphics.StrokeCap.Round,
            join = androidx.compose.ui.graphics.StrokeJoin.Round,
        )
        when (kind) {
            TabKind.WATCHERS -> {
                // Eye outline + center pupil (filled when active)
                val eye = androidx.compose.ui.graphics.Path().apply {
                    moveTo(2f * sx, 12f * sy)
                    cubicTo(2f * sx, 12f * sy, 5.5f * sx, 5f * sy, 12f * sx, 5f * sy)
                    cubicTo(18.5f * sx, 5f * sy, 22f * sx, 12f * sy, 22f * sx, 12f * sy)
                    cubicTo(22f * sx, 12f * sy, 18.5f * sx, 19f * sy, 12f * sx, 19f * sy)
                    cubicTo(5.5f * sx, 19f * sy, 2f * sx, 12f * sy, 2f * sx, 12f * sy)
                    close()
                }
                if (active) drawPath(eye, color = color.copy(alpha = 0.18f))
                drawPath(eye, color = color, style = stroke)
                if (active) {
                    drawCircle(color, radius = 3f * sx, center = androidx.compose.ui.geometry.Offset(12f * sx, 12f * sy))
                } else {
                    drawCircle(color, radius = 3f * sx, center = androidx.compose.ui.geometry.Offset(12f * sx, 12f * sy), style = stroke)
                }
            }
            TabKind.ALERTS -> {
                // Bell outline (filled when active)
                val bell = androidx.compose.ui.graphics.Path().apply {
                    moveTo(18f * sx, 16f * sy)
                    lineTo(18f * sx, 11f * sy)
                    cubicTo(18f * sx, 7.7f * sy, 15.3f * sx, 5f * sy, 12f * sx, 5f * sy)
                    cubicTo(8.7f * sx, 5f * sy, 6f * sx, 7.7f * sy, 6f * sx, 11f * sy)
                    lineTo(6f * sx, 16f * sy)
                    lineTo(4f * sx, 19f * sy)
                    lineTo(20f * sx, 19f * sy)
                    close()
                }
                if (active) drawPath(bell, color = color.copy(alpha = 0.18f))
                drawPath(bell, color = color, style = stroke)
                // Clapper
                val clapper = androidx.compose.ui.graphics.Path().apply {
                    moveTo(10f * sx, 21f * sy)
                    cubicTo(10f * sx, 22.1f * sy, 10.9f * sx, 23f * sy, 12f * sx, 23f * sy)
                    cubicTo(13.1f * sx, 23f * sy, 14f * sx, 22.1f * sy, 14f * sx, 21f * sy)
                }
                drawPath(clapper, color = color, style = stroke)
            }
            TabKind.SIGNALS -> {
                // Spark line: zigzag with a dot at top-right
                val spark = androidx.compose.ui.graphics.Path().apply {
                    moveTo(3f * sx, 17f * sy)
                    lineTo(8f * sx, 11f * sy)
                    lineTo(12f * sx, 14f * sy)
                    lineTo(17f * sx, 7f * sy)
                    lineTo(21f * sx, 12f * sy)
                }
                drawPath(spark, color = color, style = stroke)
                drawCircle(color, radius = 1.4f * sx, center = androidx.compose.ui.geometry.Offset(17f * sx, 7f * sy))
            }
            TabKind.ACCOUNT -> {
                // Person: head circle + shoulders arc
                if (active) {
                    drawCircle(
                        color = color.copy(alpha = 0.18f),
                        radius = 4f * sx,
                        center = androidx.compose.ui.geometry.Offset(12f * sx, 9f * sy),
                    )
                }
                drawCircle(
                    color = color,
                    radius = 4f * sx,
                    center = androidx.compose.ui.geometry.Offset(12f * sx, 9f * sy),
                    style = stroke,
                )
                val shoulders = androidx.compose.ui.graphics.Path().apply {
                    moveTo(4f * sx, 21f * sy)
                    cubicTo(4f * sx, 17f * sy, 8f * sx, 15f * sy, 12f * sx, 15f * sy)
                    cubicTo(16f * sx, 15f * sy, 20f * sx, 17f * sy, 20f * sx, 21f * sy)
                }
                drawPath(shoulders, color = color, style = stroke)
            }
        }
    }
}

@Composable
private fun animateColor(target: Color) =
    androidx.compose.animation.animateColorAsState(
        targetValue = target,
        animationSpec = tween(durationMillis = 220),
        label = "color",
    )
