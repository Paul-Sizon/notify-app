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
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.layout.widthIn
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.shape.CircleShape
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material3.AlertDialog
import androidx.compose.material3.ExperimentalMaterial3Api
import androidx.compose.material3.Icon
import androidx.compose.material3.ModalBottomSheet
import androidx.compose.material3.Text
import androidx.compose.material3.TextButton
import androidx.compose.material3.rememberModalBottomSheetState
import androidx.compose.runtime.Composable
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.setValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import com.notify.anything.notify.platform.UrlOpener

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun SignalDetailSheet(
    state: AppState,
    subscription: Subscription,
    urlOpener: UrlOpener,
    onDismiss: () -> Unit,
) {
    val sheetState = rememberModalBottomSheetState(skipPartiallyExpanded = true)
    val signals = state.signals(subscription.id)
    val confirmed = state.confirmedDate(subscription)
    var showDeleteConfirm by remember { mutableStateOf(false) }

    ModalBottomSheet(
        onDismissRequest = onDismiss,
        sheetState = sheetState,
        containerColor = NotifyColors.bg,
        contentColor = NotifyColors.label1,
        dragHandle = null,
    ) {
        LazyColumn(
            modifier = Modifier.fillMaxWidth(),
            contentPadding = PaddingValues(top = 14.dp, bottom = 40.dp),
        ) {
            item {
                Box(
                    modifier = Modifier
                        .fillMaxWidth()
                        .padding(bottom = 14.dp),
                    contentAlignment = Alignment.Center,
                ) {
                    Box(
                        modifier = Modifier
                            .size(width = 36.dp, height = 4.dp)
                            .clip(RoundedCornerShape(999.dp))
                            .background(NotifyColors.label4),
                    )
                }
            }
            item {
                Row(
                    modifier = Modifier
                        .fillMaxWidth()
                        .padding(horizontal = 22.dp, vertical = 4.dp),
                    horizontalArrangement = Arrangement.End,
                ) {
                    Box(
                        modifier = Modifier
                            .size(36.dp).clip(CircleShape)
                            .background(NotifyColors.surface)
                            .clickable { showDeleteConfirm = true },
                        contentAlignment = Alignment.Center,
                    ) {
                        Icon(NotifyIcons.Delete, null, tint = NotifyColors.danger, modifier = Modifier.size(15.dp))
                    }
                }
            }
            item {
                Column(
                    modifier = Modifier.padding(horizontal = 22.dp).padding(top = 8.dp, bottom = 22.dp),
                    verticalArrangement = Arrangement.spacedBy(14.dp),
                ) {
                    Row(horizontalArrangement = Arrangement.spacedBy(8.dp), verticalAlignment = Alignment.CenterVertically) {
                        TypePill(subscription.type)
                        if (confirmed != null) {
                            Row(
                                verticalAlignment = Alignment.CenterVertically,
                                horizontalArrangement = Arrangement.spacedBy(6.dp),
                                modifier = Modifier
                                    .clip(RoundedCornerShape(999.dp))
                                    .background(NotifyColors.accentSoft)
                                    .padding(horizontal = 8.dp, vertical = 5.dp),
                            ) {
                                Icon(NotifyIcons.CalendarMonth, null, tint = NotifyColors.accent, modifier = Modifier.size(11.dp))
                                Text(formatEventDate(confirmed), style = NotifyType.bodyMed.copy(fontSize = 12.sp), color = NotifyColors.accent)
                            }
                        }
                    }
                    Text(
                        subscription.query,
                        style = NotifyType.title2.copy(fontSize = 28.sp),
                        color = NotifyColors.label1,
                    )
                    CadenceChip(subscription.cadenceSeconds, subscription.lastRunAt)
                }
            }
            item {
                Row(
                    modifier = Modifier
                        .fillMaxWidth()
                        .padding(horizontal = 22.dp, vertical = 4.dp)
                        .clip(RoundedCornerShape(999.dp))
                        .background(NotifyColors.accent)
                        .clickable { state.run(subscription) }
                        .padding(horizontal = 16.dp, vertical = 14.dp),
                    verticalAlignment = Alignment.CenterVertically,
                    horizontalArrangement = Arrangement.spacedBy(10.dp),
                ) {
                    Box(
                        modifier = Modifier.size(28.dp).clip(CircleShape).background(NotifyColors.accentInk.copy(alpha = 0.2f)),
                        contentAlignment = Alignment.Center,
                    ) {
                        Icon(NotifyIcons.AutoAwesome, null, tint = NotifyColors.accentInk, modifier = Modifier.size(14.dp))
                    }
                    Text("Run agent now", style = NotifyType.bodyMed, color = NotifyColors.accentInk)
                    Spacer(Modifier.weight(1f))
                    Icon(NotifyIcons.ArrowForward, null, tint = NotifyColors.accentInk, modifier = Modifier.size(13.dp))
                }
            }

            if (signals.isEmpty()) {
                item {
                    Column(
                        modifier = Modifier.fillMaxWidth().padding(top = 30.dp),
                        horizontalAlignment = Alignment.CenterHorizontally,
                        verticalArrangement = Arrangement.spacedBy(12.dp),
                    ) {
                        Box(
                            modifier = Modifier.size(76.dp).clip(CircleShape).background(NotifyColors.accentSoft),
                            contentAlignment = Alignment.Center,
                        ) {
                            Icon(NotifyIcons.Search, null, tint = NotifyColors.accent, modifier = Modifier.size(26.dp))
                        }
                        Text("No signals yet", style = NotifyType.title3, color = NotifyColors.label1)
                        Text(
                            "Pull the trigger above to ask the agent to look right now.",
                            style = NotifyType.body,
                            color = NotifyColors.label2,
                            modifier = Modifier.widthIn(max = 280.dp),
                        )
                    }
                }
            } else {
                item {
                    Text(
                        "${signals.size} SIGNAL${if (signals.size == 1) "" else "S"}",
                        style = NotifyType.eyebrow, color = NotifyColors.label3,
                        modifier = Modifier.padding(horizontal = 22.dp, vertical = 14.dp),
                    )
                }
                for (sig in signals) {
                    item(key = sig.id) {
                        Box(modifier = Modifier.padding(horizontal = 22.dp, vertical = 5.dp)) {
                            SignalDetailRow(sig, urlOpener)
                        }
                    }
                }
            }
        }
    }

    if (showDeleteConfirm) {
        AlertDialog(
            onDismissRequest = { showDeleteConfirm = false },
            title = { Text("Delete watcher?") },
            text = { Text("This deletes the watcher and all its signals. Cannot be undone.") },
            confirmButton = {
                TextButton(onClick = {
                    showDeleteConfirm = false
                    state.delete(subscription)
                    onDismiss()
                }) { Text("Delete", color = NotifyColors.danger) }
            },
            dismissButton = { TextButton(onClick = { showDeleteConfirm = false }) { Text("Cancel") } },
            containerColor = NotifyColors.bgElevated,
            titleContentColor = NotifyColors.label1,
            textContentColor = NotifyColors.label2,
        )
    }
}

@Composable
private fun SignalDetailRow(signal: Signal, urlOpener: UrlOpener) {
    Column(
        modifier = Modifier
            .fillMaxWidth()
            .clip(RoundedCornerShape(14.dp))
            .background(NotifyColors.surface)
            .border(0.5.dp, NotifyColors.stroke, RoundedCornerShape(14.dp))
            .padding(14.dp),
        verticalArrangement = Arrangement.spacedBy(12.dp),
    ) {
        if (signal.isResolved && signal.occursAt != null) {
            Row(
                verticalAlignment = Alignment.CenterVertically,
                horizontalArrangement = Arrangement.spacedBy(6.dp),
                modifier = Modifier
                    .clip(RoundedCornerShape(999.dp))
                    .background(NotifyColors.accentSoft)
                    .padding(horizontal = 9.dp, vertical = 5.dp),
            ) {
                Icon(NotifyIcons.CalendarMonth, null, tint = NotifyColors.accent, modifier = Modifier.size(11.dp))
                Text(formatEventDate(signal.occursAt), style = NotifyType.bodyMed.copy(fontSize = 12.sp), color = NotifyColors.accent)
            }
        }
        Text(signal.title, style = NotifyType.bodyMed, color = NotifyColors.label1)
        if (!signal.body.isNullOrBlank()) {
            Text(signal.body, style = NotifyType.body, color = NotifyColors.label2)
        }
        signal.url?.let { url ->
            Row(
                modifier = Modifier
                    .clip(RoundedCornerShape(999.dp))
                    .background(NotifyColors.accentSoft)
                    .clickable { urlOpener.open(url) }
                    .padding(horizontal = 10.dp, vertical = 6.dp),
                verticalAlignment = Alignment.CenterVertically,
                horizontalArrangement = Arrangement.spacedBy(6.dp),
            ) {
                Icon(NotifyIcons.OpenInNew, null, tint = NotifyColors.accent, modifier = Modifier.size(11.dp))
                Text(
                    signal.sourceDomains.firstOrNull() ?: "source",
                    style = NotifyType.bodyMed.copy(fontSize = 12.sp), color = NotifyColors.accent,
                )
            }
        }
        Row(verticalAlignment = Alignment.CenterVertically) {
            Text(relativeLong(signal.firstSeenAt), style = NotifyType.caption.copy(fontSize = 11.sp), color = NotifyColors.label3)
            Spacer(Modifier.weight(1f))
            ConfidenceBar(signal.confidence)
        }
    }
}
