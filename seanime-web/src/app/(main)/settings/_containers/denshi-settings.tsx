import { SettingsCard } from "@/app/(main)/settings/_components/settings-card"
import { Switch } from "@/components/ui/switch"
import React from "react"
import { RiSettings3Fill } from "react-icons/ri"

export function DenshiSettings() {

    const [settings, setSettings] = React.useState<DenshiSettings | null>(null)
    const settingsRef = React.useRef<DenshiSettings | null>(null)
    const [loading, setLoading] = React.useState(true)

    React.useEffect(() => {
        if (window.electron?.denshiSettings) {
            window.electron.denshiSettings.get().then((s) => {
                setSettings(s)
                settingsRef.current = s
                setLoading(false)
            })
        }
    }, [])

    function updateSetting(key: keyof DenshiSettings, value: boolean) {
        if (!settingsRef.current || !window.electron?.denshiSettings) return

        const newSettings = { ...settingsRef.current, [key]: value }
        settingsRef.current = newSettings
        setSettings(newSettings)
        window.electron.denshiSettings.set(newSettings)
    }

    if (loading || !settings) {
        return null
    }

    return (
        <div className="space-y-4">
            <SettingsCard title="Window">
                <Switch
                    side="right"
                    value={settings.minimizeToTray}
                    onValueChange={(v) => updateSetting("minimizeToTray", v)}
                    label="Minimize to tray on close"
                    help="When enabled, closing the window will minimize the app to the system tray instead of quitting."
                />
                <Switch
                    side="right"
                    value={settings.openInBackground}
                    onValueChange={(v) => updateSetting("openInBackground", v)}
                    label="Open in background"
                    help="When enabled, the app will start hidden. You can show it from the system tray."
                />
            </SettingsCard>

            <SettingsCard title="System">
                <Switch
                    side="right"
                    value={settings.openAtLaunch}
                    onValueChange={(v) => updateSetting("openAtLaunch", v)}
                    label="Open at launch"
                    help={window.electron?.platform === "linux"
                        ? "This feature is not supported on Linux."
                        : "When enabled, the app will start automatically when you log in to your computer."}
                    disabled={window.electron?.platform === "linux"}
                />
            </SettingsCard>

            <div className="flex items-center gap-2 text-sm text-gray-500 bg-gray-50 dark:bg-gray-900/30 rounded-lg p-3 border border-gray-200 dark:border-gray-800 border-dashed">
                <RiSettings3Fill className="text-base" />
                <span>Settings are saved automatically and applied after a restart</span>
            </div>
        </div>
    )
}
