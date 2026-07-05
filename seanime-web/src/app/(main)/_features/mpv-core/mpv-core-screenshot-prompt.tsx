import { useSaveMediaPlayerSettings } from "@/api/hooks/settings.hooks"
import { useVideoCoreSaveScreenshot } from "@/api/hooks/videocore.hooks"
import { ScreenshotDirModal } from "@/app/(main)/_features/media-core/screenshot-dir-modal"
import { useServerStatus } from "@/app/(main)/_hooks/use-server-status"
import { useAtom } from "jotai"
import { useSetAtom } from "jotai/react"
import React from "react"
import { toast } from "sonner"
import { mc_overlayFeedback, mc_pendingScreenshotAtom, mc_screenshotPromptOpenAtom } from "./mpv-core.atoms"

interface MpvCoreScreenshotDirPromptProps {
    isFullscreen: boolean
    containerElement: HTMLElement | null
}

export function MpvCoreScreenshotDirPrompt({ isFullscreen, containerElement }: MpvCoreScreenshotDirPromptProps) {
    const [promptOpen, setPromptOpen] = useAtom(mc_screenshotPromptOpenAtom)
    const [pendingScreenshot, setPendingScreenshot] = useAtom(mc_pendingScreenshotAtom)

    const setOverlayFeedback = useSetAtom(mc_overlayFeedback)

    const serverStatus = useServerStatus()
    const { mutateAsync: saveSettings } = useSaveMediaPlayerSettings()
    const { mutateAsync: saveScreenshotMutation } = useVideoCoreSaveScreenshot()

    const handleSave = async (path: string) => {
        const currentMediaPlayer = serverStatus?.settings?.mediaPlayer
        if (!currentMediaPlayer) return false

        try {
            // 1. Save settings to DB
            await saveSettings({
                mediaPlayer: {
                    ...currentMediaPlayer,
                    screenshotDir: path,
                },
            })

            // 2. If there is a pending screenshot, save it now
            if (pendingScreenshot) {
                const { base64Data } = pendingScreenshot
                const filename = `seanime_screenshot_${new Date().getTime()}.png`

                await saveScreenshotMutation({
                    dir: path,
                    filename,
                    base64Data,
                })

                setOverlayFeedback({ message: `Screenshot saved to ${path}`, type: "message" })
                setPendingScreenshot(null)
            }
            toast.success("Screenshot folder saved")
            return true
        }
        catch (error) {
            console.error("Failed to setup screenshot folder:", error)
            toast.error(error instanceof Error ? error.message : "Failed to save screenshot folder")
            return false
        }
    }

    return (
        <ScreenshotDirModal
            open={promptOpen}
            onClose={() => {
                setPromptOpen(false)
                setPendingScreenshot(null)
            }}
            onSave={handleSave}
            portalContainer={isFullscreen ? containerElement : null}
        />
    )
}
