import { platform } from "@tauri-apps/plugin-os"
import React from "react"


export function TauriSidebarPaddingMacOS() {

    const [currentPlatform, setCurrentPlatform] = React.useState("")

    React.useEffect(() => {
        setCurrentPlatform(platform())
    }, [])

    if (currentPlatform !== "macos") return null

    return (
        <div className="h-4">

        </div>
    )
}
