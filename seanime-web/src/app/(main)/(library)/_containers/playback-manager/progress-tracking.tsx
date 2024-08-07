import { API_ENDPOINTS } from "@/api/generated/endpoints"
import {
    usePlaybackCancelCurrentPlaylist,
    usePlaybackPlaylistNext,
    usePlaybackPlayNextEpisode,
    usePlaybackSyncCurrentProgress,
} from "@/api/hooks/playback_manager.hooks"
import {
    PlaybackManager_PlaybackState,
    PlaybackManager_PlaylistState,
} from "@/app/(main)/(library)/_containers/playback-manager/_lib/playback-manager.types"
import { useWebsocketMessageListener } from "@/app/(main)/_hooks/handle-websockets"
import { useServerStatus } from "@/app/(main)/_hooks/use-server-status"
import { ConfirmationDialog, useConfirmationDialog } from "@/components/shared/confirmation-dialog"
import { imageShimmer } from "@/components/shared/image-helpers"
import { Button, IconButton } from "@/components/ui/button"
import { cn } from "@/components/ui/core/styling"
import { Modal } from "@/components/ui/modal"
import { ProgressBar } from "@/components/ui/progress-bar"
import { WSEvents } from "@/lib/server/ws-events"
import { useQueryClient } from "@tanstack/react-query"
import { atom } from "jotai"
import { useAtom } from "jotai/react"
import Image from "next/image"
import React, { useState } from "react"
import { BiSolidSkipNextCircle } from "react-icons/bi"
import { MdCancel } from "react-icons/md"
import { PiPopcornFill } from "react-icons/pi"
import { toast } from "sonner"


const __pt_showModalAtom = atom(false)
const __pt_isTrackingAtom = atom(false)
const __pt_isCompletedAtom = atom(false)

export function ProgressTracking() {

    const serverStatus = useServerStatus()

    const [showModal, setShowModal] = useAtom(__pt_showModalAtom)
    /**
     * Progress tracking states
     * - 'True' when tracking has started
     * - 'False' when tracking has stopped
     */
    const [isTracking, setIsTracking] = useAtom(__pt_isTrackingAtom)
    /**
     * Video completion state
     * - 'True' when the video has been completed
     * - 'False' by default
     */
    const [isCompleted, setIsCompleted] = useAtom(__pt_isCompletedAtom)

    // \/ Modal can be displayed when progress tracking or video is completed
    // Basically, keep the modal visible if there's no more tracking but the video is completed
    const shouldBeDisplayed = isTracking || isCompleted

    const [state, setState] = useState<PlaybackManager_PlaybackState | null>(null)
    const [playlistState, setPlaylistState] = useState<PlaybackManager_PlaylistState | null>(null)

    // Tracking started
    useWebsocketMessageListener<PlaybackManager_PlaybackState | null>({
        type: WSEvents.PLAYBACK_MANAGER_PROGRESS_TRACKING_STARTED,
        onMessage: data => {
            setIsTracking(true)
            setIsCompleted(false)
            setShowModal(true) // Show the modal when tracking starts
            setState(data)
        },
    })

    // Video completed
    useWebsocketMessageListener<PlaybackManager_PlaybackState | null>({
        type: WSEvents.PLAYBACK_MANAGER_PROGRESS_VIDEO_COMPLETED,
        onMessage: data => {
            setIsCompleted(true)
            setState(data)
        },
    })

    // Tracking stopped completely
    useWebsocketMessageListener<string>({
        type: WSEvents.PLAYBACK_MANAGER_PROGRESS_TRACKING_STOPPED,
        onMessage: data => {
            setIsTracking(false)
            if (state?.progressUpdated) {
                setIsCompleted(false) // If the progress has been updated, reset the completed state, so that the modal doesn't show up again
            }
            if (data === "Player closed") {
                toast.info("Player closed")
            } else if (data === "Tracking stopped") {
                toast.info("Tracking stopped")
            } else {
                toast.error(data)
            }
        },
    })

    // Playback state
    useWebsocketMessageListener<PlaybackManager_PlaybackState | null>({
        type: WSEvents.PLAYBACK_MANAGER_PROGRESS_PLAYBACK_STATE,
        onMessage: data => {
            if (!isTracking) {
                setIsTracking(true)
            }
            setState(data)
        },
    })

    const queryClient = useQueryClient()

    // Progress has been updated
    useWebsocketMessageListener<PlaybackManager_PlaybackState | null>({
        type: WSEvents.PLAYBACK_MANAGER_PROGRESS_UPDATED,
        onMessage: data => {
            if (data) {
                if (!serverStatus?.isOffline) {
                    queryClient.invalidateQueries({ queryKey: [API_ENDPOINTS.ANIME_ENTRIES.GetAnimeEntry.key, String(data.mediaId)] })
                    queryClient.invalidateQueries({ queryKey: [API_ENDPOINTS.ANIME_COLLECTION.GetLibraryCollection.key] })
                    queryClient.invalidateQueries({ queryKey: [API_ENDPOINTS.ANILIST.GetAnimeCollection.key] })
                } else {
                    queryClient.invalidateQueries({ queryKey: [API_ENDPOINTS.OFFLINE.GetOfflineSnapshot.key] })
                }
                setState(data)
                toast.success("Progress updated")
            }
        },
    })

    useWebsocketMessageListener<PlaybackManager_PlaylistState | null>({
        type: WSEvents.PLAYBACK_MANAGER_PLAYLIST_STATE,
        onMessage: data => {
            setPlaylistState(data)
        },
    })

    const { mutate: syncProgress, isPending } = usePlaybackSyncCurrentProgress()

    const { mutate: playlistNext, isSuccess: submittedPlaylistNext } = usePlaybackPlaylistNext([playlistState?.current?.name])

    const { mutate: stopPlaylist, isSuccess: submittedStopPlaylist } = usePlaybackCancelCurrentPlaylist([playlistState?.current?.name])

    const { mutate: nextEpisode, isSuccess: submittedNextEpisode, isPending: submittingNextEpisode } = usePlaybackPlayNextEpisode([state?.filename])

    const confirmPlayNext = useConfirmationDialog({
        title: "Play next episode",
        description: "Are you sure you want to play the next episode?",
        actionText: "Confirm",
        actionIntent: "success",
        onConfirm: () => {
            if (!submittedPlaylistNext) playlistNext()
        },
    })

    const confirmNextEpisode = useConfirmationDialog({
        title: "Play next episode",
        description: "Are you sure you want to play the next episode?",
        actionText: "Confirm",
        actionIntent: "success",
        onConfirm: () => {
            if (!submittedNextEpisode) nextEpisode()
        },
    })

    const confirmStopPlaylist = useConfirmationDialog({
        title: "Play next",
        actionText: "Confirm",
        actionIntent: "alert",
        description: "Are you sure you want to stop the playlist? It will be deleted.",
        onConfirm: () => {
            if (!submittedStopPlaylist) stopPlaylist()
        },
    })


    function handleUpdateProgress() {
        syncProgress()
    }

    return (
        <>
            {shouldBeDisplayed && <Button
                intent="primary"
                className={cn("animate-pulse")}
                leftIcon={<PiPopcornFill />}
                onClick={() => setShowModal(true)}
            >
                Currently watching
            </Button>}


            <Modal
                open={showModal && shouldBeDisplayed}
                onOpenChange={v => setShowModal(v)}
                title="Progress"
                titleClass="text-center"
                contentClass="!space-y-2 relative max-w-2xl"
            >
                {state && <div className="bg-gray-950 border rounded-md p-4 text-center relative overflow-hidden">
                    <p className="text-[--muted]">Currently watching</p>
                    <h3 className="text-lg font-medium line-clamp-1">{state?.mediaTitle}</h3>
                    <p className="text-2xl font-bold">Episode {state?.episodeNumber}
                        <span className="text-[--muted]">{" / "}{state?.mediaTotalEpisodes || "-"}</span></p>
                    {!!state?.completionPercentage && <div className="absolute left-0 top-0 w-full">
                        <ProgressBar className="h-2" value={state.completionPercentage * 100} />
                    </div>}
                </div>}
                {(serverStatus?.settings?.library?.autoUpdateProgress && !state?.progressUpdated) && (
                    <p className="text-[--muted] text-center">
                        Your progress will be automatically updated
                    </p>
                )}
                {(state?.progressUpdated) && (
                    <p className="text-green-300 text-center">
                        Progress updated
                    </p>
                )}
                {(
                    !!state?.completionPercentage
                    && state?.completionPercentage > 0.7
                    && !state.progressUpdated
                ) && <div className="flex gap-2 justify-center items-center">
                    <Button
                        intent="primary-subtle"
                        disabled={isPending || state?.progressUpdated}
                        onClick={handleUpdateProgress}
                        className="w-full"
                        loading={isPending}
                    >
                        Update progress now
                    </Button>
                </div>}
                {(
                    !!state?.completionPercentage
                    && state?.completionPercentage > 0.7
                    && state?.canPlayNext
                    && !playlistState
                ) && <div className="flex gap-2 justify-center items-center">
                    <Button
                        intent="white"
                        onClick={() => confirmNextEpisode.open()}
                        className="w-full"
                        disabled={submittedNextEpisode}
                        loading={submittingNextEpisode}
                        leftIcon={<BiSolidSkipNextCircle className="text-2xl" />}
                    >
                        Play next episode
                    </Button>
                </div>}
                {!!playlistState?.next && (
                    <div className="bg-gray-950 border rounded-md p-4 text-center relative overflow-hidden">
                        <div className="space-y-3">
                            <div>
                                <h4 className="text-lg font-medium text-center text-[--muted]">Playlist</h4>
                                {!!playlistState.remaining &&
                                    <p>{playlistState.remaining} episode{playlistState.remaining > 1 ? "s" : ""} after this one</p>}
                                <p className="text-center truncate line-clamp-1">Next: <span className="font-semibold">{playlistState?.next?.name}</span>
                                </p>
                            </div>
                            <div
                                className={cn(
                                    "w-full rounded-md relative overflow-hidden",
                                    submittedPlaylistNext ? "opacity-50 pointer-events-none" : "cursor-pointer",
                                )}
                                onClick={() => {
                                    if (!submittedPlaylistNext) confirmPlayNext.open()
                                }}
                            >
                                {(playlistState.next?.mediaImage) && <Image
                                    src={playlistState.next?.mediaImage || ""}
                                    placeholder={imageShimmer(700, 475)}
                                    sizes="10rem"
                                    fill
                                    alt=""
                                    className="object-center object-cover z-[1]"
                                />}
                                <div className="inset-0 relative z-[2] bg-black border bg-opacity-70 hover:bg-opacity-80 transition flex flex-col gap-2 items-center justify-center p-4">
                                    <p className="flex gap-2 items-center"><BiSolidSkipNextCircle className="block text-2xl" /> Play next</p>
                                </div>
                            </div>
                            <div className="absolute -top-0.5 right-2">
                                <IconButton
                                    intent="alert-subtle"
                                    onClick={() => {
                                        if (!submittedStopPlaylist) confirmStopPlaylist.open()
                                    }}
                                    size="sm"
                                    // className="w-full"
                                    disabled={submittedPlaylistNext}
                                    loading={submittedStopPlaylist}
                                    icon={<MdCancel />}
                                />
                            </div>
                        </div>
                    </div>
                )}
            </Modal>

            <ConfirmationDialog {...confirmPlayNext} />
            <ConfirmationDialog {...confirmStopPlaylist} />
            <ConfirmationDialog {...confirmNextEpisode} />
        </>
    )

}
