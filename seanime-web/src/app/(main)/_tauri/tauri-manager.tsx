"use client"

import { listen } from "@tauri-apps/api/event"
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
        const unlisten = listen("message", (event) => {
            const message = event.payload
            console.log("Received message from Rust:", message)
        })

        mousetrap.bind("f11", () => {
            onFullscreenChange()
        })

        mousetrap.bind("esc", () => {
            const appWindow = new Window("main")
            appWindow.isFullscreen().then((isFullscreen) => {
                if (isFullscreen) {
                    appWindow.setFullscreen(false)
                    appWindow.setAlwaysOnTop(false)
                }
            })
        })

        document.addEventListener("fullscreenchange", onFullscreenChange)

        return () => {
            unlisten.then((f) => f())
            mousetrap.unbind("f11")
            document.removeEventListener("fullscreenchange", onFullscreenChange)
        }
    }, [])

    function onFullscreenChange() {
        const appWindow = new Window("main")

        appWindow.isFullscreen().then((isFullscreen) => {
            if (isFullscreen) {
                appWindow.setDecorations(false)
            } else {
                appWindow.setDecorations(true)
            }
            appWindow.setFullscreen(!isFullscreen)
        })
    }

    return (
        <>

        </>
    )
}
