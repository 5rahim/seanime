import { useUpdateAnimeEntryProgress } from "@/api/hooks/anime_entries.hooks"
import { useHandleContinuityWithMediaPlayer, useHandleCurrentMediaContinuity } from "@/api/hooks/continuity.hooks"
import {
    useCancelDiscordActivity,
    useSetDiscordAnimeActivityWithProgress,
    useUpdateDiscordAnimeActivityWithProgress,
} from "@/api/hooks/discord.hooks"

import { useSeaCommandInject } from "@/app/(main)/_features/sea-command/use-inject"
import { useSkipData } from "@/app/(main)/_features/sea-media-player/aniskip"
import { useFullscreenHandler } from "@/app/(main)/_features/sea-media-player/macos-tauri-fullscreen"
import { SeaMediaPlayerPlaybackSubmenu } from "@/app/(main)/_features/sea-media-player/sea-media-player-components"
import {
    __seaMediaPlayer_scopedCurrentProgressAtom,
    __seaMediaPlayer_scopedProgressItemAtom,
    useSeaMediaPlayer,
} from "@/app/(main)/_features/sea-media-player/sea-media-player-provider"
import {
    __seaMediaPlayer_autoNextAtom,
    __seaMediaPlayer_autoPlayAtom,
    __seaMediaPlayer_autoSkipIntroOutroAtom,
    __seaMediaPlayer_discreteControlsAtom,
    __seaMediaPlayer_mutedAtom,
    __seaMediaPlayer_volumeAtom,
} from "@/app/(main)/_features/sea-media-player/sea-media-player.atoms"
import { useServerStatus } from "@/app/(main)/_hooks/use-server-status"
import { LuffyError } from "@/components/shared/luffy-error"
import { vidstackLayoutIcons } from "@/components/shared/vidstack"
import { Button } from "@/components/ui/button"
import { cn } from "@/components/ui/core/styling"
import { LoadingSpinner } from "@/components/ui/loading-spinner"
import { Skeleton } from "@/components/ui/skeleton"
import { logger } from "@/lib/helpers/debug"
import { __isDesktop__ } from "@/types/constants"
import {
    MediaCanPlayDetail,
    MediaCanPlayEvent,
    MediaDurationChangeEvent,
    MediaEndedEvent,
    MediaFullscreenChangeEvent,
    MediaPlayer,
    MediaPlayerInstance,
    MediaProvider,
    MediaProviderAdapter,
    MediaProviderChangeEvent,
    MediaProviderSetupEvent,
    MediaTimeUpdateEvent,
    MediaTimeUpdateEventDetail,
    Track,
    type TrackProps,
} from "@vidstack/react"
import { DefaultVideoLayout, DefaultVideoLayoutProps } from "@vidstack/react/player/layouts/default"
import { useAtomValue } from "jotai"
import { useAtom } from "jotai/react"
import capitalize from "lodash/capitalize"
import mousetrap from "mousetrap"
import Image from "next/image"
import React from "react"

export type SeaMediaPlayerProps = {
    url?: string | { src: string, type: string }
    poster?: string
    isLoading?: boolean
    isPlaybackError?: string
    playerRef: React.RefObject<MediaPlayerInstance>
    onProviderChange?: (provider: MediaProviderAdapter | null, e: MediaProviderChangeEvent) => void
    onProviderSetup?: (provider: MediaProviderAdapter, e: MediaProviderSetupEvent) => void
    onTimeUpdate?: (detail: MediaTimeUpdateEventDetail, e: MediaTimeUpdateEvent) => void
    onCanPlay?: (detail: MediaCanPlayDetail, e: MediaCanPlayEvent) => void
    onEnded?: (e: MediaEndedEvent) => void
    onDurationChange?: (detail: number, e: MediaDurationChangeEvent) => void
    tracks?: TrackProps[]
    chapters?: ChapterProps[]
    videoLayoutSlots?: Omit<DefaultVideoLayoutProps["slots"], "settingsMenuEndItems">
    settingsItems?: React.ReactElement
    loadingText?: React.ReactNode
    onGoToNextEpisode: () => void
    onGoToPreviousEpisode?: () => void
    mediaInfoDuration?: number
}

type ChapterProps = {
    title: string
    startTime: number
    endTime: number
}

export function SeaMediaPlayer(props: SeaMediaPlayerProps) {
    const {
        url,
        poster,
        isLoading,
        isPlaybackError,
        playerRef,
        tracks = [],
        chapters = [],
        videoLayoutSlots,
        loadingText,
        onCanPlay: _onCanPlay,
        onProviderChange: _onProviderChange,
        onProviderSetup: _onProviderSetup,
        onEnded: _onEnded,
        onTimeUpdate: _onTimeUpdate,
        onDurationChange: _onDurationChange,
        onGoToNextEpisode,
        onGoToPreviousEpisode,
        settingsItems,
        mediaInfoDuration,
    } = props

    const serverStatus = useServerStatus()

    const [duration, setDuration] = React.useState(0)

    const { media, progress } = useSeaMediaPlayer()

    const [trueDuration, setTrueDuration] = React.useState(0)
    React.useEffect(() => {
        if (!mediaInfoDuration) return
        const durationFixed = mediaInfoDuration || 0
        if (durationFixed) {
            setTrueDuration(durationFixed)
        }
    }, [mediaInfoDuration])

    const [progressItem, setProgressItem] = useAtom(__seaMediaPlayer_scopedProgressItemAtom) // scoped

    const autoPlay = useAtomValue(__seaMediaPlayer_autoPlayAtom)
    const autoNext = useAtomValue(__seaMediaPlayer_autoNextAtom)
    const discreteControls = useAtomValue(__seaMediaPlayer_discreteControlsAtom)
    const autoSkipIntroOutro = useAtomValue(__seaMediaPlayer_autoSkipIntroOutroAtom)
    const [volume, setVolume] = useAtom(__seaMediaPlayer_volumeAtom)
    const [muted, setMuted] = useAtom(__seaMediaPlayer_mutedAtom)

    // Store the updated progress
    const [currentProgress, setCurrentProgress] = useAtom(__seaMediaPlayer_scopedCurrentProgressAtom)
    React.useEffect(() => {
        setCurrentProgress(progress.currentProgress ?? 0)
    }, [progress.currentProgress])

    const [showSkipIntroButton, setShowSkipIntroButton] = React.useState(false)
    const [showSkipEndingButton, setShowSkipEndingButton] = React.useState(false)

    const watchHistoryRef = React.useRef<number>(0)
    const checkTimeRef = React.useRef<number>(0)
    const canPlayRef = React.useRef<boolean>(false)
    const previousUrlRef = React.useRef<string | { src: string, type: string } | undefined>(undefined)

    // Track last focused element
    const lastFocusedElementRef = React.useRef<HTMLElement | null>(null)

    /** AniSkip **/
    const { data: aniSkipData } = useSkipData(media?.idMal, progress.currentEpisodeNumber ?? -1)

    /** Progress update **/
    const { mutate: updateProgress, isPending: isUpdatingProgress, isSuccess: hasUpdatedProgress } = useUpdateAnimeEntryProgress(
        media?.id,
        currentProgress,
    )

    const onDurationChange = React.useCallback((detail: number, e: MediaDurationChangeEvent) => {
        _onDurationChange?.(detail, e)

        setDuration(detail)
    }, [])

    /**
     * Continuity
     */
    const { handleUpdateWatchHistory } = useHandleContinuityWithMediaPlayer(playerRef, progress.currentEpisodeNumber, media?.id)

    /**
     * Discord Rich Presence
     */
        // const { mutate: setAnimeDiscordActivity } = useSetDiscordLegacyAnimeActivity()
    const { mutate: setAnimeDiscordActivity } = useSetDiscordAnimeActivityWithProgress()
    const { mutate: updateAnimeDiscordActivity } = useUpdateDiscordAnimeActivityWithProgress()
    const { mutate: cancelDiscordActivity } = useCancelDiscordActivity()

    // useInterval(() => {
    //     if(!playerRef.current) return

    //     if (serverStatus?.settings?.discord?.enableRichPresence && serverStatus?.settings?.discord?.enableAnimeRichPresence) {
    //         updateAnimeDiscordActivity({
    //             progress: Math.floor(playerRef.current?.currentTime ?? 0),
    //             duration: Math.floor(playerRef.current?.duration ?? 0),
    //             paused: playerRef.current?.paused ?? false,
    //         })
    //     }
    // }, 6000)
    React.useEffect(() => {
        const interval = setInterval(() => {
            if (!playerRef.current || !canPlayRef.current) return

            if (serverStatus?.settings?.discord?.enableRichPresence && serverStatus?.settings?.discord?.enableAnimeRichPresence) {
                updateAnimeDiscordActivity({
                    progress: Math.floor(playerRef.current?.currentTime ?? 0),
                    duration: Math.floor(playerRef.current?.duration ?? 0),
                    paused: playerRef.current?.paused ?? false,
                })
            }
        }, 6000)

        return () => clearInterval(interval)
    }, [serverStatus?.settings?.discord, url, canPlayRef.current])

    React.useEffect(() => {
        if (previousUrlRef.current === url) return
        previousUrlRef.current = url

        // Reset the canPlayRef when the url changes
        canPlayRef.current = false
    }, [url])

    const onTimeUpdate = (detail: MediaTimeUpdateEventDetail, e: MediaTimeUpdateEvent) => { // let React compiler optimize
        _onTimeUpdate?.(detail, e)

        /**
         * AniSkip
         */
        if (
            aniSkipData?.op?.interval &&
            !!detail?.currentTime &&
            detail?.currentTime >= aniSkipData.op.interval.startTime &&
            detail?.currentTime <= aniSkipData.op.interval.endTime
        ) {
            setShowSkipIntroButton(true)
            if (autoSkipIntroOutro) {
                seekTo(aniSkipData?.op?.interval?.endTime || 0)
            }
        } else {
            setShowSkipIntroButton(false)
        }
        if (
            aniSkipData?.ed?.interval &&
            Math.abs(aniSkipData.ed.interval.startTime - (aniSkipData?.ed?.episodeLength)) < 500 &&
            !!detail?.currentTime &&
            detail?.currentTime >= aniSkipData.ed.interval.startTime &&
            detail?.currentTime <= aniSkipData.ed.interval.endTime
        ) {
            setShowSkipEndingButton(true)
            if (autoSkipIntroOutro) {
                seekTo(aniSkipData?.ed?.interval?.endTime || 0)
            }
        } else {
            setShowSkipEndingButton(false)
        }

        if (watchHistoryRef.current > 2000) {
            watchHistoryRef.current = 0
            handleUpdateWatchHistory()
        }

        watchHistoryRef.current++

        /**
         * Progress
         */
        if (checkTimeRef.current < 200) {
            checkTimeRef.current++
            return
        }
        checkTimeRef.current = 0

        // Use trueDuration if available, otherwise fallback to the dynamic duration
        const effectiveDuration = trueDuration || duration

        if (
            media &&
            // valid episode number
            progress.currentEpisodeNumber != null &&
            progress.currentEpisodeNumber > 0 &&
            // progress wasn't updated
            (!progressItem || progress.currentEpisodeNumber > progressItem.episodeNumber) &&
            // video is almost complete using the fixed duration
            effectiveDuration > 0 && (detail.currentTime / effectiveDuration) >= 0.8 &&
            // episode number greater than progress
            progress.currentEpisodeNumber > (currentProgress ?? 0)
        ) {
            if (serverStatus?.settings?.library?.autoUpdateProgress) {
                if (!isUpdatingProgress) {
                    updateProgress({
                        episodeNumber: progress.currentEpisodeNumber,
                        mediaId: media?.id,
                        totalEpisodes: media?.episodes || 0,
                        malId: media?.idMal || undefined,
                    }, {
                        onSuccess: () => {
                            setCurrentProgress(progress.currentEpisodeNumber!)
                        },
                    })
                }
            } else {
                setProgressItem({
                    episodeNumber: progress.currentEpisodeNumber,
                })
            }
        }
    }

    /**
     * Watch continuity
     */
    const { watchHistory, waitForWatchHistory, getEpisodeContinuitySeekTo } = useHandleCurrentMediaContinuity(media?.id)

    const wentToNextEpisodeRef = React.useRef(false)
    const onEnded = (e: MediaEndedEvent) => {
        _onEnded?.(e)

        if (autoNext && !wentToNextEpisodeRef.current) {
            onGoToNextEpisode()
            wentToNextEpisodeRef.current = true
        }
    }

    const onProviderChange = (provider: MediaProviderAdapter | null, e: MediaProviderChangeEvent) => {
        _onProviderChange?.(provider, e)
    }

    const onProviderSetup = (provider: MediaProviderAdapter, e: MediaProviderSetupEvent) => {
        _onProviderSetup?.(provider, e)
    }


    const onCanPlay = (e: MediaCanPlayDetail, event: MediaCanPlayEvent) => {
        _onCanPlay?.(e, event)

        canPlayRef.current = true

        if (__isDesktop__ && wentToNextEpisodeRef.current) {
            logger("MEDIA PLAYER").info("Restoring fullscreen")
            try {
                playerRef.current?.enterFullscreen()
                playerRef.current?.el?.focus()
            }
            catch {
            }
        }

        wentToNextEpisodeRef.current = false

        // If the watch history is found and the episode number matches, seek to the last watched time
        if (progress.currentEpisodeNumber && watchHistory?.found && watchHistory.item?.episodeNumber === progress.currentEpisodeNumber) {
            const lastWatchedTime = getEpisodeContinuitySeekTo(progress.currentEpisodeNumber,
                playerRef.current?.currentTime,
                playerRef.current?.duration)
            logger("MEDIA PLAYER").info("Watch continuity: Seeking to last watched time", { lastWatchedTime })
            if (lastWatchedTime > 0) {
                logger("MEDIA PLAYER").info("Watch continuity: Seeking to", lastWatchedTime)
                Object.assign(playerRef.current || {}, { currentTime: lastWatchedTime })
            }
        }

        if (
            serverStatus?.settings?.discord?.enableRichPresence &&
            serverStatus?.settings?.discord?.enableAnimeRichPresence &&
            media?.id
        ) {
            const videoProgress = playerRef.current?.currentTime ?? 0
            const videoDuration = playerRef.current?.duration ?? 0
            logger("MEDIA PLAYER").info("Setting discord activity", {
                videoProgress,
                videoDuration,
            })
            setAnimeDiscordActivity({
                mediaId: media?.id ?? 0,
                title: media?.title?.userPreferred || media?.title?.romaji || media?.title?.english || "Watching",
                image: media?.coverImage?.large || media?.coverImage?.medium || "",
                isMovie: media?.format === "MOVIE",
                episodeNumber: progress.currentEpisodeNumber ?? 0,
                progress: Math.floor(videoProgress),
                duration: Math.floor(videoDuration),
                totalEpisodes: media?.episodes,
                currentEpisodeCount: media?.nextAiringEpisode?.episode ? media?.nextAiringEpisode?.episode - 1 : media?.episodes,
                episodeTitle: progress?.currentEpisodeTitle || undefined,
            })
        }

        if (autoPlay) {
            playerRef.current?.play()
        }
    }

    function seekTo(time: number) {
        Object.assign(playerRef.current ?? {}, { currentTime: time })
    }

    function onSkipIntro() {
        if (!aniSkipData?.op?.interval?.endTime) return
        seekTo(aniSkipData?.op?.interval?.endTime || 0)
    }

    function onSkipOutro() {
        if (!aniSkipData?.ed?.interval?.endTime) return
        seekTo(aniSkipData?.ed?.interval?.endTime || 0)
    }

    const cues = React.useMemo(() => {
        const introStart = aniSkipData?.op?.interval?.startTime ?? 0
        const introEnd = aniSkipData?.op?.interval?.endTime ?? 0
        const outroStart = aniSkipData?.ed?.interval?.startTime ?? 0
        const outroEnd = aniSkipData?.ed?.interval?.endTime ?? 0
        const ret = []
        if (introEnd > introStart) {
            ret.push({
                startTime: introStart,
                endTime: introEnd,
                text: "Intro",
            })
        }
        if (outroEnd > outroStart) {
            ret.push({
                startTime: outroStart,
                endTime: outroEnd,
                text: "Outro",
            })
        }
        return ret
    }, [])

    React.useEffect(() => {
        mousetrap.bind("f", () => {
            logger("MEDIA PLAYER").info("Fullscreen key pressed")
            try {
                playerRef.current?.enterFullscreen()
                playerRef.current?.el?.focus()
            }
            catch {
            }
        })

        return () => {
            mousetrap.unbind("f")

            if (serverStatus?.settings?.discord?.enableRichPresence && serverStatus?.settings?.discord?.enableAnimeRichPresence) {
                cancelDiscordActivity()
            }
        }
    }, [])

    const { inject, remove } = useSeaCommandInject()

    React.useEffect(() => {

        inject("media-player-controls", {
            items: [
                {
                    id: "toggle-play",
                    value: "toggle-play",
                    heading: "Controls",
                    priority: 100,
                    render: () => (
                        <>
                            <p>Toggle Play</p>
                        </>
                    ),
                    onSelect: () => {
                        if (playerRef.current?.paused) {
                            playerRef.current?.play()
                        } else {
                            playerRef.current?.pause()
                        }
                    },
                },
                {
                    id: "fullscreen",
                    value: "fullscreen",
                    heading: "Controls",
                    priority: 99,
                    render: () => (
                        <>
                            <p>Fullscreen</p>
                        </>
                    ),
                    onSelect: () => {
                        playerRef.current?.enterFullscreen()
                    },
                },
                {
                    id: "next-episode",
                    value: "next-episode",
                    heading: "Controls",
                    priority: 98,
                    render: () => (
                        <>
                            <p>Next Episode</p>
                        </>
                    ),
                    onSelect: () => onGoToNextEpisode(),
                },
                {
                    id: "previous-episode",
                    value: "previous-episode",
                    heading: "Controls",
                    priority: 97,
                    render: () => (
                        <>
                            <p>Previous Episode</p>
                        </>
                    ),
                    onSelect: () => onGoToPreviousEpisode?.(),
                },
            ],
        })

        return () => remove("media-player-controls")
    }, [url])

    const { onMediaEnterFullscreenRequest } = useFullscreenHandler(playerRef)

    return (
        <>
            <div data-sea-media-player-container className="aspect-video relative w-full self-start mx-auto">
                {(!!isPlaybackError?.length) ? (
                    <LuffyError title="Oops!">
                        <p className="max-w-md">
                            {capitalize(isPlaybackError)}
                        </p>
                    </LuffyError>
                ) : (!!url && !isLoading) ? (
                    <MediaPlayer
                        data-sea-media-player
                        streamType="on-demand"
                        playsInline
                        ref={playerRef}
                        autoPlay={autoPlay}
                        crossOrigin
                        src={url}
                        poster={poster}
                        aspectRatio="16/9"
                        controlsDelay={discreteControls ? 500 : undefined}
                        className={cn(discreteControls && "discrete-controls")}
                        onProviderChange={onProviderChange}
                        onMediaEnterFullscreenRequest={onMediaEnterFullscreenRequest}
                        onFullscreenChange={(isFullscreen: boolean, event: MediaFullscreenChangeEvent) => {
                            if (isFullscreen) {
                                // Store the currently focused element
                                lastFocusedElementRef.current = document.activeElement as HTMLElement
                            } else {
                                // Restore focus
                                setTimeout(() => {
                                    lastFocusedElementRef.current?.focus()
                                }, 100)
                            }
                        }}
                        onProviderSetup={onProviderSetup}
                        volume={volume}
                        onVolumeChange={detail => setVolume(detail.volume)}
                        onTimeUpdate={onTimeUpdate}
                        onDurationChange={onDurationChange}
                        onCanPlay={onCanPlay}
                        onEnded={onEnded}
                        muted={muted}
                        onMediaMuteRequest={() => setMuted(true)}
                        onMediaUnmuteRequest={() => setMuted(false)}
                    >
                        <MediaProvider>
                            {tracks.map((track, index) => (
                                <Track key={`track-${index}`} {...track} />
                            ))}
                            {/*{chapters?.length > 0 ? chapters.map((chapter, index) => (*/}
                            {/*    <Track kind="chapters" key={`chapter-${index}`} {...chapter} />*/}
                            {/*)) : cues.length > 0 ? cues.map((cue, index) => (*/}
                            {/*    <Track kind="chapters" key={`cue-${index}`} {...cue} />*/}
                            {/*)) : null}*/}
                        </MediaProvider>
                        <div
                            data-sea-media-player-skip-intro-outro-container
                            className="absolute bottom-24 px-4 w-full justify-between flex items-center"
                        >
                            <div>
                                {showSkipIntroButton && (
                                    <Button intent="white" size="sm" onClick={onSkipIntro} loading={autoSkipIntroOutro}>
                                        Skip opening
                                    </Button>
                                )}
                            </div>
                            <div>
                                {showSkipEndingButton && (
                                    <Button intent="white" size="sm" onClick={onSkipOutro} loading={autoSkipIntroOutro}>
                                        Skip ending
                                    </Button>
                                )}
                            </div>
                        </div>
                        <DefaultVideoLayout
                            icons={vidstackLayoutIcons}
                            slots={{
                                ...videoLayoutSlots,
                                settingsMenuEndItems: <>
                                    {settingsItems}
                                    <SeaMediaPlayerPlaybackSubmenu />
                                </>,
                                // centerControlsGroupStart: <div>
                                //     {onGoToPreviousEpisode && (
                                //         <IconButton
                                //             intent="white-basic"
                                //             size="lg"
                                //             onClick={onGoToPreviousEpisode}
                                //             aria-label="Previous Episode"
                                //             icon={<LuArrowLeft className="size-12" />}
                                //         />
                                //     )}
                                // </div>,
                                // centerControlsGroupEnd: <div className="flex items-center justify-center gap-2">
                                //     {onGoToNextEpisode && (
                                //         <IconButton
                                //             intent="white-basic"
                                //             size="lg"
                                //             onClick={onGoToNextEpisode}
                                //             aria-label="Next Episode"
                                //             icon={<LuArrowRight className="size-12" />}
                                //         />
                                //     )}
                                // </div>
                            }}
                        />
                    </MediaPlayer>
                ) : (
                    <Skeleton
                        data-sea-media-player-loading-container
                        className="w-full h-full absolute flex justify-center items-center flex-col space-y-4"
                    >
                        <LoadingSpinner
                            spinner={
                                <div className="w-16 h-16 lg:w-[100px] lg:h-[100px] relative">
                                    <Image
                                        src="/logo_2.png"
                                        alt="Loading..."
                                        priority
                                        fill
                                        className="animate-pulse"
                                    />
                                </div>
                            }
                        />
                        <div className="text-center text-xs lg:text-sm">
                            {loadingText ?? <>
                                <p>Loading...</p>
                            </>}
                        </div>
                    </Skeleton>
                )}
            </div>
        </>
    )
}
