/**
 * Utilities to detect the browser and get information on the current environment
 * Based on https://github.com/google/shaka-player/blob/master/lib/util/platform.js
 *
 * @deprecated - Parsing User Agent is a maintenance burden and
 * should rely on external libraries only. It's also going to be replaced with Client-Hints.
 * Migration paths:
 * * Check for platform-specific features *where needed*
 * directly (i.e. Chromecast/AirPlay/MSE) instead of a per-browser basis.
 * This will always be 100% fault free.
 *
 * * Use something like https://www.npmjs.com/package/unique-names-generator to
 * distinguish between instances. Instance names could be shown and be modified by the user
 * at settings. This would make user instances distinguishable in a 100% fault-tolerant way
 * and solve incongruences like how a device is named. For instance,
 * an instance running in an Android Auto headset will be recognised as Android only, which is less
 * than ideal.
 */
export function supportsMediaSource(): boolean {
    /*
     * Browsers that lack a media source implementation will have no reference
     * to |window.MediaSource|.
     */
    return !!window.MediaSource
}

/**
 * Check if the user agent of the navigator contains a key.
 *
 * @private
 * @static
 * @param key - Key for which to perform a check.
 * @returns Determines if user agent of navigator contains a key
 */
function userAgentContains(key: string): boolean {
    const userAgent = navigator.userAgent || ""

    return userAgent.includes(key)
}

/* Desktop Browsers */

/**
 * Check if the current platform is Mozilla Firefox.
 *
 * @returns Determines if browser is Mozilla Firefox
 */
export function isFirefox(): boolean {
    return userAgentContains("Firefox/")
}

/**
 * Check if the current platform is Microsoft Edge.
 *
 * @static
 * @returns Determines if browser is Microsoft Edge
 */
export function isEdge(): boolean {
    return userAgentContains("Edg/") || userAgentContains("Edge/")
}

/**
 * Check if the current platform is Chromium based.
 *
 * @returns Determines if browser is Chromium based
 */
export function isChromiumBased(): boolean {
    return userAgentContains("Chrome")
}

/**
 * Check if the current platform is Google Chrome.
 *
 * @returns Determines if browser is Google Chrome
 */
export function isChrome(): boolean {
    /*
     * The Edge user agent will also contain the "Chrome" keyword, so we need
     * to make sure this is not Edge.
     */
    return userAgentContains("Chrome") && !isEdge() && !isWebOS()
}

/**
 * Check if the current platform is from Apple.
 *
 * Returns true on all iOS browsers and on desktop Safari.
 *
 * Returns false for non-Safari browsers on macOS, which are independent of
 * Apple.
 *
 * @returns Determines if current platform is from Apple
 */
export function isApple(): boolean {
    return navigator.vendor.includes("Apple") && !isTizen()
}

/**
 * Returns a major version number for Safari, or Safari-based iOS browsers.
 *
 * @returns The major version number for Safari
 */
export function safariVersion(): number | undefined {
    // All iOS browsers and desktop Safari will return true for isApple().
    if (!isApple()) {
        return
    }

    let userAgent = ""

    if (navigator.userAgent) {
        userAgent = navigator.userAgent
    }

    /*
     * This works for iOS Safari and desktop Safari, which contain something
     * like "Version/13.0" indicating the major Safari or iOS version.
     */
    let match = /Version\/(\d+)/.exec(userAgent)

    if (match) {
        return Number.parseInt(match[1], /* Base= */ 10)
    }

    /*
     * This works for all other browsers on iOS, which contain something like
     * "OS 13_3" indicating the major & minor iOS version.
     */
    match = /OS (\d+)(?:_\d+)?/.exec(userAgent)

    if (match) {
        return Number.parseInt(match[1], /* Base= */ 10)
    }
}

/* TV Platforms */

/**
 * Check if the current platform is Tizen.
 *
 * @returns Determines if current platform is Tizen
 */
export function isTizen(): boolean {
    return userAgentContains("Tizen")
}

/**
 * Check if the current platform is Tizen 2
 *
 * @returns Determines if current platform is Tizen 2
 */
export function isTizen2(): boolean {
    return userAgentContains("Tizen 2")
}

/**
 * Check if the current platform is Tizen 3
 *
 * @returns Determines if current platform is Tizen 3
 * @memberof BrowserDetector
 */
export function isTizen3(): boolean {
    return userAgentContains("Tizen 3")
}

/**
 * Check if the current platform is Tizen 4.
 *
 * @returns Determines if current platform is Tizen 4
 * @memberof BrowserDetector
 */
export function isTizen4(): boolean {
    return userAgentContains("Tizen 4")
}

/**
 * Check if the current platform is Tizen 5.
 *
 * @returns Determines if current platform is Tizen 5
 * @memberof BrowserDetector
 */
export function isTizen5(): boolean {
    return userAgentContains("Tizen 5")
}

/**
 * Check if the current platform is Tizen 5.5.
 *
 * @returns Determines if current platform is Tizen 5.5
 * @memberof BrowserDetector
 */
export function isTizen55(): boolean {
    return userAgentContains("Tizen 5.5")
}

/**
 * Check if the current platform is WebOS.
 *
 * @returns Determines if current platform is WebOS
 * @memberof BrowserDetector
 */
export function isWebOS(): boolean {
    return userAgentContains("Web0S")
}

/**
 * Determines if current platform is WebOS1
 */
export function isWebOS1(): boolean {
    return (
        isWebOS()
        && userAgentContains("AppleWebKit/537")
        && !userAgentContains("Chrome/")
    )
}

/**
 * Determines if current platform is WebOS2
 */
export function isWebOS2(): boolean {
    return (
        isWebOS()
        && userAgentContains("AppleWebKit/538")
        && !userAgentContains("Chrome/")
    )
}

/**
 * Determines if current platform is WebOS3
 */
export function isWebOS3(): boolean {
    return isWebOS() && userAgentContains("Chrome/38")
}

/**
 * Determines if current platform is WebOS4
 */
export function isWebOS4(): boolean {
    return isWebOS() && userAgentContains("Chrome/53")
}

/**
 * Determines if current platform is WebOS5
 */
export function isWebOS5(): boolean {
    return isWebOS() && userAgentContains("Chrome/68")
}

/* Platform Utilities */

/**
 * Determines if current platform is Android
 */
export function isAndroid(): boolean {
    return userAgentContains("Android")
}

/**
 * Guesses if the platform is a mobile one (iOS or Android).
 *
 * @returns Determines if current platform is mobile (Guess)
 */
export function isMobile(): boolean {
    let userAgent = ""

    if (navigator.userAgent) {
        userAgent = navigator.userAgent
    }

    if (/iPhone|iPad|iPod|Android/.test(userAgent)) {
        // This is Android, iOS, or iPad < 13.
        return true
    }

    /*
     * Starting with iOS 13 on iPad, the user agent string no longer has the
     * word "iPad" in it.  It looks very similar to desktop Safari.  This seems
     * to be intentional on Apple's part.
     * See: https://forums.developer.apple.com/thread/119186
     *
     * So if it's an Apple device with multi-touch support, assume it's a mobile
     * device.  If some future iOS version starts masking their user agent on
     * both iPhone & iPad, this clause should still work.  If a future
     * multi-touch desktop Mac is released, this will need some adjustment.
     */
    return isApple() && navigator.maxTouchPoints > 1
}

/**
 * Guesses if the platform is a Smart TV (Tizen or WebOS).
 *
 * @returns Determines if platform is a Smart TV
 */
export function isTv(): boolean {
    return isTizen() || isWebOS()
}

/**
 * Guesses if the platform is a PS4
 *
 * @returns Determines if the device is a PS4
 */
export function isPs4(): boolean {
    return userAgentContains("playstation 4")
}

/**
 * Guesses if the platform is a Xbox
 *
 * @returns Determines if the device is a Xbox
 */
export function isXbox(): boolean {
    return userAgentContains("xbox")
}
