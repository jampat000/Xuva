package com.xuva.tv.ui.theme

import androidx.compose.runtime.Composable
import androidx.compose.ui.graphics.Color
import androidx.tv.material3.ExperimentalTvMaterial3Api
import androidx.tv.material3.MaterialTheme
import androidx.tv.material3.darkColorScheme

private val XuvaDarkColors = darkColorScheme(
    primary = Color(0xFFB388FF),
    onPrimary = Color(0xFF1A0D33),
    background = Color(0xFF0B0B0F),
    onBackground = Color(0xFFEDEDED),
    surface = Color(0xFF15151B),
    onSurface = Color(0xFFEDEDED),
)

@OptIn(ExperimentalTvMaterial3Api::class)
@Composable
fun XuvaTVTheme(content: @Composable () -> Unit) {
    MaterialTheme(
        colorScheme = XuvaDarkColors,
        content = content,
    )
}
