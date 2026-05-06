package com.notify.anything.notify.ui

import androidx.compose.animation.core.Spring
import androidx.compose.animation.core.animateDpAsState
import androidx.compose.animation.core.animateFloatAsState
import androidx.compose.animation.core.spring
import androidx.compose.animation.core.tween
import androidx.compose.foundation.LocalIndication
import androidx.compose.foundation.clickable
import androidx.compose.foundation.interaction.MutableInteractionSource
import androidx.compose.foundation.interaction.collectIsPressedAsState
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.offset
import androidx.compose.runtime.Composable
import androidx.compose.runtime.CompositionLocalProvider
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.setValue
import androidx.compose.runtime.staticCompositionLocalOf
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.alpha
import androidx.compose.ui.draw.scale
import androidx.compose.ui.unit.Dp
import androidx.compose.ui.unit.dp
import com.notify.anything.notify.platform.Haptics
import com.notify.anything.notify.platform.NoopHaptics

/**
 * Component-level access to platform haptics. App root provides the actual
 * Android implementation; previews + JVM desktop fall back to no-op.
 */
val LocalHaptics = staticCompositionLocalOf<Haptics> { NoopHaptics }

@Composable
fun ProvideHaptics(haptics: Haptics, content: @Composable () -> Unit) {
    CompositionLocalProvider(LocalHaptics provides haptics) { content() }
}

/**
 * Tappable surface with soft press-scale + ripple + haptic. Use anywhere
 * a row, card, or chip should feel "alive" without rolling the same
 * scale-and-feedback boilerplate per call site.
 *
 * `pressScale` controls how much the element shrinks under the finger.
 * 0.97f = subtle (rows), 0.94f = button-like, 0.90f = aggressive (FAB).
 */
@Composable
fun Modifier.softClick(
    pressScale: Float = 0.97f,
    haptic: Haptics.() -> Unit = { tap() },
    onClick: () -> Unit,
): Modifier {
    val haptics = LocalHaptics.current
    val source = remember { MutableInteractionSource() }
    val pressed by source.collectIsPressedAsState()
    val scale by animateFloatAsState(
        targetValue = if (pressed) pressScale else 1f,
        animationSpec = spring(stiffness = Spring.StiffnessMediumLow, dampingRatio = 0.7f),
        label = "press-scale",
    )
    val indication = LocalIndication.current
    return this
        .scale(scale)
        .clickable(
            interactionSource = source,
            indication = indication,
            onClick = {
                haptics.haptic()
                onClick()
            },
        )
}

fun <T> notifySpring() = spring<T>(
    stiffness = Spring.StiffnessMediumLow,
    dampingRatio = 0.78f,
    visibilityThreshold = null,
)

fun <T> quickTween() = tween<T>(durationMillis = 220)

/**
 * Wraps a list item so its first appearance fades and slides upward from
 * a small offset. Driven by a one-shot `visible` flag set in
 * `LaunchedEffect(Unit)`, so each item plays the animation exactly once
 * when it enters the composition.
 */
@Composable
fun AnimatedListEntry(
    initialOffsetY: Dp = 14.dp,
    content: @Composable () -> Unit,
) {
    var visible by remember { mutableStateOf(false) }
    LaunchedEffect(Unit) { visible = true }
    val offsetY by animateDpAsState(
        targetValue = if (visible) 0.dp else initialOffsetY,
        animationSpec = spring(stiffness = Spring.StiffnessMediumLow, dampingRatio = 0.85f),
        label = "entry-offset",
    )
    val alpha by animateFloatAsState(
        targetValue = if (visible) 1f else 0f,
        animationSpec = tween(durationMillis = 280),
        label = "entry-alpha",
    )
    Box(modifier = Modifier.offset(y = offsetY).alpha(alpha)) { content() }
}
