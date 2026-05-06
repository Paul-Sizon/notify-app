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
import androidx.compose.foundation.layout.offset
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
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import com.notify.anything.notify.SubscriptionType
import kotlinx.coroutines.delay
import kotlinx.datetime.Clock
import kotlin.math.abs
import kotlin.math.sin
import kotlin.time.Duration.Companion.hours

/* ──────── Type pill ──────── */
/**
 * iOS-style pill chip — translucent grey on neutral surfaces, accent
 * variant for "live signal" cues. Sized for inline use inside watcher
 * card metadata rows. Per design tokens: 22px height, 9px h-padding.
 */
@Composable
fun NotifyChip(
    text: String,
    leadingIcon: androidx.compose.ui.graphics.vector.ImageVector? = null,
    accent: Boolean = false,
) {
    Row(
        verticalAlignment = Alignment.CenterVertically,
        horizontalArrangement = Arrangement.spacedBy(4.dp),
        modifier = Modifier
            .clip(RoundedCornerShape(999.dp))
            .background(if (accent) NotifyColors.accentSoft else NotifyColors.chipBg)
            .padding(horizontal = 9.dp, vertical = 3.dp),
    ) {
        if (leadingIcon != null) {
            Icon(
                leadingIcon, null,
                tint = if (accent) NotifyColors.accent else NotifyColors.label2,
                modifier = Modifier.size(10.dp),
            )
        }
        Text(
            text,
            style = NotifyType.footnote.copy(fontSize = 12.sp, fontWeight = FontWeight.Medium),
            color = if (accent) NotifyColors.accent else NotifyColors.label2,
        )
    }
}

@Composable
fun TypePill(type: SubscriptionType) {
    val label = if (type == SubscriptionType.EVENT) "event" else "news"
    val icon = if (type == SubscriptionType.EVENT) NotifyIcons.CalendarMonth else NotifyIcons.Newspaper
    NotifyChip(text = label, leadingIcon = icon)
}

/**
 * Cadence chip lives inside the metadata row beneath a watcher's query.
 * Wording follows the design's "every Xh" copy convention.
 */
@Composable
fun CadenceChip(cadenceSeconds: Int, lastRunAt: kotlinx.datetime.Instant?) {
    NotifyChip(text = "every ${cadenceShort(cadenceSeconds)}")
}

private fun cadenceShort(seconds: Int): String {
    val h = seconds / 3600
    val m = (seconds % 3600) / 60
    return when {
        h >= 24 -> "${h / 24}d"
        h >= 1 -> if (m == 0) "${h}h" else "${h}h${m}m"
        else -> "${m}m"
    }
}

/**
 * Watcher card — matches the Claude Design layout.
 *
 * Active state: bgElev2 surface, 16dp radius, query at 17/600, metadata
 * chips beneath (type · cadence · "new" with wave glyph if unread). Right
 * column holds the relative last-run timestamp and a chevron. Unread
 * watchers gain a 6px red-orange dot offset off the left edge with a
 * glow shadow — the iOS lockscreen "new" cue.
 *
 * Resolved state: muted bgElev (surfaceMute), strikethrough query,
 * answer line led with a tiny check icon, optional answer subtitle, and
 * an accent-stamped date in the right column with "RESOLVED" eyebrow
 * underneath. Long-press for run/delete menu.
 */
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

    Box(modifier = Modifier.fillMaxWidth()) {
        // Unread dot lives outside the card, kissing the left edge with a
        // soft glow. Position constants pulled from design (left = -4,
        // top = padV + 8 ≈ 28).
        if (unread && !isResolved) {
            Box(
                modifier = Modifier
                    .padding(start = 0.dp, top = 28.dp)
                    .size(6.dp)
                    .clip(CircleShape)
                    .background(NotifyColors.accent),
            )
        }

        @OptIn(ExperimentalFoundationApi::class)
        Row(
            verticalAlignment = Alignment.Top,
            horizontalArrangement = Arrangement.spacedBy(12.dp),
            modifier = Modifier
                .fillMaxWidth()
                .clip(RoundedCornerShape(16.dp))
                .background(if (isResolved) NotifyColors.surfaceMute else NotifyColors.surface)
                .combinedClickable(
                    onClick = { onTap() },
                    onLongClick = { showMenu = true },
                )
                .padding(horizontal = 16.dp, vertical = 18.dp),
        ) {
            Column(modifier = Modifier.weight(1f)) {
                Text(
                    subscription.query,
                    style = NotifyType.title3,
                    color = if (isResolved) NotifyColors.label2 else NotifyColors.label1,
                    textDecoration = if (isResolved)
                        androidx.compose.ui.text.style.TextDecoration.LineThrough
                    else null,
                    maxLines = 2,
                )
                Spacer(Modifier.height(8.dp))

                if (isResolved && confirmedDate != null) {
                    // Resolved: check + answer summary + date sub
                    Row(
                        verticalAlignment = Alignment.CenterVertically,
                        horizontalArrangement = Arrangement.spacedBy(6.dp),
                        modifier = Modifier.padding(bottom = 4.dp),
                    ) {
                        Icon(
                            NotifyIcons.Check, null,
                            tint = NotifyColors.accent,
                            modifier = Modifier.size(12.dp),
                        )
                        Text(
                            "Confirmed " + formatEventDate(confirmedDate),
                            style = NotifyType.caption.copy(fontWeight = FontWeight.Medium),
                            color = NotifyColors.label1,
                        )
                    }
                    Text(
                        "Agent has stopped checking — pull a card from history to re-arm.",
                        style = NotifyType.footnote,
                        color = NotifyColors.label3,
                        maxLines = 2,
                    )
                } else {
                    Row(
                        horizontalArrangement = Arrangement.spacedBy(6.dp),
                        verticalAlignment = Alignment.CenterVertically,
                    ) {
                        TypePill(subscription.type)
                        CadenceChip(subscription.cadenceSeconds, subscription.lastRunAt)
                        if (unread) {
                            NotifyChip(text = "new", accent = true)
                        }
                    }
                }
            }
            Column(
                horizontalAlignment = Alignment.End,
                verticalArrangement = Arrangement.spacedBy(4.dp),
                modifier = Modifier.padding(top = 2.dp),
            ) {
                Text(
                    text = if (isResolved && confirmedDate != null)
                        formatEventDate(confirmedDate)
                    else subscription.lastRunAt?.let { relativeShort(it) } ?: "NEVER",
                    style = NotifyType.footnote.copy(
                        fontWeight = if (isResolved) FontWeight.SemiBold else FontWeight.Normal,
                        fontFamily = FontFamily.Monospace,
                        letterSpacing = if (isResolved) 0.2.sp else 0.sp,
                    ),
                    color = if (isResolved) NotifyColors.accent else NotifyColors.label3,
                )
                if (isResolved) {
                    Text(
                        "RESOLVED",
                        style = NotifyType.eyebrow.copy(fontSize = 10.sp),
                        color = NotifyColors.label4,
                    )
                } else {
                    Icon(
                        NotifyIcons.ChevronRight, null,
                        tint = NotifyColors.label4,
                        modifier = Modifier.size(14.dp),
                    )
                }
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
    // Slow vertical bob on the icon — never settles, gives the empty
    // screen a heartbeat so it feels alive rather than abandoned.
    val transition = rememberInfiniteTransition(label = "empty")
    val bob by transition.animateFloat(
        initialValue = -4f, targetValue = 4f,
        animationSpec = infiniteRepeatable(
            tween(2200, easing = androidx.compose.animation.core.FastOutSlowInEasing),
            repeatMode = RepeatMode.Reverse,
        ),
        label = "bob",
    )
    val pulseAlpha by transition.animateFloat(
        initialValue = 0.55f, targetValue = 1f,
        animationSpec = infiniteRepeatable(
            tween(2200, easing = androidx.compose.animation.core.FastOutSlowInEasing),
            repeatMode = RepeatMode.Reverse,
        ),
        label = "alpha",
    )
    Column(
        horizontalAlignment = Alignment.CenterHorizontally,
        verticalArrangement = Arrangement.spacedBy(18.dp),
        modifier = Modifier.fillMaxWidth().padding(40.dp),
    ) {
        Box(
            modifier = Modifier
                .size(92.dp)
                .offset { androidx.compose.ui.unit.IntOffset(0, bob.toInt()) }
                .clip(CircleShape)
                .background(NotifyColors.accentSoft),
            contentAlignment = Alignment.Center,
        ) {
            Icon(
                NotifyIcons.GraphicEq, null,
                tint = NotifyColors.accent.copy(alpha = pulseAlpha),
                modifier = Modifier.size(38.dp),
            )
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
                    .softClick(pressScale = 0.95f, haptic = { tapMedium() }) { onCta() }
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
/**
 * Floating action button with two animation hooks:
 *  - `bumpKey` change triggers a one-shot scale-pulse to celebrate side
 *    effects (e.g. a watcher just landed in the list).
 *  - press scale + ripple via `softClick`.
 */
@Composable
fun NotifyFab(bumpKey: Any = Unit, onClick: () -> Unit) {
    var bumpScale by remember { mutableStateOf(1f) }
    val animatedBump by animateFloatAsState(
        targetValue = bumpScale,
        animationSpec = androidx.compose.animation.core.spring(
            dampingRatio = 0.45f,
            stiffness = androidx.compose.animation.core.Spring.StiffnessMedium,
        ),
        label = "fab-bump",
    )
    LaunchedEffect(bumpKey) {
        bumpScale = 1.18f
        kotlinx.coroutines.delay(180)
        bumpScale = 1f
    }
    Box(
        modifier = Modifier
            .size(56.dp)
            .scale(animatedBump)
            .shadow(16.dp, CircleShape, ambientColor = NotifyColors.accentGlow, spotColor = NotifyColors.accentGlow)
            .clip(CircleShape)
            .background(NotifyColors.accent)
            .softClick(pressScale = 0.9f, haptic = { tapMedium() }) { onClick() },
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
