package com.notify.anything.notify.ui

import kotlinx.datetime.Clock
import kotlinx.datetime.DateTimeUnit
import kotlinx.datetime.Instant
import kotlinx.datetime.LocalDate
import kotlinx.datetime.TimeZone
import kotlinx.datetime.atStartOfDayIn
import kotlinx.datetime.daysUntil
import kotlinx.datetime.minus
import kotlinx.datetime.toLocalDateTime

/**
 * Server emits two shapes:
 *   - RFC3339 timestamps with optional fractional seconds.
 *   - YYYY-MM-DD for occurs_at on event signals.
 */
fun parseServerDate(s: String?): Instant? {
    if (s.isNullOrBlank()) return null
    return runCatching { Instant.parse(s) }.getOrNull()
        ?: runCatching {
            val ld = LocalDate.parse(s)
            ld.atStartOfDayIn(TimeZone.UTC)
        }.getOrNull()
}

fun relativeShort(t: Instant): String {
    val now = Clock.System.now()
    val secs = (now - t).inWholeSeconds
    return when {
        secs < 60 -> "JUST NOW"
        secs < 3600 -> "${secs / 60}M AGO"
        secs < 86400 -> "${secs / 3600}H AGO"
        else -> "${secs / 86400}D AGO"
    }
}

fun relativeLong(t: Instant): String {
    val now = Clock.System.now()
    val secs = (now - t).inWholeSeconds
    if (secs < 0) return "in the future"
    return when {
        secs < 60 -> "just now"
        secs < 3600 -> "${secs / 60}m ago"
        secs < 86400 -> "${secs / 3600}h ago"
        secs < 86400 * 7 -> "${secs / 86400}d ago"
        else -> "${secs / (86400 * 7)}w ago"
    }
}

fun formatEventDate(t: Instant): String {
    val tz = TimeZone.currentSystemDefault()
    val dt = t.toLocalDateTime(tz)
    val month = listOf("Jan","Feb","Mar","Apr","May","Jun","Jul","Aug","Sep","Oct","Nov","Dec")[dt.monthNumber - 1]
    return "$month ${dt.dayOfMonth}, ${dt.year}"
}

enum class DayBucket { TODAY, YESTERDAY, THIS_WEEK, EARLIER }

fun bucketOf(t: Instant): DayBucket {
    val tz = TimeZone.currentSystemDefault()
    val today = Clock.System.now().toLocalDateTime(tz).date
    val d = t.toLocalDateTime(tz).date
    val diff = d.daysUntil(today)
    return when {
        diff <= 0 -> DayBucket.TODAY
        diff == 1 -> DayBucket.YESTERDAY
        diff < 7 -> DayBucket.THIS_WEEK
        else -> DayBucket.EARLIER
    }
}

fun startOfTodayMinusDays(days: Int): Instant {
    val tz = TimeZone.currentSystemDefault()
    val today = Clock.System.now().toLocalDateTime(tz).date
    val target = today.minus(days, DateTimeUnit.DAY)
    return target.atStartOfDayIn(tz)
}

fun isSameDay(a: Instant, b: Instant): Boolean {
    val tz = TimeZone.currentSystemDefault()
    return a.toLocalDateTime(tz).date == b.toLocalDateTime(tz).date
}

fun cadenceLabel(cadenceSeconds: Int): String {
    val h = cadenceSeconds / 3600
    val m = (cadenceSeconds % 3600) / 60
    return when {
        h >= 24 -> "EVERY ${h / 24}D"
        h >= 1 -> if (m == 0) "EVERY ${h}H" else "EVERY ${h}H ${m}M"
        else -> "EVERY ${m}M"
    }
}
