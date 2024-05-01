import { THEME_DEFAULT_VALUES, useThemeSettings } from "@/lib/theme/hooks"
import { colord, RgbColor } from "colord"
import React from "react"

type CustomColorProviderProps = {}


export function CustomThemeProvider(props: CustomColorProviderProps) {

    const {} = props

    const ts = useThemeSettings()

    function setBgColor(r: any, variable: string, defaultColor: string | null, customColor: string | RgbColor) {
        if (ts.backgroundColor === THEME_DEFAULT_VALUES.backgroundColor) {
            if (defaultColor) r.style.setProperty(variable, defaultColor)
            return
        }
        if (typeof customColor === "string") {
            r.style.setProperty(variable, customColor)
        } else {
            r.style.setProperty(variable, `${customColor.r} ${customColor.g} ${customColor.b}`)
        }
    }

    function setColor(r: any, variable: string, defaultColor: string | null, customColor: string | RgbColor) {
        if (ts.accentColor === THEME_DEFAULT_VALUES.accentColor) {
            if (defaultColor) r.style.setProperty(variable, defaultColor)
            return
        }
        if (typeof customColor === "string") {
            r.style.setProperty(variable, customColor)
        } else {
            r.style.setProperty(variable, `${customColor.r} ${customColor.g} ${customColor.b}`)
        }
    }


    // e.g. #0a050d -> dark purple
    // e.g. #11040d -> dark pink-ish purple
    // #050a0d -> dark blue
    React.useEffect(() => {
        let r = document.querySelector(":root") as any

        if (!ts.enableColorSettings) return

        // if (ts.backgroundColor === THEME_DEFAULT_VALUES.backgroundColor) {
        //     return
        // }

        setBgColor(r, "--background", "#0c0c0c", ts.backgroundColor)
        setBgColor(r, "--paper", "#101010", colord(ts.backgroundColor).lighten(0.025).toHex())
        setBgColor(r, "--media-card-popup-background", colord("rgb(16 16 16)").toHex(), colord(ts.backgroundColor).lighten(0.025).toHex())
        setBgColor(r,
            "--hover-from-background-color",
            colord("rgb(23 23 23)").toHex(),
            colord(ts.backgroundColor).lighten(0.025).desaturate(0.05).toHex())


        setBgColor(r, "--color-gray-950", "16 16 16", colord(ts.backgroundColor).lighten(0.008).desaturate(0.05).toRgb())
        setBgColor(r, "--color-gray-900", "23 23 23", colord(ts.backgroundColor).lighten(0.04).desaturate(0.05).toRgb())
        setBgColor(r, "--color-gray-800", "38 38 38", colord(ts.backgroundColor).lighten(0.06).desaturate(0.2).toRgb())
        setBgColor(r, "--color-gray-700", "64 64 64", colord(ts.backgroundColor).lighten(0.08).desaturate(0.2).toRgb())
        setBgColor(r, "--color-gray-600", "82 82 82", colord(ts.backgroundColor).lighten(0.1).desaturate(0.2).toRgb())
        setBgColor(r, "--color-gray-500", "115 115 115", colord(ts.backgroundColor).lighten(0.15).desaturate(0.2).toRgb())
        setBgColor(r, "--color-gray-400", "163 163 163", colord(ts.backgroundColor).lighten(0.3).desaturate(0.2).toRgb())
        // setColor(r, "--color-gray-300", null, colord(ts.backgroundColor).lighten(0.4).desaturate(0.2).toRgb())

    }, [ts.enableColorSettings, ts.backgroundColor])

    React.useEffect(() => {
        let r = document.querySelector(":root") as any

        if (!ts.enableColorSettings) return

        // if (ts.accentColor === THEME_DEFAULT_VALUES.accentColor) {
        //     return
        // }

        setColor(r, "--color-brand-200", "212 208 255", colord(ts.accentColor).lighten(0.2).toRgb())
        setColor(r, "--color-brand-300", "199 194 255", colord(ts.accentColor).lighten(0.15).toRgb())
        setColor(r, "--color-brand-400", "159 146 255", colord(ts.accentColor).lighten(0.1).toRgb())
        setColor(r, "--color-brand-500", "97 82 223", colord(ts.accentColor).toRgb())
        setColor(r, "--color-brand-600", "82 67 203", colord(ts.accentColor).darken(0.1).toRgb())
        setColor(r, "--color-brand-700", "63 46 178", colord(ts.accentColor).darken(0.15).toRgb())
        setColor(r, "--color-brand-800", "49 40 135", colord(ts.accentColor).darken(0.2).toRgb())
        setColor(r, "--color-brand-900", "35 28 107", colord(ts.accentColor).darken(0.25).toRgb())
        setColor(r, "--color-brand-950", "26 20 79", colord(ts.accentColor).darken(0.3).toRgb())
        setColor(r, "--brand", colord("rgba(199 194 255)").toHex(), colord(ts.accentColor).lighten(0.35).desaturate(0.1).toHex())
    }, [ts.enableColorSettings, ts.accentColor])

    return null
}

