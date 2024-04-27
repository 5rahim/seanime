import { API_ENDPOINTS } from "@/api/generated/endpoints"
import { __scanner_modalIsOpen } from "@/app/(main)/(library)/_containers/scanner-modal"

import { useWebsocketMessageListener } from "@/app/(main)/_hooks/handle-websockets"
import { useServerStatus } from "@/app/(main)/_hooks/use-server-status"
import { PageWrapper } from "@/components/shared/page-wrapper"
import { Button, CloseButton } from "@/components/ui/button"
import { Card, CardDescription, CardFooter, CardHeader, CardTitle } from "@/components/ui/card"
import { Spinner } from "@/components/ui/loading-spinner"
import { useBoolean } from "@/hooks/use-disclosure"
import { WSEvents } from "@/lib/server/ws-events"
import { useQueryClient } from "@tanstack/react-query"
import { useSetAtom } from "jotai/react"
import React, { useState } from "react"
import { BiSolidBinoculars } from "react-icons/bi"
import { FiSearch } from "react-icons/fi"
import { toast } from "sonner"

type LibraryWatcherProps = {
    children?: React.ReactNode
}

export function LibraryWatcher(props: LibraryWatcherProps) {

    const {
        children,
        ...rest
    } = props

    const qc = useQueryClient()
    const serverStatus = useServerStatus()
    const [fileEvent, setFileEvent] = useState<string | null>(null)
    const fileAdded = useBoolean(false)
    const fileRemoved = useBoolean(false)
    const autoScanning = useBoolean(false)
    const [progress, setProgress] = useState(0)

    const setScannerModalOpen = useSetAtom(__scanner_modalIsOpen)

    useWebsocketMessageListener<string>({
        type: WSEvents.LIBRARY_WATCHER_FILE_ADDED,
        onMessage: data => {
            console.log("Library watcher", data)
            if (!serverStatus?.settings?.library?.autoScan) { // Only show the notification if auto scan is disabled
                fileAdded.on()
                setFileEvent(data)
            }
        },
    })

    useWebsocketMessageListener<string>({
        type: WSEvents.LIBRARY_WATCHER_FILE_REMOVED,
        onMessage: data => {
            console.log("Library watcher", data)
            if (!serverStatus?.settings?.library?.autoScan) { // Only show the notification if auto scan is disabled
                fileRemoved.on()
                setFileEvent(data)
            }
        },
    })

    // Scan progress event
    useWebsocketMessageListener<number>({
        type: WSEvents.SCAN_PROGRESS,
        onMessage: data => {
            // Remove notification of file added or removed
            setFileEvent(null)
            fileAdded.off()
            fileRemoved.off()
            setProgress(data)
            // reset progress
            if (data === 100) {
                setTimeout(() => {
                    setProgress(0)
                }, 2000)
            }
        },
    })

    // Auto scan event started
    useWebsocketMessageListener<string>({
        type: WSEvents.AUTO_SCAN_STARTED,
        onMessage: _ => {
            autoScanning.on()
        },
    })
    // Auto scan event completed
    useWebsocketMessageListener<string>({
        type: WSEvents.AUTO_SCAN_COMPLETED,
        onMessage: _ => {
            autoScanning.off()
            toast.success("Library scanned")
            qc.invalidateQueries({ queryKey: [API_ENDPOINTS.ANIME_COLLECTION.GetLibraryCollection.key] })
            qc.invalidateQueries({ queryKey: [API_ENDPOINTS.ANIME_ENTRIES.GetMissingEpisodes.key] })
            qc.invalidateQueries({ queryKey: [API_ENDPOINTS.AUTO_DOWNLOADER.GetAutoDownloaderItems.key] })
        },
    })

    function handleCancel() {
        setFileEvent(null)
        fileAdded.off()
        fileRemoved.off()
    }

    if (autoScanning.active && progress > 0) {
        return (
            <div className="z-50 fixed bottom-4 right-4">
                <PageWrapper>
                    <Card className="w-fit max-w-[400px]">
                        <CardHeader>
                            <CardDescription className="flex items-center gap-2 text-base">
                                <Spinner className="size-6" /> {progress}% Refreshing your library...
                            </CardDescription>
                        </CardHeader>
                    </Card>
                </PageWrapper>
            </div>
        )
    } else if (!!fileEvent) {
        return (
            <div className="z-50 fixed bottom-4 right-4">
                <PageWrapper>
                    <Card className="w-full max-w-[400px] min-h-[150px] relative">
                        <CardHeader>
                            <CardTitle className="text-lg flex items-center gap-2">
                                <BiSolidBinoculars className="text-brand-400" />
                                Library watcher
                            </CardTitle>
                            <CardDescription className="flex items-center gap-2 text-base">
                                A change has been detected in your library, refresh your entries.
                            </CardDescription>
                        </CardHeader>
                        <CardFooter>
                            <Button
                                intent="primary-outline"
                                leftIcon={<FiSearch />}
                                size="sm"
                                onClick={() => setScannerModalOpen(true)}
                                className="rounded-full"
                            >
                                Scan your library
                            </Button>
                        </CardFooter>
                        <CloseButton className="absolute top-2 right-2" onClick={handleCancel} />
                    </Card>
                </PageWrapper>
            </div>
        )
    }

    return null
}
