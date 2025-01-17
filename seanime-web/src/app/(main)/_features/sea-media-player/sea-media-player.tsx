import { useUpdateAnimeEntryProgress } from "@/api/hooks/anime_entries.hooks"
import { useHandleContinuityWithMediaPlayer, useHandleCurrentMediaContinuity } from "@/api/hooks/continuity.hooks"
import { useSkipData } from "@/app/(main)/_features/sea-media-player/aniskip"
import {
    __seaMediaPlayer_scopedCurrentProgressAtom,
    __seaMediaPlayer_scopedProgressItemAtom,
    useSeaMediaPlayer,
} from "@/app/(main)/_features/sea-media-player/sea-media-player-provider"
import { useServerStatus } from "@/app/(main)/_hooks/use-server-status"
import { LuffyError } from "@/components/shared/luffy-error"
import { vidstackLayoutIcons } from "@/components/shared/vidstack"
import { Button } from "@/components/ui/button"
import { cn } from "@/components/ui/core/styling"
import { LoadingSpinner } from "@/components/ui/loading-spinner"
import { Skeleton } from "@/components/ui/skeleton"
import { logger } from "@/lib/helpers/debug"
import {
    MediaCanPlayDetail,
    MediaCanPlayEvent,
    MediaDurationChangeEvent,
    MediaEndedEvent,
    MediaPlayer,
    MediaPlayerInstance,
    MediaProvider,
    MediaProviderAdapter,
    MediaProviderChangeEvent,
    MediaProviderSetupEvent,
    MediaTimeUpdateEvent,
    MediaTimeUpdateEventDetail,
    MediaVolumeChange,
    MediaVolumeChangeEvent,
    Track,
    type TrackProps,
} from "@vidstack/react"
import { DefaultVideoLayout, DefaultVideoLayoutProps } from "@vidstack/react/player/layouts/default"
import { atom } from "jotai"
import { useAtom } from "jotai/react"
import mousetrap from "mousetrap"
import Image from "next/image"
import React from "react"

const fullscreenAtom = atom(false)

export type SeaMediaPlayerProps = {
    url?: string
    poster?: string
    isLoading?: boolean
    isPlaybackError?: boolean
    playerRef: React.RefObject<MediaPlayerInstance>
    onProviderChange?: (provider: MediaProviderAdapter | null, e: MediaProviderChangeEvent) => void
    onProviderSetup?: (provider: MediaProviderAdapter, e: MediaProviderSetupEvent) => void
    onTimeUpdate?: (detail: MediaTimeUpdateEventDetail, e: MediaTimeUpdateEvent) => void
    onCanPlay?: (detail: MediaCanPlayDetail, e: MediaCanPlayEvent) => void
    onEnded?: (e: MediaEndedEvent) => void
    volume?: number
    autoSkipIntroOutro?: boolean
    onVolumeChange?: (detail: MediaVolumeChange, e: MediaVolumeChangeEvent) => void
    onDurationChange?: (detail: number, e: MediaDurationChangeEvent) => void
    autoPlay?: boolean
    autoNext?: boolean
    discreteControls?: boolean
    tracks?: TrackProps[]
    chapters?: ChapterProps[]
    videoLayoutSlots?: DefaultVideoLayoutProps["slots"]
    loadingText?: React.ReactNode
    onGoToNextEpisode: () => void
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
        autoSkipIntroOutro = false,
        volume = 1,
        onVolumeChange,
        autoPlay = false,
        autoNext = false,
        discreteControls = false,
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
    } = props

    const serverStatus = useServerStatus()

    const [duration, setDuration] = React.useState(0)

    const { media, progress } = useSeaMediaPlayer()

    const [progressItem, setProgressItem] = useAtom(__seaMediaPlayer_scopedProgressItemAtom) // scoped

    // Store the updated progress
    const [currentProgress, setCurrentProgress] = useAtom(__seaMediaPlayer_scopedCurrentProgressAtom)
    React.useEffect(() => {
        setCurrentProgress(progress.currentProgress ?? 0)
    }, [progress.currentProgress])

    const [showSkipIntroButton, setShowSkipIntroButton] = React.useState(false)
    const [showSkipEndingButton, setShowSkipEndingButton] = React.useState(false)

    const watchHistoryRef = React.useRef<number>(0)
    const checkTimeRef = React.useRef<number>(0)

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

        if (
            media &&
            // valid episode number
            progress.currentEpisodeNumber != null &&
            progress.currentEpisodeNumber > 0 &&
            // progress wasn't updated
            (!progressItem || progress.currentEpisodeNumber > progressItem.episodeNumber) &&
            // video is almost complete
            duration > 0 && (detail.currentTime / duration) >= 0.8 &&
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

    const onEnded = (e: MediaEndedEvent) => {
        _onEnded?.(e)

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

        if (autoNext) {
            onGoToNextEpisode()
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
        }
    }, [])

    return (
        <div className="aspect-video relative w-full self-start mx-auto">
            {isPlaybackError ? (
                <LuffyError title="Playback Error" />
            ) : (!!url && !isLoading) ? (
                <MediaPlayer
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
                    onProviderSetup={onProviderSetup}
                    volume={volume}
                    onVolumeChange={onVolumeChange}
                    onTimeUpdate={onTimeUpdate}
                    onDurationChange={onDurationChange}
                    onCanPlay={onCanPlay}
                    onEnded={onEnded}
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
                    <div className="absolute bottom-24 px-4 w-full justify-between flex items-center">
                        <div>
                            {showSkipIntroButton && (
                                <Button intent="white-subtle" size="sm" onClick={onSkipIntro} loading={autoSkipIntroOutro}>
                                    Skip opening
                                </Button>
                            )}
                        </div>
                        <div>
                            {showSkipEndingButton && (
                                <Button intent="white-subtle" size="sm" onClick={onSkipOutro} loading={autoSkipIntroOutro}>
                                    Skip ending
                                </Button>
                            )}
                        </div>
                    </div>
                    <DefaultVideoLayout
                        icons={vidstackLayoutIcons}
                        slots={videoLayoutSlots}
                    />
                </MediaPlayer>
            ) : (
                <Skeleton className="w-full h-full absolute flex justify-center items-center flex-col space-y-4">
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
    )
}
