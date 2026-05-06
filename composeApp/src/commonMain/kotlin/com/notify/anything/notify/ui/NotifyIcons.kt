package com.notify.anything.notify.ui

import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.Add
import androidx.compose.material.icons.filled.ArrowForward
import androidx.compose.material.icons.filled.Check
import androidx.compose.material.icons.filled.CheckCircle
import androidx.compose.material.icons.filled.Clear
import androidx.compose.material.icons.filled.Close
import androidx.compose.material.icons.filled.DateRange
import androidx.compose.material.icons.filled.Delete
import androidx.compose.material.icons.filled.Info
import androidx.compose.material.icons.filled.KeyboardArrowRight
import androidx.compose.material.icons.filled.MoreVert
import androidx.compose.material.icons.filled.Notifications
import androidx.compose.material.icons.filled.Person
import androidx.compose.material.icons.filled.Phone
import androidx.compose.material.icons.filled.PlayArrow
import androidx.compose.material.icons.filled.Refresh
import androidx.compose.material.icons.filled.Search
import androidx.compose.material.icons.filled.Settings
import androidx.compose.material.icons.filled.Star
import androidx.compose.material.icons.filled.Warning

/**
 * Compose-Multiplatform 1.10 doesn't ship `material-icons-extended`, so we
 * back app-specific icon names with the closest core glyph. UI design
 * tradeoff: a few icons are slightly less semantic (Star instead of
 * `GraphicEq` for the waveform indicator), but the core set ships with the
 * compiler so no extra artifact resolution.
 */
object NotifyIcons {
    val Add = Icons.Filled.Add
    val ArrowForward = Icons.Filled.ArrowForward
    val AutoAwesome = Icons.Filled.Star
    val CalendarMonth = Icons.Filled.DateRange
    val Check = Icons.Filled.Check
    val CheckCircle = Icons.Filled.CheckCircle
    val ChevronRight = Icons.Filled.KeyboardArrowRight
    val Clear = Icons.Filled.Clear
    val Close = Icons.Filled.Close
    val Delete = Icons.Filled.Delete
    val DeleteForever = Icons.Filled.Delete
    val Description = Icons.Filled.Info
    val Error = Icons.Filled.Warning
    val GraphicEq = Icons.Filled.Star
    val Info = Icons.Filled.Info
    val MoreHoriz = Icons.Filled.MoreVert
    val Newspaper = Icons.Filled.Info
    val Notifications = Icons.Filled.Notifications
    val OpenInNew = Icons.Filled.ArrowForward
    val Person = Icons.Filled.Person
    val PhoneAndroid = Icons.Filled.Phone
    val PlayArrow = Icons.Filled.PlayArrow
    val Refresh = Icons.Filled.Refresh
    val Router = Icons.Filled.Settings
    val Schedule = Icons.Filled.Refresh
    val Search = Icons.Filled.Search
    val Settings = Icons.Filled.Settings
    val Sync = Icons.Filled.Refresh
    val Visibility = Icons.Filled.Search
    val AccountCircle = Icons.Filled.Person
}
