import { app } from "electron"
import log from "electron-log/main"
import fs from "node:fs"
import path from "node:path"

export type DenshiSettings = {
    minimizeToTray: boolean
    openInBackground: boolean
    openAtLaunch: boolean
    updateChannel: string
    windowBounds: Electron.Rectangle | null
    windowMaximized: boolean
    mpvPrismLogging: boolean
}

export const DENSHI_SETTINGS_DEFAULTS: DenshiSettings = {
    minimizeToTray: true,
    openInBackground: false,
    openAtLaunch: false,
    updateChannel: "github",
    windowBounds: null,
    windowMaximized: true,
    mpvPrismLogging: false,
}

function getDenshiSettingsPath(): string {
    return path.join(app.getPath("userData"), "denshi-settings.json")
}

export function loadDenshiSettings(): DenshiSettings {
    try {
        const settingsPath = getDenshiSettingsPath()
        if (fs.existsSync(settingsPath)) {
            const data = JSON.parse(fs.readFileSync(settingsPath, "utf-8"))
            return { ...DENSHI_SETTINGS_DEFAULTS, ...data }
        }
    }
    catch (error) {
        log.error("[Denshi] Failed to load settings:", error)
    }
    return { ...DENSHI_SETTINGS_DEFAULTS }
}

export function saveDenshiSettings(settings: DenshiSettings): void {
    try {
        fs.writeFileSync(getDenshiSettingsPath(), JSON.stringify(settings, null, 2), "utf-8")
    }
    catch (error) {
        log.error("[Denshi] Failed to save settings:", error)
    }
}
