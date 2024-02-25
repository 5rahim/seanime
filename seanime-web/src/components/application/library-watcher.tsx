import { _scannerModalIsOpen } from "@/app/(main)/(library)/_containers/scanner/scanner-modal"
import { useWebsocketMessageListener } from "@/atoms/websocket"
import { Button, CloseButton } from "@/components/ui/button"
import { useBoolean } from "@/hooks/use-disclosure"
import { WSEvents } from "@/lib/server/endpoints"
import { useSetAtom } from "jotai/react"
import React, { useState } from "react"
import { BiSolidBinoculars } from "react-icons/bi"
import { FiSearch } from "react-icons/fi"

type LibraryWatcherProps = {
    children?: React.ReactNode
}

export function LibraryWatcher(props: LibraryWatcherProps) {

    const {
        children,
        ...rest
    } = props

    const [event, setEvent] = useState<string | null>(null)
    const fileAdded = useBoolean(false)
    const fileRemoved = useBoolean(false)

    const setScannerModalOpen = useSetAtom(_scannerModalIsOpen)

    useWebsocketMessageListener<string>({
        type: WSEvents.LIBRARY_WATCHER_FILE_ADDED,
        onMessage: data => {
            console.log("Library watcher", data)
            fileAdded.on()
            setEvent(data)
        },
    })

    useWebsocketMessageListener<string>({
        type: WSEvents.LIBRARY_WATCHER_FILE_REMOVED,
        onMessage: data => {
            console.log("Library watcher", data)
            fileRemoved.on()
            setEvent(data)
        },
    })

    useWebsocketMessageListener<string>({
        type: WSEvents.SCAN_PROGRESS,
        onMessage: data => {
            setEvent(null)
            fileAdded.off()
            fileRemoved.off()
        },
    })

    function handleCancel() {
        setEvent(null)
        fileAdded.off()
        fileRemoved.off()
    }

    if (!event) return null

    return (
        <div className="z-50 fixed bottom-4 right-4">
            <div className="bg-gray-900 border  rounded-xl p-4 w-[400px] min-h-[150px] relative">
                <CloseButton className="absolute top-2 right-2" onClick={handleCancel} />
                <div className="pr-8 space-y-3">
                    <h4 className="flex items-center gap-2">
                        <BiSolidBinoculars className="text-brand-400" />
                        Library watcher
                    </h4>
                    <p className="text-base">
                        A change has been detected in your library, refresh your entries.
                    </p>
                    <div>
                        <Button
                            intent="primary-outline"
                            leftIcon={<FiSearch />}
                            size="sm"
                            onClick={() => setScannerModalOpen(true)}
                            className="rounded-full"
                        >
                            Scan your library
                        </Button>
                    </div>
                </div>
            </div>
        </div>
    )
}
