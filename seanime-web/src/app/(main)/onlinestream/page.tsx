"use client"
import "@vidstack/react/player/styles/default/theme.css"
import "@vidstack/react/player/styles/default/layouts/video.css"
import { ScoreProgressBadges } from "@/app/(main)/entry/_containers/meta-section/_components/score-progress-badges"
import { useMediaDetails, useMediaEntry } from "@/app/(main)/entry/_lib/media-entry"
import { OnlinestreamEpisodeListItem } from "@/app/(main)/onlinestream/_components/onlinestream-episode-list-item"
import {
    OnlinestreamProviderButton,
    OnlinestreamServerButton,
    OnlinestreamSettingsButton,
} from "@/app/(main)/onlinestream/_components/onlinestream-video-addons"
import { __onlinestream_selectedServerAtom } from "@/app/(main)/onlinestream/_lib/episodes"
import { OnlinestreamManagerProvider, useOnlinestreamManager } from "@/app/(main)/onlinestream/_lib/onlinestream-manager"
import { useSkipData } from "@/app/(main)/onlinestream/_lib/skip"
import { AnilistMediaEntryModal } from "@/components/shared/anilist-media-entry-modal"
import { PageWrapper } from "@/components/shared/styling/page-wrapper"
import { Button, IconButton } from "@/components/ui/button"
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
    Poster,
    Track,
} from "@vidstack/react"
import { defaultLayoutIcons, DefaultVideoLayout } from "@vidstack/react/player/layouts/default"
import HLS from "hls.js"
import { useAtom } from "jotai/react"
import capitalize from "lodash/capitalize"
import Image from "next/image"
import Link from "next/link"
import { useSearchParams } from "next/navigation"
import React from "react"
import { AiOutlineArrowLeft } from "react-icons/ai"
import { BiCalendarAlt } from "react-icons/bi"

export default function Page() {

    const searchParams = useSearchParams()
    const mediaId = searchParams.get("id")
    const { mediaEntry, mediaEntryLoading } = useMediaEntry(mediaId)
    const { mediaDetails } = useMediaDetails(mediaId)

    const ref = React.useRef<MediaPlayerInstance>(null)

    const [selectedServer, setSelectedServer] = useAtom(__onlinestream_selectedServerAtom)

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
        episodeNumber,
        handleChangeEpisodeNumber,
        episodeLoading,
    } = useOnlinestreamManager({
        mediaId,
        ref,
    })

    /** AniSkip **/
    const { data: aniSkipData } = useSkipData(media?.idMal, episodeNumber)

    const [showSkipIntroButton, setShowSkipIntroButton] = React.useState(false)
    const [showSkipEndingButton, setShowSkipEndingButton] = React.useState(false)

    const seekTo = React.useCallback((time: number) => {
        Object.assign(ref.current ?? {}, { currentTime: time })
    }, [])

    React.useEffect(() => {
        console.log(aniSkipData)
    }, [aniSkipData])

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


    if (!loadPage || !episodes || mediaEntryLoading) return <div className="p-4 sm:p-8 space-y-4">
        <div className="flex gap-4 items-center relative">
            <Skeleton className="h-16" />
        </div>
        <div
            className="grid xl:grid-cols-[1fr,500px] gap-4 xl:gap-4"
        >
            <div className="aspect-video relative">
                <Skeleton className="h-full w-full absolute" />
            </div>

            <Skeleton className="hidden lg:block relative h-[75dvh] overflow-y-auto pr-4 pt-0" />

        </div>
    </div>

    return (
        <>
            <PageWrapper className="p-4 sm:p-8 space-y-4">
                <OnlinestreamManagerProvider
                    opts={opts}
                >

                    <div className="flex gap-4 items-center relative">
                        <Link href={`/entry?id=${media?.id}`}>
                            <IconButton icon={<AiOutlineArrowLeft />} rounded intent="white-outline" size="md" />
                        </Link>
                        <h3>{media.title?.userPreferred}</h3>
                    </div>
                    <div
                        className="grid xl:grid-cols-[1fr,500px] gap-4 xl:gap-4"
                    >
                        <div className="space-y-4">
                            <div className="aspect-video relative">
                                {!!url ? <MediaPlayer
                                    ref={ref}
                                    crossOrigin="anonymous"
                                    src={{
                                        src: url || "",
                                        type: "application/x-mpegurl",
                                    }}
                                    onProviderChange={onProviderChange}
                                    onProviderSetup={onProviderSetup}
                                    className="w-full h-full absolute"
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
                                    }}
                                >
                                    <MediaProvider>
                                        <Poster
                                            src={currentEpisodeDetails?.image || media.coverImage?.extraLarge || ""}
                                            alt="Episode"
                                        />
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
                                    <Skeleton className="h-full w-full absolute">
                                        <LoadingSpinner containerClass="h-full absolute" />
                                    </Skeleton>
                                )}
                            </div>

                            {currentEpisodeDetails && (
                                <div className="space-y-4">
                                    <h3 className="line-clamp-1">{currentEpisodeDetails?.title}</h3>
                                    {currentEpisodeDetails?.description && <p className="text-gray-400">
                                        {currentEpisodeDetails?.description}
                                    </p>}
                                </div>
                            )}

                            <div className="flex gap-4 lg:gap-5">

                                {media.coverImage?.large && <div
                                    className="flex-none w-[200px] h-[270px] relative rounded-md overflow-hidden bg-[--background] shadow-md border block"
                                >
                                    <Image
                                        src={media.coverImage.large}
                                        alt="cover image"
                                        fill
                                        priority
                                        className="object-cover object-center"
                                    />
                                </div>}


                                <div className="space-y-2">
                                    {/*TITLE*/}
                                    <div className="space-y-2">
                                        <p
                                            className="[text-shadow:_0_1px_10px_rgb(0_0_0_/_20%)] line-clamp-1 font-bold text-pretty text-xl lg:text-3xl"
                                            children={media.title?.userPreferred || ""}
                                        />
                                        {media.title?.userPreferred?.toLowerCase() !== media.title?.english?.toLowerCase() &&
                                            <p className="text-gray-400 line-clamp-2">{media.title?.english}</p>}
                                        {media.title?.userPreferred?.toLowerCase() !== media.title?.romaji?.toLowerCase() &&
                                            <p className="text-gray-400 line-clamp-2">{media.title?.romaji}</p>}
                                    </div>

                                    {/*SEASON*/}
                                    {!!media.season ? (
                                            <div>
                                                <p className="text-lg text-gray-200 flex w-full gap-1 items-center">
                                                    <BiCalendarAlt /> {new Intl.DateTimeFormat("en-US", {
                                                    year: "numeric",
                                                    month: "short",
                                                }).format(new Date(media.startDate?.year || 0,
                                                    media.startDate?.month || 0))} - {capitalize(media.season ?? "")}
                                                </p>
                                            </div>
                                        ) :
                                        (
                                            <p className="text-lg text-gray-200 flex w-full gap-1 items-center">

                                            </p>
                                        )}

                                    {/*PROGRESS*/}
                                    <div className="flex gap-2 md:gap-4 items-center">
                                        <ScoreProgressBadges
                                            score={mediaEntry?.listData?.score}
                                            progress={mediaEntry?.listData?.progress}
                                            episodes={media.episodes}
                                        />
                                        <AnilistMediaEntryModal listData={mediaEntry?.listData} media={media} />
                                        <p className="text-base md:text-lg">{capitalize(mediaEntry?.listData?.status === "CURRENT"
                                            ? "Watching"
                                            : mediaEntry?.listData?.status)}</p>
                                    </div>

                                    {mediaDetails &&
                                        <ScrollArea className="h-32 text-[--muted] hover:text-gray-300 transition-colors duration-500 text-sm pr-2">{mediaDetails?.description?.replace(
                                            /(<([^>]+)>)/ig,
                                            "")}</ScrollArea>}
                                </div>

                            </div>

                            <p className="text-lg font-semibold block lg:hidden">
                                Episodes
                            </p>
                        </div>


                        <ScrollArea className="relative xl:sticky h-[75dvh] overflow-y-auto pr-4 pt-0">
                            <div className="space-y-4">
                                {episodes?.map((episode, idx) => {
                                    return (
                                        <div
                                            key={idx + (episode.title || "") + episode.number}
                                            className={"block cursor-pointer"}
                                            id={`episode-${String(episode.number)}`}
                                            onClick={() => handleChangeEpisodeNumber(episode.number)}
                                        >
                                            <OnlinestreamEpisodeListItem
                                                title={media.format === "MOVIE" ? "Complete movie" : `Episode ${episode.number}`}
                                                episodeTitle={`${episode.title}`}
                                                description={episode.description ?? undefined}
                                                image={episode.image}
                                                media={media}
                                                isSelected={episode.number === episodeNumber}
                                                disabled={episodeLoading}
                                                // isWatched={progress ? episode.number <= progress : undefined}
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

            </PageWrapper>

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
