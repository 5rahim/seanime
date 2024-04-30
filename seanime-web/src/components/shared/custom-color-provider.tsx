import { useThemeSettings } from "@/lib/theme/hooks"
import { colord, RgbColor } from "colord"
import React from "react"

type CustomColorProviderProps = {
    children: React.ReactNode
}

type ThemeColors = {
    mediaCardPopupBackground: string
}

const defaultThemeColors: ThemeColors = {
    mediaCardPopupBackground: "#101010",
}

const __ThemeColorsContext = React.createContext<ThemeColors>(defaultThemeColors)

export function CustomColorProvider(props: CustomColorProviderProps) {

    const {
        children,
        ...rest
    } = props

    const ts = useThemeSettings()

    const data: ThemeColors = React.useMemo(() => {
        if (ts.backgroundColor === "#0c0c0c") return defaultThemeColors
        return {
            mediaCardPopupBackground: colord(ts.backgroundColor).lighten(0.025).toHex(),
        }
    }, [ts.backgroundColor])

    function setColor(r: any, variable: string, defaultColor: string | null, customColor: string | RgbColor) {
        if (ts.backgroundColor === "#0c0c0c") {
            if (defaultColor) r.style.setProperty(variable, defaultColor)
            return
        }
        if (typeof customColor === "string") {
            r.style.setProperty(variable, customColor)
        } else {
            r.style.setProperty(variable, `${customColor.r} ${customColor.g} ${customColor.b}`)
            console.log(variable, `${customColor.r} ${customColor.g} ${customColor.b}`)
        }
    }


    // e.g. #0a050d -> dark purple
    // e.g. #11040d -> dark pink-ish purple
    // #050a0d -> dark blue
    React.useEffect(() => {
        let r = document.querySelector(":root") as any

        if (ts.backgroundColor === "#0c0c0c") {
            return
        }

        r.style.setProperty("--background", ts.backgroundColor)
        setColor(r, "--paper", "#101010", colord(ts.backgroundColor).lighten(0.025).toHex())
        setColor(r, "--media-card-popup-background", null, colord(ts.backgroundColor).lighten(0.025).toHex())
        setColor(r, "--hover-from-background-color", null, colord(ts.backgroundColor).lighten(0.025).desaturate(0.05).toHex())


        setColor(r, "--color-gray-950", null, colord(ts.backgroundColor).lighten(0.008).desaturate(0.05).toRgb())
        setColor(r, "--color-gray-900", null, colord(ts.backgroundColor).lighten(0.04).desaturate(0.05).toRgb())
        setColor(r, "--color-gray-800", null, colord(ts.backgroundColor).lighten(0.06).desaturate(0.2).toRgb())
        setColor(r, "--color-gray-700", null, colord(ts.backgroundColor).lighten(0.08).desaturate(0.2).toRgb())
        setColor(r, "--color-gray-600", null, colord(ts.backgroundColor).lighten(0.1).desaturate(0.2).toRgb())
        setColor(r, "--color-gray-500", null, colord(ts.backgroundColor).lighten(0.14).desaturate(0.2).toRgb())
        setColor(r, "--color-gray-400", null, colord(ts.backgroundColor).lighten(0.3).desaturate(0.2).toRgb())
        // setColor(r, "--color-gray-300", null, colord(ts.backgroundColor).lighten(0.4).desaturate(0.2).toRgb())

    }, [ts.backgroundColor])

    return (
        <__ThemeColorsContext.Provider value={data}>
            {children}
        </__ThemeColorsContext.Provider>
    )
}

export function useThemeColors() {
    return React.useContext(__ThemeColorsContext)
}


