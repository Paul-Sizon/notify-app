package com.notify.anything.notify.ui

import com.notify.anything.notify.ui.NotifyIcons

import androidx.compose.animation.AnimatedVisibility
import androidx.compose.animation.core.RepeatMode
import androidx.compose.animation.core.animateFloat
import androidx.compose.animation.core.animateFloatAsState
import androidx.compose.animation.core.infiniteRepeatable
import androidx.compose.animation.core.rememberInfiniteTransition
import androidx.compose.animation.core.tween
import androidx.compose.animation.fadeIn
import androidx.compose.animation.fadeOut
import androidx.compose.animation.slideInVertically
import androidx.compose.animation.slideOutVertically
import androidx.compose.foundation.ExperimentalFoundationApi
import androidx.compose.foundation.background
import androidx.compose.foundation.border
import androidx.compose.foundation.clickable
import androidx.compose.foundation.combinedClickable
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.Spacer
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.layout.widthIn
import androidx.compose.foundation.layout.width
import androidx.compose.foundation.shape.CircleShape
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material3.DropdownMenu
import androidx.compose.material3.DropdownMenuItem
import androidx.compose.material3.Icon
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.setValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.alpha
import androidx.compose.ui.draw.clip
import androidx.compose.ui.draw.scale
import androidx.compose.ui.draw.shadow
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.text.font.FontFamily
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import com.notify.anything.notify.SubscriptionType
import kotlinx.coroutines.delay
import kotlinx.datetime.Clock
import kotlin.math.abs
import kotlin.math.sin
import kotlin.time.Duration.Companion.hours

/* ──────── Type pill ──────── */
@Composable
fun TypePill(type: SubscriptionType) {
    val label = if (type == SubscriptionType.EVENT) "EVENT" else "NEWS"
    val icon = if (type == SubscriptionType.EVENT) NotifyIcons.CalendarMonth else NotifyIcons.Newspaper
    Row(
        verticalAlignment = Alignment.CenterVertically,
        horizontalArrangement = Arrangement.spacedBy(4.dp),
        modifier = Modifier
            .clip(RoundedCornerShape(999.dp))
            .background(NotifyColors.surfaceHi.copy(alpha = 0.5f))
            .padding(horizontal = 8.dp, vertical = 4.dp),
    ) {
        Icon(icon, null, tint = NotifyColors.label3, modifier = Modifier.size(11.dp))
        Text(label, style = NotifyType.eyebrow, color = NotifyColors.label3)
    }
}

/* ──────── Cadence chip ──────── */
@Composable
fun CadenceChip(cadenceSeconds: Int, lastRunAt: kotlinx.datetime.Instant?) {
    Row(
        verticalAlignment = Alignment.CenterVertically,
        horizontalArrangement = Arrangement.spacedBy(6.dp),
        modifier = Modifier
            .clip(RoundedCornerShape(999.dp))
            .background(NotifyColors.surfaceMute)
            .border(0.5.dp, NotifyColors.stroke, RoundedCornerShape(999.dp))
            .padding(horizontal = 10.dp, vertical = 6.dp),
    ) {
        Box(
            modifier = Modifier.size(5.dp).clip(CircleShape).background(NotifyColors.accent),
        )
        Text(cadenceLabel(cadenceSeconds), style = NotifyType.eyebrow, color = NotifyColors.label2)
        Text("·", style = NotifyType.eyebrow, color = NotifyColors.label4)
        Text(
            if (lastRunAt == null) "NEVER" else relativeShort(lastRunAt),
            style = NotifyType.eyebrow,
            color = NotifyColors.label3,
        )
    }
}

/* ──────── Subscription card ──────── */
@Composable
fun SubscriptionCard(
    subscription: Subscription,
    signals: List<Signal>,
    confirmedDate: kotlinx.datetime.Instant?,
    onTap: () -> Unit,
    onRun: () -> Unit,
    onDelete: () -> Unit,
) {
    val isResolved = confirmedDate != null
    val unread = (signals.firstOrNull()?.firstSeenAt
        ?: kotlinx.datetime.Instant.DISTANT_PAST) > Clock.System.now().minus(1.hours)
    var showMenu by remember { mutableStateOf(false) }

    @OptIn(ExperimentalFoundationApi::class)
    Column(
        verticalArrangement = Arrangement.spacedBy(12.dp),
        modifier = Modifier
            .fillMaxWidth()
            .clip(RoundedCornerShape(20.dp))
            .background(if (isResolved) NotifyColors.surfaceMute else NotifyColors.surface)
            .border(0.5.dp, NotifyColors.stroke, RoundedCornerShape(20.dp))
            .combinedClickable(
                onClick = { onTap() },
                onLongClick = { showMenu = true },
            )
            .padding(18.dp),
    ) {
        Row(verticalAlignment = Alignment.CenterVertically) {
            TypePill(subscription.type)
            Spacer(Modifier.weight(1f))
            when {
                isResolved -> Box(
                    modifier = Modifier
                        .clip(RoundedCornerShape(999.dp))
                        .background(NotifyColors.accentSoft)
                        .padding(horizontal = 8.dp, vertical = 4.dp),
                ) {
                    Text("RESOLVED", style = NotifyType.eyebrow, color = NotifyColors.accent)
                }
                unread -> Box(
                    modifier = Modifier.size(8.dp).clip(CircleShape).background(NotifyColors.accent),
                )
                else -> {}
            }
            Spacer(Modifier.width(6.dp))
            Box {
                Box(
                    modifier = Modifier.size(28.dp).clip(CircleShape).clickable { showMenu = true },
                    contentAlignment = Alignment.Center,
                ) {
                    Icon(NotifyIcons.MoreHoriz, null, tint = NotifyColors.label3, modifier = Modifier.size(18.dp))
                }
                DropdownMenu(expanded = showMenu, onDismissRequest = { showMenu = false }) {
                    DropdownMenuItem(
                        text = { Text("Run now") },
                        onClick = { showMenu = false; onRun() },
                        leadingIcon = { Icon(NotifyIcons.PlayArrow, null) },
                    )
                    DropdownMenuItem(
                        text = { Text("Delete", color = NotifyColors.danger) },
                        onClick = { showMenu = false; onDelete() },
                    )
                }
            }
        }
        Text(
            subscription.query,
            style = NotifyType.title3,
            color = if (isResolved) NotifyColors.label2 else NotifyColors.label1,
            maxLines = 2,
        )
        Row(verticalAlignment = Alignment.CenterVertically) {
            if (confirmedDate != null) {
                Row(
                    verticalAlignment = Alignment.CenterVertically,
                    horizontalArrangement = Arrangement.spacedBy(6.dp),
                ) {
                    Icon(NotifyIcons.CalendarMonth, null, tint = NotifyColors.accent, modifier = Modifier.size(14.dp))
                    Text(
                        formatEventDate(confirmedDate),
                        style = NotifyType.bodyMed.copy(fontSize = 13.sp),
                        color = NotifyColors.accent,
                    )
                }
            } else {
                CadenceChip(subscription.cadenceSeconds, subscription.lastRunAt)
            }
            Spacer(Modifier.weight(1f))
            Text(
                "${signals.size}",
                style = NotifyType.bodyMed.copy(fontSize = 13.sp, fontFamily = FontFamily.Monospace),
                color = NotifyColors.label2,
            )
            Spacer(Modifier.width(6.dp))
            Icon(NotifyIcons.GraphicEq, null, tint = NotifyColors.label3, modifier = Modifier.size(13.dp))
        }
    }
}

/* ──────── Alert row ──────── */
@Composable
fun AlertRow(signal: Signal, subscription: Subscription, onTap: () -> Unit) {
    Row(
        verticalAlignment = Alignment.Top,
        horizontalArrangement = Arrangement.spacedBy(12.dp),
        modifier = Modifier
            .fillMaxWidth()
            .clip(RoundedCornerShape(14.dp))
            .background(NotifyColors.surface)
            .border(0.5.dp, NotifyColors.stroke, RoundedCornerShape(14.dp))
            .clickable { onTap() }
            .padding(14.dp),
    ) {
        Box(
            modifier = Modifier.size(32.dp).clip(CircleShape).background(NotifyColors.accentSoft),
            contentAlignment = Alignment.Center,
        ) {
            Icon(
                if (subscription.type == SubscriptionType.EVENT) NotifyIcons.CalendarMonth else NotifyIcons.Newspaper,
                null, tint = NotifyColors.accent, modifier = Modifier.size(15.dp),
            )
        }
        Column(verticalArrangement = Arrangement.spacedBy(4.dp), modifier = Modifier.weight(1f)) {
            Text(
                subscription.query.uppercase(),
                style = NotifyType.eyebrow, color = NotifyColors.label3, maxLines = 1,
            )
            Text(signal.title, style = NotifyType.bodyMed, color = NotifyColors.label1, maxLines = 2)
            Row(
                horizontalArrangement = Arrangement.spacedBy(6.dp),
                verticalAlignment = Alignment.CenterVertically,
            ) {
                signal.sourceDomains.firstOrNull()?.let { dom ->
                    Text(dom, style = NotifyType.eyebrow, color = NotifyColors.label3)
                    Text("·", style = NotifyType.eyebrow, color = NotifyColors.label4)
                }
                Text(relativeShort(signal.firstSeenAt), style = NotifyType.eyebrow, color = NotifyColors.label3)
            }
        }
    }
}

/* ──────── Empty state ──────── */
@Composable
fun EmptyState(title: String, subtitle: String, cta: String? = null, onCta: (() -> Unit)? = null) {
    Column(
        horizontalAlignment = Alignment.CenterHorizontally,
        verticalArrangement = Arrangement.spacedBy(18.dp),
        modifier = Modifier.fillMaxWidth().padding(40.dp),
    ) {
        Box(
            modifier = Modifier.size(92.dp).clip(CircleShape).background(NotifyColors.accentSoft),
            contentAlignment = Alignment.Center,
        ) {
            Icon(NotifyIcons.GraphicEq, null, tint = NotifyColors.accent, modifier = Modifier.size(38.dp))
        }
        Text(title, style = NotifyType.title2, color = NotifyColors.label1)
        Text(
            subtitle,
            style = NotifyType.body,
            color = NotifyColors.label2,
            modifier = Modifier.widthIn(max = 280.dp),
        )
        if (cta != null && onCta != null) {
            Box(
                modifier = Modifier
                    .clip(RoundedCornerShape(999.dp))
                    .background(NotifyColors.accent)
                    .clickable { onCta() }
                    .padding(horizontal = 24.dp, vertical = 13.dp),
            ) {
                Text(cta, color = NotifyColors.accentInk, style = NotifyType.bodyMed)
            }
        }
    }
}

/* ──────── Toast banner ──────── */
@Composable
fun Toast(message: String?, onClear: () -> Unit, isError: Boolean = false) {
    AnimatedVisibility(
        visible = message != null,
        enter = slideInVertically(initialOffsetY = { -it }) + fadeIn(),
        exit = slideOutVertically(targetOffsetY = { -it }) + fadeOut(),
    ) {
        if (message != null) {
            LaunchedEffect(message) {
                delay(2400)
                onClear()
            }
            Row(
                verticalAlignment = Alignment.CenterVertically,
                horizontalArrangement = Arrangement.spacedBy(10.dp),
                modifier = Modifier
                    .fillMaxWidth()
                    .padding(horizontal = 16.dp)
                    .clip(RoundedCornerShape(14.dp))
                    .background(if (isError) Color(0xFF3A1418) else NotifyColors.surfaceHi)
                    .border(
                        0.75.dp,
                        if (isError) NotifyColors.danger.copy(alpha = 0.5f) else NotifyColors.stroke,
                        RoundedCornerShape(14.dp),
                    )
                    .padding(horizontal = 14.dp, vertical = 12.dp),
            ) {
                Icon(
                    if (isError) NotifyIcons.Error else NotifyIcons.Info, null,
                    tint = if (isError) NotifyColors.danger else NotifyColors.accent,
                    modifier = Modifier.size(15.dp),
                )
                Text(
                    message,
                    style = NotifyType.body,
                    color = NotifyColors.label1,
                    maxLines = 3,
                    modifier = Modifier.weight(1f),
                )
                Box(
                    modifier = Modifier.size(20.dp).clickable { onClear() },
                    contentAlignment = Alignment.Center,
                ) {
                    Icon(NotifyIcons.Close, null, tint = NotifyColors.label3, modifier = Modifier.size(11.dp))
                }
            }
        }
    }
}

/* ──────── FAB ──────── */
@Composable
fun NotifyFab(onClick: () -> Unit) {
    var pressed by remember { mutableStateOf(false) }
    val scale by animateFloatAsState(if (pressed) 0.94f else 1f, tween(180), label = "fab")
    Box(
        modifier = Modifier
            .size(56.dp)
            .scale(scale)
            .shadow(16.dp, CircleShape, ambientColor = NotifyColors.accentGlow, spotColor = NotifyColors.accentGlow)
            .clip(CircleShape)
            .background(NotifyColors.accent)
            .clickable {
                pressed = true
                onClick()
                pressed = false
            },
        contentAlignment = Alignment.Center,
    ) {
        Icon(NotifyIcons.Add, "+", tint = NotifyColors.accentInk, modifier = Modifier.size(22.dp))
    }
}

/* ──────── Pulse ring ──────── */
@Composable
fun PulseRing(modifier: Modifier = Modifier) {
    val transition = rememberInfiniteTransition(label = "pulse")
    val s by transition.animateFloat(
        initialValue = 0.85f, targetValue = 1.18f,
        animationSpec = infiniteRepeatable(tween(1600), repeatMode = RepeatMode.Reverse),
        label = "scale",
    )
    Box(modifier = modifier.size(180.dp), contentAlignment = Alignment.Center) {
        repeat(3) { i ->
            Box(
                modifier = Modifier
                    .size(110.dp)
                    .scale(s + i * 0.18f)
                    .alpha((1.4f - s - i * 0.4f).coerceIn(0f, 0.6f))
                    .clip(CircleShape)
                    .border(1.5.dp, NotifyColors.accent.copy(alpha = (0.5f - i * 0.15f).coerceAtLeast(0.05f)), CircleShape),
            )
        }
        Box(
            modifier = Modifier.size(90.dp).clip(CircleShape).background(NotifyColors.accentSoft),
            contentAlignment = Alignment.Center,
        ) {
            Icon(NotifyIcons.GraphicEq, null, tint = NotifyColors.accent, modifier = Modifier.size(34.dp))
        }
    }
}

/* ──────── Waveform ──────── */
@Composable
fun Waveform(modifier: Modifier = Modifier) {
    val transition = rememberInfiniteTransition(label = "wave")
    val phase by transition.animateFloat(
        initialValue = 0f, targetValue = 6.283f,
        animationSpec = infiniteRepeatable(tween(1400)),
        label = "phase",
    )
    Row(
        horizontalArrangement = Arrangement.spacedBy(4.dp),
        verticalAlignment = Alignment.CenterVertically,
        modifier = modifier.fillMaxWidth().height(36.dp),
    ) {
        val n = 26
        repeat(n) { i ->
            val x = i.toFloat() / n
            val h = abs(sin((phase * 2 + x * 6).toDouble())) * 0.7 + 0.3
            Box(
                modifier = Modifier
                    .width(3.dp)
                    .height((4 + 28 * h).dp)
                    .clip(RoundedCornerShape(999.dp))
                    .background(NotifyColors.accent.copy(alpha = 0.7f)),
            )
        }
    }
}

/* ──────── Confidence bar ──────── */
@Composable
fun ConfidenceBar(value: Float) {
    Row(horizontalArrangement = Arrangement.spacedBy(3.dp)) {
        repeat(3) { i ->
            Box(
                modifier = Modifier
                    .width(6.dp)
                    .height(4.dp)
                    .clip(RoundedCornerShape(999.dp))
                    .background(if ((i + 1) * 0.34f <= value) NotifyColors.accent else NotifyColors.surfaceHi),
            )
        }
    }
}
