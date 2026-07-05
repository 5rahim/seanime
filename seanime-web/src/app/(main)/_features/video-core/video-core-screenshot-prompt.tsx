import { useSaveMediaPlayerSettings } from "@/api/hooks/settings.hooks"
import { useVideoCoreSaveScreenshot } from "@/api/hooks/videocore.hooks"
import { ScreenshotDirModal } from "@/app/(main)/_features/media-core/screenshot-dir-modal"
import { useServerStatus } from "@/app/(main)/_hooks/use-server-status"
import { useAtom, useAtomValue } from "jotai"
import { useSetAtom } from "jotai/react"
import React from "react"
import { toast } from "sonner"
import { vc_containerElement, vc_isFullscreen } from "./video-core-atoms"
import { vc_showOverlayFeedback } from "./video-core-overlay-display"
import { vc_pendingScreenshotAtom, vc_screenshotPromptOpenAtom } from "./video-core-screenshot"

export function VideoCoreScreenshotDirPrompt() {
    const [promptOpen, setPromptOpen] = useAtom(vc_screenshotPromptOpenAtom)
    const [pendingScreenshot, setPendingScreenshot] = useAtom(vc_pendingScreenshotAtom)

    const isFullscreen = useAtomValue(vc_isFullscreen)
    const containerElement = useAtomValue(vc_containerElement)
    const showOverlayFeedback = useSetAtom(vc_showOverlayFeedback)

    const serverStatus = useServerStatus()
    const { mutateAsync: saveSettings } = useSaveMediaPlayerSettings()
    const { mutateAsync: saveScreenshotMutation } = useVideoCoreSaveScreenshot()

    const blobToBase64 = (blob: Blob): Promise<string> => {
        return new Promise((resolve, reject) => {
            const reader = new FileReader()
            reader.onloadend = () => {
                const base64String = (reader.result as string).split(",")[1]
                resolve(base64String)
            }
            reader.onerror = reject
            reader.readAsDataURL(blob)
        })
    }

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
                const { blob, isAnime4K } = pendingScreenshot
                const filename = `seanime_screenshot_${new Date().getTime()}${isAnime4K ? "_anime4k" : ""}.png`

                const base64Data = await blobToBase64(blob)
                await saveScreenshotMutation({
                    dir: path,
                    filename,
                    base64Data,
                })

                showOverlayFeedback({ message: "Screenshot saved", type: "message" })
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
