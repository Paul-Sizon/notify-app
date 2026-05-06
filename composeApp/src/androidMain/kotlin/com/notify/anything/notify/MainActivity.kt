package com.notify.anything.notify

import android.Manifest
import android.content.pm.PackageManager
import android.os.Build
import android.os.Bundle
import androidx.activity.ComponentActivity
import androidx.activity.compose.setContent
import androidx.activity.enableEdgeToEdge
import androidx.activity.result.contract.ActivityResultContracts
import androidx.core.content.ContextCompat
import com.notify.anything.notify.platform.AndroidHaptics
import com.notify.anything.notify.platform.AndroidNotifier
import com.notify.anything.notify.platform.AndroidPrefs
import com.notify.anything.notify.platform.AndroidUrlOpener
import com.notify.anything.notify.platform.ANDROID_DEFAULT_BASE_URL

class MainActivity : ComponentActivity() {

    private val notifPermLauncher = registerForActivityResult(
        ActivityResultContracts.RequestPermission(),
    ) { /* Granted or not — we still launch UI either way. */ }

    override fun onCreate(savedInstanceState: Bundle?) {
        enableEdgeToEdge()
        super.onCreate(savedInstanceState)

        maybeRequestNotificationPermission()

        val notifier = AndroidNotifier(applicationContext, MainActivity::class.java)
        val prefs = AndroidPrefs(applicationContext)
        val opener = AndroidUrlOpener(applicationContext)
        val haptics = AndroidHaptics(applicationContext)

        setContent {
            App(
                notifier = notifier,
                prefs = prefs,
                urlOpener = opener,
                haptics = haptics,
                initialBaseUrl = ANDROID_DEFAULT_BASE_URL,
            )
        }
    }

    private fun maybeRequestNotificationPermission() {
        if (Build.VERSION.SDK_INT < Build.VERSION_CODES.TIRAMISU) return
        val granted = ContextCompat.checkSelfPermission(
            this, Manifest.permission.POST_NOTIFICATIONS,
        ) == PackageManager.PERMISSION_GRANTED
        if (!granted) notifPermLauncher.launch(Manifest.permission.POST_NOTIFICATIONS)
    }
}
