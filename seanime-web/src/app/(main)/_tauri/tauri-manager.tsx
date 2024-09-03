"use client"

import { listen } from "@tauri-apps/api/event"
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

    const [data, setData] = React.useState()

    React.useEffect(() => {
        const unlisten = listen("message", (event) => {
            const message = event.payload
            console.log("Received message from Rust:", message)
        })
        return () => {
            unlisten.then((f) => f())
        }
    }, [])

    return (
        <>

        </>
    )
}
