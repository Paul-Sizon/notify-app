package com.notify.anything.notify

import androidx.compose.ui.unit.DpSize
import androidx.compose.ui.unit.dp
import androidx.compose.ui.window.Window
import androidx.compose.ui.window.application
import androidx.compose.ui.window.rememberWindowState
import com.notify.anything.notify.platform.JvmNotifier
import com.notify.anything.notify.platform.JvmPrefs
import com.notify.anything.notify.platform.JvmUrlOpener
import com.notify.anything.notify.platform.NoopHaptics

private const val DEFAULT_BASE_URL = "http://localhost:8080"

fun main() {
    System.setProperty("apple.awt.application.name", "notify")
    System.setProperty("apple.laf.useScreenMenuBar", "true")

    val prefs = JvmPrefs()
    val notifier = JvmNotifier()
    val urlOpener = JvmUrlOpener
    val haptics = NoopHaptics

    application {
        val state = rememberWindowState(
            size = DpSize(440.dp, 900.dp),
        )
        Window(
            onCloseRequest = ::exitApplication,
            title = "notify",
            state = state,
        ) {
            App(
                notifier = notifier,
                prefs = prefs,
                urlOpener = urlOpener,
                haptics = haptics,
                initialBaseUrl = prefs.getBaseUrl() ?: DEFAULT_BASE_URL,
            )
        }
    }
}
