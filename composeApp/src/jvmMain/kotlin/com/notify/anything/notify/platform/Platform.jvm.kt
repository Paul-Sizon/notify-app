package com.notify.anything.notify.platform

import java.awt.SystemTray
import java.awt.TrayIcon
import java.awt.Desktop
import java.awt.image.BufferedImage
import java.net.URI
import java.util.prefs.Preferences

class JvmPrefs(node: String = "com/notify/anything/notify") : Prefs {
    private val p: Preferences = Preferences.userRoot().node(node)
    override fun getDeviceId(): String? = p.get("device_id", null)
    override fun putDeviceId(id: String?) {
        if (id == null) p.remove("device_id") else p.put("device_id", id)
        p.flush()
    }
    override fun getBaseUrl(): String? = p.get("base_url", null)
    override fun putBaseUrl(url: String?) {
        if (url.isNullOrBlank()) p.remove("base_url") else p.put("base_url", url)
        p.flush()
    }
}

object JvmUrlOpener : UrlOpener {
    override fun open(url: String) {
        runCatching {
            if (Desktop.isDesktopSupported() && Desktop.getDesktop().isSupported(Desktop.Action.BROWSE)) {
                Desktop.getDesktop().browse(URI(url))
            }
        }
    }
}

/**
 * SystemTray-based notifier. macOS shows these as native banners through the
 * AWT bridge; Linux requires a tray daemon (libappindicator etc); Windows
 * shows balloon tips. On any failure we silently no-op so the UI never
 * blocks on the OS notification surface.
 */
class JvmNotifier : Notifier {
    private val trayIcon: TrayIcon? = runCatching {
        if (!SystemTray.isSupported()) return@runCatching null
        val tray = SystemTray.getSystemTray()
        val img: BufferedImage = BufferedImage(16, 16, BufferedImage.TYPE_INT_ARGB).also { bi ->
            val g = bi.createGraphics()
            g.color = java.awt.Color(0xE5, 0x5B, 0x3C)
            g.fillOval(2, 2, 12, 12)
            g.dispose()
        }
        val icon = TrayIcon(img, "notify")
        icon.isImageAutoSize = true
        tray.add(icon)
        icon
    }.getOrNull()

    override fun deliver(title: String, body: String, subscriptionId: String, signalId: String) {
        val icon = trayIcon ?: run { println("[notify] $title — $body"); return }
        runCatching { icon.displayMessage(title, body, TrayIcon.MessageType.INFO) }
    }
}
