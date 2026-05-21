import SwiftUI

public enum XuvaScale {
    public enum Platform {
        case tv, phone, pad
    }

    public static var platform: Platform {
        #if os(tvOS)
        return .tv
        #else
        return .phone
        #endif
    }

    public static var safeHorizontal: CGFloat {
        #if os(tvOS)
        return 96
        #else
        return 20
        #endif
    }

    public static var safeTop: CGFloat {
        #if os(tvOS)
        return 60
        #else
        return 20
        #endif
    }

    public static var sectionSpacing: CGFloat {
        #if os(tvOS)
        return 56
        #else
        return 28
        #endif
    }

    public static var rowSpacing: CGFloat {
        #if os(tvOS)
        return 26
        #else
        return 14
        #endif
    }

    public static var heroVerticalFraction: CGFloat {
        #if os(tvOS)
        return 0.78
        #else
        return 0.62
        #endif
    }

    public static func heroLogoMaxWidth(viewportWidth: CGFloat) -> CGFloat {
        #if os(tvOS)
        return min(720, viewportWidth * 0.4)
        #else
        return min(360, viewportWidth * 0.7)
        #endif
    }

    public static func heroLogoMaxHeight(viewportWidth: CGFloat) -> CGFloat {
        #if os(tvOS)
        return 200
        #else
        return viewportWidth < 700 ? 120 : 160
        #endif
    }

    public static func heroTitleSize(viewportWidth: CGFloat) -> CGFloat {
        #if os(tvOS)
        return 92
        #else
        return viewportWidth < 700 ? 44 : 64
        #endif
    }

    public static func sectionTitleSize() -> CGFloat {
        #if os(tvOS)
        return 38
        #else
        return 24
        #endif
    }

    public static func bodyFontSize() -> CGFloat {
        #if os(tvOS)
        return 28
        #else
        return 16
        #endif
    }

    public static func eyebrowFontSize() -> CGFloat {
        #if os(tvOS)
        return 18
        #else
        return 11
        #endif
    }

    public static func metaFontSize() -> CGFloat {
        #if os(tvOS)
        return 24
        #else
        return 14
        #endif
    }

    public static func posterWidth() -> CGFloat {
        #if os(tvOS)
        return 240
        #else
        return 132
        #endif
    }

    public static func posterHeight() -> CGFloat {
        posterWidth() * 1.5
    }

    public static func widePosterWidth() -> CGFloat {
        #if os(tvOS)
        return 420
        #else
        return 268
        #endif
    }

    public static func widePosterHeight() -> CGFloat {
        widePosterWidth() * 9 / 16
    }

    public static func navBarHeight() -> CGFloat {
        #if os(tvOS)
        return 96
        #else
        return 56
        #endif
    }

    public static func posterRowSpacing() -> CGFloat {
        #if os(tvOS)
        return 30
        #else
        return 16
        #endif
    }

    public static func detailPosterSize() -> CGSize {
        #if os(tvOS)
        return CGSize(width: 320, height: 480)
        #else
        return CGSize(width: 160, height: 240)
        #endif
    }

    public static func detailBackdropFraction() -> CGFloat {
        #if os(tvOS)
        return 0.70
        #else
        return 0.55
        #endif
    }
}
