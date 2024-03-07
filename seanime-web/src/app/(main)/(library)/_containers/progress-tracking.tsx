import { serverStatusAtom } from "@/atoms/server-status"
import { useWebsocketMessageListener } from "@/atoms/websocket"
import { Button } from "@/components/ui/button"
import { cn } from "@/components/ui/core/styling"
import { Modal } from "@/components/ui/modal"
import { useBoolean } from "@/hooks/use-disclosure"
import { SeaEndpoints, WSEvents } from "@/lib/server/endpoints"
import { useSeaMutation } from "@/lib/server/query"
import { MediaPlayerPlaybackStatus } from "@/lib/server/types"
import { useQueryClient } from "@tanstack/react-query"
import { useAtomValue } from "jotai/react"
import { useMemo, useState } from "react"
import { PiPopcornFill } from "react-icons/pi"
import { toast } from "sonner"


export function ProgressTracking() {

    const qc = useQueryClient()
    const serverStatus = useAtomValue(serverStatusAtom)

    const showModal = useBoolean(false)
    const isTracking = useBoolean(false)
    const isCompleted = useBoolean(false)

    const [status, setStatus] = useState<MediaPlayerPlaybackStatus | null>(null)

    // const { startMpvPlaybackDetection } = useStartMpvPlaybackDetection()


    const canTrackProgress = useMemo(() => {
        return true
    }, [])

    // Video started
    useWebsocketMessageListener<MediaPlayerPlaybackStatus | null>({
        type: WSEvents.MEDIA_PLAYER_TRACKING_STARTED,
        onMessage: data => {
            isTracking.on()
            showModal.on() // Open the modal
        },
    })

    // Tracking progress
    useWebsocketMessageListener<MediaPlayerPlaybackStatus | null>({
        type: WSEvents.MEDIA_PLAYER_PLAYBACK_STATUS,
        onMessage: data => {

        },
    })

    // Video is completed
    useWebsocketMessageListener<MediaPlayerPlaybackStatus | null>({
        type: WSEvents.MEDIA_PLAYER_VIDEO_COMPLETED,
        onMessage: data => {
        },
    })

    // Request to update progress
    useWebsocketMessageListener<MediaPlayerPlaybackStatus | null>({
        type: WSEvents.MEDIA_PLAYER_PROGRESS_UPDATE_REQUEST,
        onMessage: data => {
        },
    })

    // Stopped
    useWebsocketMessageListener<string>({
        type: WSEvents.MEDIA_PLAYER_TRACKING_STOPPED,
        onMessage: data => {
        },
    })

    const { mutate: updateAniListProgress, isPending } = useSeaMutation<any, { mediaId: number, progress: number, episodes: number }>({
        endpoint: SeaEndpoints.ANILIST_LIST_ENTRY_PROGRESS,
        mutationKey: ["update-anilist-list-entry-progress"],
        onSuccess: async () => {
            toast.success("Progress updated on AniList")
            // setStatus(null)
            // isCompleted.off()
            // if (!serverSideTracking.active) {
            //     showModal.off()
            //     isTracking.off()
            // }
            // await qc.refetchQueries({ queryKey: ["get-media-entry", entry.mediaId] })
            await qc.refetchQueries({ queryKey: ["get-library-collection"] })
            await qc.refetchQueries({ queryKey: ["get-anilist-collection"] })
        },
    })

    const { mutate: updateMALProgress } = useSeaMutation<any, { mediaId: number, progress: number }>({
        endpoint: SeaEndpoints.MAL_LIST_ENTRY_PROGRESS,
        mutationKey: ["update-mal-list-entry-progress"],
        onSuccess: async () => {
            toast.success("Progress updated on MAL")
        },
        onError: () => {
            // Ignore errors
        },
    })

    function handleUpdateProgress() {
        // if (episode) {
        //     updateAniListProgress({ mediaId: entry.mediaId, progress: episode!.progressNumber, episodes: entry.media?.episodes ?? 0 })
        //
        //     // If the media has a MAL ID, update the progress on MAL as well
        //     if (serverStatus?.mal && entry.media?.idMal) {
        //         updateMALProgress({ mediaId: entry.media?.idMal, progress: episode!.episodeNumber })
        //     }
        // } else {
        //     toast.error("Could not detect the episode number.")
        // }
    }

    return (
        <>
            {canTrackProgress && <Button
                intent="primary"
                className={cn("animate-pulse", { "animate-bounce": isCompleted.active })}
                leftIcon={<PiPopcornFill />}
                onClick={showModal.on}
            >
                Currently watching
            </Button>}


            <Modal
                open={showModal.active && canTrackProgress}
                onOpenChange={showModal.off}
                title="Progress"
                titleClass="text-center"
                contentClass="!space-y-2 relative"
            >
                <PiPopcornFill className="text-6xl mx-auto absolute top-4 left-4 text-yellow-100" />
                <div className="bg-gray-950 border rounded-md p-4 text-center">
                    <p className="text-[--muted]">Currently watching</p>
                    <h3 className="text-lg font-medium line-clamp-1">One Piece</h3>
                    <h3 className="text-2xl font-bold">Episode 1075</h3>
                    <p className="text-[--muted]">{}</p>
                </div>
                {!serverStatus?.settings?.library?.autoUpdateProgress && (
                    <p className="text-[--muted] text-center">
                        Your progress will be automatically updated
                    </p>
                )}
                <div className="flex gap-2 justify-center items-center">
                    <Button
                        intent="white"
                        onClick={handleUpdateProgress}
                        loading={isPending}
                        className="w-full"
                        disabled
                    >
                        Play next episode
                    </Button>
                </div>
                {(!!status && isCompleted.active && canTrackProgress) &&
                    <div className="flex gap-2 justify-center items-center">
                        <Button
                            intent="primary"
                            disabled={false}
                            onClick={handleUpdateProgress}
                            loading={isPending}
                            className="w-full"
                        >
                            Confirm
                        </Button>
                    </div>
                }
            </Modal>
        </>
    )

}

function removeSpecificFileExtension(filename: string | undefined): string {
    if (!filename) {
        return ""
    }
    const validExtensions = [".mkv", ".mp4"]
    const lastDotIndex = filename.lastIndexOf(".")

    if (lastDotIndex === -1) {
        // No extension found
        return filename
    }

    const extension = filename.slice(lastDotIndex)

    if (validExtensions.includes(extension)) {
        return filename.slice(0, lastDotIndex)
    }

    return filename
}
