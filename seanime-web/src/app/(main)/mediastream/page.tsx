"use client"
import { useGetAnimeEntry } from "@/api/hooks/anime_entries.hooks"
import { EpisodeGridItem } from "@/app/(main)/_features/anime/_components/episode-grid-item"
import { MediaEntryPageSmallBanner } from "@/app/(main)/_features/media/_components/media-entry-page-small-banner"
import { MediaEpisodeInfoModal } from "@/app/(main)/_features/media/_components/media-episode-info-modal"
import { SeaMediaPlayer } from "@/app/(main)/_features/sea-media-player/sea-media-player"
import { SeaMediaPlayerLayout } from "@/app/(main)/_features/sea-media-player/sea-media-player-layout"
import { useServerStatus } from "@/app/(main)/_hooks/use-server-status"
import { useHandleMediastream } from "@/app/(main)/mediastream/_lib/handle-mediastream"
import { useMediastreamCurrentFile, useMediastreamJassubOffscreenRender } from "@/app/(main)/mediastream/_lib/mediastream.atoms"
import { Alert } from "@/components/ui/alert"
import { AppLayoutStack } from "@/components/ui/app-layout"
import { Button } from "@/components/ui/button"
import { Checkbox } from "@/components/ui/checkbox"
import { Modal } from "@/components/ui/modal"
import { Separator } from "@/components/ui/separator"
import { Skeleton } from "@/components/ui/skeleton"
import { MediaPlayerInstance } from "@vidstack/react"
import "@/app/vidstack-theme.css"
import "@vidstack/react/player/styles/default/layouts/video.css"
import { uniq } from "lodash"
import { CaptionsFileFormat } from "media-captions"
import { useRouter, useSearchParams } from "next/navigation"
import React from "react"
import "@vidstack/react/player/styles/base.css"
import { BiInfoCircle } from "react-icons/bi"
import { SeaMediaPlayerProvider } from "../_features/sea-media-player/sea-media-player-provider"


export default function Page() {
    const serverStatus = useServerStatus()
    const router = useRouter()
    const searchParams = useSearchParams()
    const mediaId = searchParams.get("id")
    const { data: animeEntry, isLoading: animeEntryLoading } = useGetAnimeEntry(mediaId)
    const playerRef = React.useRef<MediaPlayerInstance>(null)
    const { filePath } = useMediastreamCurrentFile()

    const mainEpisodes = React.useMemo(() => {
        return animeEntry?.episodes?.filter(ep => ep.type === "main") ?? []
    }, [animeEntry?.episodes])

    const specialEpisodes = React.useMemo(() => {
        return animeEntry?.episodes?.filter(ep => ep.type === "special") ?? []
    }, [animeEntry?.episodes])

    const ncEpisodes = React.useMemo(() => {
        return animeEntry?.episodes?.filter(ep => ep.type === "nc") ?? []
    }, [animeEntry?.episodes])

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
        onCanPlay,
        playNextEpisode,
        onPlayFile,
        isCodecSupported,
        setStreamType,
        disabledAutoSwitchToDirectPlay,
        handleUpdateWatchHistory,
        episode,
        duration,
    } = useHandleMediastream({ playerRef, episodes, mediaId })

    const { jassubOffscreenRender, setJassubOffscreenRender } = useMediastreamJassubOffscreenRender()

    /**
     * The episode number of the current file
     */
    const episodeNumber = React.useMemo(() => {
        return episodes.find(ep => !!ep.localFile?.path && ep.localFile?.path === filePath)?.episodeNumber || -1
    }, [episodes, filePath])

    const progress = animeEntry?.listData?.progress

    /**
     * Effect for when media entry changes
     * - Redirect if media entry is not found
     * - Reset current progress
     */
    React.useEffect(() => {
        if (!mediaId || (!animeEntryLoading && !animeEntry) || (!animeEntryLoading && !!animeEntry && !filePath)) {
            router.push("/")
        }
    }, [mediaId, animeEntry, animeEntryLoading, filePath])

    if (animeEntryLoading || !animeEntry?.media) return <div className="px-4 lg:px-8 space-y-4">
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
        <SeaMediaPlayerProvider
            media={animeEntry?.media}
            progress={{
                currentProgress: progress ?? 0,
                currentEpisodeNumber: episodeNumber === -1 ? null : episodeNumber,
            }}
        >
            <AppLayoutStack className="p-4 lg:p-8 z-[5]">
                <SeaMediaPlayerLayout
                    mediaId={mediaId ? Number(mediaId) : undefined}
                    title={animeEntry?.media?.title?.userPreferred}
                    episodes={episodes}
                    rightHeaderActions={<>
                        <div className="">
                            <Modal
                                title="Playback"
                                trigger={
                                    <Button leftIcon={<BiInfoCircle />} className="rounded-full" intent="gray-basic" size="sm">
                                        Playback info
                                    </Button>
                                }
                                contentClass="sm:rounded-3xl"
                            >
                                <div className="space-y-2">
                                    <p className="tracking-wide text-sm text-[--muted] break-all">
                                        {mediaContainer?.mediaInfo?.path}
                                    </p>
                                    {isCodecSupported(mediaContainer?.mediaInfo?.mimeCodec || "") ? <Alert
                                        intent="success"
                                        description="File video and audio codecs are compatible with this client. Direct play is recommended."
                                    /> : <Alert
                                        intent="warning"
                                        description="File video and audio codecs are not compatible with this client. Transcoding is needed."
                                    />}

                                    <p>
                                        <span className="font-bold">Video codec: </span>
                                        <span>{mediaContainer?.mediaInfo?.video?.mimeCodec}</span>
                                    </p>
                                    <p>
                                        <span className="font-bold">Audio codec: </span>
                                        <span>{uniq(mediaContainer?.mediaInfo?.audios?.map(n => n.mimeCodec)).join(", ")}</span>
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
                                        <pre className="overflow-x-auto overflow-y-auto max-h-[calc(100dvh-300px)] whitespace-pre-wrap p-2 rounded-[--radius-md] bg-gray-900">
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
                                                intent="primary-subtle"
                                                onClick={() => setStreamType("transcode")}
                                                disabled={!disabledAutoSwitchToDirectPlay}
                                            >
                                                Switch to transcoding
                                            </Button>
                                            {!disabledAutoSwitchToDirectPlay && <p className="text-[--muted] text-sm italic opacity-50">
                                                Enable 'prefer transcoding' in the media streaming settings if you want to switch to transcoding
                                            </p>}
                                        </div>}

                                    {(mediaContainer?.streamType === "transcode" && isCodecSupported(mediaContainer?.mediaInfo?.mimeCodec || "")) &&
                                        <Button intent="success-subtle" onClick={() => setStreamType("direct")}>
                                            Switch to direct play
                                        </Button>}
                                </div>
                            </Modal>
                        </div>

                        {/* {(!!progressItem && animeEntry?.media && progressItem.episodeNumber > currentProgress) && <Button
                         className="animate-pulse"
                         loading={isUpdatingProgress}
                         disabled={hasUpdatedProgress}
                         onClick={() => {
                         updateProgress({
                         episodeNumber: progressItem.episodeNumber,
                         mediaId: animeEntry.media!.id,
                         totalEpisodes: animeEntry.media!.episodes || 0,
                         malId: animeEntry.media!.idMal || undefined,
                         }, {
                         onSuccess: () => setProgressItem(undefined),
                         })
                         setCurrentProgress(progressItem.episodeNumber)
                         }}
                         >Update progress</Button>} */}
                    </>}
                    mediaPlayer={
                        <SeaMediaPlayer
                            url={mediaContainer?.streamType === "direct" ? {
                                src: url || "",
                                type: mediaContainer?.mediaInfo?.extension === "mp4" ? "video/mp4" :
                                    mediaContainer?.mediaInfo?.extension === "avi" ? "video/x-msvideo" : "video/webm",
                            } : url}
                            isPlaybackError={isError}
                            isLoading={isMediaContainerLoading}
                            playerRef={playerRef}
                            poster={episodes?.find(n => n.localFile?.path === mediaContainer?.filePath)?.episodeMetadata?.image ||
                                animeEntry?.media?.bannerImage || animeEntry?.media?.coverImage?.extraLarge}
                            onProviderChange={onProviderChange}
                            onProviderSetup={onProviderSetup}
                            onCanPlay={onCanPlay}
                            onGoToNextEpisode={playNextEpisode}
                            tracks={subtitles?.map((sub) => ({
                                src: subtitleEndpointUri + sub.link,
                                label: sub.title || sub.language,
                                lang: sub.language,
                                type: (sub.extension?.replace(".", "") || "ass") as CaptionsFileFormat,
                                kind: "subtitles",
                                default: sub.isDefault || (!subtitles.some(n => n.isDefault) && sub.language?.startsWith("en")),
                            }))}
                            mediaInfoDuration={mediaContainer?.mediaInfo?.duration}
                            loadingText={<>
                                <p>Extracting video metadata...</p>
                                <p>This might take a while.</p>
                            </>}
                        />
                    }
                    episodeList={episodes.map((episode) => (
                        <EpisodeGridItem
                            key={episode.localFile?.path || ""}
                            id={`episode-${String(episode.episodeNumber)}`}
                            media={episode?.baseAnime as any}
                            title={episode?.displayTitle || episode?.baseAnime?.title?.userPreferred || ""}
                            image={episode?.episodeMetadata?.image || episode?.baseAnime?.coverImage?.large}
                            episodeTitle={episode?.episodeTitle}
                            fileName={episode?.localFile?.parsedInfo?.original}
                            onClick={() => {
                                if (episode.localFile?.path) {
                                    onPlayFile(episode.localFile?.path || "")
                                }
                            }}
                            isWatched={!!progress && progress >= episode?.progressNumber}
                            isFiller={episode.episodeMetadata?.isFiller}
                            isSelected={episode.localFile?.path === filePath}
                            length={episode.episodeMetadata?.length}
                            className="flex-none w-full"
                            episodeNumber={episode.episodeNumber}
                            progressNumber={episode.progressNumber}
                            action={<>
                                <MediaEpisodeInfoModal
                                    title={episode.displayTitle}
                                    image={episode.episodeMetadata?.image}
                                    episodeTitle={episode.episodeTitle}
                                    airDate={episode.episodeMetadata?.airDate}
                                    length={episode.episodeMetadata?.length}
                                    summary={episode.episodeMetadata?.summary || episode.episodeMetadata?.overview}
                                    isInvalid={episode.isInvalid}
                                    filename={episode.localFile?.parsedInfo?.original}
                                />
                            </>}
                        />
                    ))}
                />
            </AppLayoutStack>

            <MediaEntryPageSmallBanner bannerImage={animeEntry?.media?.bannerImage || animeEntry?.media?.coverImage?.extraLarge} />
        </SeaMediaPlayerProvider>
    )
}
