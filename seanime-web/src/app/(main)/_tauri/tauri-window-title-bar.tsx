"use client"
import { IconButton } from "@/components/ui/button"
import { getCurrentWebviewWindow } from "@tauri-apps/api/webviewWindow"
import { platform } from "@tauri-apps/plugin-os"
import React from "react"
import { VscChromeClose, VscChromeMaximize, VscChromeMinimize, VscChromeRestore } from "react-icons/vsc"

type TauriWindowTitleBarProps = {
    children?: React.ReactNode
}

export function TauriWindowTitleBar(props: TauriWindowTitleBarProps) {

    const {
        children,
        ...rest
    } = props

    const [showTrafficLights, setShowTrafficLights] = React.useState(false)
    const [displayDragRegion, setDisplayDragRegion] = React.useState(true)


    function handleMinimize() {
        getCurrentWebviewWindow().minimize().then()
    }

    const [maximized, setMaximized] = React.useState(true)

    function toggleMaximized() {
        getCurrentWebviewWindow().toggleMaximize().then()
    }

    function handleClose() {
        getCurrentWebviewWindow().close().then()
    }

    React.useEffect(() => {

        const listener = getCurrentWebviewWindow().onResized(() => {
            onFullscreenChange()
            // Get the current window maximized state
            getCurrentWebviewWindow().isMaximized().then((maximized) => {
                setMaximized(maximized)
            })
        })

        // Check if the window is in fullscreen mode, and hide the traffic lights & drag region if it is
        function onFullscreenChange() {
            if (getCurrentWebviewWindow().label !== "main") return

            getCurrentWebviewWindow().isFullscreen().then((fullscreen) => {
                setShowTrafficLights(!fullscreen)
                setDisplayDragRegion(!fullscreen)
            })
        }

        document.addEventListener("fullscreenchange", onFullscreenChange)

        return () => {
            listener.then((f) => f())
            document.removeEventListener("fullscreenchange", onFullscreenChange)
        }
    }, [])

    const [currentPlatform, setCurrentPlatform] = React.useState("")

    React.useEffect(() => {
        (async () => {
            setCurrentPlatform(platform())
            const win = getCurrentWebviewWindow()
            const minimizable = await win.isMinimizable()
            const maximizable = await win.isMaximizable()
            const closable = await win.isClosable()
            setShowTrafficLights(_ => {
                let showTrafficLights = false

                if (win.label === "splashscreen") {
                    return false
                }

                if (minimizable || maximizable || closable) {
                    showTrafficLights = true
                }

                return showTrafficLights
            })
        })()
    }, [])

    if (!(currentPlatform === "windows" || currentPlatform === "macos")) return null

    return (
        <>
            <div
                className="__tauri-window-traffic-lights scroll-locked-offset bg-transparent fixed top-0 left-0 h-10 z-[999] w-full bg-opacity-90 flex pointer-events-[all]"
                style={{
                    pointerEvents: "all",
                }}
            >
                {displayDragRegion && <div className="flex flex-1" data-tauri-drag-region></div>}
                {(currentPlatform === "windows" && showTrafficLights) && <div className="flex h-10 items-center justify-center gap-1 mr-2">
                    <IconButton
                        className="outline-none w-11 size-8 rounded-lg duration-0 shadow-none text-white hover:text-white bg-transparent hover:bg-[rgba(255,255,255,0.05)] active:text-white active:bg-[rgba(255,255,255,0.1)]"
                        icon={<VscChromeMinimize className="text-[0.95rem]" />}
                        onClick={handleMinimize}
                        tabIndex={-1}
                    />
                    <IconButton
                        className="outline-none w-11 size-8 rounded-lg duration-0 shadow-none text-white hover:text-white bg-transparent hover:bg-[rgba(255,255,255,0.05)] active:text-white active:bg-[rgba(255,255,255,0.1)]"
                        icon={maximized ? <VscChromeRestore className="text-[0.95rem]" /> : <VscChromeMaximize className="text-[0.95rem]" />}
                        onClick={toggleMaximized}
                        tabIndex={-1}
                    />
                    <IconButton
                        className="outline-none w-11 size-8 rounded-lg duration-0 shadow-none text-white hover:text-white bg-transparent hover:bg-red-500 active:bg-red-600 active:text-white"
                        icon={<VscChromeClose className="text-[0.95rem]" />}
                        onClick={handleClose}
                        tabIndex={-1}
                    />
                </div>}
            </div>
        </>
    )
}
