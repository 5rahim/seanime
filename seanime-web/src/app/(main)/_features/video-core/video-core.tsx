import { getServerBaseUrl } from "@/api/client/server-url"
import { API_ENDPOINTS } from "@/api/generated/endpoints"
import { useHandleCurrentMediaContinuity } from "@/api/hooks/continuity.hooks"
import { useDirectstreamConvertSubs } from "@/api/hooks/directstream.hooks"
import { useCancelDiscordActivity } from "@/api/hooks/discord.hooks"
import { useNakamaWatchParty } from "@/app/(main)/_features/nakama/nakama-manager"
import { nativePlayer_initialState, nativePlayer_stateAtom } from "@/app/(main)/_features/native-player/native-player.atoms"
import { AniSkipTime } from "@/app/(main)/_features/sea-media-player/aniskip"
import { vc_anime4kOption, VideoCoreAnime4K } from "@/app/(main)/_features/video-core/video-core-anime-4k"
import { Anime4KOption, VideoCoreAnime4KManager } from "@/app/(main)/_features/video-core/video-core-anime-4k-manager"
import { VideoCoreAudioManager } from "@/app/(main)/_features/video-core/video-core-audio"
import { VideoCoreAudioMenu } from "@/app/(main)/_features/video-core/video-core-audio-menu"
import {
    vc_hoveringControlBar,
    VideoCoreControlBar,
    VideoCoreFullscreenButton,
    VideoCoreMobileControlBar,
    VideoCorePipButton,
    VideoCorePlayButton,
    VideoCoreTimestamp,
    VideoCoreVolumeButton,
} from "@/app/(main)/_features/video-core/video-core-control-bar"
import { VideoCoreDrawer } from "@/app/(main)/_features/video-core/video-core-drawer"
import { useVideoCoreSetupEvents } from "@/app/(main)/_features/video-core/video-core-events"
import { vc_fullscreenManager, VideoCoreFullscreenManager } from "@/app/(main)/_features/video-core/video-core-fullscreen"
import {
    useVideoCoreHls,
    vc_hlsAudioTracks,
    vc_hlsCurrentAudioTrack,
    vc_hlsCurrentQuality,
    vc_hlsQualityLevels,
    vc_hlsSetAudioTrack,
    vc_hlsSetQuality,
} from "@/app/(main)/_features/video-core/video-core-hls"
import { useVideoCoreIOSFullscreenSubtitles } from "@/app/(main)/_features/video-core/video-core-ios-fullscreen-subtitles"
import { MediaCaptionsManager } from "@/app/(main)/_features/video-core/video-core-media-captions"
import { vc_mediaSessionManager, VideoCoreMediaSessionManager } from "@/app/(main)/_features/video-core/video-core-media-session"
import { vc_menuOpen, vc_menuSectionOpen } from "@/app/(main)/_features/video-core/video-core-menu"
import { useVideoCoreMobileGestures } from "@/app/(main)/_features/video-core/video-core-mobile-gestures"
import {
    vc_overlayFeedback,
    vc_overlayFeedbackTimeout,
    vc_showOverlayFeedback,
    VideoCoreOverlayDisplay,
} from "@/app/(main)/_features/video-core/video-core-overlay-display"
import { vc_pipElement, vc_pipManager, VideoCorePipManager } from "@/app/(main)/_features/video-core/video-core-pip"
import {
    useVideoCorePlaylist,
    useVideoCorePlaylistSetup,
    vc_playlistState,
    VideoCorePlaylistControl,
} from "@/app/(main)/_features/video-core/video-core-playlist"
import { VideoCoreKeybindingController, VideoCorePreferencesModal } from "@/app/(main)/_features/video-core/video-core-preferences"
import { VideoCorePreviewManager } from "@/app/(main)/_features/video-core/video-core-preview"
import { VideoCoreResolutionMenu } from "@/app/(main)/_features/video-core/video-core-resolution-menu"
import { VideoCoreSettingsMenu } from "@/app/(main)/_features/video-core/video-core-settings-menu"
import { VideoCoreSubtitleMenu } from "@/app/(main)/_features/video-core/video-core-subtitle-menu"
import { VideoCoreSubtitleManager } from "@/app/(main)/_features/video-core/video-core-subtitles"
import { vc_timeRangeElement, VideoCoreTimeRange } from "@/app/(main)/_features/video-core/video-core-time-range"
import { VideoCoreTopPlaybackInfo, VideoCoreTopSection } from "@/app/(main)/_features/video-core/video-core-top-section"
import { VideoCoreWatchPartyChat } from "@/app/(main)/_features/video-core/video-core-watch-party-chat"
import {
    vc_autoNextAtom,
    vc_autoPlayVideoAtom,
    vc_autoSkipOPEDAtom,
    vc_beautifyImageAtom,
    vc_settings,
    vc_storedMutedAtom,
    vc_storedPlaybackRateAtom,
    vc_storedVolumeAtom,
    VideoCore_VideoPlaybackInfo,
    VideoCore_VideoSource,
    VideoCore_VideoSubtitleTrack,
    VideoCoreLifecycleState,
} from "@/app/(main)/_features/video-core/video-core.atoms"
import {
    useVideoCoreBindings,
    vc_createChapterCues,
    vc_createChaptersFromAniSkip,
    vc_formatTime,
    vc_logGeneralInfo,
} from "@/app/(main)/_features/video-core/video-core.utils"
import { useServerStatus } from "@/app/(main)/_hooks/use-server-status"
import { __torrentSearch_selectedTorrentsAtom } from "@/app/(main)/entry/_containers/torrent-search/torrent-search-container"
import {
    __torrentSearch_selectionAtom,
    __torrentSearch_selectionEpisodeAtom,
} from "@/app/(main)/entry/_containers/torrent-search/torrent-search-drawer"
import { TorrentStreamOverlay } from "@/app/(main)/entry/_containers/torrent-stream/torrent-stream-overlay"
import { GradientBackground } from "@/components/shared/gradient-background"
import { LuffyError } from "@/components/shared/luffy-error"
import { Button, IconButton } from "@/components/ui/button"
import { cn } from "@/components/ui/core/styling"
import { LoadingSpinner } from "@/components/ui/loading-spinner"
import { logger } from "@/lib/helpers/debug"
import { __isDesktop__ } from "@/types/constants"
import { useQueryClient } from "@tanstack/react-query"
import { ErrorData } from "hls.js"
import { atom } from "jotai"
import { derive } from "jotai-derive"
import { ScopeProvider } from "jotai-scope"
import { useAtom, useAtomValue, useSetAtom } from "jotai/react"
import React, { useMemo, useRef, useState } from "react"
import { BiExpand, BiX } from "react-icons/bi"
import { FiMinimize2 } from "react-icons/fi"
import { ImSpinner2 } from "react-icons/im"
import { PiSpinnerDuotone } from "react-icons/pi"
import { RemoveScrollBar } from "react-remove-scroll-bar"
import { useMeasure, useUnmount, useUpdateEffect, useWindowSize } from "react-use"

const log = logger("VIDEO CORE")

export const VIDEOCORE_DEBUG_ELEMENTS = false

const DELAY_BEFORE_NOT_BUSY = 1_000 //ms

export const vc_activePlayerId = atom<string | null>(null)

export const vc_isMobile = atom(false)
export const vc_isSwiping = atom(false) // Mobile swipe state
export const vc_swipeSeekTime = atom<number | null>(null) // Mobile swipe seek time
export const vc_videoSize = atom({ width: 1, height: 1 })
export const vc_realVideoSize = atom({ width: 0, height: 0 })
export const vc_duration = atom(1)
export const vc_currentTime = atom(0)
export const vc_playbackRate = atom(1)
export const vc_readyState = atom(0)
export const vc_buffering = atom(false)
export const vc_isMuted = atom(false)
export const vc_volume = atom(1)
export const vc_subtitleDelay = atom(0)
export const vc_isFullscreen = atom(false)
export const vc_pip = derive([vc_pipElement], (pipElement) => pipElement !== null)
export const vc_seeking = atom(false)
export const vc_seekingTargetProgress = atom(0) // 0-100
export const vc_timeRanges = atom<TimeRanges | null>(null)
export const vc_closestBufferedTime = derive([vc_timeRanges, vc_currentTime], (tr, currentTime) => {
    if (!tr) return 0
    let closest = 0
    for (let i = 0; i < tr.length; i++) {
        const start = tr.start(i)
        const end = tr.end(i)
        if (currentTime >= start && currentTime <= end) {
            return end
        }
        if (end >= currentTime && closest > end) {
            closest = end
        }
    }
    return closest
})
export const vc_ended = atom(false)
export const vc_paused = atom(true)
export const vc_miniPlayer = atom(false)
export const vc_cursorBusy = derive([vc_hoveringControlBar, vc_menuOpen], (f1, f2) => {
    return f1 || !!f2
})
export const vc_cursorPosition = atom({ x: 0, y: 0 })
export const vc_busy = atom(true)

export const vc_videoElement = atom<HTMLVideoElement | null>(null)
export const vc_containerElement = atom<HTMLDivElement | null>(null)

export const vc_subtitleManager = atom<VideoCoreSubtitleManager | null>(null)
export const vc_mediaCaptionsManager = atom<MediaCaptionsManager | null>(null)
export const vc_audioManager = atom<VideoCoreAudioManager | null>(null)
export const vc_previewManager = atom<VideoCorePreviewManager | null>(null)
export const vc_anime4kManager = atom<VideoCoreAnime4KManager | null>(null)

export const vc_previousPausedState = atom(false)

export const vc_lastKnownProgress = atom<{ mediaId: number, progressNumber: number, time: number } | null>(null)

export const vc_skipOpeningTime = atom<number | null>(null)
export const vc_skipEndingTime = atom<number | null>(null)

type VideoCoreAction = "seekTo" | "seek" | "togglePlay"

export const vc_dispatchAction = atom(null, (get, set, action: { type: VideoCoreAction; payload?: any }) => {
    const videoElement = get(vc_videoElement)
    const duration = get(vc_duration)
    let t = 0
    if (videoElement) {
        switch (action.type) {
            // for smooth seeking, we don't want to peg the current time to the actual video time
            // instead act like the target time is instantly reached
            case "seekTo":
                if (isNaN(duration) || duration <= 1) return
                t = Math.min(duration, Math.max(0, action.payload.time))
                videoElement.currentTime = t
                set(vc_currentTime, t)
                if (action.payload.flashTime) {
                    set(vc_showOverlayFeedback, { message: `${vc_formatTime(t)} / ${vc_formatTime(duration)}`, type: "message" })
                }
                break
            case "seek":
                if (isNaN(duration) || duration <= 1) return
                const currentTime = get(vc_currentTime)
                t = Math.min(duration, Math.max(0, currentTime + action.payload.time))
                videoElement.currentTime = t
                set(vc_currentTime, t)
                if (action.payload.flashTime) {
                    set(vc_showOverlayFeedback, { message: `${vc_formatTime(t)} / ${vc_formatTime(duration)}`, type: "message" })
                }
                break
            case "togglePlay":
                videoElement.paused ? videoElement.play() : videoElement.pause()
                break
        }
    }
})

export function VideoCoreProvider(props: { id: string, children: React.ReactNode }) {
    const { children } = props

    const [activePlayer, setActivePlayer] = useAtom(vc_activePlayerId)
    const setNativePlayerState = useSetAtom(nativePlayer_stateAtom)

    React.useLayoutEffect(() => {
        if (props.id === "native-player") return

        setActivePlayer(props.id)
        setNativePlayerState(nativePlayer_initialState)

        return () => {
            setActivePlayer(null)
        }
    }, [])

    if (activePlayer !== null && activePlayer !== props.id) return null

    return (
        <ScopeProvider
            atoms={[
                vc_videoSize,
                vc_realVideoSize,
                vc_duration,
                vc_currentTime,
                vc_playbackRate,
                vc_readyState,
                vc_buffering,
                vc_isMuted,
                vc_volume,
                vc_subtitleDelay,
                // vc_isFullscreen, expose this
                vc_seeking,
                vc_seekingTargetProgress,
                vc_timeRanges,
                vc_ended,
                vc_paused,
                vc_miniPlayer,
                vc_cursorPosition,
                vc_busy,
                vc_videoElement,
                vc_containerElement,
                vc_subtitleManager,
                vc_audioManager,
                vc_previewManager,
                vc_anime4kManager,
                vc_pipManager,
                vc_fullscreenManager,
                vc_mediaSessionManager,
                vc_pipElement,
                vc_previousPausedState,
                vc_lastKnownProgress,
                vc_skipOpeningTime,
                vc_skipEndingTime,
                vc_dispatchAction,
                vc_hoveringControlBar,
                vc_menuOpen,
                vc_menuSectionOpen,
                vc_showOverlayFeedback,
                vc_overlayFeedback,
                vc_overlayFeedbackTimeout,
                vc_playlistState,
                vc_timeRangeElement,
                vc_hlsQualityLevels,
                vc_hlsCurrentQuality,
                vc_hlsSetQuality,
                vc_hlsAudioTracks,
                vc_hlsCurrentAudioTrack,
                vc_hlsSetAudioTrack,
                vc_isSwiping,
                vc_isMobile,
                vc_swipeSeekTime,
            ]}
        >
            {children}
        </ScopeProvider>
    )
}

export type VideoCoreChapterCue = {
    startTime: number
    endTime: number
    text: string
}

interface PlayerContentProps {
    videoRef: React.MutableRefObject<HTMLVideoElement | null>
    inline: boolean
    state: VideoCoreLifecycleState
    chapterCues: VideoCoreChapterCue[] | undefined
    aniSkipData: VideoCoreProps["aniSkipData"]
    streamUrl: string | undefined
    combineRef: (instance: HTMLVideoElement | null) => void
    combineContainerRef: (instance: HTMLDivElement | null) => void
    handleContainerPointerMove: (e: React.PointerEvent<HTMLDivElement>) => void
    handleClick: (e: React.MouseEvent<HTMLDivElement>) => void
    handleLoadedMetadata: (e: React.SyntheticEvent<HTMLVideoElement>) => void
    handleTimeUpdate: (e: React.SyntheticEvent<HTMLVideoElement>) => void
    handleEnded: (e: React.SyntheticEvent<HTMLVideoElement>) => void
    handlePlay: (e: React.SyntheticEvent<HTMLVideoElement>) => void
    handlePause: (e: React.SyntheticEvent<HTMLVideoElement>) => void
    handleDoubleClick: (e: React.MouseEvent<HTMLVideoElement>) => void
    handleLoadedData: (e: React.SyntheticEvent<HTMLVideoElement>) => void
    handleVolumeChange: (e: React.SyntheticEvent<HTMLVideoElement>) => void
    handleRateChange: (e: React.SyntheticEvent<HTMLVideoElement>) => void
    handleError: (e: React.SyntheticEvent<HTMLVideoElement>) => void
    handleWaiting: (e: React.SyntheticEvent<HTMLVideoElement>) => void
    handleCanPlay: (e: React.SyntheticEvent<HTMLVideoElement>) => void
    handleStalled: (e: React.SyntheticEvent<HTMLVideoElement>) => void
    onTerminateStream: () => void
    onVideoSourceChange: ((source: VideoCore_VideoSource) => void) | undefined
}

const PlayerContent = React.memo<PlayerContentProps>(({
    videoRef,
    inline,
    state,
    chapterCues,
    aniSkipData,
    streamUrl,
    combineRef,
    combineContainerRef,
    handleContainerPointerMove,
    handleClick,
    handleLoadedMetadata,
    handleTimeUpdate,
    handleEnded,
    handlePlay,
    handlePause,
    handleDoubleClick,
    handleLoadedData,
    handleVolumeChange,
    handleRateChange,
    handleError,
    handleWaiting,
    handleCanPlay,
    handleStalled,
    onTerminateStream,
    onVideoSourceChange,
}) => {
    const isMobile = useAtomValue(vc_isMobile)
    const isMiniPlayer = useAtomValue(vc_miniPlayer)
    const busy = useAtomValue(vc_busy)
    const paused = useAtomValue(vc_paused)
    const buffering = useAtomValue(vc_buffering)
    const settings = useAtomValue(vc_settings)
    const beautifyImage = useAtomValue(vc_beautifyImageAtom)
    const isPip = useAtomValue(vc_pip)
    const skipOpeningTime = useAtomValue(vc_skipOpeningTime)
    const skipEndingTime = useAtomValue(vc_skipEndingTime)
    const pipManager = useAtomValue(vc_pipManager)
    const action = useSetAtom(vc_dispatchAction)
    const [autoPlay] = useAtom(vc_autoPlayVideoAtom)
    const [muted] = useAtom(vc_storedMutedAtom)

    return (
        <>
            <TorrentStreamOverlay
                isNativePlayerComponent="top-section"
                show={!isMiniPlayer && !(!!state.playbackInfo?.streamUrl && !state.loadingState)}
            />

            {(state?.playbackError) && (
                <div
                    data-vc-element="playback-error-container"
                    className="h-full w-full bg-black/100 flex items-center justify-center z-[20] absolute p-4"
                >
                    <div className="text-white text-center" data-vc-element="playback-error-content">
                        {!isMiniPlayer ? (
                            <LuffyError title="Playback Error" imageContainerClass="size-[3.5rem] lg:size-[8rem]" />
                        ) : (
                            <h1 data-vc-element="playback-error-title" className={cn("text-2xl font-bold", isMiniPlayer && "text-lg")}>Playback
                                                                                                                                       Error</h1>
                        )}
                        <p
                            data-vc-element="playback-error-message"
                            className={cn("text-base text-white/50 max-w-xl", isMiniPlayer && "text-sm max-w-lg mx-auto")}
                        >
                            {state.playbackError || "An error occurred while playing the stream. Please try again later."}
                        </p>
                    </div>
                </div>
            )}

            <div
                data-vc-element="container"
                ref={combineContainerRef}
                className={cn(
                    "relative w-full h-full bg-black overflow-clip flex items-center justify-center",
                    (!busy && !isMiniPlayer) && "cursor-none",
                )}
                onPointerMove={handleContainerPointerMove}
            >
                {(!!state.playbackInfo?.streamUrl && !state.loadingState) ? (
                    <>
                        <VideoCoreKeybindingController
                            videoRef={videoRef}
                            active={state.active}
                            chapterCues={chapterCues ?? []}
                            introStartTime={aniSkipData?.op?.interval?.startTime}
                            introEndTime={aniSkipData?.op?.interval?.endTime}
                            endingStartTime={aniSkipData?.ed?.interval?.startTime}
                            endingEndTime={aniSkipData?.ed?.interval?.endTime}
                        />

                        <VideoCoreOverlayDisplay />

                        {buffering && (
                            <div
                                data-vc-element="buffering-indicator"
                                className="absolute inset-0 flex items-center justify-center z-[50] pointer-events-none"
                            >
                                <div className="bg-black/20 backdrop-blur-sm rounded-full p-4">
                                    <PiSpinnerDuotone className="size-12 text-white animate-spin" />
                                </div>
                            </div>
                        )}

                        {busy && (
                            <>
                                {!!skipOpeningTime && !isMiniPlayer && (
                                    <div
                                        data-vc-element="skip-oped-button-container"
                                        data-vc-for="opening"
                                        className="absolute left-5 bottom-28 z-[60] native-player-hide-on-fullscreen"
                                    >
                                        <Button
                                            size="sm"
                                            intent="gray-basic"
                                            onClick={e => {
                                                e.stopPropagation()
                                                action({ type: "seekTo", payload: { time: skipOpeningTime || 0 } })
                                            }}
                                            onPointerMove={e => e.stopPropagation()}
                                        >
                                            Skip Opening
                                        </Button>
                                    </div>
                                )}

                                {!!skipEndingTime && !isMiniPlayer && (
                                    <div
                                        data-vc-element="skip-oped-button-container"
                                        data-vc-for="ending"
                                        className="absolute right-5 bottom-28 z-[60] native-player-hide-on-fullscreen"
                                    >
                                        <Button
                                            size="sm"
                                            intent="gray-basic"
                                            onClick={e => {
                                                e.stopPropagation()
                                                action({ type: "seekTo", payload: { time: skipEndingTime || 0 } })
                                            }}
                                            onPointerMove={e => e.stopPropagation()}
                                        >
                                            Skip Ending
                                        </Button>
                                    </div>
                                )}
                            </>
                        )}

                        <div
                            data-vc-element="inner-container"
                            className="relative w-full h-full flex items-center justify-center"
                            onClick={handleClick}
                            onContextMenu={handleClick}
                        >
                            <video
                                data-vc-element="video"
                                data-video-core-element
                                crossOrigin="anonymous"
                                preload="auto"
                                src={streamUrl && !streamUrl.includes(".m3u8") ? streamUrl : undefined}
                                ref={combineRef}
                                onLoadedMetadata={handleLoadedMetadata}
                                onTimeUpdate={handleTimeUpdate}
                                onEnded={handleEnded}
                                onPlay={handlePlay}
                                onPause={handlePause}
                                onDoubleClick={handleDoubleClick}
                                onLoadedData={handleLoadedData}
                                onVolumeChange={handleVolumeChange}
                                onRateChange={handleRateChange}
                                onError={handleError}
                                onWaiting={handleWaiting}
                                onCanPlay={handleCanPlay}
                                onStalled={handleStalled}
                                autoPlay={autoPlay}
                                muted={muted}
                                playsInline
                                controls={false}
                                style={{
                                    border: "none",
                                    width: "100%",
                                    height: "100%",
                                    objectFit: "contain",
                                    objectPosition: "center",
                                    filter: (settings.videoEnhancement.enabled && beautifyImage)
                                        ? `contrast(${settings.videoEnhancement.contrast}) saturate(${settings.videoEnhancement.saturation}) brightness(${settings.videoEnhancement.brightness})`
                                        : "none",
                                    imageRendering: "crisp-edges",
                                }}
                            >
                                {state.playbackInfo?.mkvMetadata?.subtitleTracks?.map(track => (
                                    <track
                                        id={track.number.toString()}
                                        key={track.number}
                                        kind="subtitles"
                                        srcLang={track.language || "eng"}
                                        label={track.name}
                                    />
                                ))}
                            </video>
                        </div>

                        {!isMobile && <VideoCoreTopSection inline={inline}>
                            <VideoCoreTopPlaybackInfo state={state} />
                            {!inline && (
                                <div
                                    data-vc-element="floating-buttons-container"
                                    className={cn(
                                        "opacity-0",
                                        "transition-opacity duration-200 ease-in-out",
                                        (busy || paused) && "opacity-100",
                                    )}
                                >
                                    <FloatingButtons part="video" onTerminateStream={onTerminateStream} />
                                </div>
                            )}
                        </VideoCoreTopSection>}

                        {isPip && (
                            <div
                                data-vc-element="pip-overlay"
                                className="absolute top-0 left-0 w-full h-full z-[100] bg-black flex items-center justify-center"
                            >
                                <Button
                                    intent="gray-outline"
                                    size="xl"
                                    onClick={() => {
                                        pipManager?.togglePip()
                                    }}
                                >
                                    Exit PiP
                                </Button>
                            </div>
                        )}

                        {!isMobile ? <VideoCoreControlBar
                            timeRange={<VideoCoreTimeRange chapterCues={chapterCues ?? []} />}
                        >
                            <VideoCorePlayButton />
                            <VideoCorePlaylistControl />
                            <VideoCoreVolumeButton />
                            <VideoCoreTimestamp />
                            <div className="flex flex-1" data-vc-element="control-bar-separator" />
                            {!inline && <TorrentStreamOverlay isNativePlayerComponent="control-bar" show={!isMiniPlayer} />}
                            <VideoCoreWatchPartyChat />
                            <VideoCoreSettingsMenu />
                            <VideoCoreResolutionMenu state={state} onVideoSourceChange={onVideoSourceChange} />
                            <VideoCoreSubtitleMenu inline={inline} />
                            <VideoCoreAudioMenu />
                            <VideoCorePipButton />
                            <VideoCoreFullscreenButton />
                        </VideoCoreControlBar> : <VideoCoreMobileControlBar
                            timeRange={<VideoCoreTimeRange chapterCues={chapterCues ?? []} />}
                            topLeftSection={<>
                                <VideoCorePlaylistControl />
                            </>}
                            topRightSection={<>
                                <VideoCoreSettingsMenu />
                                <VideoCoreResolutionMenu state={state} onVideoSourceChange={onVideoSourceChange} />
                                <VideoCoreSubtitleMenu inline={inline} />
                                <VideoCoreAudioMenu />
                                <VideoCorePipButton />
                                <VideoCoreVolumeButton />
                            </>}
                            bottomRightSection={<>
                                <VideoCoreFullscreenButton />
                            </>}
                            bottomLeftSection={<>
                                <VideoCoreTimestamp />
                            </>}
                        />}
                    </>
                ) : (
                    <div
                        data-vc-element="loading-overlay"
                        className="w-full h-full absolute flex justify-center items-center flex-col space-y-4 bg-black rounded-md"
                    >
                        {!inline && <FloatingButtons part="loading" onTerminateStream={onTerminateStream} />}
                        {state.loadingState && (
                            <LoadingSpinner
                                title={state.loadingState || "Loading..."}
                                spinner={<ImSpinner2 className="size-20 text-white animate-spin" />}
                                containerClass="z-[1]"
                            />
                        )}
                        {!isMiniPlayer && !inline && (
                            <div className="opacity-50 absolute inset-0 z-[0] overflow-hidden" data-vc-element="loading-overlay-gradient">
                                <GradientBackground duration={10} breathingRange={5} />
                            </div>
                        )}
                    </div>
                )}
            </div>
        </>
    )
})

PlayerContent.displayName = "PlayerContent"

export interface VideoCoreProps {
    id: string
    state: VideoCoreLifecycleState
    aniSkipData?: {
        op: AniSkipTime | null
        ed: AniSkipTime | null
    } | undefined
    onTerminateStream: () => void
    onEnded?: () => void
    onCompleted?: () => void
    onPlay?: () => void
    onPause?: () => void
    onTimeUpdate?: (e: React.SyntheticEvent<HTMLVideoElement, Event>) => void
    onLoadedData?: (e: React.SyntheticEvent<HTMLVideoElement, Event>) => void
    onLoadedMetadata?: (e: React.SyntheticEvent<HTMLVideoElement, Event>) => void
    onVolumeChange?: () => void
    onFullscreenChange?: (fullscreen: boolean) => void
    onSeeking?: () => void
    onSeeked?: (time: number) => void
    onError?: (error: string) => void
    onPlaybackRateChange?: () => void
    // onFileUploaded: (data: { name: string, content: string }) => void
    onVideoSourceChange?: ((source: VideoCore_VideoSource) => void) | undefined
    onPlayEpisode?: (which: "previous" | "next") => void
    inlineClassName?: string
    onHlsMediaDetached?: () => void
    onHlsFatalError?: (error: ErrorData) => void
    onChangePlaybackType?: (type: VideoCore_VideoPlaybackInfo["streamType"]) => void
    inline?: boolean
    mRef?: React.MutableRefObject<HTMLVideoElement | null>
}

export function VideoCore(props: VideoCoreProps) {
    const {
        state,
        aniSkipData,
        onTerminateStream: _onTerminateStream,
        onEnded,
        onPlay,
        onCompleted,
        onPause,
        onTimeUpdate,
        onLoadedData,
        onLoadedMetadata,
        onVolumeChange,
        onFullscreenChange,
        onSeeking,
        onSeeked,
        onError,
        onPlaybackRateChange,
        // onFileUploaded,
        inline = false,
        inlineClassName,
        onVideoSourceChange,
        onHlsMediaDetached,
        onHlsFatalError,
        onPlayEpisode,
        onChangePlaybackType,
        mRef,
    } = props
    const serverStatus = useServerStatus()
    const [streamType, setStreamType] = useState<VideoCore_VideoPlaybackInfo["streamType"]>(state.playbackInfo?.streamType ?? "unknown")

    const videoRef = useRef<HTMLVideoElement | null>(null)
    const containerRef = useRef<HTMLDivElement | null>(null)

    const {
        dispatchTerminatedEvent,
        dispatchVideoLoadedEvent,
        dispatchVideoCompletedEvent,
        dispatchVideoErrorEvent,
        dispatchCanPlayEvent,
        dispatchTranslateTextEvent,
        dispatchTranslateSubtitleTrackEvent,
    } = useVideoCoreSetupEvents(props.id, state, videoRef, onTerminateStream)

    const { width: windowWidth } = useWindowSize()
    const [isMobilePlayer, setIsMobilePlayer] = useAtom(vc_isMobile)
    React.useEffect(() => {
        setIsMobilePlayer(windowWidth < 1024)
    }, [windowWidth < 1024])

    const setVideoElement = useSetAtom(vc_videoElement)
    const setRealVideoSize = useSetAtom(vc_realVideoSize)
    useVideoCoreBindings(videoRef, state.playbackInfo)
    useVideoCorePlaylistSetup(state, onPlayEpisode)

    const { isParticipant: isWatchPartyParticipant } = useNakamaWatchParty()

    const videoCompletedRef = useRef(false)
    const currentPlaybackRef = useRef<string | null>(null)

    const [, setContainerElement] = useAtom(vc_containerElement)

    const [subtitleManager, setSubtitleManager] = useAtom(vc_subtitleManager)
    const [mediaCaptionsManager, setMediaCaptionsManager] = useAtom(vc_mediaCaptionsManager)
    const [audioManager, setAudioManager] = useAtom(vc_audioManager)
    const [previewManager, setPreviewManager] = useAtom(vc_previewManager)
    const [anime4kManager, setAnime4kManager] = useAtom(vc_anime4kManager)
    const [pipManager, setPipManager] = useAtom(vc_pipManager)
    const setPipElement = useSetAtom(vc_pipElement)
    const [fullscreenManager, setFullscreenManager] = useAtom(vc_fullscreenManager)
    const setIsFullscreen = useSetAtom(vc_isFullscreen)
    const [mediaSessionManager, setMediaSessionManager] = useAtom(vc_mediaSessionManager)
    const action = useSetAtom(vc_dispatchAction)

    // States
    const qc = useQueryClient()
    const settings = useAtomValue(vc_settings)
    const [isMiniPlayer, setIsMiniPlayer] = useAtom(vc_miniPlayer)
    const [busy, setBusy] = useAtom(vc_busy)
    const [buffering, setBuffering] = useAtom(vc_buffering)
    const duration = useAtomValue(vc_duration)
    const fullscreen = useAtomValue(vc_isFullscreen)
    const showOverlayFeedback = useSetAtom(vc_showOverlayFeedback)
    const cursorBusy = useAtomValue(vc_cursorBusy)

    const [skipOpeningTime, setSkipOpeningTime] = useAtom(vc_skipOpeningTime)
    const [skipEndingTime, setSkipEndingTime] = useAtom(vc_skipEndingTime)

    const [autoNext] = useAtom(vc_autoNextAtom)
    const [autoPlay] = useAtom(vc_autoPlayVideoAtom)
    const [autoSkipOpeningOutro] = useAtom(vc_autoSkipOPEDAtom)
    const [volume] = useAtom(vc_storedVolumeAtom)
    const [muted] = useAtom(vc_storedMutedAtom)
    const [playbackRate, setPlaybackRate] = useAtom(vc_storedPlaybackRateAtom)

    const { mutate: cancelDiscordActivity } = useCancelDiscordActivity()

    const { mutate: convertSubs } = useDirectstreamConvertSubs()

    const isFirstError = React.useRef(true)
    const shouldDispatchTerminatedOnUnmount = React.useRef(false)
    const [activePlayer, setActivePlayer] = useAtom(vc_activePlayerId)

    React.useEffect(() => {
        setIsMiniPlayer(false)
    }, [])

    // Track if this player should dispatch terminated event on unmount
    React.useEffect(() => {
        shouldDispatchTerminatedOnUnmount.current = (activePlayer === props.id && !!state.playbackInfo)
    }, [activePlayer, props.id, state.playbackInfo?.id])

    // Call dispatchTerminatedEvent on unmount if this was the active player
    useUnmount(() => {
        if (shouldDispatchTerminatedOnUnmount.current) {
            dispatchTerminatedEvent()
            setActivePlayer(null)
        }
    })

    function onTerminateStream() {
        _onTerminateStream?.()
        dispatchTerminatedEvent()
    }

    // Measure video element size
    const [measureRef, { width, height }] = useMeasure<HTMLVideoElement>()
    React.useEffect(() => {
        setRealVideoSize({
            width,
            height,
        })
    }, [width, height])

    // refetch continuity data when playback info changes
    React.useEffect(() => {
        qc.invalidateQueries({ queryKey: [API_ENDPOINTS.CONTINUITY.GetContinuityWatchHistory.key] })
        qc.invalidateQueries({ queryKey: [API_ENDPOINTS.CONTINUITY.GetContinuityWatchHistoryItem.key] })

    }, [state.playbackInfo?.id])


    // Re-focus the video element when playback info changes
    React.useEffect(() => {
        if (state.active && videoRef.current && !!state.playbackInfo) {
            // Small delay to ensure the video element is fully rendered
            setTimeout(() => {
                videoRef.current?.focus()
            }, 100)
        }
    }, [state.active, state.playbackInfo?.id])

    // Merge refs
    const combineRef = (instance: HTMLVideoElement | null) => {
        videoRef.current = instance
        if (mRef) {
            mRef.current = instance
        }
        if (instance) measureRef(instance)
        setVideoElement(instance)
    }

    const combineContainerRef = (instance: HTMLDivElement | null) => {
        containerRef.current = instance
        setContainerElement(instance)
    }

    // actions
    function togglePlay() {
        if (videoRef?.current?.paused) {
            videoRef?.current?.play().catch()
            onPlay?.()
            showOverlayFeedback({ message: "PLAY", type: "icon" })
        } else {
            videoRef?.current?.pause()
            onPause?.()
            showOverlayFeedback({ message: "PAUSE", type: "icon" })
        }
    }

    function onAudioChange() {
        log.info("Audio changed", videoRef.current?.audioTracks)
        if (videoRef.current?.audioTracks) {
            for (let i = 0; i < videoRef.current.audioTracks.length; i++) {
                const track = videoRef.current.audioTracks[i]
                if (track.enabled) {
                    audioManager?.selectTrack(Number(track.id))
                    break
                }
            }
        }
        action({ type: "seek", payload: { time: -1 } })
    }

    // Continuity
    const {
        watchHistory,
        waitForWatchHistory,
        shouldWaitForWatchHistory,
        getEpisodeContinuitySeekTo,
    } = useHandleCurrentMediaContinuity(state?.playbackInfo?.media?.id)

    React.useEffect(() => {
        if (watchHistory) {
            log.info("Continuity watch history", watchHistory)
        }
    }, [watchHistory])

    const hasSoughtRef = React.useRef(false)

    // Lifecycle
    useUpdateEffect(() => {
        // Wait for watch history to be ready before starting playback (only if continuity is enabled)
        if (waitForWatchHistory && shouldWaitForWatchHistory && !state.playbackInfo?.disableRestoreFromContinuity) return

        // If the playback info is null, the stream is loading or unmounted
        if (!state.playbackInfo) {
            log.info("Cleaning up")
            dispatchTerminatedEvent()
            cancelDiscordActivity()
            hasSoughtRef.current = false
            isFirstError.current = true
            if (videoRef.current) {
                videoRef.current.pause()
                videoRef.current.removeAttribute("src")
                videoRef.current.load()
                videoRef.current = null
            }
            setVideoElement(null)
            subtitleManager?.destroy?.()
            setSubtitleManager(null)
            mediaCaptionsManager?.destroy?.()
            setMediaCaptionsManager(null)
            previewManager?.cleanup?.()
            setPreviewManager(null)
            setAudioManager(null)
            anime4kManager?.destroy?.()
            setAnime4kManager(null)
            pipManager?.destroy?.()
            setPipManager(null)
            setPipElement(null)
            fullscreenManager?.destroy?.()
            setFullscreenManager(null)
            // setIsFullscreen(false)
            if (mediaSessionManager) {
                mediaSessionManager.setVideo(null)
                mediaSessionManager.destroy()
            }
            setMediaSessionManager(null)
            currentPlaybackRef.current = null
            videoRef.current = null
        }

        // When a new playback info is received
        if (!!state.playbackInfo?.id && (!currentPlaybackRef.current || state.playbackInfo.id !== currentPlaybackRef.current)) {
            hasSoughtRef.current = false
            isFirstError.current = true
            log.info("New stream loaded", state.playbackInfo)
            setStreamType(state.playbackInfo.streamType)
            vc_logGeneralInfo(videoRef.current)
            dispatchVideoLoadedEvent()
        }
    }, [state.playbackInfo?.id, waitForWatchHistory, shouldWaitForWatchHistory])

    // Override active player, won't apply to native-player
    React.useEffect(() => {
        if (state.playbackInfo?.id && activePlayer === props.id) {
            setActivePlayer(props.id)
        }
    }, [state.playbackInfo?.id, activePlayer])

    const streamUrl = state?.playbackInfo?.streamUrl?.replace?.("{{SERVER_URL}}", getServerBaseUrl())

    // Initialize HLS
    useVideoCoreHls({
        videoElement: videoRef.current,
        streamUrl: streamUrl,
        streamType: streamType,
        onMediaDetached: onHlsMediaDetached,
        onFatalError: onHlsFatalError,
    })

    const [anime4kOption, setAnime4kOption] = useAtom(vc_anime4kOption)

    // Get HLS audio track values
    const hlsAudioTracks = useAtomValue(vc_hlsAudioTracks)
    const hlsCurrentAudioTrack = useAtomValue(vc_hlsCurrentAudioTrack)
    const hlsSetAudioTrack = useAtomValue(vc_hlsSetAudioTrack)

    // events
    const handleLoadedMetadata = (e: React.SyntheticEvent<HTMLVideoElement>) => {
        onLoadedMetadata?.(e)
        if (!videoRef.current) return
        const v = videoRef.current

        log.info("Loaded metadata", v.duration)
        log.info("Audio tracks", v.audioTracks)
        log.info("Text tracks", v.textTracks)

        setSkipOpeningTime(null)
        setSkipEndingTime(null)

        // onCaptionsChange() not needed?
        onAudioChange()

        videoCompletedRef.current = false

        if (!state.playbackInfo) return // shouldn't happen

        // setHasUpdatedProgress(false)

        currentPlaybackRef.current = state.playbackInfo.id

        /*
         * Event or file tracks using libass renderer
         */
        // const hasLibassRendererTracks = state.playbackInfo?.subtitleTracks?.some(t => t.useLibassRenderer)
        // // Initialize the subtitle manager if the stream is MKV or has useLibassRenderer tracks
        // if (!!state.playbackInfo?.mkvMetadata || hasLibassRendererTracks) {
        //
        // }

        /*
         * File subtitle tracks that don't use libass renderer
         */
        const nonLibassSubtitleTracks = state.playbackInfo?.subtitleTracks?.filter(t => !t.useLibassRenderer)
        if (nonLibassSubtitleTracks && nonLibassSubtitleTracks.length > 0) {
            setMediaCaptionsManager(p => {
                if (p) p.destroy()
                return new MediaCaptionsManager({
                    videoElement: v!,
                    tracks: nonLibassSubtitleTracks,
                    translateTargetLang: serverStatus?.settings?.mediaPlayer?.vcTranslate
                        ? serverStatus?.settings?.mediaPlayer?.vcTranslateTargetLanguage
                        : null,
                    settings: settings,
                    fetchAndConvertToVTT: (url?: string, content?: string) => {
                        return new Promise((resolve, reject) => {
                            convertSubs({ url: url ?? "", content: content ?? "", to: "vtt" }, {
                                onSuccess: (data) => resolve(data),
                                onError: (error) => reject(error),
                            })
                        })
                    },
                    sendTranslateRequest: (text?: string, track?: VideoCore_VideoSubtitleTrack) => {
                        if (text) {
                            dispatchTranslateTextEvent(text)
                        }
                        if (track) {
                            dispatchTranslateSubtitleTrackEvent(track)
                        }
                    },
                })
            })
        } else {
            setSubtitleManager(p => {
                if (p) p.destroy()
                return new VideoCoreSubtitleManager({
                    videoElement: v!,
                    playbackInfo: state.playbackInfo!,
                    jassubOffscreenRender: true,
                    translateTargetLang: serverStatus?.settings?.mediaPlayer?.vcTranslate
                        ? serverStatus?.settings?.mediaPlayer?.vcTranslateTargetLanguage
                        : null,
                    settings: settings,
                    fetchAndConvertToASS: (url?: string, content?: string) => {
                        return new Promise((resolve, reject) => {
                            convertSubs({ url: url ?? "", content: content ?? "", to: "ass" }, {
                                onSuccess: (data) => resolve(data),
                                onError: (error) => reject(error),
                            })
                        })
                    },
                    sendTranslateRequest: (text?: string, track?: VideoCore_VideoSubtitleTrack) => {
                        if (text) {
                            dispatchTranslateTextEvent(text)
                        }
                        if (track) {
                            dispatchTranslateSubtitleTrackEvent(track)
                        }
                    },
                })
            })
        }

        // Initialize audio manager for HLS streams
        if (hlsAudioTracks.length > 0 && hlsSetAudioTrack) {
            setAudioManager(new VideoCoreAudioManager({
                videoElement: v!,
                playbackInfo: state.playbackInfo,
                settings: settings,
                onError: (error) => {
                    log.error("Audio manager error", error)
                    onError?.(error)
                },
                hlsSetAudioTrack: hlsSetAudioTrack,
                hlsAudioTracks: hlsAudioTracks,
                hlsCurrentAudioTrack: hlsCurrentAudioTrack,
            }))
        } else if (!!state.playbackInfo?.mkvMetadata) {
            setAudioManager(new VideoCoreAudioManager({
                videoElement: v!,
                playbackInfo: state.playbackInfo,
                settings: settings,
                onError: (error) => {
                    log.error("Audio manager error", error)
                    onError?.(error)
                },
            }))
        }

        // Initialize Anime4K manager
        setAnime4kManager(p => {
            if (p) p.destroy()
            return new VideoCoreAnime4KManager({
                videoElement: v!,
                settings: settings,
                onFallback: (message) => {
                    showOverlayFeedback({ message, duration: 2000 })
                },
                onOptionChanged: (opt) => {
                    setAnime4kOption(opt)
                },
            })
        })

        // Initialize PIP manager
        setPipManager(p => {
            if (p) p.destroy()
            const manager = new VideoCorePipManager((element) => {
                setPipElement(element)
            })
            manager.setVideo(v!, state.playbackInfo!)
            return manager
        })

        // Initialize fullscreen manager
        setFullscreenManager(p => {
            if (p) p.destroy()
            return new VideoCoreFullscreenManager((isFullscreen: boolean) => {
                setIsFullscreen(isFullscreen)
                onFullscreenChange?.(isFullscreen)
            })
        })

        // Initialize media session manager
        setMediaSessionManager(p => {
            if (p) p.destroy()
            const manager = new VideoCoreMediaSessionManager()

            manager.on("media-play-request", () => {
                if (videoRef.current?.paused) {
                    videoRef.current.play().catch(err => {
                        log.error("Failed to play video from media session", err)
                    })
                }
            })

            manager.on("media-pause-request", () => {
                if (videoRef.current && !videoRef.current.paused) {
                    videoRef.current.pause()
                }
            })

            manager.on("media-seek-request", (event: Event) => {
                const { seekTime } = (event as CustomEvent).detail
                action({ type: "seekTo", payload: { time: seekTime } })
            })
            manager.setPlaybackInfo(state.playbackInfo!)
            manager.setVideo(v!)
            return manager
        })
    }

    function setupPreviewManager() {
        if (previewManager !== null) {
            previewManager.cleanup()
            setPreviewManager(null)
        }
        React.startTransition(() => {
            if (videoRef.current && streamUrl) {
                log.info("Initializing preview manager")
                setPreviewManager(p => {
                    if (p) p.cleanup()
                    return new VideoCorePreviewManager(videoRef.current!, streamUrl!, streamType, state.playbackInfo?.playbackType !== "onlinestream")
                })
            }
        })
    }

    // Setup preview manager when stream type changes
    React.useEffect(() => {
        if (currentPlaybackRef.current) {
            setupPreviewManager()
        }
    }, [streamType, currentPlaybackRef.current])

    const handleTimeUpdate = (e: React.SyntheticEvent<HTMLVideoElement>) => {
        onTimeUpdate?.(e)
        if (!videoRef.current) return
        const v = videoRef.current

        // Video completed event
        const percent = v.currentTime / v.duration
        if (!!v.duration && !videoCompletedRef.current && percent >= 0.8) {
            videoCompletedRef.current = true
            onCompleted?.()
            dispatchVideoCompletedEvent()
        }
    }

    const { playEpisode } = useVideoCorePlaylist()
    const handleEnded = (e: React.SyntheticEvent<HTMLVideoElement>) => {
        log.info("Video ended")
        subtitleManager?.pgsRenderer?.stop()
        onEnded?.()
        if (autoNext && !isWatchPartyParticipant) {
            // videoRef?.current?.pause()
            playEpisode("next")
        }
    }

    const [debouncedMenuOpen, setDebouncedMenuOpen] = React.useState(false)
    const menuOpen = useAtomValue(vc_menuOpen)
    React.useEffect(() => {
        if (!!menuOpen) {
            setDebouncedMenuOpen(true)
            return
        }
        let t = setTimeout(() => {
            setDebouncedMenuOpen(false)
        }, 800)
        return () => {
            clearTimeout(t)
        }
    }, [menuOpen])

    let lastClickTime = React.useRef(0)

    const handleClick = (e: React.SyntheticEvent<HTMLDivElement>) => {
        // log.info("Video clicked")
        // check if right click

        if (inline) {
            if (e.type === "click") {
                const now = Date.now()
                if (!debouncedMenuOpen) {
                    togglePlay()
                }
                if (lastClickTime.current && now - lastClickTime.current < 300) {
                    fullscreenManager?.toggleFullscreen()
                } else {
                    setTimeout(() => {
                        setBusy(false)
                    }, 100)
                }
                lastClickTime.current = now
            }

            if (e.type === "contextmenu") {
                e.preventDefault()
            }
            return
        }

        if (e.type === "click") {
            if (!debouncedMenuOpen) {
                togglePlay()
            }
            setTimeout(() => {
                setBusy(false)
            }, 100)
        }

        if (e.type === "contextmenu") {
            const now = Date.now()
            if (lastClickTime.current && now - lastClickTime.current < 500) {
                fullscreenManager?.toggleFullscreen()
            }
            lastClickTime.current = now
        }
    }

    const handleDoubleClick = (e: React.SyntheticEvent<HTMLVideoElement>) => {
        // fullscreenManager?.toggleFullscreen()
    }

    const handlePlay = (e: React.SyntheticEvent<HTMLVideoElement>) => {
        // log.info("Video resumed")
        onPlay?.()
    }

    const handlePause = (e: React.SyntheticEvent<HTMLVideoElement>) => {
        // log.info("Video paused")
        subtitleManager?.pgsRenderer?.stop()
        onPause?.()
    }

    const handleLoadedData = (e: React.SyntheticEvent<HTMLVideoElement>) => {
        log.info("Loaded data")
        onLoadedData?.(e)
    }

    const handleVolumeChange = (e: React.SyntheticEvent<HTMLVideoElement>) => {
        const v = e.currentTarget
        // log.info("Volume changed", v.volume)
        onVolumeChange?.()
    }

    const handleRateChange = (e: React.SyntheticEvent<HTMLVideoElement>) => {
        const v = e.currentTarget
        log.info("Playback rate changed", v.playbackRate)
        onPlaybackRateChange?.()

        if (v.playbackRate !== playbackRate) {
            setPlaybackRate(v.playbackRate)
        }
    }

    const handleError = (e: React.SyntheticEvent<HTMLVideoElement>) => {
        log.error("Video error", e)
        if (isFirstError.current && props.id !== "native-player") {
            // Change stream type to HLS if it failed to load
            log.warning("Video player could not load the URL, switching to HLS")
            setStreamType("hls")
            isFirstError.current = false
            return
        }

        const error = `Video playback error occurred. (Code: ${(e.currentTarget.error && e.currentTarget.error.code) || "unknown"})`
        onError?.(error)
        dispatchVideoErrorEvent(error)
    }

    const handleWaiting = (e: React.SyntheticEvent<HTMLVideoElement>) => {
        setBuffering(true)
    }

    function restoreSeekTime(time: number, showMessage: boolean, paused?: boolean) {
        if (!videoRef.current) return
        if (anime4kOption === "off" || anime4kManager?.canvas !== null) {
            if (showMessage) showOverlayFeedback({ message: "Progress restored", duration: 1500 })
            videoRef.current.currentTime = time
            if (paused && !videoRef.current.paused) {
                videoRef.current.pause()
            } else if (paused === false && videoRef.current.paused) {
                videoRef.current.play().catch()
            }
        } else if (anime4kOption !== ("off" as Anime4KOption)) {
            videoRef.current.pause()
            if (showMessage) showOverlayFeedback({ message: "Restoring progress", duration: 1500 })
            anime4kManager.registerOnCanvasCreatedOnce(() => {
                if (!videoRef.current) return
                videoRef.current.currentTime = time
                if (paused && !videoRef.current.paused) {
                    videoRef.current.pause()
                } else if (paused === false && videoRef.current.paused) {
                    videoRef.current.play().catch()
                }
            })
        }
    }

    const handleCanPlay = (e: React.SyntheticEvent<HTMLVideoElement>) => {
        setBuffering(false)

        if (!hasSoughtRef.current) {
            if (!state.playbackInfo || !videoRef.current) return
            hasSoughtRef.current = true
            // if (autoPlay) {
            //     videoRef.current.play().catch()
            // }

            // Do nothing if the stream is not seekable
            if (isWatchPartyParticipant) return

            dispatchCanPlayEvent()

            // Restore previous position if available
            if (!state.playbackInfo.disableRestoreFromContinuity && !state.playbackInfo.initialState) {
                if (state.playbackInfo?.episode?.progressNumber && watchHistory?.found && watchHistory.item?.episodeNumber === state.playbackInfo?.episode?.progressNumber) {
                    const lastWatchedTime = getEpisodeContinuitySeekTo(state.playbackInfo?.episode?.progressNumber,
                        videoRef.current?.currentTime,
                        videoRef.current?.duration)
                    log.info("Watch continuity: Fetched last watched time", { lastWatchedTime })
                    if (lastWatchedTime > 0) {
                        log.info("Watch continuity: Seeking to", lastWatchedTime)
                        restoreSeekTime(lastWatchedTime, true)
                    }
                }
            }

            if (state.playbackInfo.initialState) {
                log.info("Setting initial stream state", state.playbackInfo.initialState)
                if (state.playbackInfo.initialState.currentTime) {
                    // action({ type: "seekTo", payload: { time: state.playbackInfo.initialState.currentTime } })
                    restoreSeekTime(state.playbackInfo.initialState.currentTime, false, state.playbackInfo.initialState.paused)
                }
            } else if (autoPlay) {
                videoRef.current.play().catch()
            }
        }
    }

    const handleStalled = (e: React.SyntheticEvent<HTMLVideoElement>) => {
        setBuffering(true)
    }

    // external state

    // Handle volume changes
    React.useEffect(() => {
        if (videoRef.current && volume !== undefined && volume !== videoRef.current.volume) {
            videoRef.current.volume = volume
        }
    }, [volume, videoRef.current])

    // Handle mute changes
    React.useEffect(() => {
        if (videoRef.current && muted !== undefined && muted !== videoRef.current.muted) {
            videoRef.current.muted = muted
        }
    }, [muted, videoRef.current])

    // Handle playback rate changes
    React.useEffect(() => {
        if (videoRef.current && playbackRate !== undefined && playbackRate !== videoRef.current.playbackRate) {
            videoRef.current.playbackRate = playbackRate
        }
    }, [playbackRate, videoRef.current])

    // Update PIP manager
    React.useEffect(() => {
        if (pipManager && videoRef.current && state.playbackInfo) {
            pipManager.setVideo(videoRef.current, state.playbackInfo)
            if (subtitleManager) pipManager.setSubtitleManager(subtitleManager)
            if (mediaCaptionsManager) pipManager.setMediaCaptionsManager(mediaCaptionsManager)
        }
    }, [pipManager, subtitleManager, mediaCaptionsManager, videoRef.current, state.playbackInfo])

    // Update fullscreen manager
    React.useEffect(() => {
        if (fullscreenManager && containerRef.current) {
            fullscreenManager.setContainer(containerRef.current)
        }
        if (fullscreenManager && videoRef.current) {
            fullscreenManager.setVideoElement(videoRef.current)
        }
    }, [fullscreenManager, containerRef.current, videoRef.current])

    React.useEffect(() => {
        if (mediaSessionManager && videoRef.current && state.playbackInfo && state.active) {
            mediaSessionManager.setVideo(videoRef.current)
            mediaSessionManager.setPlaybackInfo(state.playbackInfo)
            mediaSessionManager.activate()
        } else if (mediaSessionManager && !state.active) {
            log.info("Video inactive, deactivating media session")
            mediaSessionManager.deactivate()
        }
    }, [mediaSessionManager, videoRef.current, state.playbackInfo, state.active, playEpisode])

    // Handle iOS fullscreen subtitles
    useVideoCoreIOSFullscreenSubtitles({
        videoElement: videoRef.current,
    })

    // Handle mobile gestures
    useVideoCoreMobileGestures({
        videoElement: videoRef.current,
        containerElement: containerRef.current,
        onSeek: (time) => {
            if (videoRef.current) {
                videoRef.current.currentTime = time
            }
        },
    })

    // container events
    const setNotBusyTimeout = React.useRef<NodeJS.Timeout | null>(null)
    const lastPointerPosition = React.useRef({ x: 0, y: 0 })
    const handleContainerPointerMove = (e: React.PointerEvent<HTMLDivElement>) => {
        const { x, y } = e.nativeEvent
        const dx = x - lastPointerPosition.current.x
        const dy = y - lastPointerPosition.current.y
        if (Math.abs(dx) < 15 && Math.abs(dy) < 15) return
        if (setNotBusyTimeout?.current) {
            clearTimeout(setNotBusyTimeout.current)
        }
        setBusy(true)
        setNotBusyTimeout.current = setTimeout(() => {
            if (!cursorBusy) {
                setBusy(false)
            }
        }, DELAY_BEFORE_NOT_BUSY)
        lastPointerPosition.current = { x, y }
    }

    const chapterCues = useMemo(() => {
            if (!duration || duration <= 1) return []
            // If we have MKV chapters, use them
            if (state.playbackInfo?.mkvMetadata?.chapters?.length) {
                const cues = vc_createChapterCues(state.playbackInfo.mkvMetadata.chapters, duration)
                log.info("Chapter cues from MKV", cues)
                return cues
            }

            // Otherwise, create chapters from AniSkip data if available
            if (!!aniSkipData?.op?.interval && duration > 0) {
                log.info("Creating chapter cues from AniSkip data", aniSkipData)
                const chapters = vc_createChaptersFromAniSkip(aniSkipData, duration, state?.playbackInfo?.media?.format)
                const cues = vc_createChapterCues(chapters, duration)
                log.info("Chapter cues from AniSkip", cues)
                return cues
            }

            return []
        },
        [
            state.playbackInfo?.mkvMetadata?.chapters,
            aniSkipData?.op?.interval,
            aniSkipData?.ed?.interval,
            duration,
            state?.playbackInfo?.media?.format,
        ])

    // Inline mode
    if (inline) {
        return (
            <ScopeProvider atoms={[__torrentSearch_selectionAtom, __torrentSearch_selectionEpisodeAtom, __torrentSearch_selectedTorrentsAtom]}>
                <VideoCoreAnime4K />
                <VideoCorePreferencesModal />
                {fullscreen && <RemoveScrollBar />}
                <div
                    data-vc-element="inline-container"
                    className={cn(
                        "relative w-full h-full",
                        inlineClassName,
                        fullscreen && "fixed z-[99999] inset-0",
                    )}
                >
                    <PlayerContent
                        inline={inline}
                        state={state}
                        videoRef={videoRef}
                        chapterCues={chapterCues}
                        aniSkipData={aniSkipData}
                        streamUrl={streamUrl}
                        combineRef={combineRef}
                        combineContainerRef={combineContainerRef}
                        handleContainerPointerMove={handleContainerPointerMove}
                        handleClick={handleClick}
                        handleLoadedMetadata={handleLoadedMetadata}
                        handleTimeUpdate={handleTimeUpdate}
                        handleEnded={handleEnded}
                        handlePlay={handlePlay}
                        handlePause={handlePause}
                        handleDoubleClick={handleDoubleClick}
                        handleLoadedData={handleLoadedData}
                        handleVolumeChange={handleVolumeChange}
                        handleRateChange={handleRateChange}
                        handleError={handleError}
                        handleWaiting={handleWaiting}
                        handleCanPlay={handleCanPlay}
                        handleStalled={handleStalled}
                        onTerminateStream={onTerminateStream}
                        onVideoSourceChange={onVideoSourceChange}
                    />
                </div>
            </ScopeProvider>
        )
    }

    // Drawer mode
    return (
        <>
            <ScopeProvider atoms={[__torrentSearch_selectionAtom, __torrentSearch_selectionEpisodeAtom, __torrentSearch_selectedTorrentsAtom]}>

                <VideoCoreAnime4K />
                <VideoCorePreferencesModal />
                {state.active && !isMiniPlayer && <RemoveScrollBar />}

                <TorrentStreamOverlay isNativePlayerComponent="overlay" show={(state.active && isMiniPlayer)} />

                <VideoCoreDrawer
                    open={state.active}
                    onOpenChange={(v) => {
                        if (!v) {
                            if (!isMiniPlayer) {
                                setIsMiniPlayer(true)
                                fullscreenManager?.exitFullscreen()
                            } else {
                                onTerminateStream()
                            }
                        }
                    }}
                    borderToBorder
                    miniPlayer={isMiniPlayer}
                    size={isMiniPlayer ? "md" : "full"}
                    side={isMiniPlayer ? "right" : "bottom"}
                    contentClass={cn(
                        "p-0 m-0",
                        !isMiniPlayer && "h-full",
                    )}
                    allowOutsideInteraction={true}
                    overlayClass={cn(
                        isMiniPlayer && "hidden",
                    )}
                    hideCloseButton
                    closeClass={cn(
                        "z-[99]",
                        __isDesktop__ && !isMiniPlayer && "top-8",
                        isMiniPlayer && "left-4",
                    )}
                    data-native-player-drawer
                    onMiniPlayerClick={() => {
                        togglePlay()
                    }}
                >
                    <PlayerContent
                        inline={inline}
                        state={state}
                        videoRef={videoRef}
                        chapterCues={chapterCues}
                        aniSkipData={aniSkipData}
                        streamUrl={streamUrl}
                        combineRef={combineRef}
                        combineContainerRef={combineContainerRef}
                        handleContainerPointerMove={handleContainerPointerMove}
                        handleClick={handleClick}
                        handleLoadedMetadata={handleLoadedMetadata}
                        handleTimeUpdate={handleTimeUpdate}
                        handleEnded={handleEnded}
                        handlePlay={handlePlay}
                        handlePause={handlePause}
                        handleDoubleClick={handleDoubleClick}
                        handleLoadedData={handleLoadedData}
                        handleVolumeChange={handleVolumeChange}
                        handleRateChange={handleRateChange}
                        handleError={handleError}
                        handleWaiting={handleWaiting}
                        handleCanPlay={handleCanPlay}
                        handleStalled={handleStalled}
                        onTerminateStream={onTerminateStream}
                        onVideoSourceChange={onVideoSourceChange}
                    />
                </VideoCoreDrawer>

            </ScopeProvider>
        </>
    )
}

function FloatingButtons(props: { part: "video" | "loading", onTerminateStream: () => void }) {
    const { part, onTerminateStream } = props
    const fullscreen = useAtomValue(vc_isFullscreen)
    const [isMiniPlayer, setIsMiniPlayer] = useAtom(vc_miniPlayer)
    if (fullscreen) return null
    const Content = () => (
        <>
            {!isMiniPlayer && <>
                <IconButton
                    data-vc-element="floating-button-miniplayer"
                    data-vc-for={part}
                    icon={<FiMinimize2 className="text-2xl" />}
                    intent="gray-basic"
                    className="rounded-full absolute top-0 flex-none right-4 z-[999]"
                    onClick={() => {
                        setIsMiniPlayer(true)
                    }}
                />
            </>}

            {isMiniPlayer && <>
                <IconButton
                    data-vc-element="floating-button-expand"
                    data-vc-for={part}
                    type="button"
                    intent="gray"
                    size="sm"
                    className={cn(
                        "rounded-full text-2xl flex-none absolute z-[999] right-4 top-4 pointer-events-auto bg-black/30 hover:bg-black/40",
                        isMiniPlayer && "text-xl",
                    )}
                    icon={<BiExpand />}
                    onClick={() => {
                        setIsMiniPlayer(false)
                    }}
                />
                <IconButton
                    data-vc-element="floating-button-terminate"
                    data-vc-for={part}
                    type="button"
                    intent="alert-subtle"
                    size="sm"
                    className={cn(
                        "rounded-full text-2xl flex-none absolute z-[999] left-4 top-4 pointer-events-auto",
                        isMiniPlayer && "text-xl",
                    )}
                    icon={<BiX />}
                    onClick={() => {
                        onTerminateStream()
                    }}
                />
            </>}
        </>
    )

    if (part === "loading") {
        return (
            <div
                data-vc-element="loading-floating-buttons-container"
                data-vc-for={part}
                className={cn(
                    "absolute top-8 w-full z-[100]",
                    isMiniPlayer && "top-0",
                )}
            >
                <Content />
            </div>
        )
    }

    return <Content />
}
