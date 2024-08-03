import { useServerStatus } from "@/app/(main)/_hooks/use-server-status"
import { useAtom } from "jotai/react"
import { atomWithStorage } from "jotai/utils"
import React from "react"

const __mediastream_filePath = atomWithStorage<string | undefined>("sea-mediastream-filepath", undefined)

export function useMediastreamCurrentFile() {
    const [filePath, setFilePath] = useAtom(__mediastream_filePath)

    return {
        filePath,
        setFilePath,
    }
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

const __mediastream_jassubOffscreenRender = atomWithStorage<boolean>("sea-mediastream-jassub-offscreen-render", false)

export function useMediastreamJassubOffscreenRender() {
    const [jassubOffscreenRender, setJassubOffscreenRender] = useAtom(__mediastream_jassubOffscreenRender)

    return {
        jassubOffscreenRender,
        setJassubOffscreenRender,
    }
}


//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

export const __mediastream_autoPlayAtom = atomWithStorage("sea-mediastream-autoplay", false)

export const __mediastream_autoNextAtom = atomWithStorage("sea-mediastream-autonext", false)

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

/**
 * Holds the list of media IDs that are to be transcoded on the current device
 * @DEPRECATED
 */
const __mediastream_mediaToTranscode = atomWithStorage<string[]>("sea-mediastream-media-to-transcode", [])

export function useMediastreamMediaToTranscode() {
    const [mediaToTranscode, setMediaToTranscode] = useAtom(__mediastream_mediaToTranscode)

    function addMediaToTranscode(mediaId: number) {
        setMediaToTranscode((prev) => [...prev, String(mediaId)])
    }

    function removeMediaToTranscode(mediaId: number) {
        setMediaToTranscode((prev) => prev.filter((id) => id !== String(mediaId)))
    }

    function clearMediaToTranscode() {
        setMediaToTranscode([])
    }

    return {
        mediaToTranscode,
        addMediaToTranscode,
        removeMediaToTranscode,
        clearMediaToTranscode,
    }
}

/**
 * Whether media streaming should be done on this device
 */
const __mediastream_activeOnDevice = atomWithStorage<boolean | null>("sea-mediastream-active-on-device", null)

export function useMediastreamActiveOnDevice() {
    const serverStatus = useServerStatus()
    const [activeOnDevice, setActiveOnDevice] = useAtom(__mediastream_activeOnDevice)

    // Set default behavior
    React.useLayoutEffect(() => {
        if (activeOnDevice !== null) return

        if (serverStatus?.clientDevice !== "desktop") {
            setActiveOnDevice(true) // Always active on mobile devices
        } else {
            setActiveOnDevice(false) // Always inactive on desktop devices
        }

    }, [serverStatus?.clientUserAgent, activeOnDevice])

    return {
        activeOnDevice,
        setActiveOnDevice,
    }
}
