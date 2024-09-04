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
            getCurrentWebviewWindow().isMaximized().then((maximized) => {
                setMaximized(maximized)
            })
        })

        return () => {
            listener.then((f) => f())
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
            <div className="__tauri-window-traffic-light scroll-locked-offset bg-transparent fixed top-0 left-0 h-10 z-[999] w-full bg-opacity-90 flex">
                <div className="flex flex-1" data-tauri-drag-region></div>
                {(currentPlatform === "windows" && showTrafficLights) && <div className="flex">
                    <IconButton
                        className="w-11 h-10 duration-0 shadow-none text-white hover:text-white bg-transparent hover:bg-[rgba(255,255,255,0.05)] active:text-white active:bg-[rgba(255,255,255,0.1)] rounded-none"
                        icon={<VscChromeMinimize className="text-[0.95rem]" />}
                        onClick={handleMinimize}
                    />
                    <IconButton
                        className="w-11 h-10 duration-0 shadow-none text-white hover:text-white bg-transparent hover:bg-[rgba(255,255,255,0.05)] active:text-white active:bg-[rgba(255,255,255,0.1)] rounded-none"
                        icon={maximized ? <VscChromeRestore className="text-[0.95rem]" /> : <VscChromeMaximize className="text-[0.95rem]" />}
                        onClick={toggleMaximized}
                    />
                    <IconButton
                        className="w-11 h-10 duration-0 shadow-none text-white hover:text-white bg-transparent hover:bg-red-500 active:bg-red-600 active:text-white rounded-none"
                        icon={<VscChromeClose className="text-[0.95rem]" />}
                        onClick={handleClose}
                    />
                </div>}
            </div>
        </>
    )
}
