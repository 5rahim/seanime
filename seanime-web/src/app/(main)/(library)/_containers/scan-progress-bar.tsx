"use client"
import { __scanner_isScanningAtom } from "@/app/(main)/(library)/_containers/scanner-modal"

import { useWebsocketMessageListener } from "@/app/(main)/_hooks/websocket.hooks"
import { Spinner } from "@/components/ui/loading-spinner"
import { WSEvents } from "@/lib/server/ws-events"
import { useAtom } from "jotai/react"
import { useState } from "react"

export function ScanProgressBar() {

    const [isScanning] = useAtom(__scanner_isScanningAtom)

    const [progress, setProgress] = useState(0)
    const [status, setStatus] = useState("Scanning...")

    useWebsocketMessageListener<number>({
        type: WSEvents.SCAN_PROGRESS,
        onMessage: data => {
            console.log("Scan progress", data)
            setProgress(data)
            // reset progress
            if (data === 100) {
                setTimeout(() => {
                    setProgress(0)
                }, 2000)
            }
        },
    })

    useWebsocketMessageListener<string>({
        type: WSEvents.SCAN_STATUS,
        onMessage: data => {
            console.log("Scan status", data)
            setStatus(data)
            // reset progress
            if (data === "Scan completed") {
                setTimeout(() => {
                    setStatus("Scanning...")
                }, 2000)
            }
        },
    })

    if (!isScanning) return null

    return (
        <>
            <div className="w-full bg-gray-900 fixed top-0 left-0 z-[100]">
                <div
                    className="bg-brand h-3 text-xs font-medium text-blue-100 text-center p-0.5 leading-none transition-all"
                    style={{ width: progress + "%" }}
                />
            </div>
            <div className="fixed left-0 top-8 w-full flex justify-center z-[100]">
                <div className="bg-gray-900 rounded-full border  py-3 px-6 flex gap-2 items-center">
                    <Spinner className="w-4 h-4" />
                    <p>{progress}% - {status}</p>
                </div>
            </div>
        </>
    )

}
