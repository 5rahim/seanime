"use client"

import { ElectronWindowTitleBar } from "@/app/(main)/_electron/electron-window-title-bar"
import { TauriWindowTitleBar } from "@/app/(main)/_tauri/tauri-window-title-bar"
import { __isElectronDesktop__, __isTauriDesktop__ } from "@/types/constants"
import React from "react"

export default function Template({ children }: { children: React.ReactNode }) {
    return (
        <>
            {__isTauriDesktop__ && <TauriWindowTitleBar />}
            {__isElectronDesktop__ && <ElectronWindowTitleBar />}
            {children}
        </>
    )
}
