"use client"

import { useGetAnimeEntry, useUpdateAnimeEntryProgress } from "@/api/hooks/anime_entries.hooks"
import { EpisodeGridItem } from "@/app/(main)/_features/anime/_components/episode-grid-item"
import { MediaEpisodeInfoModal } from "@/app/(main)/_features/media/_components/media-episode-info-modal"
import { MediastreamPlaybackSubmenu } from "@/app/(main)/mediastream/_components/mediastream-video-addons"
import {
    __mediastream_currentProgressAtom,
    __mediastream_progressItemAtom,
    useHandleMediastream,
} from "@/app/(main)/mediastream/_lib/handle-mediastream"
import {
    __mediastream_autoPlayAtom,
    useMediastreamCurrentFile,
    useMediastreamJassubOffscreenRender,
} from "@/app/(main)/mediastream/_lib/mediastream.atoms"
import { useSkipData } from "@/app/(main)/onlinestream/_lib/skip"
import { LuffyError } from "@/components/shared/luffy-error"
import { Alert } from "@/components/ui/alert"
import { AppLayoutStack } from "@/components/ui/app-layout"
import { Button, IconButton } from "@/components/ui/button"
import { Checkbox } from "@/components/ui/checkbox"
import { cn } from "@/components/ui/core/styling"
import { LoadingSpinner } from "@/components/ui/loading-spinner"
import { Modal } from "@/components/ui/modal"
import { ScrollArea } from "@/components/ui/scroll-area"
import { Separator } from "@/components/ui/separator"
import { Skeleton } from "@/components/ui/skeleton"
import { MediaPlayer, MediaPlayerInstance, MediaProvider, Track } from "@vidstack/react"
import "@vidstack/react/player/styles/default/theme.css"
import "@vidstack/react/player/styles/default/layouts/video.css"
import { DefaultAudioLayout, defaultLayoutIcons, DefaultVideoLayout } from "@vidstack/react/player/layouts/default"
import { useAtomValue } from "jotai"
import { useAtom } from "jotai/react"
import { uniq } from "lodash"
import { CaptionsFileFormat } from "media-captions"
import Image from "next/image"
import Link from "next/link"
import { useRouter, useSearchParams } from "next/navigation"
import React from "react"
import { AiOutlineArrowLeft } from "react-icons/ai"
import "@vidstack/react/player/styles/base.css"
import { BiInfoCircle } from "react-icons/bi"


export default function Page() {

    const router = useRouter()
    const searchParams = useSearchParams()
    const mediaId = searchParams.get("id")
    const { data: mediaEntry, isLoading: mediaEntryLoading } = useGetAnimeEntry(mediaId)
    const playerRef = React.useRef<MediaPlayerInstance>(null)
    const { filePath } = useMediastreamCurrentFile()

    const mainEpisodes = React.useMemo(() => {
        return mediaEntry?.episodes?.filter(ep => ep.type === "main") ?? []
    }, [mediaEntry?.episodes])

    const specialEpisodes = React.useMemo(() => {
        return mediaEntry?.episodes?.filter(ep => ep.type === "special") ?? []
    }, [mediaEntry?.episodes])

    const ncEpisodes = React.useMemo(() => {
        return mediaEntry?.episodes?.filter(ep => ep.type === "nc") ?? []
    }, [mediaEntry?.episodes])

    const episodes = React.useMemo(() => {
        return [...mainEpisodes, ...specialEpisodes, ...ncEpisodes]
    }, [mainEpisodes, specialEpisodes, ncEpisodes])

    const {
        url,
        isError,
        isMediaContainerLoading,
        mediaContainer,
        subtitles,
        subtitleEndpointUri,
        onProviderChange,
        onProviderSetup,
        onTimeUpdate,
        onCanPlay,
        onEnded,
        onPlayFile,
        isCodecSupported,
        setStreamType,
        disabledAutoSwitchToDirectPlay,
    } = useHandleMediastream({ playerRef, episodes })

    const autoPlay = useAtomValue(__mediastream_autoPlayAtom)
    const { jassubOffscreenRender, setJassubOffscreenRender } = useMediastreamJassubOffscreenRender()

    /**
     * The episode number of the current file
     */
    const episodeNumber = React.useMemo(() => {
        return episodes.find(ep => !!ep.localFile?.path && ep.localFile?.path === filePath)?.episodeNumber || -1
    }, [episodes, filePath])

    /** AniSkip **/
    const { data: aniSkipData } = useSkipData(mediaEntry?.media?.idMal, episodeNumber)
    const [showSkipIntroButton, setShowSkipIntroButton] = React.useState(false)
    const [showSkipEndingButton, setShowSkipEndingButton] = React.useState(false)

    const seekTo = React.useCallback((time: number) => {
        Object.assign(playerRef.current ?? {}, { currentTime: time })
    }, [])

    /**
     * Progress update
     */
    const {
        mutate: updateProgress,
        isPending: isUpdatingProgress,
        isSuccess: hasUpdatedProgress,
    } = useUpdateAnimeEntryProgress(mediaId, episodeNumber)

    const [progressItem, setProgressItem] = useAtom(__mediastream_progressItemAtom)

    const [currentProgress, setCurrentProgress] = useAtom(__mediastream_currentProgressAtom)

    /**
     * Effect for when media entry changes
     * - Redirect if media entry is not found
     * - Reset current progress
     */
    React.useEffect(() => {
        if (!mediaId || (!mediaEntryLoading && !mediaEntry) || (!mediaEntryLoading && !!mediaEntry && !filePath)) {
            router.push("/")
        }
        if (mediaEntry) {
            setCurrentProgress(mediaEntry.listData?.progress ?? 0)
        }
    }, [mediaId, mediaEntry, mediaEntryLoading, filePath])

    /** Scroll to selected episode element when the episode list changes (on mount) **/
    const episodeListContainerRef = React.useRef<HTMLDivElement>(null)
    React.useLayoutEffect(() => {
        if (episodeListContainerRef.current) {
            React.startTransition(() => {
                const element = document.getElementById(`episode-${episodeNumber}`)
                if (element) {
                    element.scrollIntoView()
                    React.startTransition(() => {
                        window.scrollTo({ top: 0 })
                    })
                }
            })
        }
    }, [episodeListContainerRef.current, episodes, episodeNumber])

    const checkTimeRef = React.useRef<number>(0)

    if (mediaEntryLoading) return <div className="px-4 lg:px-8 space-y-4">
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
            <AppLayoutStack className="px-4 lg:px-8 z-[5]">

                <div className="flex flex-col lg:flex-row gap-2 w-full justify-between">
                    <div className="flex gap-4 items-center relative w-full">
                        <Link href={`/entry?id=${mediaEntry?.mediaId}`}>
                            <IconButton icon={<AiOutlineArrowLeft />} rounded intent="white-outline" size="md" />
                        </Link>
                        <h3 className="max-w-full lg:max-w-[50%] text-ellipsis truncate">{mediaEntry?.media?.title?.userPreferred}</h3>
                    </div>

                    <div className="flex gap-2 items-center">

                        {!!mediaContainer?.mediaInfo?.mimeCodec && (
                            <div className="">
                                <Modal
                                    title="Playback"
                                    trigger={
                                        <Button leftIcon={<BiInfoCircle />} className="rounded-full" intent="gray-outline">
                                            Playback info
                                        </Button>
                                    }
                                >
                                    <div className="space-y-2">
                                        <p className="line-clamp-1 text-[--muted]">
                                            {mediaContainer?.mediaInfo?.path}
                                        </p>
                                        {isCodecSupported(mediaContainer.mediaInfo.mimeCodec) ? <Alert
                                            intent="success"
                                            description="File video and audio codecs are compatible with this client"
                                        /> : <Alert
                                            intent="alert"
                                            description="File video and audio codecs are not compatible with this client"
                                        />}

                                        <p>
                                            <span className="font-bold">Video codec: </span>
                                            <span>{mediaContainer.mediaInfo.video?.mimeCodec}</span>
                                        </p>
                                        <p>
                                            <span className="font-bold">Audio codec: </span>
                                            <span>{uniq(mediaContainer.mediaInfo.audios?.map(n => n.mimeCodec)).join(", ")}</span>
                                        </p>

                                        <Modal
                                            title="Playback"
                                            trigger={
                                                <Button size="sm" className="rounded-full" intent="gray-outline">
                                                    More data
                                                </Button>
                                            }
                                            contentClass="max-w-3xl"
                                        >
                                           <pre className="overflow-x-auto overflow-y-auto max-h-[calc(100dvh-300px)] whitespace-pre-wrap p-2 rounded-md bg-gray-900">
                                                {JSON.stringify(mediaContainer, null, 2)}
                                           </pre>
                                        </Modal>


                                        <Separator />

                                        <p className="font-semibold text-lg">
                                            Jassub
                                        </p>

                                        <Checkbox
                                            label="Offscreen rendering"
                                            value={jassubOffscreenRender}
                                            onValueChange={v => setJassubOffscreenRender(v as boolean)}
                                            help="Enable this if you are experiencing performance issues"
                                        />

                                        <Separator />

                                        {(mediaContainer?.streamType === "direct") &&
                                            <div className="space-y-2">
                                                <Button
                                                    intent="alert-outline"
                                                    onClick={() => setStreamType("transcode")}
                                                    disabled={!disabledAutoSwitchToDirectPlay}
                                                >
                                                    Switch to transcoding
                                                </Button>
                                                {!disabledAutoSwitchToDirectPlay && <p className="text-[--muted]">
                                                    Disable 'auto switch to direct play' if you need to switch to transcoding
                                                </p>}
                                            </div>}

                                        {(mediaContainer?.streamType === "transcode" && isCodecSupported(mediaContainer.mediaInfo.mimeCodec)) &&
                                            <Button intent="alert-outline" onClick={() => setStreamType("direct")}>
                                                Switch to direct play
                                            </Button>}

                                    </div>
                                </Modal>
                            </div>
                        )}

                        {(!!progressItem && mediaEntry?.media && progressItem.episodeNumber > currentProgress) && <Button
                            className="animate-pulse"
                            loading={isUpdatingProgress}
                            disabled={hasUpdatedProgress}
                            onClick={() => {
                                updateProgress({
                                    episodeNumber: progressItem.episodeNumber,
                                    mediaId: mediaEntry.media!.id,
                                    totalEpisodes: mediaEntry.media!.episodes || 0,
                                    malId: mediaEntry.media!.idMal || undefined,
                                }, {
                                    onSuccess: () => setProgressItem(undefined),
                                })
                                setCurrentProgress(progressItem.episodeNumber)
                            }}
                        >Update progress</Button>}
                    </div>
                </div>

                <div
                    className={cn(
                        "flex gap-4 w-full flex-col 2xl:flex-row",
                    )}
                >

                    <div className="relative w-full">
                        <div
                            className={cn(
                                "aspect-video relative w-full self-start mx-auto",
                            )}
                        >
                            {isError ?
                                <LuffyError title="Playback Error" /> :
                                (!!url && !isMediaContainerLoading) ? <MediaPlayer
                                    key={mediaContainer?.filePath || ""}
                                    streamType="on-demand" // force VOD
                                    playsInline
                                    ref={playerRef}
                                    autoPlay={autoPlay}
                                    crossOrigin
                                    src={mediaContainer?.streamType === "direct" ? {
                                        src: url,
                                        type: mediaContainer?.mediaInfo?.extension === "mp4" ? "video/mp4" :
                                            mediaContainer?.mediaInfo?.extension === "avi" ? "video/x-msvideo" : "video/webm",
                                    } : url}
                                    aspectRatio="16/9"
                                    // poster={episodes?.find(n => n.localFile?.path === mediaContainer?.filePath)?.episodeMetadata?.image ||
                                    // mediaEntry?.media?.bannerImage || mediaEntry?.media?.coverImage?.extraLarge || ""}
                                    onProviderChange={onProviderChange}
                                    onProviderSetup={onProviderSetup}
                                    onTimeUpdate={e => {
                                        if (checkTimeRef.current < 200) {
                                            checkTimeRef.current++
                                            return
                                        }
                                        checkTimeRef.current = 0

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
                                        onTimeUpdate(e)
                                    }}
                                    onCanPlay={onCanPlay}
                                    onEnded={onEnded}
                                >
                                    <MediaProvider>
                                        {subtitles?.map((sub) => (
                                            <Track
                                                key={String(sub.index)}
                                                src={subtitleEndpointUri + sub.link}
                                                label={sub.title || sub.language}
                                                lang={sub.language}
                                                type={(sub.extension?.replace(".", "") || "ass") as CaptionsFileFormat}
                                                kind="subtitles"
                                                default={sub.isDefault || (!subtitles.some(n => n.isDefault) && sub.language?.startsWith("en"))}
                                            />
                                        ))}
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
                                            settingsMenuEndItems: <>
                                                <MediastreamPlaybackSubmenu />
                                            </>,
                                        }}
                                    />
                                    <DefaultAudioLayout
                                        icons={defaultLayoutIcons}
                                    />
                                </MediaPlayer> : (
                                    <Skeleton className="w-full h-full absolute flex justify-center items-center flex-col space-y-4">
                                        <LoadingSpinner
                                            containerClass=""
                                            spinner={<div className="w-16 h-16 lg:w-[100px] lg:h-[100px] relative">
                                                <Image
                                                    src="/logo_2.png"
                                                    alt="Loading..."
                                                    priority
                                                    fill
                                                    className="animate-pulse"
                                                />
                                            </div>}
                                        />
                                        <div className="text-center text-xs lg:text-sm">
                                            <p>
                                                Extracting video metadata...
                                            </p>
                                            <p>
                                                This might take a while.
                                            </p>
                                        </div>
                                    </Skeleton>
                                )}
                        </div>
                    </div>

                    <ScrollArea
                        ref={episodeListContainerRef}
                        className="2xl:max-w-[450px] w-full relative 2xl:sticky 2xl:h-[75dvh] overflow-y-auto 2xl:pr-4 pt-0"
                    >
                        <div className="space-y-4">
                            {episodes.map((episode) => (
                                <EpisodeGridItem
                                    key={episode.localFile?.path || ""}
                                    id={`episode-${String(episode.episodeNumber)}`}
                                    media={episode?.basicMedia as any}
                                    title={episode?.displayTitle || episode?.basicMedia?.title?.userPreferred || ""}
                                    image={episode?.episodeMetadata?.image || episode?.basicMedia?.coverImage?.large}
                                    episodeTitle={episode?.episodeTitle}
                                    fileName={episode?.localFile?.parsedInfo?.original}
                                    onClick={() => {
                                        if (episode.localFile?.path) {
                                            onPlayFile(episode.localFile?.path || "")
                                        }
                                    }}
                                    // description={episode?.absoluteEpisodeNumber !== episodeNumber
                                    //     ? `(Episode ${episode?.absoluteEpisodeNumber})`
                                    //     : undefined}
                                    isWatched={!!currentProgress && currentProgress >= episode?.progressNumber}
                                    isFiller={episode.episodeMetadata?.isFiller}
                                    isSelected={episode.localFile?.path === filePath}
                                    imageContainerClassName="w-20 h-20"
                                    className="flex-none w-full"
                                    action={<>
                                        <MediaEpisodeInfoModal
                                            title={episode.displayTitle}
                                            image={episode.episodeMetadata?.image}
                                            episodeTitle={episode.episodeTitle}
                                            airDate={episode.episodeMetadata?.airDate}
                                            length={episode.episodeMetadata?.length}
                                            summary={episode.episodeMetadata?.summary}
                                            isInvalid={episode.isInvalid}
                                        />
                                    </>}
                                />
                            ))}
                            <div className="hidden 2xl:block h-[1rem]">

                            </div>
                        </div>
                        <div
                            className={"hidden 2xl:block z-[5] absolute bottom-0 w-full h-[2rem] bg-gradient-to-t from-[--background] to-transparent"}
                        />
                    </ScrollArea>

                </div>


            </AppLayoutStack>
            <div
                className="h-[30rem] w-full flex-none object-cover object-center absolute -top-[5rem] overflow-hidden bg-[--background]"
            >
                <div
                    className="w-full absolute z-[2] top-0 h-[8rem] opacity-40 bg-gradient-to-b from-[--background] to-transparent via"
                />
                <div className="absolute w-full h-full">
                    {(!!mediaEntry?.media?.bannerImage || !!mediaEntry?.media?.coverImage?.extraLarge) && <Image
                        src={mediaEntry?.media?.bannerImage || mediaEntry?.media?.coverImage?.extraLarge || ""}
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
