package com.notify.anything.notify.ui

import androidx.compose.animation.core.RepeatMode
import androidx.compose.animation.core.animateFloat
import androidx.compose.animation.core.infiniteRepeatable
import androidx.compose.animation.core.rememberInfiniteTransition
import androidx.compose.animation.core.tween
import androidx.compose.foundation.background
import androidx.compose.foundation.border
import androidx.compose.foundation.clickable
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.Spacer
import androidx.compose.foundation.layout.fillMaxHeight
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.heightIn
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.items
import androidx.compose.foundation.rememberScrollState
import androidx.compose.foundation.shape.CircleShape
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.foundation.text.BasicTextField
import androidx.compose.foundation.verticalScroll
import androidx.compose.material3.ExperimentalMaterial3Api
import androidx.compose.material3.Icon
import androidx.compose.material3.ModalBottomSheet
import androidx.compose.material3.Text
import androidx.compose.material3.rememberModalBottomSheetState
import androidx.compose.runtime.Composable
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateMapOf
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.rememberCoroutineScope
import androidx.compose.runtime.setValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.draw.scale
import androidx.compose.ui.graphics.SolidColor
import androidx.compose.ui.text.TextStyle
import androidx.compose.ui.text.font.FontFamily
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import com.notify.anything.notify.OnboardingSuggestion
import kotlinx.coroutines.launch

private val PREFILL_CONTEXT = """
    San Francisco, Senior Software Engineer at a Series B startup (~150 people). Five years in, three at current company. Backend-leaning full-stack — Python and Go daily, currently leading a Postgres-to-distributed-storage migration. Reads Hacker News in the morning, Pragmatic Engineer on weekends. Vaguely thinking about leaving for a smaller team or starting something herself in 2 years.

    Lives in the Mission. Runs 4x/week, training for the SF Marathon. Member of a local run club. Hot yoga twice a week at a studio nearby. Climbs at Mission Cliffs on weekends. Mostly cooks — Whole Foods plus farmers' market on Saturdays. Eats out 2x/week at places with real vegetable programs (Souvla, Reem's, Nopa). Doesn't drink much; will go to a natural wine bar with friends. Sober-curious adjacent, interested in NA cocktails.

    Cultural taste: indie and electronic — Bon Iver, Caribou, Floating Points, Mitski. Catches small shows at The Independent, The Chapel, Great American. Avoids arena tours. Watches A24 movies. Genuinely interested in AI/ML developments but skeptical of hype cycles, occasionally reads papers. Goes to maybe one tech meetup a month — picks them carefully, hates recruiting-bait events.

    Civic life: watches SF politics closely — housing policy, public transit, Prop measures. Concerned about Mission gentrification, BART funding, street safety. She votes and reads the voter guide.
""".trimIndent()

private const val CHAR_LIMIT = 2000
private const val MIN_CHARS = 10

private enum class AIPhase { Input, Loading, Reveal }

/**
 * AI-from-context watcher suggester — Compose parity with iOS AISuggestionsView.
 * Three phases in one sheet: free-text input → loading pulse → toggleable reveal cards.
 */
@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun AISuggestionsSheet(state: AppState, onDismiss: () -> Unit) {
    val sheetState = rememberModalBottomSheetState(skipPartiallyExpanded = true)
    val scope = rememberCoroutineScope()

    var phase by remember { mutableStateOf(AIPhase.Input) }
    var contextText by remember { mutableStateOf(PREFILL_CONTEXT) }
    var suggestions by remember { mutableStateOf(emptyList<OnboardingSuggestion>()) }
    var fallback by remember { mutableStateOf(false) }
    var error by remember { mutableStateOf<String?>(null) }
    val activated = remember { mutableStateMapOf<String, String>() } // localKey → subId
    val inFlight = remember { mutableStateMapOf<String, Boolean>() }

    ModalBottomSheet(
        onDismissRequest = onDismiss,
        sheetState = sheetState,
        containerColor = NotifyColors.bg,
        contentColor = NotifyColors.label1,
        dragHandle = null,
    ) {
        Column(modifier = Modifier.fillMaxHeight().padding(top = 8.dp)) {
            // Top bar
            Row(
                modifier = Modifier.fillMaxWidth().padding(horizontal = 16.dp, vertical = 8.dp),
                verticalAlignment = Alignment.CenterVertically,
            ) {
                Box(
                    modifier = Modifier
                        .size(36.dp)
                        .clip(CircleShape)
                        .background(NotifyColors.surface)
                        .clickable { onDismiss() },
                    contentAlignment = Alignment.Center,
                ) {
                    Icon(NotifyIcons.ChevronLeft, null, tint = NotifyColors.label1, modifier = Modifier.size(18.dp))
                }
                Spacer(Modifier.weight(1f))
                Text("AI Suggestions", style = NotifyType.bodyMed, color = NotifyColors.label1)
                Spacer(Modifier.weight(1f))
                Box(modifier = Modifier.size(36.dp))
            }

            when (phase) {
                AIPhase.Input -> InputContent(
                    contextText = contextText,
                    onChange = { contextText = if (it.length > CHAR_LIMIT) it.take(CHAR_LIMIT) else it },
                    error = error,
                    onSubmit = {
                        val trimmed = contextText.trim()
                        if (trimmed.length < MIN_CHARS) return@InputContent
                        scope.launch {
                            phase = AIPhase.Loading
                            error = null
                            try {
                                val (sugs, fb) = state.fetchAISuggestions(trimmed)
                                suggestions = sugs
                                fallback = fb
                                phase = AIPhase.Reveal
                            } catch (t: Throwable) {
                                error = "Couldn't reach AI — ${t.message ?: t.toString()}"
                                phase = AIPhase.Input
                            }
                        }
                    },
                )
                AIPhase.Loading -> LoadingContent()
                AIPhase.Reveal -> RevealContent(
                    suggestions = suggestions,
                    fallback = fallback,
                    isActive = { activated.containsKey(it.query + it.cadenceSeconds) },
                    isInFlight = { inFlight[it.query + it.cadenceSeconds] == true },
                    onToggle = { s ->
                        val key = s.query + s.cadenceSeconds
                        if (inFlight[key] == true) return@RevealContent
                        scope.launch {
                            inFlight[key] = true
                            try {
                                val existing = activated[key]
                                if (existing != null) {
                                    state.deleteById(existing)
                                    activated.remove(key)
                                } else {
                                    val id = state.activateSuggestion(s)
                                    if (id != null) activated[key] = id
                                }
                            } finally {
                                inFlight.remove(key)
                            }
                        }
                    },
                    onDone = onDismiss,
                    activatedCount = activated.size,
                )
            }
        }
    }
}

@Composable
private fun InputContent(
    contextText: String,
    onChange: (String) -> Unit,
    error: String?,
    onSubmit: () -> Unit,
) {
    val scroll = rememberScrollState()
    val valid = contextText.trim().length >= MIN_CHARS
    Column(
        modifier = Modifier
            .fillMaxWidth()
            .verticalScroll(scroll)
            .padding(horizontal = 22.dp)
            .padding(top = 12.dp, bottom = 24.dp),
        verticalArrangement = Arrangement.spacedBy(18.dp),
    ) {
        // Sparkle badge
        Box(
            modifier = Modifier
                .align(Alignment.CenterHorizontally)
                .size(64.dp)
                .clip(RoundedCornerShape(14.dp))
                .background(NotifyColors.surface)
                .border(0.5.dp, NotifyColors.stroke, RoundedCornerShape(14.dp)),
            contentAlignment = Alignment.Center,
        ) {
            Icon(NotifyIcons.Sparkles, null, tint = NotifyColors.accent, modifier = Modifier.size(28.dp))
        }

        Column(horizontalAlignment = Alignment.CenterHorizontally, verticalArrangement = Arrangement.spacedBy(10.dp)) {
            Text(
                "Tell us what to look for.",
                style = NotifyType.title2.copy(fontSize = 28.sp, fontWeight = FontWeight.Bold),
                color = NotifyColors.label1,
            )
            Text(
                "Describe your current interests, projects, or specific signals you want the AI to track.",
                style = NotifyType.body, color = NotifyColors.label2,
            )
        }

        Column(verticalArrangement = Arrangement.spacedBy(10.dp)) {
            Text("CONTEXT & PREFERENCES", style = NotifyType.eyebrow, color = NotifyColors.label3)
            Box(
                modifier = Modifier
                    .fillMaxWidth()
                    .clip(RoundedCornerShape(16.dp))
                    .background(NotifyColors.surface)
                    .border(0.5.dp, NotifyColors.stroke, RoundedCornerShape(16.dp))
                    .heightIn(min = 200.dp)
                    .padding(14.dp),
            ) {
                BasicTextField(
                    value = contextText,
                    onValueChange = onChange,
                    cursorBrush = SolidColor(NotifyColors.accent),
                    textStyle = TextStyle(color = NotifyColors.label1, fontSize = 14.sp),
                    modifier = Modifier.fillMaxSize(),
                )
            }
            Row(modifier = Modifier.fillMaxWidth(), horizontalArrangement = Arrangement.End) {
                Text(
                    "${contextText.length} / $CHAR_LIMIT",
                    style = NotifyType.caption.copy(fontFamily = FontFamily.Monospace),
                    color = NotifyColors.label3,
                )
            }
        }

        if (error != null) {
            Text(error, style = NotifyType.caption, color = NotifyColors.danger)
        }

        Box(
            modifier = Modifier
                .fillMaxWidth()
                .clip(RoundedCornerShape(28.dp))
                .background(if (valid) NotifyColors.accent else NotifyColors.accent.copy(alpha = 0.4f))
                .clickable(enabled = valid) { onSubmit() }
                .padding(vertical = 16.dp),
            contentAlignment = Alignment.Center,
        ) {
            Row(verticalAlignment = Alignment.CenterVertically, horizontalArrangement = Arrangement.spacedBy(10.dp)) {
                Icon(NotifyIcons.Search, null, tint = NotifyColors.accentInk, modifier = Modifier.size(18.dp))
                Text("Find Signals", style = NotifyType.bodyMed, color = NotifyColors.accentInk)
            }
        }
    }
}

@Composable
private fun LoadingContent() {
    val transition = rememberInfiniteTransition(label = "ai-pulse")
    val pulse by transition.animateFloat(
        initialValue = 0.85f, targetValue = 1.05f,
        animationSpec = infiniteRepeatable(animation = tween(900), repeatMode = RepeatMode.Reverse),
        label = "scale",
    )
    Box(modifier = Modifier.fillMaxSize(), contentAlignment = Alignment.Center) {
        Column(horizontalAlignment = Alignment.CenterHorizontally, verticalArrangement = Arrangement.spacedBy(20.dp)) {
            Box(
                modifier = Modifier
                    .scale(pulse)
                    .size(80.dp)
                    .clip(RoundedCornerShape(18.dp))
                    .background(NotifyColors.surface)
                    .border(0.5.dp, NotifyColors.stroke, RoundedCornerShape(18.dp)),
                contentAlignment = Alignment.Center,
            ) {
                Icon(NotifyIcons.Sparkles, null, tint = NotifyColors.accent, modifier = Modifier.size(32.dp))
            }
            Text("Scanning the web for signals…", style = NotifyType.body, color = NotifyColors.label2)
        }
    }
}

@Composable
private fun RevealContent(
    suggestions: List<OnboardingSuggestion>,
    fallback: Boolean,
    isActive: (OnboardingSuggestion) -> Boolean,
    isInFlight: (OnboardingSuggestion) -> Boolean,
    onToggle: (OnboardingSuggestion) -> Unit,
    onDone: () -> Unit,
    activatedCount: Int,
) {
    Column(modifier = Modifier.fillMaxSize()) {
        Column(modifier = Modifier.padding(horizontal = 22.dp).padding(top = 12.dp, bottom = 8.dp)) {
            Text(
                if (fallback) "Popular watchers for you" else "Suggested watchers",
                style = NotifyType.title2, color = NotifyColors.label1,
            )
            Spacer(Modifier.size(4.dp))
            Text("Tap to add. Edit cadence later.", style = NotifyType.caption, color = NotifyColors.label2)
        }
        LazyColumn(
            modifier = Modifier.weight(1f).fillMaxWidth(),
            contentPadding = androidx.compose.foundation.layout.PaddingValues(horizontal = 22.dp, vertical = 8.dp),
            verticalArrangement = Arrangement.spacedBy(10.dp),
        ) {
            items(suggestions, key = { it.query + it.cadenceSeconds }) { s ->
                SuggestionRow(s, isActive(s), isInFlight(s)) { onToggle(s) }
            }
        }
        Box(
            modifier = Modifier
                .fillMaxWidth()
                .padding(horizontal = 22.dp)
                .padding(top = 12.dp, bottom = 28.dp)
                .clip(RoundedCornerShape(28.dp))
                .background(NotifyColors.accent)
                .clickable { onDone() }
                .padding(vertical = 16.dp),
            contentAlignment = Alignment.Center,
        ) {
            Text(
                if (activatedCount == 0) "Done" else "Done · $activatedCount added",
                style = NotifyType.bodyMed, color = NotifyColors.accentInk,
            )
        }
    }
}

@Composable
private fun SuggestionRow(s: OnboardingSuggestion, active: Boolean, busy: Boolean, onClick: () -> Unit) {
    Row(
        modifier = Modifier
            .fillMaxWidth()
            .clip(RoundedCornerShape(16.dp))
            .background(NotifyColors.surface)
            .border(if (active) 1.2.dp else 0.5.dp,
                if (active) NotifyColors.accent.copy(alpha = 0.6f) else NotifyColors.stroke,
                RoundedCornerShape(16.dp))
            .clickable(enabled = !busy) { onClick() }
            .padding(14.dp),
        horizontalArrangement = Arrangement.spacedBy(12.dp),
        verticalAlignment = Alignment.Top,
    ) {
        Box(
            modifier = Modifier
                .size(36.dp)
                .clip(CircleShape)
                .background(if (active) NotifyColors.accent else NotifyColors.accentSoft),
            contentAlignment = Alignment.Center,
        ) {
            Icon(
                if (active) NotifyIcons.Check
                else if (s.type == "event") NotifyIcons.Event
                else NotifyIcons.News,
                null,
                tint = if (active) NotifyColors.accentInk else NotifyColors.accent,
                modifier = Modifier.size(16.dp),
            )
        }
        Column(verticalArrangement = Arrangement.spacedBy(4.dp), modifier = Modifier.weight(1f)) {
            Text(s.query, style = NotifyType.bodyMed, color = NotifyColors.label1, maxLines = 2)
            Text(s.reason, style = NotifyType.caption, color = NotifyColors.label3, maxLines = 2)
            Text(
                "${if (s.type == "event") "Event" else "News"} · ${aiCadenceLabel(s.cadenceSeconds)}",
                style = NotifyType.caption, color = NotifyColors.label3,
            )
        }
    }
}

private fun aiCadenceLabel(s: Int): String = when (s) {
    in Int.MIN_VALUE..3599 -> "${s / 60}m"
    3600 -> "Hourly"
    21600 -> "Every 6h"
    86400 -> "Daily"
    else -> "${s / 3600}h"
}
