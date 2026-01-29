import { __isDesktop__ } from "@/types/constants"
import { MediaEnterFullscreenRequestEvent, MediaFullscreenRequestTarget, MediaPlayerInstance } from "@vidstack/react"
import React from "react"

export function useFullscreenHandler(playerRef: React.RefObject<MediaPlayerInstance>) {

    React.useEffect(() => {
        let unlisten: any | null = null
        if ((window as any)?.__TAURI__) {
            const currentWindow: any | undefined = (window as any)?.__TAURI__?.window?.getCurrentWindow?.()
            if (currentWindow) {
                (async () => {
                    unlisten = await currentWindow.listen("macos-activation-policy-accessory-done", () => {
                        console.log("macos policy accessory event done")
                        try {
                            console.log("requesting fullscreen")
                            playerRef.current?.enterFullscreen()
                        }
                        catch (e) {
                            console.log("failed to enter fullscreen from 'macos-activation-policy-accessory-done'", e)
                        }
                    })
                })()
            }
        }
        return () => {
            unlisten?.()
        }
    }, [])

    function onMediaEnterFullscreenRequest(detail: MediaFullscreenRequestTarget, nativeEvent: MediaEnterFullscreenRequestEvent) {
        if (__isDesktop__) {
            try {
                if ((window as any)?.__TAURI__) {
                    const platform: string | undefined = (window as any)?.__TAURI__?.os?.platform?.()
                    const currentWindow: any | undefined = (window as any)?.__TAURI__?.window?.getCurrentWindow?.()
                    if (!!platform && platform === "macos") {
                        nativeEvent.preventDefault()
                        console.log("native fullscreen event prevented, sending macos policy accessory event")
                        currentWindow.emit("macos-activation-policy-accessory").then(() => {
                            console.log("macos policy accessory event sent")
                            if (nativeEvent.defaultPrevented) {
                                // console.log("requesting fullscreen")
                                try {
                                    // playerRef.current?.enterFullscreen()
                                }
                                catch (e) {
                                    console.log("failed to enter fullscreen from onMediaEnterFullscreenRequest", e)
                                }
                            }
                        })
                    }
                }
            }
            catch {

            }
        }
    }

    return {
        onMediaEnterFullscreenRequest,
    }
}
