package com.notify.anything.notify.ui

import androidx.compose.animation.core.RepeatMode
import androidx.compose.animation.core.animateFloat
import androidx.compose.animation.core.animateFloatAsState
import androidx.compose.animation.core.animateIntAsState
import androidx.compose.animation.core.infiniteRepeatable
import androidx.compose.animation.core.rememberInfiniteTransition
import androidx.compose.animation.core.tween
import androidx.compose.foundation.background
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.PaddingValues
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.Spacer
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.LazyItemScope
import androidx.compose.foundation.lazy.items
import androidx.compose.foundation.shape.CircleShape
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.getValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.alpha
import androidx.compose.ui.draw.clip
import androidx.compose.ui.draw.scale
import androidx.compose.ui.text.font.FontFamily
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp

@Composable
fun WatchersScreen(state: AppState) {
    val active = state.activeSubscriptions
    val resolved = state.resolvedSubscriptions
    val animatedCount by animateIntAsState(
        targetValue = active.size, animationSpec = tween(450), label = "watching-count",
    )

    LazyColumn(
        modifier = Modifier.fillMaxSize().background(NotifyColors.bg),
        contentPadding = PaddingValues(top = 8.dp, bottom = 120.dp),
    ) {
        item {
            Column(
                modifier = Modifier.padding(horizontal = 22.dp).padding(top = 8.dp, bottom = 16.dp),
                verticalArrangement = Arrangement.spacedBy(6.dp),
            ) {
                Row(verticalAlignment = Alignment.Bottom) {
                    Text("Watching", style = NotifyType.title1, color = NotifyColors.label1)
                    Spacer(Modifier.size(8.dp))
                    Text(
                        "$animatedCount",
                        style = NotifyType.title2.copy(fontFamily = FontFamily.Monospace, fontSize = 22.sp),
                        color = NotifyColors.label3,
                    )
                    Spacer(Modifier.weight(1f))
                    StatusOrb(loading = state.loading)
                }
                Text(
                    "Quiet by default. Loud when it matters.",
                    style = NotifyType.body, color = NotifyColors.label2,
                )
            }
        }

        if (state.subscriptions.isEmpty()) {
            item {
                Box(modifier = Modifier.fillMaxWidth().padding(top = 48.dp), contentAlignment = Alignment.Center) {
                    EmptyState(
                        title = "Nothing to watch yet",
                        subtitle = "Add a topic and the agent quietly checks the web on your cadence — pinging only when something changes.",
                        cta = "Add your first watcher",
                        onCta = { state.showAddSheet = true },
                    )
                }
            }
        } else {
            items(active, key = { it.id }) { sub ->
                AnimatedListEntry {
                    Box(modifier = Modifier.padding(horizontal = 22.dp, vertical = 6.dp)) {
                        SubscriptionCard(
                            subscription = sub,
                            signals = state.signals(sub.id),
                            confirmedDate = null,
                            onTap = { state.detailSubscriptionId = sub.id },
                            onRun = { state.run(sub) },
                            onDelete = { state.delete(sub) },
                        )
                    }
                }
            }

            if (resolved.isNotEmpty()) {
                item {
                    Row(
                        modifier = Modifier.padding(horizontal = 22.dp, vertical = 14.dp),
                        verticalAlignment = Alignment.CenterVertically,
                        horizontalArrangement = Arrangement.spacedBy(8.dp),
                    ) {
                        Text("RESOLVED · PAUSED", style = NotifyType.eyebrow, color = NotifyColors.label3)
                        Box(modifier = Modifier.weight(1f).height(1.dp).background(NotifyColors.stroke))
                    }
                }
                items(resolved, key = { it.id }) { sub ->
                    AnimatedListEntry {
                        Box(modifier = Modifier.padding(horizontal = 22.dp, vertical = 6.dp)) {
                            SubscriptionCard(
                                subscription = sub,
                                signals = state.signals(sub.id),
                                confirmedDate = state.confirmedDate(sub),
                                onTap = { state.detailSubscriptionId = sub.id },
                                onRun = { state.run(sub) },
                                onDelete = { state.delete(sub) },
                            )
                        }
                    }
                }
            }
        }
    }
}

/**
 * Status orb in the screen header. Idle state = solid green dot. Loading
 * state = same dot with two ripple rings expanding and fading outward.
 * Mirrors the iOS-side concentric-pulse cue: motion = "we're working on it".
 */
@Composable
private fun StatusOrb(loading: Boolean) {
    Box(modifier = Modifier.size(28.dp), contentAlignment = Alignment.Center) {
        if (loading) {
            val transition = rememberInfiniteTransition(label = "orb")
            for (i in 0..1) {
                val phase by transition.animateFloat(
                    initialValue = 0f, targetValue = 1f,
                    animationSpec = infiniteRepeatable(
                        tween(1400, delayMillis = i * 700, easing = androidx.compose.animation.core.LinearEasing),
                    ),
                    label = "ring$i",
                )
                Box(
                    modifier = Modifier
                        .size(8.dp)
                        .scale(1f + phase * 1.6f)
                        .alpha(1f - phase)
                        .clip(CircleShape)
                        .background(NotifyColors.accent),
                )
            }
        }
        Box(modifier = Modifier.size(8.dp).clip(CircleShape).background(NotifyColors.accent))
    }
}
