"use client"

import { TauriWindowTitleBar } from "@/app/(main)/_tauri/tauri-window-title-bar"
import React from "react"

export default function Template({ children }: { children: React.ReactNode }) {
    return (
        <>
            {process.env.NEXT_PUBLIC_PLATFORM === "desktop" && <TauriWindowTitleBar />}
            {children}
        </>
    )
}
