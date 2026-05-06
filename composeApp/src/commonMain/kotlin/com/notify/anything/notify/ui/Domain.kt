package com.notify.anything.notify.ui

import com.notify.anything.notify.SignalDTO
import com.notify.anything.notify.SubscriptionDTO
import com.notify.anything.notify.SubscriptionType
import kotlinx.datetime.Clock
import kotlinx.datetime.Instant

data class Subscription(
    val id: String,
    val query: String,
    val type: SubscriptionType,
    val cadenceSeconds: Int,
    val lastRunAt: Instant?,
    val nextRunAt: Instant,
    val createdAt: Instant,
)

data class Signal(
    val id: String,
    val subscriptionId: String,
    val title: String,
    val body: String?,
    val url: String?,
    val occursAt: Instant?,
    val sourceDomains: List<String>,
    val confidence: Float,
    val firstSeenAt: Instant,
) {
    val isResolved: Boolean
        get() = occursAt?.let { it > Clock.System.now() } ?: false
}

fun SubscriptionDTO.toDomain(): Subscription = Subscription(
    id = id,
    query = query,
    type = SubscriptionType.fromWire(type),
    cadenceSeconds = cadenceSeconds,
    lastRunAt = parseServerDate(lastRunAt),
    nextRunAt = parseServerDate(nextRunAt) ?: Clock.System.now(),
    createdAt = parseServerDate(createdAt) ?: Clock.System.now(),
)

fun SignalDTO.toDomain(): Signal = Signal(
    id = id,
    subscriptionId = subscriptionId,
    title = title,
    body = body,
    url = url,
    occursAt = parseServerDate(occursAt),
    sourceDomains = sourceDomains,
    confidence = confidence,
    firstSeenAt = parseServerDate(firstSeenAt) ?: Clock.System.now(),
)
