import { API_ENDPOINTS } from "@/api/generated/endpoints"
import {
    usePlaybackAutoPlayNextEpisode,
    usePlaybackCancelCurrentPlaylist,
    usePlaybackGetNextEpisode,
    usePlaybackPlaylistNext,
    usePlaybackPlayNextEpisode,
    usePlaybackSyncCurrentProgress,
} from "@/api/hooks/playback_manager.hooks"
import { PlaybackManager_PlaybackState, PlaybackManager_PlaylistState } from "@/app/(main)/_features/progress-tracking/_lib/playback-manager.types"
import { useWebsocketMessageListener } from "@/app/(main)/_hooks/handle-websockets"
import { useServerStatus } from "@/app/(main)/_hooks/use-server-status"
import { useDebridStreamAutoplay, useTorrentStreamAutoplay } from "@/app/(main)/entry/_containers/torrent-stream/_lib/handle-torrent-stream"
import { ConfirmationDialog, useConfirmationDialog } from "@/components/shared/confirmation-dialog"
import { imageShimmer } from "@/components/shared/image-helpers"
import { Button, IconButton } from "@/components/ui/button"
import { cn } from "@/components/ui/core/styling"
import { Modal } from "@/components/ui/modal"
import { ProgressBar } from "@/components/ui/progress-bar"
import { logger } from "@/lib/helpers/debug"
import { WSEvents } from "@/lib/server/ws-events"
import { useQueryClient } from "@tanstack/react-query"
import { atom, useAtomValue } from "jotai"
import { useAtom } from "jotai/react"
import mousetrap from "mousetrap"
import Image from "next/image"
import React from "react"
import { BiSolidSkipNextCircle } from "react-icons/bi"
import { MdCancel } from "react-icons/md"
import { PiPopcornFill } from "react-icons/pi"
import { toast } from "sonner"

const __pt_showModalAtom = atom(false)
const __pt_showAutoPlayCountdownModalAtom = atom(false)
const __pt_isTrackingAtom = atom(false)
const __pt_isCompletedAtom = atom(false)

const AUTOPLAY_COUNTDOWN = 6

type Props = {
    asSidebarButton?: boolean
}

export function PlaybackManagerProgressTrackingButton({ asSidebarButton }: Props) {
    const [showModal, setShowModal] = useAtom(__pt_showModalAtom)

    const isTracking = useAtomValue(__pt_isTrackingAtom)

    const isCompleted = useAtomValue(__pt_isCompletedAtom)

    // \/ Modal can be displayed when progress tracking or video is completed
    // Basically, keep the modal visible if there's no more tracking but the video is completed
    const shouldBeDisplayed = isTracking || isCompleted

    return (
        <>
            {shouldBeDisplayed && (
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

export function PlaybackManagerProgressTracking() {
    const serverStatus = useServerStatus()
    const qc = useQueryClient()

    const [showModal, setShowModal] = useAtom(__pt_showModalAtom)
    const [showAutoPlayCountdownModal, setShowAutoPlayCountdownModal] = useAtom(__pt_showAutoPlayCountdownModalAtom)


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

    const [state, setState] = React.useState<PlaybackManager_PlaybackState | null>(null)
    const [playlistState, setPlaylistState] = React.useState<PlaybackManager_PlaylistState | null>(null)


    const [willAutoPlay, setWillAutoPlay] = React.useState(false)
    const [autoPlayInXSeconds, setAutoPlayInXSeconds] = React.useState(AUTOPLAY_COUNTDOWN)
    const {
        data: nextEpisodeForAutoplay,
        mutate: getNextEpisode,
        isPending: isFetchingNextEpisode,
        reset: resetGetNextEp,
    } = usePlaybackGetNextEpisode()

    const { mutate: autoPlayNextEpisode, isPending: isAutoPlayingNextEpisode } = usePlaybackAutoPlayNextEpisode()

    const { mutate: syncProgress, isPending } = usePlaybackSyncCurrentProgress()

    const { mutate: playlistNext, isSuccess: submittedPlaylistNext } = usePlaybackPlaylistNext([playlistState?.current?.name])

    const { mutate: stopPlaylist, isSuccess: submittedStopPlaylist } = usePlaybackCancelCurrentPlaylist([playlistState?.current?.name])

    const { mutate: nextEpisode, isSuccess: submittedNextEpisode, isPending: submittingNextEpisode } = usePlaybackPlayNextEpisode([state?.filename])

    // Tracking started
    useWebsocketMessageListener<PlaybackManager_PlaybackState | null>({
        type: WSEvents.PLAYBACK_MANAGER_PROGRESS_TRACKING_STARTED,
        onMessage: data => {
            logger("PlaybackManagerProgressTracking").info("Tracking started", data)
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
            logger("PlaybackManagerProgressTracking").info("Video completed", data)
            setIsCompleted(true)
            setState(data)
        },
    })

    // TORRENT STREAMING Autoplay
    const { hasNextTorrentstreamEpisode, autoplayNextTorrentstreamEpisode, resetTorrentstreamAutoplayInfo } = useTorrentStreamAutoplay()
    const { hasNextDebridstreamEpisode, autoplayNextDebridstreamEpisode, resetDebridstreamAutoplayInfo } = useDebridStreamAutoplay()

    // Tracking stopped completely
    useWebsocketMessageListener<string>({
        type: WSEvents.PLAYBACK_MANAGER_PROGRESS_TRACKING_STOPPED,
        onMessage: data => {
            logger("PlaybackManagerProgressTracking").info("Tracking stopped", data, "Completion percentage:", state?.completionPercentage)
            setIsTracking(false)
            // Letting 'isCompleted' be true if the progress hasn't been updated
            // so the modal is left available for the user to update the progress manually
            if (state?.progressUpdated) {
                // Setting 'isCompleted' to 'false' to hide the modal
                logger("PlaybackManagerProgressTracking").info("Progress updated, setting isCompleted to false")
                setIsCompleted(false)
            }

            if (data === "Player closed") {
                toast.info("Player closed")

                if ((state?.completionPercentage || 0) <= 0.8) {
                    setIsCompleted(false)
                }
            } else if (data === "Tracking stopped") {
                toast.info("Tracking stopped")
            } else {
                toast.error(data)
            }

            qc.invalidateQueries({ queryKey: [API_ENDPOINTS.CONTINUITY.GetContinuityWatchHistory.key] }).then()

            if (!playlistState && state?.completionPercentage && state?.completionPercentage > 0.7) {
                if (serverStatus?.settings?.library?.autoPlayNextEpisode && !willAutoPlay) {
                    if (!isFetchingNextEpisode) {
                        resetGetNextEp()
                        React.startTransition(() => {
                            setWillAutoPlay(true)
                            getNextEpisode()
                            setAutoPlayInXSeconds(AUTOPLAY_COUNTDOWN)
                        })
                    }
                }
            }
        },
    })

    const timerRef = React.useRef<NodeJS.Timeout | null>(null)
    const countdownRef = React.useRef<NodeJS.Timeout | null>(null)

    function clearTimers() {
        try {
            if (countdownRef.current) {
                clearInterval(countdownRef.current)
            }
            if (timerRef.current) {
                clearTimeout(timerRef.current)
            }
        }
        catch (e) {
        }
    }

    // LOCAL MEDIA Autoplay
    React.useEffect(() => {
        // If the next episode is available and autoplay is enabled
        if (willAutoPlay && !isFetchingNextEpisode && (nextEpisodeForAutoplay || hasNextTorrentstreamEpisode || hasNextDebridstreamEpisode)) {

            // Start the countdown
            setShowAutoPlayCountdownModal(true)
            countdownRef.current = setInterval(() => {
                setAutoPlayInXSeconds(prev => prev - 1)
            }, 1000)

            timerRef.current = setTimeout(() => {
                setShowAutoPlayCountdownModal(false)

                if (!!nextEpisodeForAutoplay) {
                    autoPlayNextEpisode()
                } else if (hasNextTorrentstreamEpisode) {
                    autoplayNextTorrentstreamEpisode()
                } else if (hasNextDebridstreamEpisode) {
                    autoplayNextDebridstreamEpisode()
                }

                setWillAutoPlay(false)

                clearTimers()
            }, AUTOPLAY_COUNTDOWN * 1000)

            return () => {
                clearTimers()
            }
        }
    }, [nextEpisodeForAutoplay, hasNextTorrentstreamEpisode, hasNextDebridstreamEpisode, willAutoPlay, isFetchingNextEpisode])


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
                queryClient.invalidateQueries({ queryKey: [API_ENDPOINTS.ANIME_ENTRIES.GetAnimeEntry.key, String(data.mediaId)] })
                queryClient.invalidateQueries({ queryKey: [API_ENDPOINTS.ANIME_COLLECTION.GetLibraryCollection.key] })
                queryClient.invalidateQueries({ queryKey: [API_ENDPOINTS.ANILIST.GetAnimeCollection.key] })

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


    const confirmPlayNext = useConfirmationDialog({
        title: "Play next episode",
        description: "Are you sure you want to play the next episode?",
        actionText: "Confirm",
        actionIntent: "success",
        onConfirm: () => {
            if (!submittedPlaylistNext) playlistNext()
        },
    })

    // Progress update keyboard shortcuts
    React.useEffect(() => {
        mousetrap.bind("u", () => {
            if (!isPending && state?.completionPercentage && state?.completionPercentage > 0.7) {
                syncProgress()
            }
        })

        mousetrap.bind("space", () => {
            if (!isPending && state?.completionPercentage && state?.completionPercentage > 0.7) {
                setWillAutoPlay(false)
                if (state?.canPlayNext && !playlistState) {
                    nextEpisode()
                }
                if (!!playlistState?.next) {
                    playlistNext()
                }
                // if (hasNextTorrentstreamEpisode) {
                //     autoplayNextTorrentstreamEpisode()
                // }
            }
        })

        return () => {
            mousetrap.unbind("u")
            mousetrap.unbind("space")
        }
    }, [state?.completionPercentage && state?.completionPercentage > 0.7, state?.canPlayNext, !!playlistState?.next])

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

    function cancelAutoPlay() {
        setShowAutoPlayCountdownModal(false)
        setWillAutoPlay(false)
        resetTorrentstreamAutoplayInfo()
        resetDebridstreamAutoplayInfo()
        clearTimers()
    }


    function handleUpdateProgress() {
        syncProgress()
    }

    return (
        <>
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
                        onClick={() => {
                            cancelAutoPlay()
                            confirmNextEpisode.open()
                        }}
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
                                    if (!submittedPlaylistNext) {
                                        cancelAutoPlay()
                                        confirmPlayNext.open()
                                    }
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
                                        if (!submittedStopPlaylist) {
                                            cancelAutoPlay()
                                            confirmStopPlaylist.open()
                                        }
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

            <Modal
                open={showAutoPlayCountdownModal && willAutoPlay}
                onOpenChange={v => {
                    if (!v) {
                        logger("PlaybackManagerProgressTracking").info("Auto play cancelled")
                        cancelAutoPlay()
                    }
                }}
                title="Auto playing next episode"
                titleClass="text-center"
                contentClass="!space-y-2 relative max-w-xl"
            >

                <div className="bg-gray-950 border rounded-md p-4 text-center relative overflow-hidden">
                    <p className="text-[--muted]">Playing next episode in</p>
                    <h3 className="text-2xl font-bold">{autoPlayInXSeconds}</h3>
                </div>

                <div className="flex gap-2 justify-center items-center">
                    <Button
                        intent="alert-subtle"
                        onClick={() => {
                            logger("PlaybackManagerProgressTracking").info("Auto play cancelled")
                            cancelAutoPlay()
                        }}
                        className="w-full"
                    >
                        Cancel
                    </Button>
                </div>

            </Modal>

            <ConfirmationDialog {...confirmPlayNext} />
            <ConfirmationDialog {...confirmStopPlaylist} />
            <ConfirmationDialog {...confirmNextEpisode} />
        </>
    )

}
