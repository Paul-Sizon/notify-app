package com.notify.anything.notify

import androidx.compose.ui.window.Window
import androidx.compose.ui.window.application
import com.notify.anything.notify.platform.InMemoryPrefs
import com.notify.anything.notify.platform.NoopNotifier
import com.notify.anything.notify.platform.NoopUrlOpener

fun main() = application {
    Window(onCloseRequest = ::exitApplication, title = "notify") {
        App(
            notifier = NoopNotifier,
            prefs = InMemoryPrefs(),
            urlOpener = NoopUrlOpener,
            initialBaseUrl = "http://localhost:8080",
        )
    }
}
