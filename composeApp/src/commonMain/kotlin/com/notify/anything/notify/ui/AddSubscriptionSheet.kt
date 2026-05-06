package com.notify.anything.notify.ui

import com.notify.anything.notify.ui.NotifyIcons

import androidx.compose.animation.core.tween
import androidx.compose.animation.core.animateFloatAsState
import androidx.compose.foundation.background
import androidx.compose.foundation.border
import androidx.compose.foundation.clickable
import androidx.compose.foundation.gestures.detectHorizontalDragGestures
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.BoxWithConstraints
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.Spacer
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.offset
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.layout.width
import androidx.compose.foundation.shape.CircleShape
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.foundation.text.BasicTextField
import androidx.compose.foundation.text.KeyboardOptions
import androidx.compose.material3.ExperimentalMaterial3Api
import androidx.compose.material3.Icon
import androidx.compose.material3.ModalBottomSheet
import androidx.compose.material3.SheetState
import androidx.compose.material3.Text
import androidx.compose.material3.rememberModalBottomSheetState
import androidx.compose.runtime.Composable
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.setValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.graphics.SolidColor
import androidx.compose.ui.input.pointer.pointerInput
import androidx.compose.ui.platform.LocalDensity
import androidx.compose.ui.text.TextStyle
import androidx.compose.ui.text.input.ImeAction
import androidx.compose.ui.text.input.KeyboardType
import androidx.compose.ui.unit.IntOffset
import androidx.compose.ui.unit.dp
import com.notify.anything.notify.SubscriptionType
import kotlin.math.roundToInt

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun AddSubscriptionSheet(
    onDismiss: () -> Unit,
    onCreate: (String, SubscriptionType, Int) -> Unit,
    onAI: () -> Unit = {},
) {
    val sheetState = rememberModalBottomSheetState(skipPartiallyExpanded = true)
    var query by remember { mutableStateOf("") }
    var type by remember { mutableStateOf(SubscriptionType.EVENT) }
    var cadenceIndex by remember { mutableStateOf(2) }

    val cadences = listOf(
        "15 min" to 15 * 60,
        "30 min" to 30 * 60,
        "1 hour" to 3600,
        "3 hours" to 3 * 3600,
        "6 hours" to 6 * 3600,
        "Daily" to 86400,
    )

    ModalBottomSheet(
        onDismissRequest = onDismiss,
        sheetState = sheetState,
        containerColor = NotifyColors.bgElevated,
        contentColor = NotifyColors.label1,
        dragHandle = null,
    ) {
        Column(
            modifier = Modifier
                .fillMaxWidth()
                .padding(horizontal = 22.dp)
                .padding(top = 22.dp, bottom = 16.dp),
            verticalArrangement = Arrangement.spacedBy(26.dp),
        ) {
            Box(
                modifier = Modifier
                    .width(36.dp).height(4.dp)
                    .clip(RoundedCornerShape(999.dp))
                    .background(NotifyColors.label4)
                    .align(Alignment.CenterHorizontally),
            )

            Column(verticalArrangement = Arrangement.spacedBy(4.dp)) {
                Text("New watcher", style = NotifyType.title2, color = NotifyColors.label1)
                Text(
                    "The agent will check the web on your cadence.",
                    style = NotifyType.body, color = NotifyColors.label2,
                )
            }

            Column(verticalArrangement = Arrangement.spacedBy(8.dp)) {
                Text("WATCH FOR", style = NotifyType.eyebrow, color = NotifyColors.label3)
                Box(
                    modifier = Modifier
                        .fillMaxWidth()
                        .clip(RoundedCornerShape(14.dp))
                        .background(NotifyColors.surface)
                        .border(0.5.dp, NotifyColors.stroke, RoundedCornerShape(14.dp))
                        .padding(14.dp),
                ) {
                    BasicTextField(
                        value = query,
                        onValueChange = { query = it },
                        singleLine = true,
                        textStyle = NotifyType.title3.copy(color = NotifyColors.label1),
                        cursorBrush = SolidColor(NotifyColors.accent),
                        keyboardOptions = KeyboardOptions(
                            keyboardType = KeyboardType.Uri,
                            imeAction = ImeAction.Done,
                            autoCorrectEnabled = false,
                            capitalization = androidx.compose.ui.text.input.KeyboardCapitalization.None,
                        ),
                        decorationBox = { inner ->
                            if (query.isEmpty()) {
                                Text(
                                    "Coldplay tour São Paulo 2026",
                                    style = NotifyType.title3, color = NotifyColors.label4,
                                )
                            }
                            inner()
                        },
                    )
                }
            }

            Column(verticalArrangement = Arrangement.spacedBy(8.dp)) {
                Text("KIND", style = NotifyType.eyebrow, color = NotifyColors.label3)
                Row(horizontalArrangement = Arrangement.spacedBy(8.dp)) {
                    TypeChip(SubscriptionType.EVENT, type, "Event", true) { type = it }
                    TypeChip(SubscriptionType.NEWS, type, "News", false) { type = it }
                }
            }

            Column(verticalArrangement = Arrangement.spacedBy(12.dp)) {
                Text("CADENCE", style = NotifyType.eyebrow, color = NotifyColors.label3)
                Row(verticalAlignment = Alignment.CenterVertically) {
                    Text(
                        "Every ${cadences[cadenceIndex].first}",
                        style = NotifyType.title3, color = NotifyColors.label1,
                    )
                    Spacer(Modifier.weight(1f))
                    Box(
                        modifier = Modifier.size(38.dp).clip(CircleShape).background(NotifyColors.accentSoft),
                        contentAlignment = Alignment.Center,
                    ) {
                        Icon(NotifyIcons.Schedule, null, tint = NotifyColors.accent, modifier = Modifier.size(16.dp))
                    }
                }
                CadenceSlider(
                    count = cadences.size, index = cadenceIndex,
                    onChange = { cadenceIndex = it },
                )
            }

            Row(
                modifier = Modifier.fillMaxWidth().clickable { onAI() }.padding(vertical = 8.dp),
                horizontalArrangement = Arrangement.Center,
                verticalAlignment = Alignment.CenterVertically,
            ) {
                Icon(NotifyIcons.Sparkles, null, tint = NotifyColors.accent, modifier = Modifier.size(14.dp))
                Spacer(Modifier.size(8.dp))
                Text("Need suggestions? ", style = NotifyType.body, color = NotifyColors.label2)
                Text("Try AI", style = NotifyType.bodyMed, color = NotifyColors.accent)
            }

            Row(horizontalArrangement = Arrangement.spacedBy(10.dp), modifier = Modifier.fillMaxWidth()) {
                Box(
                    modifier = Modifier.weight(1f)
                        .height(50.dp)
                        .clip(RoundedCornerShape(999.dp))
                        .background(NotifyColors.surface)
                        .border(0.5.dp, NotifyColors.stroke, RoundedCornerShape(999.dp))
                        .clickable { onDismiss() },
                    contentAlignment = Alignment.Center,
                ) {
                    Text("Cancel", style = NotifyType.bodyMed, color = NotifyColors.label1)
                }
                val enabled = query.trim().length >= 3
                val alpha by animateFloatAsState(if (enabled) 1f else 0.4f, tween(150), label = "btn")
                Box(
                    modifier = Modifier.weight(1f)
                        .height(50.dp)
                        .clip(RoundedCornerShape(999.dp))
                        .background(NotifyColors.accent.copy(alpha = alpha))
                        .clickable(enabled = enabled) {
                            onCreate(query.trim(), type, cadences[cadenceIndex].second)
                            onDismiss()
                        },
                    contentAlignment = Alignment.Center,
                ) {
                    Row(
                        verticalAlignment = Alignment.CenterVertically,
                        horizontalArrangement = Arrangement.spacedBy(8.dp),
                    ) {
                        Text("Start watching", style = NotifyType.bodyMed, color = NotifyColors.accentInk)
                        Icon(NotifyIcons.ArrowForward, null, tint = NotifyColors.accentInk, modifier = Modifier.size(13.dp))
                    }
                }
            }
        }
    }
}

@Composable
private fun TypeChip(
    target: SubscriptionType,
    current: SubscriptionType,
    label: String,
    isEvent: Boolean,
    onSelect: (SubscriptionType) -> Unit,
) {
    val active = current == target
    val bg by androidx.compose.animation.animateColorAsState(
        if (active) NotifyColors.accent else NotifyColors.surface,
        animationSpec = androidx.compose.animation.core.tween(220),
        label = "chip-bg",
    )
    val fg by androidx.compose.animation.animateColorAsState(
        if (active) NotifyColors.accentInk else NotifyColors.label1,
        animationSpec = androidx.compose.animation.core.tween(220),
        label = "chip-fg",
    )
    Box(
        modifier = Modifier
            .clip(RoundedCornerShape(999.dp))
            .background(bg)
            .border(0.5.dp, NotifyColors.stroke, RoundedCornerShape(999.dp))
            .softClick(pressScale = 0.95f, haptic = { selection() }) { onSelect(target) }
            .padding(horizontal = 16.dp, vertical = 11.dp),
    ) {
        Row(
            verticalAlignment = Alignment.CenterVertically,
            horizontalArrangement = Arrangement.spacedBy(8.dp),
        ) {
            Icon(
                if (isEvent) NotifyIcons.CalendarMonth else NotifyIcons.Newspaper, null,
                tint = fg,
                modifier = Modifier.size(15.dp),
            )
            Text(label, style = NotifyType.bodyMed, color = fg)
        }
    }
}

@Composable
private fun CadenceSlider(count: Int, index: Int, onChange: (Int) -> Unit) {
    // Animate the rendered position of both the filled track and the knob,
    // so dragging snaps the underlying integer index but the visual glides
    // between stops with a spring.
    val animatedIndex by androidx.compose.animation.core.animateFloatAsState(
        targetValue = index.toFloat(),
        animationSpec = androidx.compose.animation.core.spring(
            dampingRatio = 0.78f,
            stiffness = androidx.compose.animation.core.Spring.StiffnessMedium,
        ),
        label = "cadence-knob",
    )
    BoxWithConstraints(modifier = Modifier.fillMaxWidth().height(22.dp)) {
        val w = maxWidth
        val density = LocalDensity.current
        val widthPx = with(density) { w.toPx() }
        val stepPx = widthPx / (count - 1).coerceAtLeast(1)
        val knobPx = stepPx * animatedIndex

        Box(
            modifier = Modifier
                .fillMaxWidth()
                .height(6.dp)
                .clip(RoundedCornerShape(999.dp))
                .background(NotifyColors.surface)
                .align(Alignment.CenterStart),
        )
        Box(
            modifier = Modifier
                .width(with(density) { knobPx.toDp() }.coerceAtLeast(8.dp))
                .height(6.dp)
                .clip(RoundedCornerShape(999.dp))
                .background(NotifyColors.accent)
                .align(Alignment.CenterStart),
        )
        Box(
            modifier = Modifier
                .offset { IntOffset(x = (knobPx - with(density) { 11.dp.toPx() }).roundToInt(), y = 0) }
                .size(22.dp)
                .clip(CircleShape)
                .background(NotifyColors.label1)
                .border(3.dp, NotifyColors.accent, CircleShape)
                .align(Alignment.CenterStart),
        )
        Box(
            modifier = Modifier
                .fillMaxWidth()
                .height(36.dp)
                .pointerInput(count) {
                    detectHorizontalDragGestures { change, _ ->
                        val raw = change.position.x.coerceIn(0f, widthPx)
                        val newIdx = (raw / stepPx).roundToInt().coerceIn(0, count - 1)
                        if (newIdx != index) onChange(newIdx)
                    }
                },
        )
    }
}
