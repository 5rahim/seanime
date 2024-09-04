"use client"
import { IconButton } from "@/components/ui/button"
import { getCurrentWindow } from "@tauri-apps/api/window"
import { platform } from "@tauri-apps/plugin-os"
import React from "react"
import { VscChromeClose, VscChromeMaximize, VscChromeMinimize, VscChromeRestore } from "react-icons/vsc"

type TauriWindowsTitleBarProps = {
    children?: React.ReactNode
}

export function TauriWindowsTitleBar(props: TauriWindowsTitleBarProps) {

    const {
        children,
        ...rest
    } = props


    function handleMinimize() {
        getCurrentWindow().minimize()
    }

    const [maximized, setMaximized] = React.useState(true)

    async function toggleMaximized() {
        getCurrentWindow().toggleMaximize()
    }

    function handleClose() {
        getCurrentWindow().close()
    }

    React.useEffect(() => {

        const listener = getCurrentWindow().onResized(() => {
            getCurrentWindow().isMaximized().then((maximized) => {
                setMaximized(maximized)
            })
        })

        return () => {
            listener.then((f) => f())
        }
    }, [])

    const [currentPlatform, setCurrentPlatform] = React.useState("")

    React.useEffect(() => {
        setCurrentPlatform(platform())
    }, [])

    if (currentPlatform !== "windows") return null

    return (
        <>
            <div className="__tauri-windows-traffic-light scroll-locked-offset bg-transparent fixed top-0 left-0 h-10 z-[999] w-full bg-opacity-90 flex">
                <div className="flex flex-1" data-tauri-drag-region></div>
                <div className="flex">
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
                </div>
            </div>
        </>
    )
}
