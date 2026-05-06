package com.notify.anything.notify.ui

import androidx.compose.animation.core.animateIntAsState
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
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.getValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.text.font.FontFamily
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp

@Composable
fun AlertsScreen(state: AppState) {
    val items = state.allSignalsRecent
    val grouped = items.groupBy { bucketOf(it.first.firstSeenAt) }
    val animatedCount by animateIntAsState(items.size, tween(450), label = "alerts-count")

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
                    Text("Alerts", style = NotifyType.largeTitle, color = NotifyColors.label1)
                    Spacer(Modifier.size(8.dp))
                    Text(
                        "$animatedCount",
                        style = NotifyType.title2.copy(fontFamily = FontFamily.Monospace, fontSize = 22.sp),
                        color = NotifyColors.label3,
                    )
                }
                Text("Every signal that buzzed your phone.", style = NotifyType.body, color = NotifyColors.label2)
            }
        }

        if (items.isEmpty()) {
            item {
                Box(modifier = Modifier.fillMaxWidth().padding(top = 48.dp), contentAlignment = Alignment.Center) {
                    EmptyState(
                        title = "No alerts yet",
                        subtitle = "Watchers will buzz your phone here when something new is found.",
                    )
                }
            }
        } else {
            for (bucket in DayBucket.values()) {
                val list = grouped[bucket] ?: continue
                item(key = "h-$bucket") {
                    Text(
                        bucketLabel(bucket),
                        style = NotifyType.eyebrow,
                        color = NotifyColors.label3,
                        modifier = Modifier.padding(horizontal = 22.dp, vertical = 10.dp),
                    )
                }
                for ((sig, sub) in list) {
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
}

private fun bucketLabel(b: DayBucket): String = when (b) {
    DayBucket.TODAY -> "TODAY"
    DayBucket.YESTERDAY -> "YESTERDAY"
    DayBucket.THIS_WEEK -> "THIS WEEK"
    DayBucket.EARLIER -> "EARLIER"
}
