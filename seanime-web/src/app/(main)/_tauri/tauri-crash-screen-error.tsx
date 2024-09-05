import { emit, listen } from "@tauri-apps/api/event"
import React from "react"

export function TauriCrashScreenError() {

    const [msg, setMsg] = React.useState("")

    React.useEffect(() => {
        emit("crash-screen-loaded").then(() => {})

        const u = listen<string>("crash", (event) => {
            console.log("Received crash event", event.payload)
            setMsg(event.payload)
        })
        return () => {
            u.then((f) => f())
        }
    }, [])

    return (
        <p>
            {msg || "An error occurred"}
        </p>
    )
}
