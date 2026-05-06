package com.notify.anything.notify.platform

interface Notifier {
    fun deliver(title: String, body: String, subscriptionId: String, signalId: String)
}

interface Haptics {
    fun tap()
    fun tapMedium()
    fun selection()
    fun success()
    fun warning()
    fun error()
}

interface Prefs {
    fun getDeviceId(): String?
    fun putDeviceId(id: String?)
}

interface UrlOpener {
    fun open(url: String)
}

object NoopNotifier : Notifier {
    override fun deliver(title: String, body: String, subscriptionId: String, signalId: String) {}
}
object NoopHaptics : Haptics {
    override fun tap() {}
    override fun tapMedium() {}
    override fun selection() {}
    override fun success() {}
    override fun warning() {}
    override fun error() {}
}
object NoopUrlOpener : UrlOpener { override fun open(url: String) {} }
class InMemoryPrefs : Prefs {
    private var deviceId: String? = null
    override fun getDeviceId() = deviceId
    override fun putDeviceId(id: String?) { deviceId = id }
}
