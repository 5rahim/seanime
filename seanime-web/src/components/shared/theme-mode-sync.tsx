import { useServerStatus } from "@/app/(main)/_hooks/use-server-status"
import { useTheme } from "next-themes"
import React from "react"

/**
 * Syncs the server-persisted theme mode into next-themes (the server value is the source
 * of truth). Until server status loads we leave next-themes on its own localStorage value,
 * to avoid a flash on first paint.
 */
export function ThemeModeSync() {
    const serverStatus = useServerStatus()
    const { theme, setTheme } = useTheme()

    const mode = serverStatus?.themeSettings?.themeMode

    React.useEffect(() => {
        if (!mode) return // server settings not loaded yet
        if (theme !== mode) {
            setTheme(mode)
        }
    }, [mode])

    return null
}
