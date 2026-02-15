import { ElectronManager } from "@/app/(main)/_electron/electron-manager"
import { ElectronWindowTitleBar } from "@/app/(main)/_electron/electron-window-title-bar"
import { __isElectronDesktop__ } from "@/types/constants"
import React from "react"

export default function Template({ children }: { children: React.ReactNode }) {
    return (
        <>
            {__isElectronDesktop__ && <ElectronWindowTitleBar />}
            {__isElectronDesktop__ && <ElectronManager />}
            {children}
        </>
    )
}
