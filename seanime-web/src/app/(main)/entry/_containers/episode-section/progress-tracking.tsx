import { MediaEntry, MediaPlayerPlaybackStatus } from "@/lib/server/types"
import { useWebsocketMessageListener } from "@/atoms/websocket"
import { SeaEndpoints, WSEvents } from "@/lib/server/endpoints"
import { useBoolean } from "@/hooks/use-disclosure"
import { Modal } from "@/components/ui/modal"
import { Button } from "@/components/ui/button"
import { useMemo, useState } from "react"
import { Badge } from "@/components/ui/badge"
import { useSeaMutation } from "@/lib/server/queries/utils"
import toast from "react-hot-toast"
import { useQueryClient } from "@tanstack/react-query"
import { cn } from "@/components/ui/core"

export function ProgressTracking({ entry }: { entry: MediaEntry }) {

    const qc = useQueryClient()

    const trackerModal = useBoolean(false)
    const isTracking = useBoolean(false)
    const isCompleted = useBoolean(false)
    const [status, setStatus] = useState<MediaPlayerPlaybackStatus | null>(null)


    const episode = useMemo(() => {
        if (status) {
            return entry.episodes?.find(ep => ep.localFile?.name === status.filename)
        }
        return undefined
    }, [status, entry.episodes])

    const canTrackProgress = useMemo(() => {
        return !!episode && (!entry.listData?.progress || (!!entry.listData?.progress && entry.listData?.progress < episode.progressNumber))
    }, [episode, entry.listData?.progress])

    // Video started
    useWebsocketMessageListener<MediaPlayerPlaybackStatus | null>({
        type: WSEvents.MEDIA_PLAYER_TRACKING_STARTED,
        onMessage: data => {
            isTracking.on()
            trackerModal.on()
            // Do not override previous progress tracking when we start another video (for playlists)
            if (!isCompleted.active) {
                isCompleted.off()
                setStatus(data)
            }
        },
    })

    // Video is completed
    useWebsocketMessageListener<MediaPlayerPlaybackStatus | null>({
        type: WSEvents.MEDIA_PLAYER_VIDEO_COMPLETED,
        onMessage: data => {
            isTracking.on()
            trackerModal.on()
            isCompleted.on()
            setStatus(data) // This will override previous tracking (good for playlists)
        },
    })

    // Stopped
    useWebsocketMessageListener<string>({
        type: WSEvents.MEDIA_PLAYER_TRACKING_STOPPED,
        onMessage: data => {
            if (data === "Closed") {
                toast.success("Player closed")
            } else {
                toast.error(data)
            }
            // We reset everything when the player is closed ONLY when the video was not completed
            if (!isCompleted.active) {
                isTracking.off()
                trackerModal.off()
                setStatus(null)
            }
        },
    })

    const { mutate, isPending } = useSeaMutation<any, { mediaId: number, progress: number }>({
        endpoint: SeaEndpoints.ANILIST_LIST_ENTRY_PROGRESS,
        mutationKey: ["update-anilist-list-entry-progress"],
        onSuccess: async () => {
            toast.success("Progress updated")
            setStatus(null)
            isCompleted.off()
            trackerModal.off()
            await qc.refetchQueries({ queryKey: ["get-media-entry", entry.mediaId] })
            await qc.refetchQueries({ queryKey: ["get-library-collection"] })
            await qc.refetchQueries({ queryKey: ["get-anilist-collection"] })
        },
    })

    return (
        <>
            {isTracking.active && <Button
                intent={"success"}
                className={cn({ "animate-pulse": isCompleted.active })}
                onClick={trackerModal.on}
            >
                Update progress
            </Button>}

            <Modal
                isOpen={trackerModal.active}
                onClose={trackerModal.off}
                // isClosable
                // title="Progress"
                // titleClassName={"text-center"}
            >
                <div className="bg-[--background-color] border border-[--border] rounded-md p-4 mb-4 text-center">
                    {(!!status && isCompleted.active && !!episode) ? (
                        <p className={"text-xl"}>Current progress: <Badge size={"lg"}>{episode.progressNumber} <span
                            className={"opacity-60"}>/ {entry.currentEpisodeCount}</span></Badge>
                        </p>
                    ) : (
                        <p>Currently watching</p>
                    )}
                </div>
                <div className={"flex gap-2 justify-center items-center"}>
                    {(!!status && isCompleted.active && canTrackProgress) && <Button
                        intent={"primary"}
                        isDisabled={false}
                        onClick={() => {
                            mutate({ mediaId: entry.mediaId, progress: episode!.progressNumber })
                        }}
                        isLoading={isPending}
                        className="w-full"
                    >
                        Confirm
                    </Button>}
                </div>
            </Modal>
        </>
    )

}