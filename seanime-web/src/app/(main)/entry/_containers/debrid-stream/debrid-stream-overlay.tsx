import { DebridClient_StreamState } from "@/api/generated/types"
import { useDebridCancelStream } from "@/api/hooks/debrid.hooks"
import { PlaybackManager_PlaybackState } from "@/app/(main)/_features/progress-tracking/_lib/playback-manager.types"
import { useWebsocketMessageListener } from "@/app/(main)/_hooks/handle-websockets"
import { ConfirmationDialog, useConfirmationDialog } from "@/components/shared/confirmation-dialog"
import { AppLayoutStack } from "@/components/ui/app-layout"
import { Button } from "@/components/ui/button"
import { LoadingSpinner, Spinner } from "@/components/ui/loading-spinner"
import { Modal } from "@/components/ui/modal"
import { ProgressBar } from "@/components/ui/progress-bar"
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

    const [minimized, setMinimized] = React.useState(true)

    const [showMediaPlayerLoading, setShowMediaPlayerLoading] = React.useState(false)

    // Reset showMediaPlayerLoading after 3 minutes
    React.useEffect(() => {
        const timeout = setTimeout(() => {
            setShowMediaPlayerLoading(false)
        }, 2 * 60 * 1000)
        return () => clearTimeout(timeout)
    }, [showMediaPlayerLoading])

    useWebsocketMessageListener<DebridClient_StreamState>({
        type: WSEvents.DEBRID_STREAM_STATE,
        onMessage: data => {
            if (data) {
                if (data.status === "downloading" || data.status === "started") {
                    setState(data)
                    setShowMediaPlayerLoading(false)
                    return
                }
                if (data.status === "failed") {
                    setState(null)
                    toast.error(data.message)
                    setShowMediaPlayerLoading(false)
                    return
                }
                if (data.status === "ready") {
                    setState(null)
                    toast.info("Sending stream to player...", { duration: 1 })
                    setShowMediaPlayerLoading(true)
                    return
                }
            }
        },
    })

    useWebsocketMessageListener<PlaybackManager_PlaybackState | null>({
        type: WSEvents.PLAYBACK_MANAGER_PROGRESS_TRACKING_STARTED,
        onMessage: data => {
            if (data) {
                setShowMediaPlayerLoading(false)
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

    if (!state) return (
        <>
            {(showMediaPlayerLoading) && <div className="w-full bg-gray-950 fixed top-0 left-0 z-[100]">
                <ProgressBar size="xs" isIndeterminate />
            </div>}
        </>
    )

    return (
        <>

        {minimized && (
                <div className="fixed z-[100] bottom-8 w-full h-fit flex justify-center">
                    <div
                        className="shadow-2xl p-4 bg-gray-900 border text-white rounded-3xl cursor-pointer hover:border-gray-600"
                        onClick={() => setMinimized(false)}
                    >
                        <div className="flex items-center justify-center gap-4">
                            <HiOutlineServerStack className="text-2xl text-[--brand]" />
                            <div className="">
                                <p>
                                    Awaiting stream from Debrid service
                                </p>
                                <p className="text-[--muted] text-sm">
                                    {state?.message}
                                </p>
                            </div>
                            <Spinner className="size-5" />
                        </div>
                    </div>
                </div>
            )}

            {state?.status === "downloading" && <div className="w-full bg-gray-950 fixed top-0 left-0 z-[100]">
                <ProgressBar size="xs" isIndeterminate />
            </div>}

            <Modal
                contentClass="max-w-xl sm:rounded-3xl"
                // title="Awaiting stream"
                open={!minimized && !!state}
                onOpenChange={v => setMinimized(!v)}
            >

                <AppLayoutStack>

                    <p className="text-[--muted] italic text-sm">
                        Closing this modal will not cancel the stream
                    </p>

                    <div className="p-4 pb-0">
                        <p className="text-center text-sm line-clamp-1 tracking-wide">
                            {state?.torrentName}
                        </p>

                        <LoadingSpinner
                            title={state?.message}
                        />
                    </div>

                    <div className="flex justify-center gap-1 mt-4">
                        <Button
                            onClick={() => confirmCancelStream.open()}
                            intent="alert-basic"
                            disabled={isCancelling || state?.status !== "downloading" || state?.message === "Downloading torrent..."}
                            size="sm"
                        >
                            Cancel
                        </Button>
                        <Button
                            onClick={() => confirmCancelAndRemoveTorrent.open()}
                            intent="alert-basic"
                            disabled={isCancelling || state?.status !== "downloading" || state?.message === "Downloading torrent..."}
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
