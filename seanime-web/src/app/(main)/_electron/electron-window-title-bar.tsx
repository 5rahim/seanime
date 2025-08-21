"use client"

import { IconButton } from "@/components/ui/button"
import React from "react"
import { VscChromeClose, VscChromeMaximize, VscChromeMinimize, VscChromeRestore } from "react-icons/vsc"

type ElectronWindowTitleBarProps = {
    children?: React.ReactNode
}

export function ElectronWindowTitleBar(props: ElectronWindowTitleBarProps) {
    const {
        children,
        ...rest
    } = props

    const [showControls, setShowControls] = React.useState(true)
    const [displayDragRegion, setDisplayDragRegion] = React.useState(true)
    const [maximized, setMaximized] = React.useState(false)
    const [currentPlatform, setCurrentPlatform] = React.useState("")

    // Handle window control actions
    function handleMinimize() {
        if (window.electron?.window) {
            window.electron.window.minimize()
        }
    }

    function toggleMaximized() {
        if (window.electron?.window) {
            window.electron.window.toggleMaximize()
        }
    }

    function handleClose() {
        if (window.electron?.window) {
            window.electron.window.close()
        }
    }

    // Check fullscreen state
    function onFullscreenChange() {
        if (window.electron?.window) {
            window.electron.window.isFullscreen().then((fullscreen: boolean) => {
                setShowControls(!fullscreen)
                setDisplayDragRegion(!fullscreen)
            })
        }
    }

    React.useEffect(() => {
        // Get platform
        if (window.electron) {
            setCurrentPlatform(window.electron.platform)
        }

        // Setup window event listeners
        const removeMaximizedListener = window.electron?.on("window:maximized", () => {
            setMaximized(true)
        })

        const removeUnmaximizedListener = window.electron?.on("window:unmaximized", () => {
            setMaximized(false)
        })

        const removeFullscreenListener = window.electron?.on("window:fullscreen", (isFullscreen: boolean) => {
            setShowControls(!isFullscreen)
            setDisplayDragRegion(!isFullscreen)
        })

        // Check window capabilities
        // if (window.electron?.window) {
        //     Promise.all([
        //         window.electron.window.isMinimizable(),
        //         window.electron.window.isMaximizable(),
        //         window.electron.window.isClosable(),
        //         window.electron.window.isMaximized()
        //     ]).then(([minimizable, maximizable, closable, isMaximized]) => {
        //         setMaximized(isMaximized)
        //         setShowControls(minimizable || maximizable || closable)
        //     })
        // }

        document.addEventListener("fullscreenchange", onFullscreenChange)

        // Cleanup
        return () => {
            if (removeMaximizedListener) removeMaximizedListener()
            if (removeUnmaximizedListener) removeUnmaximizedListener()
            if (removeFullscreenListener) removeFullscreenListener()
            document.removeEventListener("fullscreenchange", onFullscreenChange)
        }
    }, [])

    // Only show on Windows and macOS
    if (!(currentPlatform === "win32" || currentPlatform === "darwin")) return null

    return (
        <>
            <div
                className="__electron-window-title-bar scroll-locked-offset bg-transparent fixed top-0 left-0 h-10 z-[999] w-full bg-opacity-90 flex pointer-events-[all]"
                style={{
                    pointerEvents: "all",
                }}
            >
                {displayDragRegion &&
                    <div className="flex flex-1 cursor-grab active:cursor-grabbing" style={{ WebkitAppRegion: "drag" } as any}></div>}
                {(currentPlatform === "win32" && showControls) &&
                    <div className="flex h-10 items-center justify-center gap-1 mr-2 !cursor-default">
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
