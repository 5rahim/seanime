import React from "react"

export function ElectronSidebarPaddingMacOS() {
    if ((window as any).electron.platform !== "darwin") return null

    return (
        <div className="h-4">
            {/* Extra padding for macOS */}
        </div>
    )
}
