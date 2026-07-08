import { DebridClient_StreamState, StreamAutoSelectStatusPayload, Torrentstream_TorrentStatus } from "@/api/generated/types"
import { useDebridCancelStream } from "@/api/hooks/debrid.hooks"
import { useTorrentstreamStopStream } from "@/api/hooks/torrentstream.hooks"
import { mc_currentTime, mpvCore_stateAtom } from "@/app/(main)/_features/mpv-core/mpv-core.atoms"
import { nativePlayer_stateAtom } from "@/app/(main)/_features/native-player/native-player.atoms"
import { PlaybackManager_PlaybackState } from "@/app/(main)/_features/progress-tracking/_lib/playback-manager.types"
import { vc_currentTime, vc_globalMiniPlayerAtom, vc_miniPlayer, vc_videoElement } from "@/app/(main)/_features/video-core/video-core-atoms"
import { useWebsocketSender } from "@/app/(main)/_hooks/handle-websockets"
import { useWebsocketMessageListener } from "@/app/(main)/_hooks/handle-websockets"
import { clientIdAtom } from "@/app/websocket-provider"
import { ConfirmationDialog, useConfirmationDialog } from "@/components/shared/confirmation-dialog"
import { cn } from "@/components/ui/core/styling"
import { Spinner } from "@/components/ui/loading-spinner"
import { WSEvents } from "@/lib/server/ws-events"
import { atom } from "jotai"
import { useAtom, useAtomValue, useSetAtom } from "jotai/react"
import { AnimatePresence, motion } from "motion/react"
import React, { useEffect, useRef, useState } from "react"
import { BiChevronDown, BiChevronUp, BiDownArrow, BiGroup, BiStop, BiUpArrow } from "react-icons/bi"
import { toast } from "sonner"

export const enum TorrentStreamEvents {
    TorrentLoading = "loading",
    TorrentLoadingFailed = "loading-failed",
    TorrentLoadingStatus = "loading-status",
    TorrentLoaded = "loaded",
    TorrentStartedPlaying = "started-playing",
    TorrentStatus = "status",
    TorrentStopped = "stopped",
    PreloadNextStream = "preload-next-stream",
}

export const __torrentstream__loadingStateAtom = atom<string | null>(null)
export const __torrentstream__isLoadedAtom = atom<boolean>(false)
export const __debridstream_stateAtom = atom<DebridClient_StreamState | null>(null)
export const playPillMinimizedAtom = atom<boolean>(true)

export function PlaybackPlayPill({ isNativePlayerComponent, show }: {
    isNativePlayerComponent?: "control-bar" | "top-section" | "overlay",
    show?: boolean
}) {
    const [mpvCoreState, setMpvCoreState] = useAtom(mpvCore_stateAtom)
    const [nativePlayerState, setNativePlayerState] = useAtom(nativePlayer_stateAtom)
    const videoElement = useAtomValue(vc_videoElement)
    const setMiniPlayer = useSetAtom(vc_miniPlayer)
    const builtInPlayerActive = mpvCoreState.active || nativePlayerState.active
    const clientId = useAtomValue(clientIdAtom)
    const { sendMessage } = useWebsocketSender()

    const [loadingState, setLoadingState] = useAtom(__torrentstream__loadingStateAtom)
    const [isLoaded, setIsLoaded] = useAtom(__torrentstream__isLoadedAtom)
    const [debridState, setDebridState] = useAtom(__debridstream_stateAtom)
    const [minimized, setMinimized] = useAtom(playPillMinimizedAtom)

    const [status, setStatus] = useState<Torrentstream_TorrentStatus | null>(null)
    const [torrentBeingLoaded, setTorrentBeingLoaded] = useState<string | null>(null)
    const [mediaPlayerStartedPlaying, setMediaPlayerStartedPlaying] = useState<boolean>(false)

    const [autoSelectState, setAutoSelectState] = useState<StreamAutoSelectStatusPayload | null>(null)

    const { mutate: stopTorrent, isPending: isStoppingTorrent } = useTorrentstreamStopStream()
    const { mutate: cancelDebrid, isPending: isCancellingDebrid } = useDebridCancelStream()

    const t = useRef<NodeJS.Timeout | null>(null)
    const [showMediaPlayerLoading, setShowMediaPlayerLoading] = useState(false)

    const pillRef = useRef<HTMLDivElement>(null)

    useEffect(() => {
        if (minimized) return

        function handleClickOutside(e: MouseEvent) {
            if (pillRef.current && !pillRef.current.contains(e.target as Node)) {
                setMinimized(true)
            }
        }

        const timer = setTimeout(() => {
            document.addEventListener("mousedown", handleClickOutside)
        }, 100)
        return () => {
            clearTimeout(timer)
            document.removeEventListener("mousedown", handleClickOutside)
        }
    }, [minimized, setMinimized])

    const confirmStop = useConfirmationDialog({
        title: "Stop streaming?",
        description: "Are you sure you want to stop and close the stream?",
        actionText: "Stop stream",
        actionIntent: "alert",
        onConfirm: () => {
            handleStopStream()
        },
    })

    useEffect(() => {
        const timeout = setTimeout(() => {
            setShowMediaPlayerLoading(false)
        }, 2 * 60 * 1000)
        return () => clearTimeout(timeout)
    }, [showMediaPlayerLoading])

    const handleStopStream = React.useCallback(() => {
            if (mpvCoreState.active && clientId) {
                const playbackId = mpvCoreState.playbackInfo?.id || ""
                const playbackType = mpvCoreState.playbackInfo?.playbackType || ""

                setMpvCoreState(draft => {
                    draft.playbackInfo = null
                    draft.playbackError = null
                    draft.loadingState = "Ending stream..."
                    draft.miniPlayer = false
                })

                setTimeout(() => {
                    setMpvCoreState(draft => {
                        draft.active = false
                    })
                }, 700)

                sendMessage({
                    type: WSEvents.MPVCORE,
                    payload: {
                        clientId,
                        type: "terminated",
                        payload: {
                            id: playbackId,
                            clientId,
                            playbackType,
                        },
                    },
                })

                return
            }

            if (nativePlayerState.active && clientId) {
                const playbackId = nativePlayerState.playbackInfo?.id || ""
                const playbackType = nativePlayerState.playbackInfo?.streamType || ""
                videoElement?.pause()
                setMiniPlayer(true)
                setNativePlayerState(draft => {
                    draft.playbackInfo = null
                    draft.playbackError = null
                    draft.loadingState = "Ending stream..."
                })
                setTimeout(() => setNativePlayerState(draft => {
                    draft.active = false
                }), 700)
                sendMessage({
                    type: WSEvents.VIDEOCORE,
                    payload: {
                        clientId,
                        type: "video-terminated",
                        payload: { id: playbackId, clientId, playerType: "native", playbackType },
                    },
                })
                return
            }

            if (debridState) {
                cancelDebrid({
                    options: {
                        removeTorrent: true,
                    },
                }, {
                    onSuccess: () => {
                        setDebridState(null)
                    },
                })
                return
            }

            stopTorrent()
        },
        [clientId, mpvCoreState, nativePlayerState, sendMessage, setMiniPlayer, setMpvCoreState, setNativePlayerState, stopTorrent, videoElement,
            debridState, cancelDebrid, setDebridState])

    useWebsocketMessageListener({
        type: WSEvents.TORRENTSTREAM_STATE,
        onMessage: ({ state, data }: { state: string, data: any }) => {
            if (state !== TorrentStreamEvents.TorrentLoading) {
                if (t.current) clearTimeout(t.current)
            }
            switch (state) {
                case TorrentStreamEvents.TorrentLoading:
                    if (!data) {
                        t.current = setTimeout(() => {
                            setLoadingState("SEARCHING_TORRENTS")
                            setStatus(null)
                            setMediaPlayerStartedPlaying(false)
                        }, 500)
                    } else {
                        setLoadingState(data.state)
                        setTorrentBeingLoaded(data.torrentBeingLoaded)
                        setMediaPlayerStartedPlaying(false)
                    }
                    break
                case TorrentStreamEvents.TorrentLoadingFailed:
                    setLoadingState(null)
                    setStatus(null)
                    setMediaPlayerStartedPlaying(false)
                    break
                case TorrentStreamEvents.TorrentLoaded:
                    setLoadingState("SENDING_STREAM_TO_MEDIA_PLAYER")
                    setIsLoaded(true)
                    setMediaPlayerStartedPlaying(false)
                    break
                case TorrentStreamEvents.TorrentStartedPlaying:
                    setLoadingState(null)
                    setIsLoaded(true)
                    setMediaPlayerStartedPlaying(true)
                    break
                case TorrentStreamEvents.TorrentStopped:
                    setLoadingState(null)
                    setIsLoaded(false)
                    setStatus(null)
                    setMediaPlayerStartedPlaying(false)
                    break
                case TorrentStreamEvents.TorrentStatus:
                    setIsLoaded(true)
                    setStatus(data)
                    break
            }
        },
    })

    useEffect(() => {
        if (process.env.NODE_ENV !== "development") return
            ;
        (window as any).__debugDebridStream = (data: DebridClient_StreamState | null) => {
            if (data) {
                if (data.status === "downloading") {
                    setDebridState(data)
                    setShowMediaPlayerLoading(false)
                    return
                }
                if (data.status === "started") {
                    setDebridState(null)
                    setAutoSelectState(null)
                    setShowMediaPlayerLoading(false)
                    return
                }
                if (data.status === "failed") {
                    setDebridState(null)
                    setAutoSelectState(null)
                    toast.error(data.message)
                    setShowMediaPlayerLoading(false)
                    return
                }
                if (data.status === "ready") {
                    setDebridState(null)
                    setAutoSelectState(null)
                    toast.info("Sending stream to player...", { duration: 1 })
                    setShowMediaPlayerLoading(true)
                    return
                }
            }

            setDebridState(null)
            setShowMediaPlayerLoading(false)
        }

        return () => {
            delete (window as any).__debugDebridStream
        }
    }, [setDebridState])

    useWebsocketMessageListener<DebridClient_StreamState>({
        type: WSEvents.DEBRID_STREAM_STATE,
        onMessage: data => {
            if (data) {
                if (data.status === "downloading") {
                    setDebridState(data)
                    setShowMediaPlayerLoading(false)
                    return
                }
                if (data.status === "started") {
                    setDebridState(null)
                    setAutoSelectState(null)
                    setShowMediaPlayerLoading(false)
                    return
                }
                if (data.status === "failed") {
                    setDebridState(null)
                    setAutoSelectState(null)
                    toast.error(data.message)
                    setShowMediaPlayerLoading(false)
                    return
                }
                if (data.status === "ready") {
                    setDebridState(null)
                    setAutoSelectState(null)
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

    useWebsocketMessageListener<StreamAutoSelectStatusPayload>({
        type: WSEvents.STREAM_AUTO_SELECT_STATUS,
        onMessage: data => {
            if (data) {
                setAutoSelectState(data)
            }
        },
    })

    // Inline native player control-bar formatting
    if (isNativePlayerComponent) {
        if (isNativePlayerComponent === "control-bar" && isLoaded && (mpvCoreState.active || nativePlayerState.active) && status) {
            return (
                <div className={cn("relative justify-left w-fit top-0 h-full flex items-center px-2 truncate", show === false && "hidden")}>
                    <div className="flex-wrap w-fit h-14 flex gap-3 items-center text-sm pointer-events-auto text-[--foreground]">
                        <div className="space-x-1">
                            <BiGroup className="inline-block text-lg" />
                            <span>{status.seeders}</span>
                        </div>
                        <div className="space-x-1">
                            <BiDownArrow className="inline-block mr-1" />
                            {status.downloadSpeed !== "" ? status.downloadSpeed : "0 B/s"}
                        </div>
                        <span className={cn("text-[--muted] font-medium", status.progressPercentage < 5 && "animate-pulse")}>
                            {status.progressPercentage.toFixed(2)}%
                        </span>
                        <div className="space-x-1">
                            <BiUpArrow className="inline-block mr-1" />
                            {status.uploadSpeed !== "" ? status.uploadSpeed : "0 B/s"}
                        </div>
                    </div>
                </div>
            )
        }
        return null
    }

    // Determine if the pill is active
    const isAutoSelecting = !!autoSelectState?.active
    const isTorrentLoading = !!loadingState && loadingState !== "SENDING_STREAM_TO_MEDIA_PLAYER"
    const isTorrentLoaded = isLoaded && !!status
    const isDebridLoading = !!debridState && (debridState.status === "downloading" || debridState.status === "started")
    const isActive = isAutoSelecting || isTorrentLoading || isTorrentLoaded || isDebridLoading

    // Determine active player states
    const videoCoreMiniPlayer = useAtomValue(vc_globalMiniPlayerAtom)
    const isMiniPlayer = (mpvCoreState.active && mpvCoreState.miniPlayer) || (nativePlayerState.active && videoCoreMiniPlayer)
    const isExpandedPlayerActive = (mpvCoreState.active && !mpvCoreState.miniPlayer) || (nativePlayerState.active && !videoCoreMiniPlayer)

    const mpvCurrentTime = useAtomValue(mc_currentTime)
    const videoCoreCurrentTime = useAtomValue(vc_currentTime)
    const isPlayerLoading = (mpvCoreState.active && !!mpvCoreState.loadingState) || (nativePlayerState.active && !!nativePlayerState.loadingState)

    // Hide the floating pill if the player is active in expanded (fullscreen) mode and the media has started playing
    const shouldHideFloating = isExpandedPlayerActive && (mediaPlayerStartedPlaying || mpvCurrentTime > 0 || videoCoreCurrentTime > 0 || !isPlayerLoading)
    const showFloatingPill = isActive && !shouldHideFloating

    // Reset mediaPlayerStartedPlaying to false when loading/selecting starts
    useEffect(() => {
        if (isAutoSelecting || isTorrentLoading || isDebridLoading) {
            setMediaPlayerStartedPlaying(false)
        }
    }, [isAutoSelecting, isTorrentLoading, isDebridLoading])

    const loadingStateStr = React.useMemo(() => {
        if (!loadingState) return ""
        switch (loadingState) {
            case "LOADING":
                return "Loading..."
            case "SEARCHING_TORRENTS":
                return "Selecting file..."
            case "ADDING_TORRENT":
                return torrentBeingLoaded ? `Adding torrent "${torrentBeingLoaded}"` : "Adding torrent..."
            case "CHECKING_TORRENT":
                return torrentBeingLoaded ? `Checking torrent "${torrentBeingLoaded}"` : "Checking torrent..."
            case "SELECTING_FILE":
                return "Selecting file..."
            case "SENDING_STREAM_TO_MEDIA_PLAYER":
                return "Sending stream to player..."
            default:
                return loadingState
        }
    }, [loadingState, torrentBeingLoaded])

    const currentStepDetail = autoSelectState?.stepDetail || debridState?.message || loadingStateStr || ""

    if (!showFloatingPill) return null

    return (
        <div className="fixed top-6 left-1/2 -translate-x-1/2 z-[1000] w-auto pointer-events-auto" ref={pillRef}>
            <motion.div
                layout
                transition={{ type: "spring", stiffness: 300, damping: 28 }}
                className={cn(
                    "bg-gray-950/95 border border-[--border] text-[--foreground] shadow-2xl backdrop-blur-md select-none overflow-hidden",
                    minimized
                        ? (isTorrentLoaded && status ? "rounded-full h-12 w-fit max-w-[420px]" : "rounded-full h-12 w-[320px]")
                        : "rounded-[2rem] p-5 w-[95vw] md:w-[400px]",
                )}
            >
                <AnimatePresence mode="wait" initial={false}>
                    {minimized ? (
                        <motion.div
                            key="minimized"
                            layout="position"
                            initial={{ opacity: 0 }}
                            animate={{ opacity: 1 }}
                            exit={{ opacity: 0 }}
                            transition={{ duration: 0.15 }}
                            className="flex items-center justify-between w-full h-full px-4 cursor-pointer"
                            onClick={() => setMinimized(false)}
                        >
                            {!isTorrentLoaded && <div className="flex items-center gap-2 min-w-0 flex-1 overflow-hidden relative">
                                <AnimatePresence mode="wait" initial={false}>
                                    <motion.span
                                        key={currentStepDetail}
                                        initial={{ opacity: 0, y: 8, filter: "blur(4px)" }}
                                        animate={{ opacity: 1, y: 0, filter: "blur(0px)" }}
                                        exit={{ opacity: 0, y: -8, filter: "blur(4px)" }}
                                        transition={{ duration: 0.25 }}
                                        className="text-[--foreground] text-xs font-semibold truncate block"
                                    >
                                        {currentStepDetail || "Loading..."}
                                    </motion.span>
                                </AnimatePresence>
                            </div>}

                            {isTorrentLoaded && status && (
                                <div className="flex items-center gap-2 text-[12px] font-medium text-[--muted] flex-shrink-0 mr-1 bg-gray-950/40 px-2.5 py-1 rounded-full">
                                    <span className="font-bold text-[--foreground]">{status.progressPercentage.toFixed(1)}%</span>
                                    <span className="flex items-center gap-0.5">
                                        <BiGroup className="size-3" />
                                        {status.seeders || "0"}
                                    </span>
                                    <span className="flex items-center gap-0.5">
                                        <BiDownArrow className="size-3" />
                                        {status.downloadSpeed || "0 B/s"}
                                    </span>
                                    <span className="flex items-center gap-0.5">
                                        <BiUpArrow className="size-3" />
                                        {status.uploadSpeed || "0 B/s"}
                                    </span>
                                </div>
                            )}

                            <div className="flex items-center gap-1 flex-shrink-0">
                                <button
                                    className="p-1 rounded-full hover:bg-[--subtle] text-[--muted] hover:text-[--foreground] transition-colors"
                                    onClick={(e) => {
                                        e.stopPropagation()
                                        setMinimized(false)
                                    }}
                                >
                                    <BiChevronDown className="text-lg" />
                                </button>
                                <button
                                    className="p-1 rounded-full hover:bg-red-900 text-[--muted] hover:text-[--red] transition-colors"
                                    onClick={(e) => {
                                        e.stopPropagation()
                                        confirmStop.open()
                                    }}
                                >
                                    <BiStop className="text-lg" />
                                </button>
                            </div>
                        </motion.div>
                    ) : (
                        <motion.div
                            key="maximized"
                            layout="position"
                            initial={{ opacity: 0 }}
                            animate={{ opacity: 1 }}
                            exit={{ opacity: 0 }}
                            transition={{ duration: 0.15 }}
                            className="flex flex-col gap-4 w-full"
                        >
                            <div className="flex items-center justify-between gap-4 border-b border-[--border]/60 pb-2">
                                <div className="min-w-0 flex-1">
                                    <h3 className="text-sm font-bold text-[--foreground] truncate">
                                        {autoSelectState?.mediaTitle || debridState?.torrentName || "Active Streaming"}
                                    </h3>
                                    <p className="text-[11px] text-[--muted] mt-0.5">
                                        {autoSelectState ? `Episode ${autoSelectState.episode}` : (debridState?.message || "Loading...")}
                                    </p>
                                </div>
                                <div className="flex items-center gap-1 flex-shrink-0">
                                    <button
                                        className="p-1 rounded-full hover:bg-[--subtle] text-[--muted] hover:text-[--foreground] transition-colors"
                                        onClick={() => setMinimized(true)}
                                    >
                                        <BiChevronUp className="text-lg" />
                                    </button>
                                    <button
                                        className="p-1 rounded-full hover:bg-red-900 text-[--red] transition-colors"
                                        onClick={(e) => {
                                            e.stopPropagation()
                                            confirmStop.open()
                                        }}
                                    >
                                        <BiStop className="text-lg" />
                                    </button>
                                </div>
                            </div>

                            {currentStepDetail && !isTorrentLoaded && (
                                <div className="flex items-center gap-2 bg-gray-950/40 border border-[--border] px-3 py-2.5 rounded-xl text-xs">
                                    <Spinner className="size-3.5 text-[--purple] flex-shrink-0" />
                                    <span className="text-[--foreground]/90 font-medium truncate flex-1">
                                        {currentStepDetail}
                                    </span>
                                </div>
                            )}

                            {autoSelectState?.candidates && autoSelectState.candidates.length > 0 && (
                                <div className="flex flex-col gap-1.5 mt-1">
                                    <h4 className="text-[10px] font-bold text-[--muted] uppercase tracking-wider px-1">Top Candidates</h4>
                                    <div className="flex flex-col gap-1.5 max-h-[160px] overflow-y-auto border border-[--border] rounded-xl bg-gray-950 p-1">
                                        {autoSelectState.candidates.map((cand, idx) => {
                                            const isSkipped = cand.status === "skipped"
                                            const isSelected = cand.status === "selected"
                                            const isAnalyzing = cand.status === "analyzing"
                                            return (
                                                <div
                                                    key={cand.name + idx}
                                                    className={cn(
                                                        "flex items-center justify-between gap-3 px-3 py-2 rounded-lg text-xs transition-colors",
                                                        isSkipped && "opacity-40",
                                                        isSelected && "bg-[--subtle] border",
                                                    )}
                                                >
                                                    <div className="min-w-0 flex-1">
                                                        <p
                                                            className={cn("font-medium text-[--foreground] truncate",
                                                                isSkipped && "line-through text-[--muted]")}
                                                        >
                                                            {cand.name}
                                                        </p>
                                                        <p className="text-[9px] text-[--muted] mt-0.5 font-mono">{cand.provider}</p>
                                                    </div>
                                                    <div className="flex items-center gap-2.5 flex-shrink-0">
                                                        <span className="text-[10px] font-bold text-[--muted]">Score: {cand.score}</span>
                                                        {!isTorrentLoaded && <span
                                                            className={cn(
                                                                "px-2 py-0.5 rounded-full text-[9px] font-bold uppercase tracking-wider border border-transparent",
                                                                cand.status === "waiting" && "bg-gray-900 text-[--muted]",
                                                                isSkipped && "bg-red-950 text-[--red] ",
                                                                isAnalyzing && "bg-amber-950 text-[--amber] ",
                                                                isSelected && "bg-green-950 text-[--green] ",
                                                            )}
                                                        >
                                                            {cand.status}
                                                        </span>}
                                                    </div>
                                                </div>
                                            )
                                        })}
                                    </div>
                                </div>
                            )}

                            {isTorrentLoaded && status && (
                                <div className="flex flex-col gap-2 bg-gray-950/30 border border-[--border] p-3 rounded-xl mt-1">
                                    <div className="flex justify-between items-center text-[11px] text-[--muted] font-semibold">
                                        <span>Speed: <strong className="text-[--foreground]">{status.downloadSpeed || "0 B/s"}</strong></span>
                                        <span>Seeders: <strong className="text-[--foreground]">{status.seeders}</strong></span>
                                    </div>
                                    <div className="w-full bg-gray-950 border border-[--border] rounded-full h-1.5 overflow-hidden">
                                        <div
                                            className="bg-[--green] h-full rounded-full transition-all duration-300"
                                            style={{ width: `${status.progressPercentage}%` }}
                                        />
                                    </div>
                                    <div className="flex justify-between items-center text-[10px] text-[--muted] font-medium">
                                        <span>{status.progressPercentage.toFixed(1)}% complete</span>
                                        <span>Upload: {status.uploadSpeed || "0 B/s"}</span>
                                    </div>
                                </div>
                            )}
                        </motion.div>
                    )}
                </AnimatePresence>
            </motion.div>
            <ConfirmationDialog {...confirmStop} />
        </div>
    )
}
