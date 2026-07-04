import { useThemeSettings } from "@/lib/theme/theme-hooks"
import { colord, extend } from "colord"
import mixPlugin from "colord/plugins/mix"
import React from "react"

extend([mixPlugin])

type CustomColorProviderProps = {}

export function CustomThemeProvider(props: CustomColorProviderProps) {
    const {} = props
    const ts = useThemeSettings()

    React.useEffect(() => {
        const r = document.querySelector(":root") as any
        if (!r) return

        const variables = [
            "--background",
            "--paper",
            "--paper-lighter",
            "--media-card-popup-background",
            "--hover-from-background-color",
            "--border",
            "--muted",
            "--muted-highlight",
            "--subtle",
            "--subtle-highlight",
            "--media-accent-color",
            "--color-gray-50",
            "--color-gray-100",
            "--color-gray-200",
            "--color-gray-300",
            "--color-gray-400",
            "--color-gray-500",
            "--color-gray-600",
            "--color-gray-700",
            "--color-gray-800",
            "--color-gray-900",
            "--color-gray-950",
            "--color-brand-50",
            "--color-brand-100",
            "--color-brand-200",
            "--color-brand-300",
            "--color-brand-400",
            "--color-brand-500",
            "--color-brand-600",
            "--color-brand-700",
            "--color-brand-800",
            "--color-brand-900",
            "--color-brand-950",
            "--brand",
        ]

        if (!ts.enableColorSettings) {
            variables.forEach(v => r.style.removeProperty(v))
            return
        }

        const setVal = (variable: string, val: string) => {
            r.style.setProperty(variable, val)
        }

        const setRgbVal = (variable: string, color: any) => {
            const rgb = colord(color).toRgb()
            r.style.setProperty(variable, `${rgb.r} ${rgb.g} ${rgb.b}`)
        }

        // BACKGROUND CALCULATIONS
        const bg = colord(ts.backgroundColor)
        const darkest = bg
        const lightest = colord("#f8fafc").mix(bg, 0.15) // 15% background tint into near-white
        const _lightest = colord("#f8fafc").mix(bg, 0.7)

        setVal("--background", bg.toHex())
        setVal("--paper", darkest.mix(_lightest, 0.03).toHex())
        setVal("--paper-lighter", darkest.mix(_lightest, 0.05).toHex())
        setVal("--media-card-popup-background", darkest.mix(_lightest, 0.04).toHex())
        setVal("--hover-from-background-color", darkest.mix(_lightest, 0.06).toHex())

        // Default dark-mode transparency overlays
        setVal("--border", "rgba(255, 255, 255, 0.08)")
        setVal("--muted", "rgba(255, 255, 255, 0.4)")
        setVal("--muted-highlight", "rgba(255, 255, 255, 0.6)")
        setVal("--subtle", "rgba(255, 255, 255, 0.06)")
        setVal("--subtle-highlight", "rgba(255, 255, 255, 0.08)")

        // GRAY SCALE GENERATION (Interpolated & tinted)
        setRgbVal("--color-gray-950", darkest.mix(_lightest, 0.02))
        setRgbVal("--color-gray-900", darkest.mix(_lightest, 0.05))
        setRgbVal("--color-gray-800", darkest.mix(_lightest, 0.10))
        setRgbVal("--color-gray-700", darkest.mix(lightest, 0.18))
        setRgbVal("--color-gray-600", darkest.mix(lightest, 0.32))
        setRgbVal("--color-gray-500", darkest.mix(lightest, 0.48))
        setRgbVal("--color-gray-400", darkest.mix(lightest, 0.64))
        setRgbVal("--color-gray-300", darkest.mix(lightest, 0.78))
        setRgbVal("--color-gray-200", darkest.mix(lightest, 0.88))
        setRgbVal("--color-gray-100", darkest.mix(lightest, 0.94))
        setRgbVal("--color-gray-50", darkest.mix(lightest, 0.98))

        // BRAND SCALE GENERATION
        const accent = colord(ts.accentColor)

        setRgbVal("--color-brand-50", accent.mix("#ffffff", 0.90))
        setRgbVal("--color-brand-100", accent.mix("#ffffff", 0.75))
        setRgbVal("--color-brand-200", accent.mix("#ffffff", 0.55))
        setRgbVal("--color-brand-300", accent.mix("#ffffff", 0.35))
        setRgbVal("--color-brand-400", accent.mix("#ffffff", 0.15))
        setRgbVal("--color-brand-500", accent)
        setRgbVal("--color-brand-600", accent.mix("#000000", 0.15))
        setRgbVal("--color-brand-700", accent.mix("#000000", 0.30))
        setRgbVal("--color-brand-800", accent.mix("#000000", 0.45))
        setRgbVal("--color-brand-900", accent.mix("#000000", 0.60))
        setRgbVal("--color-brand-950", accent.mix("#000000", 0.75))

        // Set accent-related brand theme color variables
        setVal("--media-accent-color", accent.toHex())
        setVal("--brand", accent.mix("#ffffff", 0.35).toHex()) // equivalent to brand-300
    }, [ts.enableColorSettings, ts.backgroundColor, ts.accentColor])

    return null
}


// import { THEME_DEFAULT_VALUES, useThemeSettings } from "@/lib/theme/theme-hooks"
// import { colord, extend, RgbColor } from "colord"
// import mixPlugin from "colord/plugins/mix"
// import React from "react"
//
// extend([mixPlugin])
//
//
// type CustomColorProviderProps = {}
//
//
// export function CustomThemeProvider(props: CustomColorProviderProps) {
//
//     const {} = props
//
//     const ts = useThemeSettings()
//
//     function setBgColor(r: any, variable: string, defaultColor: string | null, customColor: string | RgbColor) {
//         if (ts.backgroundColor === THEME_DEFAULT_VALUES.backgroundColor) {
//             if (defaultColor) r.style.setProperty(variable, defaultColor)
//             return
//         }
//         if (typeof customColor === "string") {
//             r.style.setProperty(variable, customColor)
//         } else {
//             r.style.setProperty(variable, `${customColor.r} ${customColor.g} ${customColor.b}`)
//         }
//     }
//
//     function setColor(r: any, variable: string, defaultColor: string | null, customColor: string | RgbColor) {
//         if (ts.accentColor === THEME_DEFAULT_VALUES.accentColor) {
//             if (defaultColor) r.style.setProperty(variable, defaultColor)
//             return
//         }
//         if (typeof customColor === "string") {
//             r.style.setProperty(variable, customColor)
//         } else {
//             r.style.setProperty(variable, `${customColor.r} ${customColor.g} ${customColor.b}`)
//         }
//     }
//
//
//     // e.g. #0a050d -> dark purple
//     // e.g. #11040d -> dark pink-ish purple
//     // #050a0d -> dark blue
//     React.useEffect(() => {
//         let r = document.querySelector(":root") as any
//
//         if (!ts.enableColorSettings) return
//
//         setBgColor(r, "--background", "#070707", ts.backgroundColor)
//         setBgColor(r, "--paper", colord("rgba(11 11 11)").toHex(), colord(ts.backgroundColor).lighten(0.025).toHex())
//         setBgColor(r, "--media-card-popup-background", colord("rgb(16 16 16)").toHex(), colord(ts.backgroundColor).lighten(0.025).toHex())
//         setBgColor(r,
//             "--hover-from-background-color",
//             colord("rgb(23 23 23)").toHex(),
//             colord(ts.backgroundColor).lighten(0.025).desaturate(0.05).toHex())
//
//
//         setBgColor(r, "--color-gray-400", "143 143 143", colord(ts.backgroundColor).lighten(0.3).desaturate(0.2).toRgb())
//         setBgColor(r, "--color-gray-500", "90 90 90", colord(ts.backgroundColor).lighten(0.15).desaturate(0.2).toRgb())
//         setBgColor(r, "--color-gray-600", "72 72 72", colord(ts.backgroundColor).lighten(0.1).desaturate(0.2).toRgb())
//         setBgColor(r, "--color-gray-700", "54 54 54", colord(ts.backgroundColor).lighten(0.08).desaturate(0.2).toRgb())
//         setBgColor(r, "--color-gray-800", "28 28 28", colord(ts.backgroundColor).lighten(0.06).desaturate(0.2).toRgb())
//         setBgColor(r, "--color-gray-900", "16 16 16", colord(ts.backgroundColor).lighten(0.04).desaturate(0.05).toRgb())
//         setBgColor(r, "--color-gray-950", "11 11 11", colord(ts.backgroundColor).lighten(0.008).desaturate(0.05).toRgb())
//         // setColor(r, "--color-gray-300", null, colord(ts.backgroundColor).lighten(0.4).desaturate(0.2).toRgb())
//
//     }, [ts.enableColorSettings, ts.backgroundColor])
//
//     React.useEffect(() => {
//         let r = document.querySelector(":root") as any
//
//         if (!ts.enableColorSettings) return
//
//         setColor(r, "--color-brand-200", "212 208 255", colord(ts.accentColor).lighten(0.35).desaturate(0.05).toRgb())
//         setColor(r, "--color-brand-300", "199 194 255", colord(ts.accentColor).lighten(0.3).desaturate(0.05).toRgb())
//         setColor(r, "--color-brand-400", "159 146 255", colord(ts.accentColor).lighten(0.1).toRgb())
//         setColor(r, "--color-brand-500", "97 82 223", colord(ts.accentColor).toRgb())
//         setColor(r, "--color-brand-600", "82 67 203", colord(ts.accentColor).darken(0.1).toRgb())
//         setColor(r, "--color-brand-700", "63 46 178", colord(ts.accentColor).darken(0.15).toRgb())
//         setColor(r, "--color-brand-800", "49 40 135", colord(ts.accentColor).darken(0.2).toRgb())
//         setColor(r, "--color-brand-900", "35 28 107", colord(ts.accentColor).darken(0.25).toRgb())
//         setColor(r, "--color-brand-950", "26 20 79", colord(ts.accentColor).darken(0.3).toRgb())
//         setColor(r, "--brand", colord("rgba(199 194 255)").toHex(), colord(ts.accentColor).lighten(0.35).desaturate(0.1).toHex())
//     }, [ts.enableColorSettings, ts.accentColor])
//
//     return null
// }
//
