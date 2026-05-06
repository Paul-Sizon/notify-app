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
    // Surfaces — pure black ground, neutral grey elevations
    val bg          = Color(0xFF000000)
    val bgElevated  = Color(0xFF0E0E10)   // bgElev1 (sectioned background)
    val surface     = Color(0xFF1C1C1E)   // bgElev2 (cards)
    val surfaceHi   = Color(0xFF2C2C2E)   // bgElev3 (track / divider fills)
    val surfaceMute = Color(0xFF141416)   // resolved card

    val separator   = Color(0x73545458)   // 0.45 opacity grey
    val stroke      = Color(0x0FFFFFFF)   // hairline 6%
    val strokeHi    = Color(0x14FFFFFF)   // glassBorder 8%

    // iOS label hierarchy — pre-multiplied alpha against white
    val label1      = Color(0xFFFFFFFF)
    val label2      = Color(0xC7EBEBF5)
    val label3      = Color(0x80EBEBF5)
    val label4      = Color(0x4CEBEBF5)

    val chipBg      = Color(0x3D767680)   // systemGray translucent

    // Accent — deep amber per design handoff
    val accent      = Color(0xFFFF9F1C)
    val accentSoft  = Color(0x29FF9F1C)
    val accentGlow  = Color(0x73FF9F1C)
    val accentInk   = Color(0xFF000000)   // text on accent fills (dark mode = black)

    val danger      = Color(0xFFFF453A)
    val warn        = Color(0xFFFFCB6B)

    // Glass surface — used for FAB + agent inline cue
    val glassBg     = Color(0xB81C1C1E)
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
