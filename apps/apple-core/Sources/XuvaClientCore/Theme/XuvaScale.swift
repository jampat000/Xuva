import SwiftUI

/// Viewport-relative sizing tokens that scale fluidly across iPhone, iPad, Mac, and tvOS.
///
/// Desktop uses `clamp(min, vw, max)` to keep typography readable at every width;
/// we mirror that here with a tiny `clamped(...)` helper. Pass the actual
/// `GeometryProxy.size` in and every screen consumes the same formula — there
/// is no separate "tvOS" or "iOS" branch for layout values.
public enum XuvaScale {
    public enum Platform {
        case tv, phone, pad

        public static var current: Platform {
            #if os(tvOS)
            return .tv
            #else
            #if targetEnvironment(macCatalyst)
            return .pad
            #else
            if UIScreen.main.traitCollection.userInterfaceIdiom == .pad {
                return .pad
            }
            return .phone
            #endif
            #endif
        }
    }

    public static var platform: Platform { Platform.current }

    /// Best-effort current viewport size. Use this only inside leaf views
    /// (button styles, poster tiles) that don't have their own GeometryReader.
    /// Top-level screens should read `GeometryReader.size` and pass it down.
    public static var screenSize: CGSize {
        #if os(tvOS)
        return UIScreen.main.bounds.size
        #else
        return UIScreen.main.bounds.size
        #endif
    }

    /// Fluid clamp: `clamp(min, base + slope*width, max)` — same pattern as
    /// CSS `clamp(min, vw, max)` but linear instead of strictly viewport-vw.
    public static func clamped(_ minimum: CGFloat, _ value: CGFloat, _ maximum: CGFloat) -> CGFloat {
        return max(minimum, min(maximum, value))
    }

    // ─── Padding & gutters ────────────────────────────────────────────────────
    public static func safeHorizontal(_ size: CGSize) -> CGFloat {
        // 5.5% of width, between 16pt (iPhone) and 120pt (4K Apple TV).
        return clamped(16, size.width * 0.055, 120)
    }

    public static func safeTop(_ size: CGSize) -> CGFloat {
        return clamped(12, size.height * 0.035, 72)
    }

    public static func sectionSpacing(_ size: CGSize) -> CGFloat {
        // Reduced from 0.055/72 — 59pt gaps between rows felt cavernous on tvOS.
        // Web equivalent is ~32px between rows; 0.038/48 gives ~41pt on 1080p TV.
        return clamped(20, size.height * 0.038, 48)
    }

    public static func rowSpacing(_ size: CGSize) -> CGFloat {
        // Gap between section title and poster rail. Web uses ~20px; 0.018/24
        // gives ~19pt on tvOS vs the old 27pt.
        return clamped(10, size.height * 0.018, 24)
    }

    public static func posterRowSpacing(_ size: CGSize) -> CGFloat {
        return clamped(12, size.width * 0.015, 36)
    }

    // ─── Hero proportions ────────────────────────────────────────────────────
    /// How tall the hero/backdrop should be relative to viewport height.
    public static func heroVerticalFraction(_ size: CGSize) -> CGFloat {
        // 72-80vh on desktop, similar on iPad/tvOS, 62vh on phone.
        if size.width < 600 { return 0.65 }
        if size.width < 1200 { return 0.78 }
        return 0.80
    }

    /// Where the hero text/CTA block sits within the hero (from top).
    public static func heroContentTopFraction(_ size: CGSize) -> CGFloat {
        if size.width < 600 { return 0.22 }
        return 0.32
    }

    public static func heroContentMaxWidth(_ size: CGSize) -> CGFloat {
        // Content column fills 60-72% of viewport; never wider than 1100pt.
        let target = size.width < 600 ? size.width * 0.92 : size.width * 0.62
        return min(target, 1100)
    }

    // ─── Hero logo / title ───────────────────────────────────────────────────
    public static func heroLogoMaxWidth(_ size: CGSize) -> CGFloat {
        return clamped(220, size.width * 0.32, 760)
    }

    public static func heroLogoMaxHeight(_ size: CGSize) -> CGFloat {
        return clamped(80, size.height * 0.18, 220)
    }

    public static func heroTitleSize(_ size: CGSize) -> CGFloat {
        // ~7vw, clamped between 38 (iPhone) and 110 (TV).
        return clamped(38, size.width * 0.058, 110)
    }

    // ─── Typography ──────────────────────────────────────────────────────────
    public static func sectionTitleSize(_ size: CGSize) -> CGFloat {
        return clamped(20, size.width * 0.022, 42)
    }

    public static func bodyFontSize(_ size: CGSize) -> CGFloat {
        return clamped(15, size.width * 0.016, 30)
    }

    public static func metaFontSize(_ size: CGSize) -> CGFloat {
        return clamped(13, size.width * 0.014, 26)
    }

    public static func eyebrowFontSize(_ size: CGSize) -> CGFloat {
        return clamped(10, size.width * 0.010, 18)
    }

    // ─── Poster sizes ────────────────────────────────────────────────────────
    public static func posterWidth(_ size: CGSize) -> CGFloat {
        return clamped(108, size.width * 0.115, 260)
    }

    /// Column count for the Movies / TV LibraryGrid. Mirrors the web's
    /// `grid-cols-3 sm:grid-cols-4 md:grid-cols-6 lg:grid-cols-8 xl:grid-cols-9`.
    /// On a 4K Apple TV (1920×1080 logical) this lands at 7 columns.
    public static func libraryGridColumnCount(_ size: CGSize) -> Int {
        switch size.width {
        case ..<640:  return 3
        case ..<900:  return 4
        case ..<1280: return 5
        case ..<1600: return 6
        case ..<1920: return 7
        default:      return 8
        }
    }

    public static func posterHeight(_ size: CGSize) -> CGFloat {
        return posterWidth(size) * 1.5
    }

    public static func widePosterWidth(_ size: CGSize) -> CGFloat {
        return clamped(220, size.width * 0.20, 460)
    }

    public static func widePosterHeight(_ size: CGSize) -> CGFloat {
        return widePosterWidth(size) * 9 / 16
    }

    // ─── Chrome ──────────────────────────────────────────────────────────────
    public static func navBarHeight(_ size: CGSize) -> CGFloat {
        return clamped(52, size.height * 0.07, 110)
    }

    public static func buttonHeight(_ size: CGSize) -> CGFloat {
        return clamped(44, size.width * 0.045, 80)
    }

    public static func buttonHorizontalPadding(_ size: CGSize) -> CGFloat {
        return clamped(18, size.width * 0.022, 40)
    }

    public static func buttonFontSize(_ size: CGSize) -> CGFloat {
        return clamped(15, size.width * 0.016, 28)
    }

    public static func iconButtonSize(_ size: CGSize) -> CGFloat {
        return clamped(38, size.width * 0.038, 64)
    }

    // ─── Detail screen ───────────────────────────────────────────────────────
    public static func detailBackdropFraction(_ size: CGSize) -> CGFloat {
        if size.width < 600 { return 0.58 }
        return 0.72
    }
}
