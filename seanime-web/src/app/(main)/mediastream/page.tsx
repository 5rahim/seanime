import { getServerBaseUrl } from "@/api/client/server-url"
import { Anime_Episode, Mediastream_StreamType } from "@/api/generated/types"
import { Mediastream_MediaContainer, Models_MediastreamSettings } from "@/api/generated/types"
import { useGetAnimeEntry } from "@/api/hooks/anime_entries.hooks"
import { useGetMediastreamSettings, useMediastreamShutdownTranscodeStream, useRequestMediastreamMediaContainer } from "@/api/hooks/mediastream.hooks"
import { EpisodeGridItem } from "@/app/(main)/_features/anime/_components/episode-grid-item"
import { MediaEpisodeInfoModal } from "@/app/(main)/_features/media/_components/media-episode-info-modal"
import { EpisodePillsGrid } from "@/app/(main)/_features/video-core/_components/episode-pills-grid.tsx"
import { useIsCodecSupported } from "@/app/(main)/_features/video-core/_lib/hooks.ts"
import { VideoCore, VideoCoreProvider } from "@/app/(main)/_features/video-core/video-core"
import {
    VideoCoreInlineHelpers,
    VideoCoreInlineHelperUpdateProgressButton,
    VideoCoreInlineLayout,
} from "@/app/(main)/_features/video-core/video-core-inline-helpers"
import { VideoCoreLifecycleState } from "@/app/(main)/_features/video-core/video-core.atoms"
import { VideoCore_VideoSubtitleTrack } from "@/app/(main)/_features/video-core/video-core.atoms"
import { useWebsocketMessageListener } from "@/app/(main)/_hooks/handle-websockets"
import { useServerStatus } from "@/app/(main)/_hooks/use-server-status"
import { useMediastreamCurrentFile } from "@/app/(main)/mediastream/_lib/mediastream.atoms"
import { clientIdAtom } from "@/app/websocket-provider"
import { LuffyError } from "@/components/shared/luffy-error"
import { PageWrapper } from "@/components/shared/page-wrapper.tsx"

import { Alert } from "@/components/ui/alert"
import { Button, IconButton } from "@/components/ui/button"
import { Modal } from "@/components/ui/modal"
import { Separator } from "@/components/ui/separator"
import { Skeleton } from "@/components/ui/skeleton"
import { logger, useLatestFunction } from "@/lib/helpers/debug"
import { usePathname, useRouter, useSearchParams } from "@/lib/navigation.ts"
import { WSEvents } from "@/lib/server/ws-events"
import { useAtom, useAtomValue } from "jotai"
import { atomWithStorage } from "jotai/utils"
import { uniq } from "lodash"
import { AnimatePresence, motion } from "motion/react"
import React from "react"
import { BiInfoCircle } from "react-icons/bi"
import { BsFillGrid3X3GapFill } from "react-icons/bs"
import { toast } from "sonner"

const log = logger("MEDIASTREAM")

// Episode view mode atom
const __mediastream_episodeViewModeAtom = atomWithStorage<"list" | "grid">("sea-mediastream-episode-view-mode", "list")

export default function Page() {
    return (
        <PageWrapper className="px-4">
            <MediastreamPage />
        </PageWrapper>
    )
}

function uuidv4(): string {
    // @ts-ignore
    return ([1e7] + -1e3 + -4e3 + -8e3 + -1e11).replace(/[018]/g, (c) =>
        (c ^ (crypto.getRandomValues(new Uint8Array(1))[0] & (15 >> (c / 4)))).toString(16),
    )
}

function MediastreamPage() {
    const serverStatus = useServerStatus()
    const router = useRouter()
    const pathname = usePathname()
    const searchParams = useSearchParams()
    const mediaId = searchParams.get("id")
    const { data: animeEntry, isLoading: animeEntryLoading } = useGetAnimeEntry(mediaId)
    const { filePath, setFilePath } = useMediastreamCurrentFile()

    const { data: mediastreamSettings, isFetching: mediastreamSettingsLoading } = useGetMediastreamSettings(true)

    const [episodeViewMode, setEpisodeViewMode] = useAtom(__mediastream_episodeViewModeAtom)

    const media = animeEntry?.media
    const episodes = React.useMemo(() => {
        return animeEntry?.episodes ?? []
    }, [animeEntry?.episodes])

    // get current episode
    const currentEpisode = React.useMemo(() => {
        return episodes.find(ep => !!ep.localFile?.path && ep.localFile?.path === filePath)
    }, [episodes, filePath])

    const episodeNumber = currentEpisode?.episodeNumber ?? 1
    const progress = animeEntry?.listData?.progress ?? 0

    React.useEffect(() => {
        if (!pathname.startsWith("/mediastream")) return
        if (!mediaId || (!animeEntryLoading && !animeEntry) || (!animeEntryLoading && !!animeEntry && !filePath)) {
            router.push("/")
        }
    }, [mediaId, pathname, animeEntry, animeEntryLoading, filePath])

    const sessionId = useAtomValue(clientIdAtom)
    const clientId = React.useMemo(() => sessionId ?? uuidv4(), [sessionId])

    // stream state
    const [streamType, setStreamType] = React.useState<Mediastream_StreamType>("transcode")
    const [url, setUrl] = React.useState<string | null>(null)
    const [playbackError, setPlaybackError] = React.useState<string | null>(null)


    const playerRef = React.useRef<HTMLVideoElement | null>(null)
    const prevUrlRef = React.useRef<string | undefined>(undefined)

    // get codec support
    const { isCodecSupported } = useIsCodecSupported()

    // request media container
    const {
        data: mediaContainer,
        isError: isMediaContainerError,
        isPending: isMediaContainerPending,
        refetch: refetchMediaContainer,
    } = useRequestMediastreamMediaContainer({
        path: filePath,
        streamType: streamType,
        clientId: clientId,
    }, !!mediastreamSettings && !mediastreamSettingsLoading && !!filePath)

    const { mutate: shutdownTranscode } = useMediastreamShutdownTranscodeStream()

    // handle stream url change
    const changeUrl = React.useCallback((newUrl: string | null) => {
        if (prevUrlRef.current !== newUrl) {
            setPlaybackError(null)
        }
        setUrl(newUrl)
        if (newUrl) {
            prevUrlRef.current = newUrl
        }
    }, [])

    // process media container
    React.useEffect(() => {
        if (isMediaContainerPending || !mediaContainer) {
            changeUrl(null)
            return
        }

        log.info("Media container changed", mediaContainer)

        const codecSupported = isCodecSupported(mediaContainer?.mediaInfo?.mimeCodec ?? "")
        log.info("Is codec supported?", codecSupported)

        // switch to direct play if supported
        if (mediaContainer.streamType === "transcode") {
            if (!codecSupported && mediastreamSettings?.directPlayOnly) {
                toast.warning("Codec not supported for direct play")
                changeUrl(null)
                return
            }

            if (codecSupported && (!mediastreamSettings?.disableAutoSwitchToDirectPlay || mediastreamSettings?.directPlayOnly)) {
                log.info("Switching to direct play")
                setStreamType("direct")
                changeUrl(null)
                return
            }
        }

        // switch to transcode if direct play not supported
        if (mediaContainer.streamType === "direct") {
            if (!codecSupported) {
                log.warning("Codec not supported for direct play, switching to transcode")
                setStreamType("transcode")
                changeUrl(null)
                return
            }
        }

        if (mediaContainer.streamUrl) {
            const _newUrl = `${getServerBaseUrl()}${mediaContainer.streamUrl}`
            log.info("Setting stream URL", _newUrl)
            changeUrl(_newUrl)
        } else {
            changeUrl(null)
        }

    }, [mediaContainer, isMediaContainerPending, mediastreamSettings, isCodecSupported])


    // handle fatal errors
    const onFatalError = React.useCallback((error: any) => {
        log.error("Fatal error", error)
        if (mediaContainer?.streamType === "transcode") {
            shutdownTranscode()
        }
        setPlaybackError("Playback error triggered. Please try again or switch stream type.")
        changeUrl(null) // reset url
        toast.error("Playback error occurred")
    }, [mediaContainer?.streamType])


    // listen for shutdown stream event
    useWebsocketMessageListener<string | null>({
        type: WSEvents.MEDIASTREAM_SHUTDOWN_STREAM,
        onMessage: msg => {
            if (msg) toast.error(msg)
            log.warning("Shutdown stream event received")
            changeUrl(null)
        },
    })


    // subtitles
    const subtitleTracks = React.useMemo<VideoCore_VideoSubtitleTrack[] | undefined>(() => {
        if (!mediaContainer?.mediaInfo?.subtitles) return undefined
        return mediaContainer.mediaInfo.subtitles.map((sub: any) => ({
            index: sub.index,
            label: sub.title || sub.language || `Track ${sub.index}`,
            language: sub.language || "eng",
            src: `${getServerBaseUrl()}/api/v1/mediastream/subs` + sub.link,
            content: undefined, // Content fetching handled by VideoCore if needed, but src is preferred
            type: sub.extension,
            default: sub.isDefault,
            useLibassRenderer: true,
        }))
    }, [mediaContainer?.mediaInfo?.subtitles])

    // navigation helpers
    const goToEpisode = useLatestFunction((ep: Anime_Episode) => {
        if (ep.localFile?.path) {
            setFilePath(ep.localFile.path)
            setStreamType("transcode") // reset to transcode if user prefers direct, the effect will switch it back
        }
    })

    const onPlayEpisode = (action: "next" | "previous") => {
        if (!currentEpisode) return
        const currentIndex = episodes.findIndex(e => e.localFile?.path === currentEpisode.localFile?.path)
        if (currentIndex === -1) return

        let targetEp = action === "next" ? episodes[currentIndex + 1] : episodes[currentIndex - 1]

        if (targetEp?.localFile?.path) {
            goToEpisode(targetEp)
        }
    }

    const state = React.useMemo(() => {
        return {
            active: true,
            playbackInfo: (url && filePath) ? {
                id: filePath,
                playbackType: "localfile",
                streamUrl: url,
                media: media!,
                episode: currentEpisode,
                localFile: currentEpisode?.localFile,
                streamType: mediaContainer?.streamType === "direct" && mediaContainer?.mediaInfo?.extension === "mkv"
                    ? "native"
                    : (mediaContainer?.streamType === "transcode" ? "hls" : "native"), // Simple heuristic
                subtitleTracks: subtitleTracks,
                libassFonts: mediaContainer?.mediaInfo?.fonts?.map(name => ({ src: `${getServerBaseUrl()}/api/v1/mediastream/att/${name}` })) || [],
                initialState: undefined,
            } : null,
            loadingState: !url ? "Loading stream..." : null,
            playbackError: playbackError,
        } satisfies VideoCoreLifecycleState
    }, [url, filePath, media, currentEpisode, mediaContainer, playbackError, subtitleTracks])

    if (animeEntryLoading || !animeEntry?.media) return <div className="px-4 lg:px-8 space-y-4">
        <div className="flex gap-4 items-center relative">
            <Skeleton className="h-12" />
        </div>
        <div className="grid 2xl:grid-cols-[1fr,450px] gap-4 xl:gap-4">
            <div className="w-full min-h-[70dvh] relative">
                <Skeleton className="h-full w-full absolute" />
            </div>
            <Skeleton className="hidden 2xl:block relative h-[78dvh] overflow-y-auto pr-4 pt-0" />
        </div>
    </div>

    return (
        <>
            <VideoCoreInlineHelpers
                playerRef={playerRef}
                currentEpisodeNumber={episodeNumber}
                currentProgress={progress}
                media={media!}
                url={url}
            />

            <VideoCoreInlineLayout
                mediaId={mediaId ? Number(mediaId) : undefined}
                currentEpisodeNumber={episodeNumber}
                title={media?.title?.userPreferred}
                episodes={episodes}
                loadingEpisodeList={animeEntryLoading}
                hideBackButton={false}
                rightHeaderActions={<>
                    <VideoCoreInlineHelperUpdateProgressButton />
                    <MediastreamPlaybackInfo
                        mediaContainer={mediaContainer}
                        isCodecSupported={isCodecSupported}
                        streamType={streamType}
                        setStreamType={setStreamType}
                        mediastreamSettings={mediastreamSettings}
                    />
                    <IconButton
                        size="sm"
                        intent={episodeViewMode === "list" ? "gray-basic" : "white-subtle"}
                        icon={<BsFillGrid3X3GapFill />}
                        onClick={() => setEpisodeViewMode(prev => prev === "list" ? "grid" : "list")}
                        title={episodeViewMode === "list" ? "Switch to grid view" : "Switch to list view"}
                    />
                </>}
                mediaPlayer={
                    <VideoCoreProvider id="mediastream" key={filePath}>
                        <div className="w-full aspect-video mx-auto border rounded-lg overflow-hidden bg-black relative z-20">
                            {isMediaContainerError || playbackError ? (
                                <div className="flex flex-col items-center justify-center h-full w-full">
                                    <LuffyError title="Playback Error">
                                        {playbackError || "Could not load media container."}
                                    </LuffyError>
                                    <button
                                        onClick={() => refetchMediaContainer()}
                                        className="mt-4 px-4 py-2 bg-gray-800 text-white rounded hover:bg-gray-700"
                                    >
                                        Retry
                                    </button>
                                </div>
                            ) : (
                                <VideoCore
                                    id="mediastream"
                                    mRef={playerRef}
                                    state={state}
                                    inline
                                    onError={onFatalError}
                                    onHlsFatalError={(e) => onFatalError(e)}
                                    onTerminateStream={() => {
                                        changeUrl(null)
                                        router.back() // or just stop?
                                    }}
                                    onPlayEpisode={onPlayEpisode}
                                />
                            )}
                        </div>
                    </VideoCoreProvider>
                }
                episodeList={<>
                    <AnimatePresence mode="wait" initial={false}>
                        {episodeViewMode === "list" ? (
                            <motion.div
                                key="list-view"
                                initial={{ opacity: 0, y: 20 }}
                                animate={{ opacity: 1, y: 0 }}
                                exit={{ opacity: 0, y: -20 }}
                                transition={{ duration: 0.3 }}
                                className="space-y-3"
                            >
                                {episodes?.map((episode, idx) => {
                                    const isSelected = episode.localFile?.path === filePath
                                    return (
                                        <EpisodeGridItem
                                            key={idx + (episode.episodeTitle || "") + episode.episodeNumber}
                                            media={media!}
                                            onClick={() => {
                                                if (episode.localFile?.path) {
                                                    setFilePath(episode.localFile.path)
                                                } else {
                                                    toast.error("File path not found for this episode")
                                                }
                                            }}
                                            title={media?.format === "MOVIE" ? "Complete movie" : `Episode ${episode.episodeNumber}`}
                                            episodeTitle={episode.episodeTitle}
                                            description={episode.episodeMetadata?.summary}
                                            image={episode.episodeMetadata?.image}
                                            isSelected={isSelected}
                                            isWatched={progress ? episode.episodeNumber <= progress : undefined}
                                            className="flex-none w-full"
                                            isFiller={episode.episodeMetadata?.isFiller}
                                            episodeNumber={episode.episodeNumber}
                                            progressNumber={episode.episodeNumber}
                                            action={<>
                                                <MediaEpisodeInfoModal
                                                    title={media?.format === "MOVIE" ? "Complete movie" : `Episode ${episode.episodeNumber}`}
                                                    image={episode.episodeMetadata?.image}
                                                    episodeTitle={episode.episodeTitle}
                                                    summary={episode.episodeMetadata?.summary}
                                                    filename={episode.localFile?.name}
                                                />
                                            </>}
                                        />
                                    )
                                })}
                                {!!episodes?.length && <p className="text-center text-[--muted] py-2">End</p>}
                            </motion.div>
                        ) : (
                            <EpisodePillsGrid
                                key="grid-view"
                                episodes={episodes?.map(ep => ({
                                    id: ep.localFile?.path || "",
                                    number: ep.episodeNumber,
                                    title: ep.episodeTitle,
                                    isFiller: ep.episodeMetadata?.isFiller,
                                })) || []}
                                currentEpisodeNumber={episodeNumber}
                                onEpisodeSelect={(num, id) => {
                                    const ep = episodes.find(e => e.localFile?.path === id)
                                    if (ep?.localFile?.path) {
                                        setFilePath(ep.localFile.path)
                                    }
                                }}
                                progress={progress}
                                getEpisodeId={(ep) => `episode-${ep.id}`}
                            />
                        )}
                    </AnimatePresence>
                </>}
            />
        </>
    )
}

type MediastreamPlaybackInfoProps = {
    mediaContainer: Mediastream_MediaContainer | undefined
    isCodecSupported: (codec: string) => boolean
    streamType: Mediastream_StreamType
    setStreamType: (type: Mediastream_StreamType) => void
    mediastreamSettings: Models_MediastreamSettings | undefined
}

function MediastreamPlaybackInfo({
    mediaContainer,
    isCodecSupported,
    streamType,
    setStreamType,
    mediastreamSettings,
}: MediastreamPlaybackInfoProps) {

    if (!mediaContainer) return null

    return (
        <Modal
            title="Playback"
            trigger={
                <Button leftIcon={<BiInfoCircle />} className="rounded-full" intent="gray-basic" size="sm">
                    Playback info
                </Button>
            }
            contentClass="sm:rounded-3xl"
        >
            <div className="space-y-4">
                <p className="tracking-wide text-sm text-[--muted] break-all">
                    {mediaContainer.mediaInfo?.path}
                </p>
                {isCodecSupported(mediaContainer.mediaInfo?.mimeCodec || "") ? <Alert
                    intent="success"
                    description="File video and audio codecs are compatible with this client. Direct play is recommended."
                /> : <Alert
                    intent="warning"
                    description="File video and audio codecs are not compatible with this client. Transcoding is needed."
                />}

                <div className="text-sm space-y-1">
                    <p>
                        <span className="font-bold">Stream type: </span>
                        <span className="uppercase">{streamType}</span>
                    </p>
                    <p>
                        <span className="font-bold">Video codec: </span>
                        <span>{mediaContainer.mediaInfo?.video?.mimeCodec}</span>
                    </p>
                    <p>
                        <span className="font-bold">Audio codec: </span>
                        <span>{uniq(mediaContainer.mediaInfo?.audios?.map(n => n.mimeCodec)).join(", ")}</span>
                    </p>
                </div>

                <Modal
                    title="Media Container Data"
                    trigger={
                        <Button size="sm" className="rounded-full" intent="gray-outline">
                            More data
                        </Button>
                    }
                    contentClass="max-w-3xl"
                >
                    <pre className="overflow-x-auto overflow-y-auto max-h-[calc(100dvh-300px)] whitespace-pre-wrap p-2 rounded-[--radius-md] bg-gray-900 text-xs text-white">
                        {JSON.stringify(mediaContainer, null, 2)}
                    </pre>
                </Modal>

                <Separator />

                {(streamType === "direct") &&
                    <div className="space-y-2">
                        <Button
                            intent="primary-subtle"
                            onClick={() => setStreamType("transcode")}
                            disabled={!mediastreamSettings?.disableAutoSwitchToDirectPlay}
                            className="w-full"
                        >
                            Switch to transcoding
                        </Button>
                        {!mediastreamSettings?.disableAutoSwitchToDirectPlay && <p className="text-[--muted] text-sm italic opacity-50">
                            Enable 'Prefer transcoding' in the media streaming settings if you want to switch to transcoding
                        </p>}
                    </div>}

                {(streamType === "transcode" && isCodecSupported(mediaContainer.mediaInfo?.mimeCodec || "")) &&
                    <Button
                        intent="success-subtle" onClick={() => setStreamType("direct")}
                        className="w-full"
                    >
                        Switch to direct play
                    </Button>}
            </div>
        </Modal>
    )
}
