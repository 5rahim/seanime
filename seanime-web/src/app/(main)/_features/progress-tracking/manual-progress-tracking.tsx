import { usePlaybackCancelManualTracking, usePlaybackStartManualTracking, usePlaybackSyncCurrentProgress } from "@/api/hooks/playback_manager.hooks"
import { useWebsocketMessageListener } from "@/app/(main)/_hooks/handle-websockets"
import { Button, IconButton } from "@/components/ui/button"
import { cn } from "@/components/ui/core/styling"
import { Modal } from "@/components/ui/modal"
import { WSEvents } from "@/lib/server/ws-events"
import { atom } from "jotai"
import { useAtom } from "jotai/react"
import React from "react"
import { PiPopcornFill } from "react-icons/pi"
import { PlaybackManager_PlaybackState } from "./_lib/playback-manager.types"

type ManualProgressTrackingProps = {
    asSidebarButton?: boolean
}

const __mpt_isWatchingAtom = atom<boolean>(false)
const __mpt_showModalAtom = atom<boolean>(false)

export function ManualProgressTrackingButton(props: ManualProgressTrackingProps) {

    const {
        asSidebarButton,
        ...rest
    } = props

    const [isWatching, setIsWatching] = useAtom(__mpt_isWatchingAtom)
    const [showModal, setShowModal] = useAtom(__mpt_showModalAtom)

    return (
        <>
            {isWatching && (
                <>
                    {asSidebarButton ? (
                        <IconButton
                            intent="primary-subtle"
                            className={cn("animate-pulse")}
                            icon={<PiPopcornFill />}
                            onClick={() => setShowModal(true)}
                        />
                    ) : (
                        <Button
                            intent="primary"
                            className={cn("animate-pulse")}
                            leftIcon={<PiPopcornFill />}
                            onClick={() => setShowModal(true)}
                        >
                            Currently watching
                        </Button>)}
                </>
            )}
        </>
    )
}

export function ManualProgressTracking() {

    const [isWatching, setIsWatching] = useAtom(__mpt_isWatchingAtom)
    const stateRef = React.useRef<PlaybackManager_PlaybackState | null>(null)
    const [state, setState] = React.useState<PlaybackManager_PlaybackState | null>(null)
    const [showModal, setShowModal] = useAtom(__mpt_showModalAtom)

    // Playback state
    useWebsocketMessageListener<PlaybackManager_PlaybackState | null>({
        type: WSEvents.PLAYBACK_MANAGER_MANUAL_TRACKING_PLAYBACK_STATE,
        onMessage: data => {
            if (!isWatching) {
                setIsWatching(true)
            }
            setState(data)
        },
    })

    React.useEffect(() => {
        if (stateRef.current === null) {
            setShowModal(true)
        }
        stateRef.current = state
    }, [state])

    useWebsocketMessageListener({
        type: WSEvents.PLAYBACK_MANAGER_MANUAL_TRACKING_STOPPED,
        onMessage: () => {
            setIsWatching(false)
            setShowModal(false)
            setState(null)
        },
    })

    const { mutate: syncProgress, isPending: isSyncing } = usePlaybackSyncCurrentProgress()

    const { mutate: startManualTracking, isPending: isStarting } = usePlaybackStartManualTracking()

    // Get the server to stop reporting the progress
    const { mutate: cancelManualTracking, isPending: isCanceling } = usePlaybackCancelManualTracking({
        onSuccess: () => {
            setShowModal(false)
            setState(null)
            setIsWatching(false)
        },
    })

    return (
        <>
            <Modal
                open={showModal && isWatching}
                onOpenChange={v => setShowModal(v)}
                title="Progress"
                titleClass="text-center"
                contentClass="!space-y-2 relative max-w-2xl"
            >
                {state && <div className="bg-gray-950 border rounded-md p-4 text-center relative overflow-hidden">
                    <p className="text-[--muted]">Currently watching</p>
                    <h3 className="text-lg font-medium line-clamp-1">{state?.mediaTitle}</h3>
                    <p className="text-2xl font-bold">Episode {state?.episodeNumber}
                        <span className="text-[--muted]">{" / "}{(!!state?.mediaTotalEpisodes && state?.mediaTotalEpisodes > 0)
                            ? state?.mediaTotalEpisodes
                            : "-"}</span></p>
                </div>}

                <div className="flex gap-2 w-full">
                    <Button
                        intent="white"
                        disabled={isSyncing || isStarting || isCanceling}
                        onClick={() => syncProgress()}
                        className="w-full"
                        loading={isSyncing}
                    >
                        Update progress now
                    </Button>
                    <Button
                        intent="alert-subtle"
                        disabled={isSyncing || isStarting || isCanceling}
                        onClick={() => cancelManualTracking()}
                        className="w-full"
                        loading={isCanceling}
                    >
                        Cancel
                    </Button>
                </div>
            </Modal>

        </>
    )
}
