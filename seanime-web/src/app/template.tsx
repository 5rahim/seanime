import { ElectronManager } from "@/app/(main)/_electron/electron-manager"
import { ElectronWindowTitleBar } from "@/app/(main)/_electron/electron-window-title-bar"
import { websocketConnectedAtom } from "@/app/websocket-provider.tsx"
import { __isElectronDesktop__ } from "@/types/constants"
import { useAtom } from "jotai"
import React from "react"
import { ImSpinner2 } from "react-icons/im"

export default function Template({ children }: { children: React.ReactNode }) {
    const [isConnected] = useAtom(websocketConnectedAtom)
    const pathname = typeof window !== "undefined" ? window.location.pathname : "/"
    const showSpinner = pathname !== "/issue-report" && pathname !== "/scan-log-viewer" && pathname !== "/public/auth"

    return (
        <>
            {__isElectronDesktop__ && <ElectronWindowTitleBar />}
            {__isElectronDesktop__ && <ElectronManager />}
            {!isConnected && showSpinner && <div
                className="fixed right-4 bottom-4 bg-gray-950 border text-sm py-2 px-4 font-semibold rounded-xl z-[100] flex gap-2 items-center opacity-70"
            >
                <ImSpinner2 className="animate-spin text-base" />
                Connecting...
            </div>}
            {children}
        </>
    )
}
