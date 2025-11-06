"use client"

import { Alert } from "@/components/ui/alert"
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
        <div className="px-4 space-y-4">
            <p>
                {msg || "An error occurred. Closing in 10 seconds."}
            </p>

            <Alert
                intent="warning"
                description="Make sure another instance of Seanime is not running."
            />
        </div>
    )
}
