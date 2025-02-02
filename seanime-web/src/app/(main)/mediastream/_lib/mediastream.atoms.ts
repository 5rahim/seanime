import { useServerStatus } from "@/app/(main)/_hooks/use-server-status"
import { useAtom } from "jotai/react"
import { atomWithStorage } from "jotai/utils"
import React from "react"

const __mediastream_filePath = atomWithStorage<string | undefined>("sea-mediastream-filepath", undefined, undefined, { getOnInit: true })

export function useMediastreamCurrentFile() {
    const [filePath, setFilePath] = useAtom(__mediastream_filePath)

    return {
        filePath,
        setFilePath,
    }
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

const __mediastream_jassubOffscreenRender = atomWithStorage<boolean>("sea-mediastream-jassub-offscreen-render", false, undefined, { getOnInit: true })

export function useMediastreamJassubOffscreenRender() {
    const [jassubOffscreenRender, setJassubOffscreenRender] = useAtom(__mediastream_jassubOffscreenRender)

    return {
        jassubOffscreenRender,
        setJassubOffscreenRender,
    }
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

/**
 * Whether media streaming should be done on this device
 */
const __mediastream_activeOnDevice = atomWithStorage<boolean | null>("sea-mediastream-active-on-device", null, undefined, { getOnInit: true })

export function useMediastreamActiveOnDevice() {
    const serverStatus = useServerStatus()
    const [activeOnDevice, setActiveOnDevice] = useAtom(__mediastream_activeOnDevice)

    // Set default behavior
    React.useLayoutEffect(() => {
        if (!!serverStatus) {

            if (activeOnDevice === null) {

                if (serverStatus?.clientDevice !== "desktop") {
                    setActiveOnDevice(true) // Always active on mobile devices
                } else {
                    setActiveOnDevice(false) // Always inactive on desktop devices
                }

            }
        }
    }, [serverStatus?.clientUserAgent, activeOnDevice])

    return {
        activeOnDevice,
        setActiveOnDevice,
    }
}
