import { useState } from "react"
import { useWebsocketMessageListener } from "@/atoms/websocket"
import { WSEvents } from "@/lib/server/endpoints"
import { useAtom } from "jotai/react"
import { _scannerIsScanningAtom } from "@/app/(main)/(library)/_components/scanner-modal"

export function ScanProgressBar() {

    const [isScanning] = useAtom(_scannerIsScanningAtom)

    const [progress, setProgress] = useState(0)

    useWebsocketMessageListener<number>({
        type: WSEvents.SCAN_PROGRESS,
        onMessage: data => {
            setProgress(data)
            // reset progress
            if (data === 100) {
                setTimeout(() => {
                    setProgress(0)
                }, 2000)
            }
        },
    })

    if (!isScanning) return null

    return (
        <div className="w-full bg-gray-800 fixed top-0 left-0 z-[100]">
            <div className="bg-brand text-xs font-medium text-blue-100 text-center p-0.5 leading-none transition-all"
                 style={{ width: progress + "%" }}> {progress}%
            </div>
        </div>
    )

}