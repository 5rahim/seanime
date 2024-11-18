import { Anime_Entry } from "@/api/generated/types"
import { useUpdateAnimeEntryProgress } from "@/api/hooks/anime_entries.hooks"
import { __mediaplayer_discreteControlsAtom } from "@/app/(main)/_atoms/builtin-mediaplayer.atoms"
import { EpisodeGridItem } from "@/app/(main)/_features/anime/_components/episode-grid-item"
import { MediaEpisodeInfoModal } from "@/app/(main)/_features/media/_components/media-episode-info-modal"
import {
    OnlinestreamParametersButton,
    OnlinestreamPlaybackSubmenu,
    OnlinestreamProviderButton,
    OnlinestreamVideoQualitySubmenu,
    SwitchSubOrDubButton,
} from "@/app/(main)/onlinestream/_components/onlinestream-video-addons"
import { OnlinestreamManualMappingModal } from "@/app/(main)/onlinestream/_containers/onlinestream-manual-matching"
import { useHandleOnlinestream } from "@/app/(main)/onlinestream/_lib/handle-onlinestream"
import { OnlinestreamManagerProvider } from "@/app/(main)/onlinestream/_lib/onlinestream-manager"
import {
    __onlinestream_autoNextAtom,
    __onlinestream_autoPlayAtom,
    __onlinestream_autoSkipIntroOutroAtom,
    __onlinestream_volumeAtom,
} from "@/app/(main)/onlinestream/_lib/onlinestream.atoms"
import { useSkipData } from "@/app/(main)/onlinestream/_lib/skip"
import { LuffyError } from "@/components/shared/luffy-error"
import { SeaLink } from "@/components/shared/sea-link"
import { Button, IconButton } from "@/components/ui/button"
import { cn } from "@/components/ui/core/styling"
import { LoadingSpinner } from "@/components/ui/loading-spinner"
import { ScrollArea } from "@/components/ui/scroll-area"
import { Skeleton } from "@/components/ui/skeleton"
import {
    isHLSProvider,
    MediaPlayer,
    MediaPlayerInstance,
    MediaProvider,
    MediaProviderAdapter,
    MediaProviderChangeEvent,
    MediaProviderSetupEvent,
    Track,
} from "@vidstack/react"
import { defaultLayoutIcons, DefaultVideoLayout } from "@vidstack/react/player/layouts/default"
import HLS from "hls.js"
import { atom } from "jotai/index"
import { useAtom, useAtomValue } from "jotai/react"
import { atomWithStorage } from "jotai/utils"
import { usePathname, useRouter, useSearchParams } from "next/navigation"
import React from "react"
import { AiOutlineArrowLeft } from "react-icons/ai"
import { FaSearch } from "react-icons/fa"
import { TbLayoutSidebarRightCollapse, TbLayoutSidebarRightExpand } from "react-icons/tb"
import { useUpdateEffect, useWindowSize } from "react-use"
import "@vidstack/react/player/styles/default/theme.css"
import "@vidstack/react/player/styles/default/layouts/video.css"

type OnlinestreamPageProps = {
    animeEntry?: Anime_Entry
    animeEntryLoading?: boolean
    hideBackButton?: boolean
}

type ProgressItem = {
    episodeNumber: number
}
const progressItemAtom = atom<ProgressItem | undefined>(undefined)

const theaterModeAtom = atomWithStorage("sea-onlinestream-theater-mode", false)


export function OnlinestreamPage({ animeEntry, animeEntryLoading, hideBackButton }: OnlinestreamPageProps) {

    const router = useRouter()
    const pathname = usePathname()
    const searchParams = useSearchParams()
    const mediaId = searchParams.get("id")
    const urlEpNumber = searchParams.get("episode")

    const ref = React.useRef<MediaPlayerInstance>(null)

    const [theaterMode, setTheaterMode] = useAtom(theaterModeAtom)
    const [volume, setVolume] = useAtom(__onlinestream_volumeAtom)

    const autoPlay = useAtomValue(__onlinestream_autoPlayAtom)
    const autoNext = useAtomValue(__onlinestream_autoNextAtom)
    const autoSkipIntroOutro = useAtomValue(__onlinestream_autoSkipIntroOutroAtom)
    const discreteControls = useAtomValue(__mediaplayer_discreteControlsAtom)
    const [progressItem, setProgressItem] = useAtom(progressItemAtom)

    const [currentProgress, setCurrentProgress] = React.useState(animeEntry?.listData?.progress ?? 0)

    const progress = React.useMemo(() => {
        setCurrentProgress(animeEntry?.listData?.progress ?? 0)
        return animeEntry?.listData?.progress ?? 0
    }, [animeEntry?.listData?.progress])

    const {
        episodes,
        currentEpisodeDetails,
        opts,
        url,
        onMediaDetached,
        onProviderSetup: _onProviderSetup,
        onCanPlay: _onCanPlay,
        onFatalError,
        loadPage,
        media,
        episodeSource,
        currentEpisodeNumber,
        handleChangeEpisodeNumber,
        episodeLoading,
        isErrorEpisodeSource,
        isErrorProvider,
        provider,
        handleUpdateWatchHistory,
    } = useHandleOnlinestream({
        mediaId,
        ref,
    })

    const maxEp = media?.nextAiringEpisode?.episode ? (media?.nextAiringEpisode?.episode - 1) : media?.episodes || 0

    /** AniSkip **/
    const { data: aniSkipData } = useSkipData(media?.idMal, currentEpisodeNumber)

    const [showSkipIntroButton, setShowSkipIntroButton] = React.useState(false)
    const [showSkipEndingButton, setShowSkipEndingButton] = React.useState(false)
    const [duration, setDuration] = React.useState(0)

    const seekTo = React.useCallback((time: number) => {
        Object.assign(ref.current ?? {}, { currentTime: time })
    }, [])

    /**
     * Set episode number on mount
     */
    const firstRenderRef = React.useRef(true)
    useUpdateEffect(() => {
        if (!!media && firstRenderRef.current) {
            const maxEp = media?.nextAiringEpisode?.episode ? (media?.nextAiringEpisode?.episode - 1) : media?.episodes || 0
            const _urlEpNumber = urlEpNumber ? Number(urlEpNumber) : undefined
            const progress = animeEntry?.listData?.progress ?? 0
            const nextProgressNumber = maxEp ? (progress + 1 < maxEp ? progress + 1 : maxEp) : 1
            handleChangeEpisodeNumber(_urlEpNumber || nextProgressNumber || 1)
            firstRenderRef.current = false
        }
    }, [media])

    React.useEffect(() => {
        const t = setTimeout(() => {
            if (urlEpNumber) {
                router.replace(pathname + `?id=${mediaId}`)
            }
        }, 500)

        return () => clearTimeout(t)
    }, [mediaId])

    function goToNextEpisode() {
        handleChangeEpisodeNumber(currentEpisodeNumber + 1 < maxEp ? currentEpisodeNumber + 1 : currentEpisodeNumber)
    }

    function onProviderChange(
        provider: MediaProviderAdapter | null,
        nativeEvent: MediaProviderChangeEvent,
    ) {
        if (isHLSProvider(provider)) {
            provider.library = HLS
            provider.config = {
                // debug: true,
            }
        }
    }

    function onProviderSetup(provider: MediaProviderAdapter, nativeEvent: MediaProviderSetupEvent) {
        if (isHLSProvider(provider)) {
            if (HLS.isSupported()) {
                _onProviderSetup()
                provider.instance?.on(HLS.Events.MEDIA_DETACHED, (event) => {
                    onMediaDetached()
                })
                provider.instance?.on(HLS.Events.ERROR, (event, data) => {
                    if (data.fatal) {
                        onFatalError()
                    }
                })
            } else if (provider.video.canPlayType("application/vnd.apple.mpegurl")) {
                provider.video.src = url || ""
            }
        }
    }

    const { width } = useWindowSize()

    /** Scroll to selected episode element when the episode list changes (on mount) **/
    const episodeListContainerRef = React.useRef<HTMLDivElement>(null)
    React.useEffect(() => {
        if (episodeListContainerRef.current && width > 1024 && !theaterMode) {
            React.startTransition(() => {
                const element = document.getElementById(`episode-${currentEpisodeNumber}`)
                if (element) {
                    element.scrollIntoView({ behavior: "smooth" })
                    // React.startTransition(() => {
                    //     window.scrollTo({ top: 0 })
                    // })
                }
            })
        }
    }, [episodeListContainerRef.current, episodes, currentEpisodeNumber, theaterMode])

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
        const t = setTimeout(() => {
            const element = document.querySelector(".vds-quality-menu")
            if (opts.hasCustomQualities) {
                // Toggle the class
                element?.classList?.add("force-hidden")
            } else {
                // Toggle the class
                element?.classList?.remove("force-hidden")
            }
        }, 1000)
        return () => clearTimeout(t)
    }, [opts.hasCustomQualities, url])

    const { mutate: updateProgress, isPending: isUpdatingProgress, isSuccess: hasUpdatedProgress } = useUpdateAnimeEntryProgress(
        mediaId,
        currentEpisodeNumber,
    )

    const checkTimeRef = React.useRef<number>(0)

    const watchHistoryRef = React.useRef<number>(0)

    if (!loadPage || !media || animeEntryLoading) return <div className="space-y-4">
        <div className="flex gap-4 items-center relative">
            <Skeleton className="h-12" />
        </div>
        <div
            className="grid 2xl:grid-cols-[1fr,450px] gap-4 xl:gap-4"
        >
            <div className="w-full min-h-[70dvh] relative">
                <Skeleton className="h-full w-full absolute" />
            </div>

            <Skeleton className="hidden 2xl:block relative h-[78dvh] overflow-y-auto pr-4 pt-0" />

        </div>
    </div>

    return (
        <>
            <OnlinestreamManagerProvider
                opts={opts}
            >
                <div className="flex flex-col lg:flex-row gap-2 w-full justify-between">
                    {!hideBackButton && <div className="flex w-full gap-4 items-center relative">
                        <SeaLink href={`/entry?id=${media?.id}`}>
                            <IconButton icon={<AiOutlineArrowLeft />} rounded intent="white-outline" size="md" />
                        </SeaLink>
                        <h3 className="max-w-full lg:max-w-[50%] text-ellipsis truncate">{media.title?.userPreferred}</h3>
                    </div>}

                    <div className="flex gap-2 items-center w-full">
                        {(!!progressItem && progressItem.episodeNumber > currentProgress) && <Button
                            className="animate-pulse"
                            loading={isUpdatingProgress}
                            disabled={hasUpdatedProgress}
                            onClick={() => {
                                updateProgress({
                                    episodeNumber: progressItem.episodeNumber,
                                    mediaId: media.id,
                                    totalEpisodes: media.episodes || 0,
                                    malId: media.idMal || undefined,
                                }, {
                                    onSuccess: () => {
                                        setProgressItem(undefined)
                                    },
                                })
                                setCurrentProgress(progressItem.episodeNumber)
                            }}
                        >Update progress</Button>}

                        {animeEntry && <OnlinestreamManualMappingModal entry={animeEntry}>
                            <IconButton
                                size="sm"
                                intent="gray-basic"
                                icon={<FaSearch />}
                            />
                        </OnlinestreamManualMappingModal>}

                        <SwitchSubOrDubButton />

                        {!!mediaId && <OnlinestreamParametersButton mediaId={Number(mediaId)} />}

                        <div className="flex flex-1"></div>

                        <IconButton
                            onClick={() => setTheaterMode(p => !p)}
                            intent="gray-basic"
                            icon={theaterMode ? <TbLayoutSidebarRightExpand /> : <TbLayoutSidebarRightCollapse />}
                        />
                    </div>
                </div>

                <div
                    className={cn(
                        "flex gap-4 w-full flex-col 2xl:flex-row",
                        theaterMode && "2xl:flex-col",
                    )}
                >
                    <div
                        className={cn(
                            "aspect-video relative w-full self-start mx-auto",
                            theaterMode && "max-h-[90vh] !w-auto aspect-video mx-auto",
                        )}
                    >
                        {!provider ? (
                            <div className="flex items-center flex-col justify-center w-full h-full">
                                <LuffyError title="No provider selected" />
                                {!!mediaId && <OnlinestreamParametersButton mediaId={Number(mediaId)} />}
                            </div>
                        ) : isErrorProvider ? <LuffyError title="Provider error" /> : !!url ? <MediaPlayer
                            streamType="on-demand"
                            playsInline
                            ref={ref}
                            autoPlay={autoPlay}
                            crossOrigin="anonymous"
                            controlsDelay={discreteControls ? 500 : undefined}
                            src={{
                                src: url || "",
                                type: "application/x-mpegurl",
                            }}
                            poster={currentEpisodeDetails?.image || media.coverImage?.extraLarge || ""}
                            aspectRatio="16/9"
                            onProviderChange={onProviderChange}
                            onProviderSetup={onProviderSetup}
                            className={cn(discreteControls && "discrete-controls")}
                            volume={volume}
                            onVolumeChange={(e, n) => {
                                setVolume(n.detail.volume)
                            }}
                            onTimeUpdate={(e) => {
                                if (watchHistoryRef.current > 2000) {
                                    watchHistoryRef.current = 0

                                    handleUpdateWatchHistory()
                                }
                                watchHistoryRef.current++

                                if (checkTimeRef.current < 200) {
                                    checkTimeRef.current++
                                    return
                                }
                                checkTimeRef.current = 0

                                if (aniSkipData?.op && e?.currentTime && e?.currentTime >= aniSkipData.op.interval.startTime && e?.currentTime <= aniSkipData.op.interval.endTime) {
                                    setShowSkipIntroButton(true)
                                    if (autoSkipIntroOutro) {
                                        seekTo(aniSkipData?.op?.interval?.endTime || 0)
                                    }
                                } else {
                                    setShowSkipIntroButton(false)
                                }
                                if (aniSkipData?.ed &&
                                    Math.abs(aniSkipData.ed.interval.startTime - (aniSkipData?.ed?.episodeLength)) < 500 &&
                                    e?.currentTime &&
                                    e?.currentTime >= aniSkipData.ed.interval.startTime &&
                                    e?.currentTime <= aniSkipData.ed.interval.endTime
                                ) {
                                    setShowSkipEndingButton(true)
                                    if (autoSkipIntroOutro) {
                                        seekTo(aniSkipData?.ed?.interval?.endTime || 0)
                                    }
                                } else {
                                    setShowSkipEndingButton(false)
                                }

                                if (
                                    (!progressItem || currentEpisodeNumber > progressItem.episodeNumber) &&
                                    duration > 0 && (e.currentTime / duration) >= 0.8 &&
                                    currentEpisodeNumber > currentProgress
                                ) {
                                    setProgressItem({
                                        episodeNumber: currentEpisodeNumber,
                                    })
                                }
                            }}
                            onEnded={(e) => {
                                console.log("onEnded", e)
                                if (autoNext) {
                                    goToNextEpisode()
                                }
                            }}
                            onCanPlay={(e) => {
                                if (e.duration && e.duration > 0) {
                                    setDuration(e.duration)
                                } else {
                                    setDuration(0)
                                }
                                if (autoPlay) {
                                    ref.current?.play()
                                }
                                _onCanPlay()
                            }}
                        >
                            <MediaProvider>
                                {episodeSource?.subtitles?.map((sub) => {
                                    return <Track
                                        key={sub.url}
                                        {...{
                                            id: sub.language,
                                            label: sub.language,
                                            kind: "subtitles",
                                            src: sub.url,
                                            language: sub.language,
                                            default: sub.language
                                                ? sub.language?.toLowerCase() === "english" || sub.language?.toLowerCase() === "en-us"
                                                : sub.language?.toLowerCase() === "english" || sub.language?.toLowerCase() === "en-us",
                                        }}
                                    />
                                })}
                                {cues?.map((cue) => {
                                    return <Track
                                        key={cue.text}
                                        {...{
                                            id: cue.text,
                                            label: cue.text,
                                            kind: "chapters",
                                            src: "",
                                            language: "",
                                            default: false,
                                            srcLang: "",
                                            startTime: cue.startTime,
                                            endTime: cue.endTime,
                                        }}
                                    />
                                })}
                            </MediaProvider>
                            <div className="absolute bottom-24 px-4 w-full justify-between flex items-center">
                                <div>
                                    {(showSkipIntroButton) && (
                                        <Button
                                            intent="white"
                                            onClick={() => seekTo(aniSkipData?.op?.interval?.endTime || 0)}
                                            loading={autoSkipIntroOutro}
                                        >
                                            Skip intro
                                        </Button>
                                    )}
                                </div>
                                <div>
                                    {(showSkipEndingButton) && (
                                        <Button
                                            intent="white"
                                            onClick={() => seekTo(aniSkipData?.ed?.interval?.endTime || 0)}
                                            loading={autoSkipIntroOutro}
                                        >
                                            Skip ending
                                        </Button>
                                    )}
                                </div>
                            </div>
                            <DefaultVideoLayout
                                icons={defaultLayoutIcons}
                                slots={{
                                    settingsMenuEndItems: (<>
                                        {opts.hasCustomQualities ? (
                                            <OnlinestreamVideoQualitySubmenu />
                                        ) : null}
                                        <OnlinestreamPlaybackSubmenu />
                                    </>),
                                    beforeCaptionButton: (
                                        <div className="flex items-center">
                                            <OnlinestreamProviderButton />
                                        </div>
                                    ),
                                }}
                            />
                        </MediaPlayer> : (
                            !isErrorEpisodeSource ? <Skeleton className="h-full w-full absolute">
                                <LoadingSpinner containerClass="h-full absolute" />
                            </Skeleton> : <div>
                                <LuffyError
                                    title="Error"
                                >
                                    <p>
                                        Failed to load episode
                                    </p>
                                    <p>
                                        Try changing the provider or refresh the page
                                    </p>
                                </LuffyError>
                            </div>
                        )}
                    </div>

                    <ScrollArea
                        ref={episodeListContainerRef}
                        className={cn(
                            "2xl:max-w-[450px] w-full relative 2xl:sticky h-[75dvh] overflow-y-auto pr-4 pt-0",
                            theaterMode && "2xl:max-w-full",
                        )}
                    >
                        <div className="space-y-4">
                            {(!episodes?.length && !loadPage) && <p>
                                No episodes found
                            </p>}
                            {episodes?.filter(Boolean)?.sort((a, b) => a!.number - b!.number)?.map((episode, idx) => {
                                return (
                                    <EpisodeGridItem
                                        key={idx + (episode.title || "") + episode.number}
                                        id={`episode-${String(episode.number)}`}
                                        onClick={() => handleChangeEpisodeNumber(episode.number)}
                                        title={media.format === "MOVIE" ? "Complete movie" : `Episode ${episode.number}`}
                                        episodeTitle={episode.title}
                                        description={episode.description ?? undefined}
                                        image={episode.image}
                                        media={media}
                                        isSelected={episode.number === currentEpisodeNumber}
                                        disabled={episodeLoading}
                                        isWatched={progress ? episode.number <= progress : undefined}
                                        className="flex-none w-full"
                                        isFiller={episode.isFiller}
                                        action={<>
                                            <MediaEpisodeInfoModal
                                                title={media.format === "MOVIE" ? "Complete movie" : `Episode ${episode.number}`}
                                                image={episode?.image}
                                                episodeTitle={episode.title}
                                                summary={episode?.description}
                                            />
                                        </>}
                                    />
                                )
                            })}
                            <p className="text-center text-[--muted] py-2">End</p>
                        </div>
                        <div
                            className={"z-[5] absolute bottom-0 w-full h-[2rem] bg-gradient-to-t from-[--background] to-transparent"}
                        />
                    </ScrollArea>
                </div>
            </OnlinestreamManagerProvider>
        </>
    )
}
