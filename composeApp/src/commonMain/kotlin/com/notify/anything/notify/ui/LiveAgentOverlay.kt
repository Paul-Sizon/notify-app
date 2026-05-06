package com.notify.anything.notify.ui

import com.notify.anything.notify.ui.NotifyIcons

import androidx.compose.foundation.background
import androidx.compose.foundation.border
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.Spacer
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.shape.CircleShape
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
import androidx.compose.ui.draw.clip
import androidx.compose.ui.graphics.vector.ImageVector
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import kotlinx.coroutines.delay

private val phases = listOf(
    "Searching" to NotifyIcons.Search,
    "Reading" to NotifyIcons.Description,
    "Extracting" to NotifyIcons.AutoAwesome,
    "Wrapping up" to NotifyIcons.CheckCircle,
)

@Composable
fun LiveAgentOverlay(query: String) {
    var phase by remember { mutableStateOf(0) }

    LaunchedEffect(Unit) {
        for (i in phases.indices) {
            phase = i
            delay(600)
        }
    }

    Box(
        modifier = Modifier
            .fillMaxSize()
            .background(NotifyColors.bg.copy(alpha = 0.94f)),
        contentAlignment = Alignment.Center,
    ) {
        Column(
            verticalArrangement = Arrangement.spacedBy(32.dp),
            horizontalAlignment = Alignment.CenterHorizontally,
            modifier = Modifier.fillMaxWidth().padding(horizontal = 22.dp),
        ) {
            PulseRing()
            Column(
                horizontalAlignment = Alignment.CenterHorizontally,
                verticalArrangement = Arrangement.spacedBy(8.dp),
            ) {
                Text(
                    "LIVE AGENT",
                    style = NotifyType.eyebrow.copy(letterSpacing = 2.sp),
                    color = NotifyColors.accent,
                )
                Text(
                    query,
                    style = NotifyType.title2.copy(fontSize = 22.sp),
                    color = NotifyColors.label1,
                )
            }
            Column(
                modifier = Modifier.padding(horizontal = 40.dp),
                verticalArrangement = Arrangement.spacedBy(10.dp),
            ) {
                for ((i, item) in phases.withIndex()) {
                    PhaseRow(label = item.first, icon = item.second, state = when {
                        i < phase -> PhaseState.DONE
                        i == phase -> PhaseState.ACTIVE
                        else -> PhaseState.PENDING
                    })
                }
            }
            Spacer(Modifier.size(16.dp))
            Waveform(modifier = Modifier.padding(horizontal = 60.dp))
        }
    }
}

private enum class PhaseState { DONE, ACTIVE, PENDING }

@Composable
private fun PhaseRow(label: String, icon: ImageVector, state: PhaseState) {
    Row(
        verticalAlignment = Alignment.CenterVertically,
        horizontalArrangement = Arrangement.spacedBy(10.dp),
    ) {
        Box(
            modifier = Modifier
                .size(16.dp).clip(CircleShape)
                .border(1.5.dp, if (state != PhaseState.PENDING) NotifyColors.accent else NotifyColors.label4, CircleShape),
            contentAlignment = Alignment.Center,
        ) {
            when (state) {
                PhaseState.DONE -> Icon(NotifyIcons.Check, null, tint = NotifyColors.accent, modifier = Modifier.size(9.dp))
                PhaseState.ACTIVE -> Box(
                    modifier = Modifier.size(6.dp).clip(CircleShape).background(NotifyColors.accent),
                )
                PhaseState.PENDING -> {}
            }
        }
        Text(
            label,
            style = NotifyType.body,
            color = if (state != PhaseState.PENDING) NotifyColors.label1 else NotifyColors.label3,
        )
    }
}
