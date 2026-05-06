package com.notify.anything.notify.ui

import androidx.compose.animation.core.animateIntAsState
import androidx.compose.animation.core.tween
import androidx.compose.foundation.background
import androidx.compose.foundation.border
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
import androidx.compose.foundation.layout.width
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.getValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.text.font.FontFamily
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import kotlinx.datetime.Clock
import kotlinx.datetime.DateTimeUnit
import kotlinx.datetime.TimeZone
import kotlinx.datetime.atStartOfDayIn
import kotlinx.datetime.minus
import kotlinx.datetime.toLocalDateTime

@Composable
fun SignalsScreen(state: AppState) {
    val all = state.allSignalsRecent
    val buckets = computeDayBuckets(state)

    LazyColumn(
        modifier = Modifier.fillMaxSize().background(NotifyColors.bg),
        contentPadding = PaddingValues(top = 8.dp, bottom = 120.dp),
    ) {
        item {
            Column(
                modifier = Modifier.padding(horizontal = 22.dp).padding(top = 8.dp, bottom = 16.dp),
                verticalArrangement = Arrangement.spacedBy(6.dp),
            ) {
                Text("Signals", style = NotifyType.largeTitle, color = NotifyColors.label1)
                Text(
                    "The firehose. Everything the agent surfaced.",
                    style = NotifyType.caption,
                    color = NotifyColors.label3,
                )
            }
        }

        item { StatsCard(state, modifier = Modifier.padding(horizontal = 22.dp).padding(bottom = 16.dp)) }

        item {
            ActivityBars(buckets, modifier = Modifier.padding(horizontal = 22.dp).padding(bottom = 16.dp))
        }

        if (all.isNotEmpty()) {
            item {
                Text(
                    "LATEST",
                    style = NotifyType.eyebrow,
                    color = NotifyColors.label3,
                    modifier = Modifier.padding(horizontal = 22.dp, vertical = 10.dp),
                )
            }
            for ((sig, sub) in all.take(40)) {
                item(key = sig.id) {
                    AnimatedListEntry {
                        Box(modifier = Modifier.padding(horizontal = 22.dp, vertical = 5.dp)) {
                            AlertRow(sig, sub) { state.detailSubscriptionId = sub.id }
                        }
                    }
                }
            }
        }
    }
}

@Composable
private fun StatsCard(state: AppState, modifier: Modifier = Modifier) {
    val watching by animateIntAsState(state.activeSubscriptions.size, tween(450), label = "w")
    val resolved by animateIntAsState(state.resolvedSubscriptions.size, tween(450), label = "r")
    val signals by animateIntAsState(state.allSignalsRecent.size, tween(450), label = "s")
    Row(
        modifier = modifier
            .fillMaxWidth()
            .clip(RoundedCornerShape(20.dp))
            .background(NotifyColors.surface)
            .border(0.5.dp, NotifyColors.stroke, RoundedCornerShape(20.dp))
            .padding(vertical = 18.dp, horizontal = 4.dp),
        verticalAlignment = Alignment.CenterVertically,
    ) {
        StatCell("WATCHING", "$watching", Modifier.weight(1f))
        Divider38()
        StatCell("RESOLVED", "$resolved", Modifier.weight(1f))
        Divider38()
        StatCell("SIGNALS", "$signals", Modifier.weight(1f))
    }
}

@Composable
private fun StatCell(label: String, value: String, modifier: Modifier = Modifier) {
    Column(
        modifier = modifier,
        horizontalAlignment = Alignment.CenterHorizontally,
        verticalArrangement = Arrangement.spacedBy(6.dp),
    ) {
        Text(
            value,
            style = NotifyType.title2.copy(fontSize = 28.sp, fontFamily = FontFamily.Monospace),
            color = NotifyColors.label1,
        )
        Text(label, style = NotifyType.eyebrow, color = NotifyColors.label3)
    }
}

@Composable
private fun Divider38() {
    Box(modifier = Modifier.width(0.5.dp).height(36.dp).background(NotifyColors.stroke))
}

@Composable
private fun ActivityBars(buckets: List<Pair<kotlinx.datetime.Instant, Int>>, modifier: Modifier = Modifier) {
    val maxV = (buckets.maxOfOrNull { it.second } ?: 1).coerceAtLeast(1)
    val today = buckets.lastOrNull()?.second ?: 0
    Column(modifier = modifier.fillMaxWidth(), verticalArrangement = Arrangement.spacedBy(10.dp)) {
        Row(verticalAlignment = Alignment.CenterVertically) {
            Text("LAST 14 DAYS", style = NotifyType.eyebrow, color = NotifyColors.label3)
            Spacer(Modifier.weight(1f))
            Text("$today today", style = NotifyType.eyebrow, color = NotifyColors.accent)
        }
        Row(
            modifier = Modifier.fillMaxWidth().height(60.dp),
            verticalAlignment = Alignment.Bottom,
            horizontalArrangement = Arrangement.spacedBy(6.dp),
        ) {
            val now = Clock.System.now()
            for ((day, count) in buckets) {
                val isToday = isSameDay(day, now)
                val h = (count.toFloat() / maxV * 56f).coerceAtLeast(4f)
                Box(
                    modifier = Modifier
                        .weight(1f)
                        .height(h.dp)
                        .clip(RoundedCornerShape(999.dp))
                        .background(if (isToday) NotifyColors.accent else NotifyColors.surfaceHi),
                )
            }
        }
    }
}

private fun computeDayBuckets(state: AppState): List<Pair<kotlinx.datetime.Instant, Int>> {
    val tz = TimeZone.currentSystemDefault()
    val today = Clock.System.now().toLocalDateTime(tz).date
    val days = (13 downTo 0).map { today.minus(it, DateTimeUnit.DAY).atStartOfDayIn(tz) }
    val counts = mutableMapOf<kotlinx.datetime.Instant, Int>()
    for ((sig, _) in state.allSignalsRecent) {
        val start = sig.firstSeenAt.toLocalDateTime(tz).date.atStartOfDayIn(tz)
        counts[start] = (counts[start] ?: 0) + 1
    }
    return days.map { it to (counts[it] ?: 0) }
}
