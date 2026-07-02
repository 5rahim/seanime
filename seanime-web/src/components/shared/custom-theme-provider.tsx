import { THEME_DEFAULT_VALUES, useThemeSettings } from "@/lib/theme/theme-hooks"
import { colord, extend } from "colord"
import mixPlugin from "colord/plugins/mix"
import { useTheme } from "next-themes"
import React from "react"

extend([mixPlugin])

type Colord = ReturnType<typeof colord>

// Tailwind `--color-*` tokens expect space-separated "R G B" channels (so `<alpha-value>`
// modifiers work); the standalone color tokens (--background, --brand, ...) take a full color.
const rgb = (c: Colord) => {
    const { r, g, b } = c.toRgb()
    return `${r} ${g} ${b}`
}

type Derivation = { variable: string; derive: (base: Colord) => string }

// When inactive these are cleared (removeProperty), not just skipped: inline styles beat class
// selectors, so a stale inline `--color-gray-*` would otherwise override the `.light` ramp.
const BG_DERIVATIONS: Derivation[] = [
    { variable: "--background", derive: c => c.toHex() },
    { variable: "--paper", derive: c => c.lighten(0.025).toHex() },
    { variable: "--media-card-popup-background", derive: c => c.lighten(0.025).toHex() },
    { variable: "--hover-from-background-color", derive: c => c.lighten(0.025).desaturate(0.05).toHex() },
    { variable: "--color-gray-400", derive: c => rgb(c.lighten(0.3).desaturate(0.2)) },
    { variable: "--color-gray-500", derive: c => rgb(c.lighten(0.15).desaturate(0.2)) },
    { variable: "--color-gray-600", derive: c => rgb(c.lighten(0.1).desaturate(0.2)) },
    { variable: "--color-gray-700", derive: c => rgb(c.lighten(0.08).desaturate(0.2)) },
    { variable: "--color-gray-800", derive: c => rgb(c.lighten(0.06).desaturate(0.2)) },
    { variable: "--color-gray-900", derive: c => rgb(c.lighten(0.04).desaturate(0.05)) },
    { variable: "--color-gray-950", derive: c => rgb(c.lighten(0.008).desaturate(0.05)) },
]

const ACCENT_DERIVATIONS: Derivation[] = [
    { variable: "--color-brand-200", derive: c => rgb(c.lighten(0.35).desaturate(0.05)) },
    { variable: "--color-brand-300", derive: c => rgb(c.lighten(0.3).desaturate(0.05)) },
    { variable: "--color-brand-400", derive: c => rgb(c.lighten(0.1)) },
    { variable: "--color-brand-500", derive: c => rgb(c) },
    { variable: "--color-brand-600", derive: c => rgb(c.darken(0.1)) },
    { variable: "--color-brand-700", derive: c => rgb(c.darken(0.15)) },
    { variable: "--color-brand-800", derive: c => rgb(c.darken(0.2)) },
    { variable: "--color-brand-900", derive: c => rgb(c.darken(0.25)) },
    { variable: "--color-brand-950", derive: c => rgb(c.darken(0.3)) },
    { variable: "--brand", derive: c => c.lighten(0.35).desaturate(0.1).toHex() },
]

function applyDerivations(derivations: Derivation[], base: Colord | null) {
    const root = document.documentElement
    for (const d of derivations) {
        if (base) {
            root.style.setProperty(d.variable, d.derive(base))
        } else {
            root.style.removeProperty(d.variable)
        }
    }
}

export function CustomThemeProvider() {
    const ts = useThemeSettings()
    const { resolvedTheme } = useTheme()

    // The customization derives its ramp by *lightening* the chosen color, which only makes
    // sense on a dark base. In light mode (or when disabled / left at default) the props are
    // cleared and the static globals.css ramp takes over.
    const enabled = resolvedTheme != null && resolvedTheme !== "light" && ts.enableColorSettings

    React.useEffect(() => {
        // Wait for next-themes to resolve before touching :root, else we flash dark-derived
        // colors on a light page during hydration.
        if (resolvedTheme == null) return
        const custom = enabled && ts.backgroundColor !== THEME_DEFAULT_VALUES.backgroundColor
        applyDerivations(BG_DERIVATIONS, custom ? colord(ts.backgroundColor) : null)
    }, [resolvedTheme, enabled, ts.backgroundColor])

    React.useEffect(() => {
        if (resolvedTheme == null) return
        const custom = enabled && ts.accentColor !== THEME_DEFAULT_VALUES.accentColor
        applyDerivations(ACCENT_DERIVATIONS, custom ? colord(ts.accentColor) : null)
    }, [resolvedTheme, enabled, ts.accentColor])

    return null
}
