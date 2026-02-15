import { ElectronCrashScreenError } from "@/app/(main)/_electron/electron-crash-screen"
import { LuffyError } from "@/components/shared/luffy-error"
import { LoadingOverlay } from "@/components/ui/loading-spinner"
import { __isElectronDesktop__ } from "@/types/constants"
import React from "react"

export default function Page() {

    return (
        <LoadingOverlay showSpinner={false}>
            <LuffyError title="Something went wrong">
                {__isElectronDesktop__ && <ElectronCrashScreenError />}
            </LuffyError>
        </LoadingOverlay>
    )

}
