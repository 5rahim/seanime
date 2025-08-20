"use client"

import React from "react"

export function ElectronCrashScreenError() {
    const [msg, setMsg] = React.useState("")

    React.useEffect(() => {

        if (window.electron) {
            const u = window.electron.on("crash", (msg: string) => {
                console.log("Received crash event", msg)
                setMsg(msg)
            })
            return () => {
                u?.()
            }
        }
    }, [])

    return (
        <p className="px-4">
            {msg || "An error occurred"}
        </p>
    )
}
