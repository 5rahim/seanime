"use client"

import { ElectronCrashScreenError } from "@/app/(main)/_electron/electron-crash-screen"
import { TauriCrashScreenError } from "@/app/(main)/_tauri/tauri-crash-screen-error"
import { LuffyError } from "@/components/shared/luffy-error"
import { LoadingOverlay } from "@/components/ui/loading-spinner"
import { __isElectronDesktop__, __isTauriDesktop__ } from "@/types/constants"
import React from "react"

export default function Page() {

    return (
        <LoadingOverlay showSpinner={false}>
            <LuffyError title="Something went wrong">
                {__isTauriDesktop__ && <TauriCrashScreenError />}
                {__isElectronDesktop__ && <ElectronCrashScreenError />}
            </LuffyError>
        </LoadingOverlay>
    )

}
