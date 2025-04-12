"use client"
import { __scanner_isScanningAtom } from "@/app/(main)/(library)/_containers/scanner-modal"

import { useWebsocketMessageListener } from "@/app/(main)/_hooks/handle-websockets"
import { PageWrapper } from "@/components/shared/page-wrapper"
import { Card, CardDescription, CardHeader } from "@/components/ui/card"
import { Spinner } from "@/components/ui/loading-spinner"
import { ProgressBar } from "@/components/ui/progress-bar"
import { WSEvents } from "@/lib/server/ws-events"
import { useAtom } from "jotai/react"
import React, { useState } from "react"

export function ScanProgressBar() {

    const [isScanning] = useAtom(__scanner_isScanningAtom)

    const [progress, setProgress] = useState(0)
    const [status, setStatus] = useState("Scanning...")

    React.useEffect(() => {
        if (!isScanning) {
            setProgress(0)
            setStatus("Scanning...")
        }
    }, [isScanning])

    useWebsocketMessageListener<number>({
        type: WSEvents.SCAN_PROGRESS,
        onMessage: data => {
            console.log("Scan progress", data)
            setProgress(data)
        },
    })

    useWebsocketMessageListener<string>({
        type: WSEvents.SCAN_STATUS,
        onMessage: data => {
            console.log("Scan status", data)
            setStatus(data)
        },
    })

    if (!isScanning) return null

    return (
        <>
            <div className="w-full bg-gray-950 fixed top-0 left-0 z-[100]" data-scan-progress-bar-container>
                <ProgressBar size="xs" value={progress} />
            </div>
            {/*<div className="fixed left-0 top-8 w-full flex justify-center z-[100]">*/}
            {/*    <div className="bg-gray-900 rounded-full border h-14 px-6 flex gap-2 items-center">*/}
            {/*        <Spinner className="w-4 h-4" />*/}
            {/*        <p>{progress}% - {status}</p>*/}
            {/*    </div>*/}
            {/*</div>*/}
            <div className="z-50 fixed bottom-4 right-4" data-scan-progress-bar-card-container>
                <PageWrapper>
                    <Card className="w-fit max-w-[400px] relative" data-scan-progress-bar-card>
                        <CardHeader>
                            <CardDescription className="flex items-center gap-2 text-base text-[--foregorund]">
                                <Spinner className="size-6" /> {progress}% - {status}
                            </CardDescription>
                        </CardHeader>
                    </Card>
                </PageWrapper>
            </div>
        </>
    )

}
