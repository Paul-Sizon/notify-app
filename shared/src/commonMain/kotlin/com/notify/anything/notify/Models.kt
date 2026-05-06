package com.notify.anything.notify

import kotlinx.serialization.SerialName
import kotlinx.serialization.Serializable

@Serializable
data class RegisterDeviceRequest(
    @SerialName("apns_token") val apnsToken: String,
)

@Serializable
data class RegisterDeviceResponse(
    @SerialName("device_id") val deviceId: String,
)

@Serializable
data class CreateSubscriptionRequest(
    val query: String,
    val type: String,
    @SerialName("cadence_seconds") val cadenceSeconds: Int,
)

@Serializable
data class SubscriptionDTO(
    val id: String,
    val query: String,
    val type: String,
    @SerialName("cadence_seconds") val cadenceSeconds: Int,
    @SerialName("last_run_at") val lastRunAt: String? = null,
    @SerialName("next_run_at") val nextRunAt: String,
    @SerialName("created_at") val createdAt: String,
)

@Serializable
data class SignalDTO(
    val id: String,
    @SerialName("subscription_id") val subscriptionId: String,
    val title: String,
    val body: String? = null,
    val url: String? = null,
    @SerialName("occurs_at") val occursAt: String? = null,
    @SerialName("source_domains") val sourceDomains: List<String> = emptyList(),
    val confidence: Float,
    @SerialName("first_seen_at") val firstSeenAt: String,
)

@Serializable
data class RunResponse(
    @SerialName("new_signals") val newSignals: Int,
)

@Serializable
data class ErrorResponse(val error: String)

enum class SubscriptionType(val wire: String) {
    EVENT("event"),
    NEWS("news");

    companion object {
        fun fromWire(s: String): SubscriptionType = when (s) {
            "event" -> EVENT
            "news" -> NEWS
            else -> EVENT
        }
    }
}

// --- onboarding ---

@Serializable
data class OnboardingRequest(
    val city: String,
    val country: String,
    val role: String,
    @SerialName("role_other") val roleOther: String? = null,
    val interests: List<String>,
)

@Serializable
data class OnboardingSuggestion(
    val query: String,
    val type: String,
    @SerialName("cadence_seconds") val cadenceSeconds: Int,
    val reason: String,
)

@Serializable
data class OnboardingResponse(
    val suggestions: List<OnboardingSuggestion>,
    val fallback: Boolean = false,
)

@Serializable
data class ContextSuggestRequest(
    val context: String,
)
