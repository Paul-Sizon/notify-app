package com.notify.anything.notify.ui

import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.setValue
import com.notify.anything.notify.ApiClient
import com.notify.anything.notify.SubscriptionType
import com.notify.anything.notify.platform.Notifier
import com.notify.anything.notify.platform.Prefs
import kotlinx.coroutines.CoroutineScope
import kotlinx.coroutines.async
import kotlinx.coroutines.awaitAll
import kotlinx.coroutines.delay
import kotlinx.coroutines.launch

enum class Tab { WATCHERS, ALERTS, SIGNALS, ACCOUNT }

/**
 * Mirrors the iOS `AppState`. Single owner of API client + observable
 * subscription/signal lists. Compose `mutableStateOf` triggers recomposition
 * when state changes — equivalent to SwiftUI's `@Observable`.
 */
class AppState(
    val scope: CoroutineScope,
    private val prefs: Prefs,
    private val notifier: Notifier,
    val defaultBaseUrl: String,
) {
    var baseUrl: String by mutableStateOf(prefs.getBaseUrl() ?: defaultBaseUrl)
        private set

    private var api: ApiClient = ApiClient(baseUrl)

    var deviceId: String? by mutableStateOf(prefs.getDeviceId()?.also { api.deviceId = it })
        private set

    var subscriptions: List<Subscription> by mutableStateOf(emptyList())
        private set

    var signalsBySub: Map<String, List<Signal>> by mutableStateOf(emptyMap())
        private set

    private val seenSignalIds: MutableSet<String> = mutableSetOf()

    var loading: Boolean by mutableStateOf(false)
        private set

    var lastError: String? by mutableStateOf(null)
    var toast: String? by mutableStateOf(null)

    var selectedTab: Tab by mutableStateOf(Tab.WATCHERS)
    var liveAgentSubId: String? by mutableStateOf(null)
        private set

    var detailSubscriptionId: String? by mutableStateOf(null)
    var showAddSheet: Boolean by mutableStateOf(false)

    fun bootstrap() {
        scope.launch {
            try {
                if (api.deviceId == null) {
                    val id = api.registerDevice("android-mock-${randomToken()}")
                    deviceId = id
                    prefs.putDeviceId(id)
                }
                refresh()
            } catch (t: Throwable) {
                lastError = "bootstrap: ${t.message ?: t.toString()}"
            }
        }
    }

    fun refresh() {
        scope.launch {
            loading = true
            try {
                val subs = api.listSubscriptions().map { it.toDomain() }
                    .sortedByDescending { it.createdAt }
                subscriptions = subs
                val pairs = subs.map { sub ->
                    async {
                        sub.id to api.listSignals(sub.id, limit = 30).map { it.toDomain() }
                    }
                }.awaitAll()
                val newMap = pairs.toMap()
                for ((_, sigs) in newMap) for (s in sigs) seenSignalIds.add(s.id)
                signalsBySub = newMap
            } catch (t: Throwable) {
                lastError = "refresh: ${t.message ?: t.toString()}"
            } finally {
                loading = false
            }
        }
    }

    fun create(query: String, type: SubscriptionType, cadenceSeconds: Int) {
        scope.launch {
            try {
                val s = api.createSubscription(query, type, cadenceSeconds).toDomain()
                subscriptions = listOf(s) + subscriptions
                signalsBySub = signalsBySub + (s.id to emptyList())
                toast = "Watcher added: ${s.query}"
                refresh()
            } catch (t: Throwable) {
                lastError = "create: ${t.message ?: t.toString()}"
            }
        }
    }

    fun delete(sub: Subscription) {
        scope.launch {
            try {
                api.deleteSubscription(sub.id)
                subscriptions = subscriptions.filterNot { it.id == sub.id }
                signalsBySub = signalsBySub - sub.id
            } catch (t: Throwable) {
                lastError = "delete: ${t.message ?: t.toString()}"
            }
        }
    }

    fun run(sub: Subscription) {
        scope.launch {
            liveAgentSubId = sub.id
            try {
                api.runSubscription(sub.id)
                val sigs = api.listSignals(sub.id, limit = 30).map { it.toDomain() }
                val newOnes = sigs.filter { it.id !in seenSignalIds }
                signalsBySub = signalsBySub + (sub.id to sigs)
                for (s in newOnes) {
                    seenSignalIds.add(s.id)
                    notifier.deliver(
                        title = sub.query,
                        body = s.title,
                        subscriptionId = sub.id,
                        signalId = s.id,
                    )
                }
                delay(600)
            } catch (t: Throwable) {
                lastError = "run: ${t.message ?: t.toString()}"
            } finally {
                liveAgentSubId = null
            }
        }
    }

    fun switchBackend(url: String) {
        scope.launch {
            try {
                api.close()
            } catch (_: Throwable) {}
            baseUrl = url
            prefs.putBaseUrl(url)
            prefs.putDeviceId(null)
            api = ApiClient(url)
            deviceId = null
            subscriptions = emptyList()
            signalsBySub = emptyMap()
            seenSignalIds.clear()
            toast = "Backend → $url"
            bootstrap()
        }
    }

    fun forgetDevice() {
        scope.launch {
            prefs.putDeviceId(null)
            deviceId = null
            api.deviceId = null
            bootstrap()
        }
    }

    fun signals(subId: String): List<Signal> = signalsBySub[subId] ?: emptyList()

    fun confirmedDate(sub: Subscription): Instant? =
        signals(sub.id).mapNotNull { it.occursAt }
            .filter { it > kotlinx.datetime.Clock.System.now() }
            .minOrNull()

    val resolvedSubscriptions: List<Subscription>
        get() = subscriptions.filter { sub ->
            signals(sub.id).any { it.isResolved }
        }

    val activeSubscriptions: List<Subscription>
        get() {
            val resolved = resolvedSubscriptions.map { it.id }.toSet()
            return subscriptions.filterNot { it.id in resolved }
        }

    val allSignalsRecent: List<Pair<Signal, Subscription>>
        get() = subscriptions.flatMap { sub ->
            signals(sub.id).map { it to sub }
        }.sortedByDescending { it.first.firstSeenAt }
}

private fun randomToken(): String {
    val chars = "abcdef0123456789"
    return (1..16).map { chars.random() }.joinToString("")
}

// Re-expose for callers that don't import kotlinx.datetime
typealias Instant = kotlinx.datetime.Instant
