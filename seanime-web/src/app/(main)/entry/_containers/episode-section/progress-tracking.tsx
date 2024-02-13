import { serverStatusAtom } from "@/atoms/server-status"
import { useWebsocketMessageListener } from "@/atoms/websocket"
import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import { cn } from "@/components/ui/core"
import { Modal } from "@/components/ui/modal"
import { useBoolean } from "@/hooks/use-disclosure"
import { SeaEndpoints, WSEvents } from "@/lib/server/endpoints"
import { useSeaMutation } from "@/lib/server/queries/utils"
import { MediaEntry, MediaEntryEpisode, MediaPlayerPlaybackStatus } from "@/lib/server/types"
import { useQueryClient } from "@tanstack/react-query"
import { useAtomValue } from "jotai/react"
import { useEffect, useMemo, useState } from "react"
import toast from "react-hot-toast"


export function ProgressTracking({ entry }: { entry: MediaEntry }) {

    const qc = useQueryClient()

    const serverStatus = useAtomValue(serverStatusAtom)
    const trackerModal = useBoolean(false)
    const isTracking = useBoolean(false)
    const isCompleted = useBoolean(false)
    const serverSideTracking = useBoolean(false)
    const [status, setStatus] = useState<MediaPlayerPlaybackStatus | null>(null)

    // const { startMpvPlaybackDetection } = useStartMpvPlaybackDetection()

    const [episode, setEpisode] = useState<MediaEntryEpisode | undefined>(!!status
        ? entry.episodes?.find(ep => ep.localFile?.name === status.filename)
        : undefined)

    useEffect(() => {
        if (status) {
            setEpisode(entry.episodes?.find(ep => ep.localFile?.name === status.filename))
        }
    }, [status, entry.episodes])

    const canTrackProgress = useMemo(() => {
        return !!episode && (!entry.listData?.progress || (!!entry.listData?.progress && entry.listData?.progress < episode.progressNumber))
    }, [episode, entry.listData?.progress])

    // Video started
    useWebsocketMessageListener<MediaPlayerPlaybackStatus | null>({
        type: WSEvents.MEDIA_PLAYER_TRACKING_STARTED,
        onMessage: data => {
            // Handle MPV playback detection
            if (data && !episode) {
                setEpisode(entry.episodes?.find(ep => removeSpecificFileExtension(ep.localFile?.name) === removeSpecificFileExtension(data.filename)))
            }
            isTracking.on()
            trackerModal.on()
            serverSideTracking.on()
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

    // Request to update progress
    useWebsocketMessageListener<MediaPlayerPlaybackStatus | null>({
        type: WSEvents.MEDIA_PLAYER_PROGRESS_UPDATE_REQUEST,
        onMessage: data => {
            console.log("Automatic progress update requested")
            handleUpdateProgress()
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
            serverSideTracking.off()
            // We reset everything when the player is closed ONLY when the video was not completed
            if (!isCompleted.active) {
                isTracking.off()
                trackerModal.off()
                setStatus(null)
            }
        },
    })

    const { mutate: updateAniListProgress, isPending } = useSeaMutation<any, { mediaId: number, progress: number, episodes: number }>({
        endpoint: SeaEndpoints.ANILIST_LIST_ENTRY_PROGRESS,
        mutationKey: ["update-anilist-list-entry-progress"],
        onSuccess: async () => {
            toast.success("Progress updated on AniList")
            setStatus(null)
            isCompleted.off()
            if (!serverSideTracking.active) {
                trackerModal.off()
                isTracking.off()
            }
            await qc.refetchQueries({ queryKey: ["get-media-entry", entry.mediaId] })
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
        if (episode) {
            updateAniListProgress({ mediaId: entry.mediaId, progress: episode!.progressNumber, episodes: entry.media?.episodes ?? 0 })

            // If the media has a MAL ID, update the progress on MAL as well
            if (serverStatus?.mal && entry.media?.idMal) {
                updateMALProgress({ mediaId: entry.media?.idMal, progress: episode!.episodeNumber })
            }
        } else {
            toast.error("Could not detect the episode number.")
        }
    }

    return (
        <>
            {isTracking.active && canTrackProgress && <Button
                intent={"success"}
                className={cn({ "animate-pulse": isCompleted.active })}
                onClick={trackerModal.on}
            >
                Update progress
            </Button>}


            <Modal
                isOpen={trackerModal.active && canTrackProgress}
                onClose={trackerModal.off}
                // isClosable
                // title="Progress"
                // titleClassName={"text-center"}
            >
                <div className="bg-[--background-color] border border-[--border] rounded-md p-4 mb-4 text-center">
                    {(!!status && isCompleted.active && !!episode) ? (
                        <p className={"text-xl"}>Current progress: <Badge size={"lg"}>{episode.progressNumber} <span
                            className={"opacity-60"}
                        >/ {entry.currentEpisodeCount}</span></Badge>
                        </p>
                    ) : (
                        <p>Currently watching</p>
                    )}
                </div>
                {serverStatus?.settings?.library?.autoUpdateProgress && (
                    <p className={"text-[--muted] py-2 text-center"}>
                        Your progress will be automatically updated
                    </p>
                )}
                <div className={"flex gap-2 justify-center items-center"}>
                    {(!!status && isCompleted.active && canTrackProgress) && <Button
                        intent={"primary"}
                        isDisabled={false}
                        onClick={handleUpdateProgress}
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
