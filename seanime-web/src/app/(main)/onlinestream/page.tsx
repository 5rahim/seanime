"use client"
import "@vidstack/react/player/styles/default/theme.css"
import "@vidstack/react/player/styles/default/layouts/video.css"
import { useGetAnilistMediaDetails } from "@/api/hooks/anilist.hooks"
import { useGetAnimeEntry, useUpdateAnimeEntryProgress } from "@/api/hooks/anime_entries.hooks"
import { OnlinestreamEpisodeListItem } from "@/app/(main)/onlinestream/_components/onlinestream-episode-list-item"
import {
    OnlinestreamParametersButton,
    OnlinestreamProviderButton,
    OnlinestreamServerButton,
    OnlinestreamSettingsButton,
} from "@/app/(main)/onlinestream/_components/onlinestream-video-addons"
import { OnlinestreamManagerProvider, useOnlinestreamManager } from "@/app/(main)/onlinestream/_lib/onlinestream-manager"
import {
    __onlinestream_autoNextAtom,
    __onlinestream_autoPlayAtom,
    __onlinestream_selectedEpisodeNumberAtom,
} from "@/app/(main)/onlinestream/_lib/onlinestream.atoms"
import { useSkipData } from "@/app/(main)/onlinestream/_lib/skip"
import { LuffyError } from "@/components/shared/luffy-error"
import { Button, IconButton } from "@/components/ui/button"
import { cn } from "@/components/ui/core/styling"
import { LoadingSpinner } from "@/components/ui/loading-spinner"
import { ScrollArea } from "@/components/ui/scroll-area"
import { Skeleton } from "@/components/ui/skeleton"
import { useQueryClient } from "@tanstack/react-query"
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
import { atom } from "jotai"
import { useAtom, useAtomValue } from "jotai/react"
import { atomWithStorage } from "jotai/utils"
import Image from "next/image"
import Link from "next/link"
import { usePathname, useRouter, useSearchParams } from "next/navigation"
import React from "react"
import { AiOutlineArrowLeft } from "react-icons/ai"
import { useUpdateEffect } from "react-use"

const theaterModeAtom = atomWithStorage("sea-onlinestream-theater-mode", false)
type ProgressItem = {
    episodeNumber: number
}
const progressItemAtom = atom<ProgressItem | undefined>(undefined)

export const dynamic = "force-static"

export default function Page() {

    const router = useRouter()
    const pathname = usePathname()
    const qc = useQueryClient()
    const searchParams = useSearchParams()
    const mediaId = searchParams.get("id")
    const urlEpNumber = searchParams.get("episode")
    const { data: mediaEntry, isLoading: mediaEntryLoading } = useGetAnimeEntry(mediaId)
    const { data: mediaDetails } = useGetAnilistMediaDetails(mediaId)

    const ref = React.useRef<MediaPlayerInstance>(null)

    const mediaIdRef = React.useRef(mediaId)

    const [theaterMode, setTheaterMode] = useAtom(theaterModeAtom)
    const [_episodeNumber, _setEpisodeNumber] = useAtom(__onlinestream_selectedEpisodeNumberAtom)

    const autoPlay = useAtomValue(__onlinestream_autoPlayAtom)
    const autoNext = useAtomValue(__onlinestream_autoNextAtom)
    const [progressItem, setProgressItem] = useAtom(progressItemAtom)

    const [currentProgress, setCurrentProgress] = React.useState(mediaEntry?.listData?.progress ?? 0)

    const progress = React.useMemo(() => {
        setCurrentProgress(mediaEntry?.listData?.progress ?? 0)
        return mediaEntry?.listData?.progress ?? 0
    }, [mediaEntry?.listData?.progress])

    const {
        episodes,
        currentEpisodeDetails,
        opts,
        url,
        onMediaDetached,
        onProviderSetup: _onProviderSetup,
        onFatalError,
        loadPage,
        media,
        episodeSource,
        currentEpisodeNumber,
        handleChangeEpisodeNumber,
        episodeLoading,
        isErrorEpisodeSource,
        isErrorProvider,
    } = useOnlinestreamManager({
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
     * Set episode number
     */
    useUpdateEffect(() => {
        if (!!media) {
            const maxEp = media?.nextAiringEpisode?.episode ? (media?.nextAiringEpisode?.episode - 1) : media?.episodes || 0
            const _urlEpNumber = urlEpNumber ? Number(urlEpNumber) : undefined
            const progress = mediaEntry?.listData?.progress ?? 0
            const nextProgressNumber = maxEp ? (progress + 1 < maxEp ? progress + 1 : maxEp) : 1
            console.log(nextProgressNumber, progress)
            handleChangeEpisodeNumber(_urlEpNumber || nextProgressNumber || 1)
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

    /** Scroll to selected episode element when the episode list changes (on mount) **/
    React.useEffect(() => {
        React.startTransition(() => {
            const element = document.getElementById(`episode-${currentEpisodeNumber}`)
            if (element) {
                element.scrollIntoView()
                React.startTransition(() => {
                    window.scrollTo({ top: 0 })
                })
            }
        })
    }, [episodes, currentEpisodeNumber])

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

    const { mutate: updateProgress, isPending: isUpdatingProgress, isSuccess: hasUpdatedProgress } = useUpdateAnimeEntryProgress(mediaId)

    if (!loadPage || !media || mediaEntryLoading) return <div className="px-4 lg:px-8 space-y-4">
        <div className="flex gap-4 items-center relative">
            <Skeleton className="h-12" />
        </div>
        <div
            className="grid xl:grid-cols-[1fr,400px] 2xl:grid-cols-[1fr,500px] gap-4 xl:gap-4"
        >
            <div className="w-full min-h-[70dvh] relative">
                <Skeleton className="h-full w-full absolute" />
            </div>

            <Skeleton className="hidden lg:block relative h-[78dvh] overflow-y-auto pr-4 pt-0" />

        </div>
    </div>

    return (
        <>
            <div className="relative px-4 lg:px-8 z-[5] space-y-4">
                <OnlinestreamManagerProvider
                    opts={opts}
                >

                    <div className="flex w-full justify-between">
                        <div className="flex gap-4 items-center relative">
                            <Link href={`/entry?id=${media?.id}`}>
                                <IconButton icon={<AiOutlineArrowLeft />} rounded intent="white-outline" size="md" />
                            </Link>
                            <h3>{media.title?.userPreferred}</h3>

                            {/*<IconButton*/}
                            {/*    icon={!theaterMode ? <GiTheater /> : <TbResize />}*/}
                            {/*    onClick={() => setTheaterMode(p => !p)}*/}
                            {/*    intent="gray-basic"*/}
                            {/*    size="lg"*/}
                            {/*/>*/}

                        </div>

                        <div className="flex gap-2 items-center">
                            {(!!progressItem) && <Button
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
                                        onSuccess: () => setProgressItem(undefined),
                                    })
                                    setCurrentProgress(progressItem.episodeNumber)
                                }}
                            >Update progress</Button>}

                            {!!mediaId && <OnlinestreamParametersButton mediaId={Number(mediaId)} />}
                        </div>
                    </div>

                    <div
                        className={cn(
                            "flex gap-4 w-full flex-col 2xl:flex-row",
                        )}
                    >
                        {/*<div className="space-y-4">*/}
                            <div
                                className={cn(
                                    "aspect-video relative w-full self-start max-h-[78dvh] mx-auto",
                                )}
                            >
                                {isErrorProvider ? <LuffyError title="Provider error" /> : !!url ? <MediaPlayer
                                    playsInline
                                    ref={ref}
                                    crossOrigin="anonymous"
                                    src={{
                                        src: url || "",
                                        type: "application/x-mpegurl",
                                    }}
                                    poster={currentEpisodeDetails?.image || media.coverImage?.extraLarge || ""}
                                    aspectRatio="16/9"
                                    onProviderChange={onProviderChange}
                                    onProviderSetup={onProviderSetup}
                                    // className="max-h-[75dvh] aspect-video"
                                    onTimeUpdate={(e) => {
                                        if (aniSkipData?.op && e?.currentTime && e?.currentTime >= aniSkipData.op.interval.startTime && e?.currentTime <= aniSkipData.op.interval.endTime) {
                                            setShowSkipIntroButton(true)
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
                                        } else {
                                            setShowSkipEndingButton(false)
                                        }

                                        if (currentEpisodeNumber > 0 &&
                                            duration > 0 &&
                                            e?.currentTime &&
                                            ((e.currentTime / duration)) >= 0.8 &&
                                            currentEpisodeNumber > currentProgress
                                        ) {
                                            if (!progressItem) {
                                                setProgressItem({
                                                    episodeNumber: currentEpisodeNumber,
                                                })
                                            }
                                        }
                                    }}
                                    onEnded={(e) => {
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
                                                <Button intent="white" onClick={() => seekTo(aniSkipData?.op?.interval?.endTime || 0)}>Skip
                                                                                                                                       intro</Button>
                                            )}
                                        </div>
                                        <div>
                                            {(showSkipEndingButton) && (
                                                <Button intent="white" onClick={() => seekTo(aniSkipData?.ed?.interval?.endTime || 0)}>Skip
                                                                                                                                       ending</Button>
                                            )}
                                        </div>
                                    </div>
                                    <DefaultVideoLayout
                                        icons={defaultLayoutIcons}
                                        slots={{
                                            settingsMenu: (
                                                <OnlinestreamSettingsButton />
                                            ),
                                            beforeCaptionButton: (
                                                <div className="flex items-center">
                                                    <OnlinestreamProviderButton />
                                                    <OnlinestreamServerButton />
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

                        {/*{currentEpisodeDetails && (*/}
                        {/*    <div className="space-y-4">*/}
                        {/*        <h3 className="line-clamp-1">{currentEpisodeDetails?.title?.replaceAll("`", "'")}</h3>*/}
                        {/*        {currentEpisodeDetails?.description && <p className="text-gray-400">*/}
                        {/*            {currentEpisodeDetails?.description?.replaceAll("`", "'")}*/}
                        {/*        </p>}*/}
                        {/*    </div>*/}
                        {/*)}*/}

                        {/*<div className="flex gap-4 lg:gap-5">*/}

                        {/*    {media.coverImage?.large && <div*/}
                        {/*        className="flex-none w-[200px] h-[270px] relative rounded-md overflow-hidden bg-[--background] shadow-md border block"*/}
                        {/*    >*/}
                        {/*        <Image*/}
                        {/*            src={media.coverImage.large}*/}
                        {/*            alt="cover image"*/}
                        {/*            fill*/}
                        {/*            priority*/}
                        {/*            className="object-cover object-center"*/}
                        {/*        />*/}
                        {/*    </div>}*/}


                        {/*    <div className="space-y-2">*/}
                        {/*        /!*TITLE*!/*/}
                        {/*        <div className="space-y-2">*/}
                        {/*            <p*/}
                        {/*                className="[text-shadow:_0_1px_10px_rgb(0_0_0_/_20%)] line-clamp-1 font-bold text-pretty text-xl lg:text-3xl"*/}
                        {/*                children={media.title?.userPreferred || ""}*/}
                        {/*            />*/}
                        {/*            {media.title?.userPreferred?.toLowerCase() !== media.title?.english?.toLowerCase() &&*/}
                        {/*                <p className="text-gray-400 line-clamp-2">{media.title?.english}</p>}*/}
                        {/*            {media.title?.userPreferred?.toLowerCase() !== media.title?.romaji?.toLowerCase() &&*/}
                        {/*                <p className="text-gray-400 line-clamp-2">{media.title?.romaji}</p>}*/}
                        {/*        </div>*/}

                        {/*        /!*SEASON*!/*/}
                        {/*        {!!media.season ? (*/}
                        {/*                <div>*/}
                        {/*                    <p className="text-lg text-gray-200 flex w-full gap-1 items-center">*/}
                        {/*                        <BiCalendarAlt /> {new Intl.DateTimeFormat("en-US", {*/}
                        {/*                        year: "numeric",*/}
                        {/*                        month: "short",*/}
                        {/*                    }).format(new Date(media.startDate?.year || 0,*/}
                        {/*                        media.startDate?.month || 0))} - {capitalize(media.season ?? "")}*/}
                        {/*                    </p>*/}
                        {/*                </div>*/}
                        {/*            ) :*/}
                        {/*            (*/}
                        {/*                <p className="text-lg text-gray-200 flex w-full gap-1 items-center">*/}

                        {/*                </p>*/}
                        {/*            )}*/}

                        {/*        /!*PROGRESS*!/*/}
                        {/*        <div className="flex gap-2 md:gap-4 items-center">*/}
                        {/*            <MediaPageHeaderScoreAndProgress*/}
                        {/*                score={mediaEntry?.listData?.score}*/}
                        {/*                progress={mediaEntry?.listData?.progress}*/}
                        {/*                episodes={media.episodes}*/}
                        {/*            />*/}
                        {/*            <AnilistMediaEntryModal listData={mediaEntry?.listData} media={media} />*/}
                        {/*            <p className="text-base md:text-lg">{capitalize(mediaEntry?.listData?.status === "CURRENT"*/}
                        {/*                ? "Watching"*/}
                        {/*                : mediaEntry?.listData?.status)}</p>*/}
                        {/*        </div>*/}

                        {/*        {mediaDetails &&*/}
                        {/*            <ScrollArea className="h-32 text-[--muted] hover:text-gray-300 transition-colors duration-500 text-sm pr-2">{mediaDetails?.description?.replace(*/}
                        {/*                /(<([^>]+)>)/ig,*/}
                        {/*                "")}</ScrollArea>}*/}
                        {/*    </div>*/}

                        {/*</div>*/}

                        {/*<p className="text-lg font-semibold block lg:hidden">*/}
                        {/*    Episodes*/}
                        {/*</p>*/}
                        {/*</div>*/}


                        <ScrollArea className="xl:max-w-[400px] 2xl:max-w-[500px] relative xl:sticky h-[75dvh] overflow-y-auto pr-4 pt-0">
                            <div className="space-y-4">
                                {(!episodes?.length && !loadPage) && <p>
                                    No episodes found
                                </p>}
                                {episodes?.filter(Boolean)?.sort((a, b) => a!.number - b!.number)?.map((episode, idx) => {
                                    return (
                                        <div
                                            key={idx + (episode.title || "") + episode.number}
                                            className={"block cursor-pointer"}
                                            id={`episode-${String(episode.number)}`}
                                            onClick={() => handleChangeEpisodeNumber(episode.number)}
                                        >
                                            <OnlinestreamEpisodeListItem
                                                title={media.format === "MOVIE" ? "Complete movie" : `Episode ${episode.number}`}
                                                episodeTitle={episode.title}
                                                // description={episode.description ?? undefined}
                                                image={episode.image}
                                                media={media}
                                                isSelected={episode.number === currentEpisodeNumber}
                                                disabled={episodeLoading}
                                                isWatched={progress ? episode.number <= progress : undefined}
                                            />
                                        </div>
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

            </div>

            <div
                className="h-[30rem] w-full flex-none object-cover object-center absolute -top-[5rem] overflow-hidden bg-[--background]"
            >
                <div
                    className="w-full absolute z-[2] top-0 h-[8rem] opacity-40 bg-gradient-to-b from-[--background] to-transparent via"
                />
                <div className="absolute w-full h-full">
                    {(!!media?.bannerImage || !!media?.coverImage?.extraLarge) && <Image
                        src={media?.bannerImage || media?.coverImage?.extraLarge || ""}
                        alt="banner image"
                        fill
                        quality={100}
                        priority
                        sizes="100vw"
                        className="object-cover object-center z-[1]"
                    />}
                </div>
                <div
                    className="w-full z-[3] absolute bottom-0 h-[32rem] bg-gradient-to-t from-[--background] via-[--background] via-50% to-transparent"
                />

            </div>
        </>
    )

}
