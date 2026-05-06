package com.notify.anything.notify.ui

import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Shapes
import androidx.compose.material3.Typography
import androidx.compose.material3.darkColorScheme
import androidx.compose.runtime.Composable
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.text.TextStyle
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp

/**
 * Tokens come from the Claude Design handoff bundle (`tokens.jsx`).
 * Dark-first iOS aesthetic with the deep-amber accent the user landed
 * on after iterating in the design canvas. Colors map 1:1 to iOS
 * groupedBackground / secondarySystemGroupedBackground roles so this
 * theme matches the SwiftUI sibling app pixel-for-pixel.
 */
object NotifyColors {
    // Surfaces — mirror iosApp/Theme.swift dark palette
    val bg          = Color(0xFF0A0A0C)
    val bgElevated  = Color(0xFF141417)
    val surface     = Color(0xFF18181C)
    val surfaceHi   = Color(0xFF1F1F24)
    val surfaceMute = Color(0xFF101013)

    val separator   = Color(0x1FFFFFFF)   // strokeHi 12%
    val stroke      = Color(0x0FFFFFFF)   // hairline 6%
    val strokeHi    = Color(0x1FFFFFFF)   // 12%

    // White-tinted label hierarchy matches iOS Color.white.opacity(...)
    val label1      = Color(0xFFFFFFFF)
    val label2      = Color(0x9EFFFFFF)   // 62%
    val label3      = Color(0x61FFFFFF)   // 38%
    val label4      = Color(0x38FFFFFF)   // 22%

    val chipBg      = Color(0x14FFFFFF)   // 8% white

    // Accent — signal green
    val accent      = Color(0xFF3DD68C)
    val accentSoft  = Color(0x293DD68C)
    val accentGlow  = Color(0x733DD68C)
    val accentInk   = Color(0xFF062814)

    val danger      = Color(0xFFFF5D6E)
    val warn        = Color(0xFFFFCB6B)

    // Glass surface — bgElevated translucent
    val glassBg     = Color(0xB8141417)
    val glassBorder = Color(0x14FFFFFF)
}

object NotifyType {
    // Large title — 34/700/-1.0 from the design tokens
    val largeTitle = TextStyle(fontSize = 34.sp, fontWeight = FontWeight.Bold, letterSpacing = (-1.0).sp)
    val title1 = TextStyle(fontSize = 28.sp, fontWeight = FontWeight.Bold, letterSpacing = (-0.6).sp)
    val title2 = TextStyle(fontSize = 22.sp, fontWeight = FontWeight.SemiBold, letterSpacing = (-0.4).sp)
    val title3 = TextStyle(fontSize = 17.sp, fontWeight = FontWeight.SemiBold, letterSpacing = (-0.3).sp)
    val body   = TextStyle(fontSize = 15.sp, fontWeight = FontWeight.Normal)
    val bodyMed = TextStyle(fontSize = 15.sp, fontWeight = FontWeight.Medium, letterSpacing = (-0.1).sp)
    val caption = TextStyle(fontSize = 13.sp, fontWeight = FontWeight.Normal)
    val footnote = TextStyle(fontSize = 12.sp, fontWeight = FontWeight.Normal)
    // Section header eyebrow — uppercase 11/600/0.6 letter-spacing
    val eyebrow = TextStyle(fontSize = 11.sp, fontWeight = FontWeight.SemiBold, letterSpacing = 0.6.sp)
}

private val notifyShapes = Shapes(
    extraSmall = RoundedCornerShape(8.dp),
    small      = RoundedCornerShape(12.dp),
    medium     = RoundedCornerShape(14.dp),
    large      = RoundedCornerShape(16.dp),
    extraLarge = RoundedCornerShape(28.dp),
)

private val notifyScheme = darkColorScheme(
    primary = NotifyColors.accent,
    onPrimary = NotifyColors.accentInk,
    secondary = NotifyColors.accent,
    onSecondary = NotifyColors.accentInk,
    background = NotifyColors.bg,
    onBackground = NotifyColors.label1,
    surface = NotifyColors.surface,
    onSurface = NotifyColors.label1,
    surfaceVariant = NotifyColors.surfaceHi,
    onSurfaceVariant = NotifyColors.label2,
    error = NotifyColors.danger,
    onError = NotifyColors.label1,
    outline = NotifyColors.strokeHi,
    outlineVariant = NotifyColors.stroke,
)

@Composable
fun NotifyTheme(content: @Composable () -> Unit) {
    MaterialTheme(
        colorScheme = notifyScheme,
        typography = Typography(
            titleLarge = NotifyType.largeTitle,
            titleMedium = NotifyType.title2,
            titleSmall = NotifyType.title3,
            bodyLarge = NotifyType.bodyMed,
            bodyMedium = NotifyType.body,
            bodySmall = NotifyType.caption,
            labelSmall = NotifyType.eyebrow,
        ),
        shapes = notifyShapes,
        content = content,
    )
}
