package com.notify.anything.notify

import io.ktor.client.HttpClient
import io.ktor.client.HttpClientConfig
import io.ktor.client.call.body
import io.ktor.client.engine.HttpClientEngine
import io.ktor.client.plugins.HttpResponseValidator
import io.ktor.client.plugins.contentnegotiation.ContentNegotiation
import io.ktor.client.plugins.defaultRequest
import io.ktor.client.plugins.logging.Logger
import io.ktor.client.plugins.logging.Logging
import io.ktor.client.request.delete
import io.ktor.client.request.get
import io.ktor.client.request.header
import io.ktor.client.request.headers
import io.ktor.client.request.parameter
import io.ktor.client.request.post
import io.ktor.client.request.setBody
import io.ktor.client.statement.bodyAsText
import io.ktor.http.ContentType
import io.ktor.http.HttpStatusCode
import io.ktor.http.URLProtocol
import io.ktor.http.contentType
import io.ktor.http.isSuccess
import io.ktor.serialization.kotlinx.json.json
import kotlinx.serialization.json.Json

/**
 * KMP HTTP client for the notify Go backend.
 *
 * Mirrors `server/e2e/client.go` route surface so behavior parity is easy to
 * verify between Swift integration and Go E2E tests.
 *
 * Auth model: a device registers once (returns deviceId), then every
 * subsequent call carries `X-Device-Id`. Persist deviceId on the iOS side
 * (UserDefaults) — losing it = orphaned subscriptions.
 */
class ApiClient(
    private val baseUrl: String,
    engine: HttpClientEngine? = null,
    enableLogging: Boolean = false,
) {
    private val json = Json {
        ignoreUnknownKeys = true
        explicitNulls = false
        encodeDefaults = true
    }

    private val http: HttpClient = run {
        val cfg: HttpClientConfig<*>.() -> Unit = {
            install(ContentNegotiation) { json(this@ApiClient.json) }
            if (enableLogging) {
                install(Logging) {
                    logger = object : Logger {
                        override fun log(message: String) { println("ktor: $message") }
                    }
                }
            }
            defaultRequest {
                contentType(ContentType.Application.Json)
                if (baseUrl.startsWith("https://")) url.protocol = URLProtocol.HTTPS
            }
            HttpResponseValidator {
                validateResponse { resp ->
                    if (!resp.status.isSuccess()) {
                        val msg = runCatching { resp.bodyAsText() }.getOrDefault("")
                        throw ApiError(resp.status.value, msg.ifBlank { resp.status.description })
                    }
                }
            }
        }
        if (engine != null) HttpClient(engine, cfg) else HttpClient(cfg)
    }

    var deviceId: String? = null

    private fun url(path: String) = "$baseUrl$path"

    @Throws(Throwable::class)
    suspend fun registerDevice(apnsToken: String): String {
        val resp: RegisterDeviceResponse = http.post(url("/v1/devices")) {
            setBody(RegisterDeviceRequest(apnsToken))
        }.body()
        deviceId = resp.deviceId
        return resp.deviceId
    }

    @Throws(Throwable::class)
    suspend fun listSubscriptions(): List<SubscriptionDTO> =
        http.get(url("/v1/subscriptions")) { authHeader() }.body()

    @Throws(Throwable::class)
    suspend fun createSubscription(
        query: String,
        type: SubscriptionType,
        cadenceSeconds: Int,
    ): SubscriptionDTO = http.post(url("/v1/subscriptions")) {
        authHeader()
        setBody(CreateSubscriptionRequest(query, type.wire, cadenceSeconds))
    }.body()

    @Throws(Throwable::class)
    suspend fun deleteSubscription(id: String) {
        http.delete(url("/v1/subscriptions/$id")) { authHeader() }
    }

    @Throws(Throwable::class)
    suspend fun runSubscription(id: String): RunResponse =
        http.post(url("/v1/subscriptions/$id/run")) { authHeader() }.body()

    @Throws(Throwable::class)
    suspend fun listSignals(subscriptionId: String, limit: Int = 50, before: String? = null): List<SignalDTO> =
        http.get(url("/v1/subscriptions/$subscriptionId/signals")) {
            authHeader()
            parameter("limit", limit)
            if (before != null) parameter("before", before)
        }.body()

    private fun io.ktor.client.request.HttpRequestBuilder.authHeader() {
        val id = deviceId ?: error("ApiClient.deviceId not set — call registerDevice() first")
        headers { append("X-Device-Id", id) }
    }

    fun close() = http.close()
}

class ApiError(val status: Int, message: String) : RuntimeException("[$status] $message")
