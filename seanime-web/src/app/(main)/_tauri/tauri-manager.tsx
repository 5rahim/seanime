"use client"

import { listen } from "@tauri-apps/api/event"
import { getCurrentWebviewWindow } from "@tauri-apps/api/webviewWindow"
import { Window } from "@tauri-apps/api/window"
import mousetrap from "mousetrap"
import React from "react"

type TauriManagerProps = {
    children?: React.ReactNode
}

// This is only rendered on the Desktop client
export function TauriManager(props: TauriManagerProps) {

    const {
        children,
        ...rest
    } = props

    React.useEffect(() => {
        const u = listen("message", (event) => {
            const message = event.payload
            console.log("Received message from Rust:", message)
        })

        mousetrap.bind("f11", () => {
            toggleFullscreen()
        })

        mousetrap.bind("esc", () => {
            const appWindow = new Window("main")
            appWindow.isFullscreen().then((isFullscreen) => {
                if (isFullscreen) {
                    toggleFullscreen()
                }
            })
        })

        document.addEventListener("fullscreenchange", toggleFullscreen)

        return () => {
            u.then((f) => f())
            mousetrap.unbind("f11")
            document.removeEventListener("fullscreenchange", toggleFullscreen)
        }
    }, [])

    function toggleFullscreen() {
        const appWindow = new Window("main")

        // Only toggle fullscreen on the main window
        if (getCurrentWebviewWindow().label !== "main") return

        appWindow.isFullscreen().then((fullscreen) => {
            // DEVNOTE: When decorations are not shown in fullscreen move there will be a gap at the bottom of the window (Windows)
            // Hide the decorations when exiting fullscreen
            // Show the decorations when entering fullscreen
            appWindow.setDecorations(!fullscreen)

            appWindow.setFullscreen(!fullscreen)
        })
    }

    return (
        <>

        </>
    )
}
