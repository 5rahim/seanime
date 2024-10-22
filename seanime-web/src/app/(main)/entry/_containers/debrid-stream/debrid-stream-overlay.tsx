import { DebridClient_StreamState } from "@/api/generated/types"
import { useDebridCancelStream } from "@/api/hooks/debrid.hooks"
import { useWebsocketMessageListener } from "@/app/(main)/_hooks/handle-websockets"
import { ConfirmationDialog, useConfirmationDialog } from "@/components/shared/confirmation-dialog"
import { AppLayoutStack } from "@/components/ui/app-layout"
import { Button } from "@/components/ui/button"
import { LoadingSpinner, Spinner } from "@/components/ui/loading-spinner"
import { Modal } from "@/components/ui/modal"
import { WSEvents } from "@/lib/server/ws-events"
import { atom } from "jotai/index"
import { useAtom } from "jotai/react"
import React from "react"
import { HiOutlineServerStack } from "react-icons/hi2"
import { toast } from "sonner"

// export const __debridstream_stateAtom = atom<DebridClient_StreamState | null>({
//     status: "downloading",
//     torrentName: "[Seanime] Some Anime - S01E03.mkv",
//     message: "Downloading torrent...",
// })

export const __debridstream_stateAtom = atom<DebridClient_StreamState | null>(null)

export function DebridStreamOverlay() {

    const [state, setState] = useAtom(__debridstream_stateAtom)

    const { mutate: cancelStream, isPending: isCancelling } = useDebridCancelStream()

    const [minimized, setMinimized] = React.useState(false)

    useWebsocketMessageListener<DebridClient_StreamState>({
        type: WSEvents.DEBRID_STREAM_STATE,
        onMessage: data => {
            if (data) {
                if (data.status === "downloading" || data.status === "started") {
                    setState(data)
                    return
                }
                if (data.status === "failed") {
                    setState(null)
                    toast.error(data.message)
                    return
                }
                if (data.status === "ready") {
                    setState(null)
                    toast.info("Sending stream to player...", { duration: 5000 })
                    return
                }
            }
        },
    })

    const confirmCancelAndRemoveTorrent = useConfirmationDialog({
        title: "Cancel and remove torrent",
        description: "Are you sure you want to cancel the stream and remove the torrent?",
        onConfirm: () => {
            cancelStream({
                options: {
                    removeTorrent: true,
                },
            }, {
                onSuccess: () => {
                    setState(null)
                },
            })
        },
    })

    const confirmCancelStream = useConfirmationDialog({
        title: "Cancel stream",
        description: "Are you sure you want to cancel the stream?",
        onConfirm: () => {
            cancelStream({
                options: {
                    removeTorrent: false,
                },
            }, {
                onSuccess: () => {
                    setState(null)
                },
            })
        },
    })

    if (!state) return null

    return (
        <>
            {minimized && (
                <div className="fixed z-[100] bottom-8 w-full h-fit flex justify-center">
                    <div
                        className=" p-4 bg-gray-900 border text-white rounded-xl cursor-pointer hover:border-gray-600"
                        onClick={() => setMinimized(false)}
                    >
                        <div className="flex items-center justify-center gap-4">
                            <HiOutlineServerStack className="text-2xl text-[--brand]" />
                            <div className="">
                                <p>
                                    Awaiting stream from Debrid service
                                </p>
                                <p className="text-[--muted] text-sm text-center">
                                    The stream will launch once it's downloaded
                                </p>
                            </div>
                            <Spinner className="size-5" />
                        </div>
                    </div>
                </div>
            )}
            <Modal
                contentClass="max-w-3xl"
                title="Awaiting stream"
                open={!minimized && !!state}
                onOpenChange={v => setMinimized(!v)}
            >

                <AppLayoutStack>

                    <p className="text-[--muted] italic text-sm">
                        Closing this modal will not cancel the stream
                    </p>

                    <div className="rounded-md border bg-gray-950 p-4 pb-0">
                        <p className="text-center text-sm">
                            {state?.torrentName}
                        </p>

                        <LoadingSpinner
                            title={state?.message}
                        />
                    </div>

                    <div className="flex justify-center gap-4 mt-4">
                        <Button
                            onClick={() => confirmCancelStream.open()}
                            disabled={isCancelling}
                            intent="alert"
                            size="sm"
                        >
                            Cancel
                        </Button>
                        <Button
                            onClick={() => confirmCancelAndRemoveTorrent.open()}
                            disabled={isCancelling}
                            intent="alert"
                            size="sm"
                        >
                            Cancel and remove torrent
                        </Button>
                    </div>

                </AppLayoutStack>


            </Modal>

            <ConfirmationDialog {...confirmCancelStream} />
            <ConfirmationDialog {...confirmCancelAndRemoveTorrent} />
        </>
    )
}
