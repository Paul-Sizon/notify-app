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

object NotifyColors {
    val bg          = Color(0xFF0A0A0C)
    val bgElevated  = Color(0xFF141417)
    val surface     = Color(0xFF18181C)
    val surfaceHi   = Color(0xFF1F1F24)
    val surfaceMute = Color(0xFF101013)

    val stroke      = Color(0x0FFFFFFF)
    val strokeHi    = Color(0x1FFFFFFF)

    val label1      = Color(0xFFFFFFFF)
    val label2      = Color(0x9EFFFFFF)
    val label3      = Color(0x61FFFFFF)
    val label4      = Color(0x38FFFFFF)

    val accent      = Color(0xFF3DD68C)
    val accentSoft  = Color(0x293DD68C)
    val accentGlow  = Color(0x733DD68C)
    val accentInk   = Color(0xFF062814)

    val danger      = Color(0xFFFF5D6E)
    val warn        = Color(0xFFFFCB6B)
}

object NotifyType {
    val title1 = TextStyle(fontSize = 30.sp, fontWeight = FontWeight.SemiBold, letterSpacing = (-0.4).sp)
    val title2 = TextStyle(fontSize = 22.sp, fontWeight = FontWeight.SemiBold)
    val title3 = TextStyle(fontSize = 18.sp, fontWeight = FontWeight.SemiBold)
    val body   = TextStyle(fontSize = 15.sp, fontWeight = FontWeight.Normal)
    val bodyMed = TextStyle(fontSize = 15.sp, fontWeight = FontWeight.Medium)
    val caption = TextStyle(fontSize = 13.sp, fontWeight = FontWeight.Normal)
    val eyebrow = TextStyle(fontSize = 11.sp, fontWeight = FontWeight.SemiBold, letterSpacing = 1.5.sp)
}

private val notifyShapes = Shapes(
    extraSmall = RoundedCornerShape(8.dp),
    small      = RoundedCornerShape(12.dp),
    medium     = RoundedCornerShape(14.dp),
    large      = RoundedCornerShape(20.dp),
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
            titleLarge = NotifyType.title1,
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
