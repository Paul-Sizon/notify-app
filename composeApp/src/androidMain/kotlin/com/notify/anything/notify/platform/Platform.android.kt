package com.notify.anything.notify.platform

import android.app.NotificationChannel
import android.app.NotificationManager
import android.app.PendingIntent
import android.content.Context
import android.content.Intent
import android.content.SharedPreferences
import android.net.Uri
import android.os.Build
import android.os.VibrationEffect
import android.os.Vibrator
import android.os.VibratorManager
import androidx.core.app.NotificationCompat
import androidx.core.content.ContextCompat

const val CHANNEL_ID = "signals"

class AndroidNotifier(private val ctx: Context, private val launcher: Class<*>) : Notifier {
    init { ensureChannel(ctx) }

    override fun deliver(title: String, body: String, subscriptionId: String, signalId: String) {
        val intent = Intent(ctx, launcher).apply {
            flags = Intent.FLAG_ACTIVITY_SINGLE_TOP
            putExtra("subscription_id", subscriptionId)
            putExtra("signal_id", signalId)
        }
        val pi = PendingIntent.getActivity(
            ctx, signalId.hashCode(), intent,
            PendingIntent.FLAG_UPDATE_CURRENT or PendingIntent.FLAG_IMMUTABLE,
        )
        val n = NotificationCompat.Builder(ctx, CHANNEL_ID)
            .setSmallIcon(android.R.drawable.ic_dialog_info)
            .setContentTitle(title)
            .setContentText(body)
            .setStyle(NotificationCompat.BigTextStyle().bigText(body))
            .setAutoCancel(true)
            .setContentIntent(pi)
            .setGroup(subscriptionId)
            .setPriority(NotificationCompat.PRIORITY_HIGH)
            .build()
        val mgr = ContextCompat.getSystemService(ctx, NotificationManager::class.java) ?: return
        mgr.notify(signalId.hashCode(), n)
    }
}

private fun ensureChannel(ctx: Context) {
    if (Build.VERSION.SDK_INT < Build.VERSION_CODES.O) return
    val mgr = ContextCompat.getSystemService(ctx, NotificationManager::class.java) ?: return
    if (mgr.getNotificationChannel(CHANNEL_ID) != null) return
    val ch = NotificationChannel(CHANNEL_ID, "Signals", NotificationManager.IMPORTANCE_HIGH).apply {
        description = "New signals from your watchers"
        enableVibration(true)
    }
    mgr.createNotificationChannel(ch)
}

class AndroidHaptics(private val ctx: Context) : Haptics {
    private fun vibrator(): Vibrator? {
        return if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.S) {
            (ctx.getSystemService(Context.VIBRATOR_MANAGER_SERVICE) as? VibratorManager)?.defaultVibrator
        } else {
            @Suppress("DEPRECATION")
            ctx.getSystemService(Context.VIBRATOR_SERVICE) as? Vibrator
        }
    }
    private fun pulse(ms: Long, amplitude: Int) {
        val v = vibrator() ?: return
        if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.O) {
            v.vibrate(VibrationEffect.createOneShot(ms, amplitude))
        } else {
            @Suppress("DEPRECATION") v.vibrate(ms)
        }
    }
    override fun tap() = pulse(8, 60)
    override fun tapMedium() = pulse(14, 120)
    override fun selection() = pulse(6, 40)
    override fun success() {
        pulse(10, 100); pulse(20, 180)
    }
    override fun warning() = pulse(30, 200)
    override fun error() = pulse(50, 220)
}

class AndroidPrefs(ctx: Context) : Prefs {
    private val sp: SharedPreferences = ctx.getSharedPreferences("notify_prefs", Context.MODE_PRIVATE)
    override fun getDeviceId(): String? = sp.getString("device_id", null)
    override fun putDeviceId(id: String?) {
        sp.edit().apply { if (id == null) remove("device_id") else putString("device_id", id) }.apply()
    }
    override fun getBaseUrl(): String? = sp.getString("base_url", null)
    override fun putBaseUrl(url: String?) {
        sp.edit().apply { if (url.isNullOrBlank()) remove("base_url") else putString("base_url", url) }.apply()
    }
}

class AndroidUrlOpener(private val ctx: Context) : UrlOpener {
    override fun open(url: String) {
        val intent = Intent(Intent.ACTION_VIEW, Uri.parse(url)).apply {
            addFlags(Intent.FLAG_ACTIVITY_NEW_TASK)
        }
        runCatching { ctx.startActivity(intent) }
    }
}

/**
 * 10.0.2.2 is the Android emulator's loopback alias to the host machine.
 * Physical devices need the cloudflared tunnel — switch via Account tab.
 */
const val ANDROID_DEFAULT_BASE_URL = "http://10.0.2.2:8080"
