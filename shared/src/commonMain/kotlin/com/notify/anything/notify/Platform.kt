package com.notify.anything.notify

interface Platform {
    val name: String
}

expect fun getPlatform(): Platform