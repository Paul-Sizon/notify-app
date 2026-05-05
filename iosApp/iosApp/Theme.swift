import SwiftUI

/// Design tokens — dark-first, signal-green accent.
/// Mirrors the web prototype's `accent: 'green'` default.
/// Light mode hooks present but unused at MVP — toggle later via Account.
enum Theme {
    // Surfaces
    static let bg          = Color(hex: 0x0A0A0C)
    static let bgElevated  = Color(hex: 0x141417)
    static let surface     = Color(hex: 0x18181C)
    static let surfaceHi   = Color(hex: 0x1F1F24)
    static let surfaceMute = Color(hex: 0x101013)

    // Strokes / dividers
    static let stroke      = Color.white.opacity(0.06)
    static let strokeHi    = Color.white.opacity(0.12)

    // Labels
    static let label1      = Color.white
    static let label2      = Color.white.opacity(0.62)
    static let label3      = Color.white.opacity(0.38)
    static let label4      = Color.white.opacity(0.22)

    // Accent — signal green
    static let accent      = Color(hex: 0x3DD68C)
    static let accentSoft  = Color(hex: 0x3DD68C).opacity(0.16)
    static let accentGlow  = Color(hex: 0x3DD68C).opacity(0.45)
    static let accentInk   = Color(hex: 0x062814)

    // Semantic
    static let danger      = Color(hex: 0xFF5D6E)
    static let warn        = Color(hex: 0xFFCB6B)

    // Radii
    static let rSmall: CGFloat = 8
    static let rMed:   CGFloat = 14
    static let rLg:    CGFloat = 20
    static let rXl:    CGFloat = 28
    static let rPill:  CGFloat = 999

    // Typography
    static func title1() -> Font  { .system(size: 30, weight: .semibold, design: .default) }
    static func title2() -> Font  { .system(size: 22, weight: .semibold) }
    static func title3() -> Font  { .system(size: 18, weight: .semibold) }
    static func body()   -> Font  { .system(size: 15, weight: .regular) }
    static func bodyMed() -> Font { .system(size: 15, weight: .medium) }
    static func caption() -> Font { .system(size: 13, weight: .regular) }
    static func eyebrow() -> Font { .system(size: 11, weight: .semibold).monospaced() }
}

extension Color {
    init(hex: UInt32, alpha: Double = 1.0) {
        let r = Double((hex >> 16) & 0xFF) / 255.0
        let g = Double((hex >>  8) & 0xFF) / 255.0
        let b = Double( hex        & 0xFF) / 255.0
        self.init(.sRGB, red: r, green: g, blue: b, opacity: alpha)
    }
}
