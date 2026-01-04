import { getServerBaseUrl } from "@/api/client/server-url"
import { Anime_Entry } from "@/api/generated/types"
import { useGetOnlineStreamEpisodeList, useGetOnlineStreamEpisodeSource, useOnlineStreamEmptyCache } from "@/api/hooks/onlinestream.hooks"
import { serverStatusAtom } from "@/app/(main)/_atoms/server-status.atoms"
import { EpisodeGridItem } from "@/app/(main)/_features/anime/_components/episode-grid-item"
import { MediaEpisodeInfoModal } from "@/app/(main)/_features/media/_components/media-episode-info-modal"
import { useNakamaStatus, useNakamaWatchParty } from "@/app/(main)/_features/nakama/nakama-manager"
import { usePlaylistManager } from "@/app/(main)/_features/playlists/_containers/global-playlist-manager"
import { useSkipData } from "@/app/(main)/_features/sea-media-player/aniskip"
import { VideoCore, VideoCoreProvider } from "@/app/(main)/_features/video-core/video-core"
import { isHLSSrc, isNativeVideoExtension, isProbablyHls } from "@/app/(main)/_features/video-core/video-core-hls"
import {
    VideoCoreInlineHelpers,
    VideoCoreInlineHelperUpdateProgressButton,
    VideoCoreInlineLayout,
} from "@/app/(main)/_features/video-core/video-core-inline-helpers"
import { vc_useLibassRendererAtom, VideoCore_VideoPlaybackInfo, VideoCore_VideoSource } from "@/app/(main)/_features/video-core/video-core.atoms"
import { useServerHMACAuth } from "@/app/(main)/_hooks/use-server-status"
import { EpisodePillsGrid } from "@/app/(main)/onlinestream/_components/episode-pills-grid"
import { OnlinestreamManualMappingModal } from "@/app/(main)/onlinestream/_containers/onlinestream-manual-matching"
import { useNakamaOnlineStreamWatchParty } from "@/app/(main)/onlinestream/_lib/handle-onlinestream"
import { useHandleOnlinestreamProviderExtensions } from "@/app/(main)/onlinestream/_lib/handle-onlinestream-providers"
import {
    __onlinestream_qualityAtom,
    __onlinestream_selectedDubbedAtom,
    __onlinestream_selectedEpisodeNumberAtom,
    __onlinestream_selectedProviderAtom,
    __onlinestream_selectedServerAtom,
} from "@/app/(main)/onlinestream/_lib/onlinestream.atoms"
import { LuffyError } from "@/components/shared/luffy-error"
import { Button, IconButton } from "@/components/ui/button"
import { Modal, ModalProps } from "@/components/ui/modal"
import { Popover, PopoverProps } from "@/components/ui/popover"
import { Select } from "@/components/ui/select"
import { Skeleton } from "@/components/ui/skeleton"
import { logger, useLatestFunction } from "@/lib/helpers/debug"
import { useWindowSize } from "@uidotdev/usehooks"
import { AxiosError } from "axios"
import { useAtom, useAtomValue } from "jotai/react"
import { atomWithStorage } from "jotai/utils"
import { uniq, uniqBy } from "lodash"
import { AnimatePresence, motion } from "motion/react"
import { usePathname, useRouter, useSearchParams } from "next/navigation"
import React from "react"
import { BsFillGrid3X3GapFill } from "react-icons/bs"
import { CgMediaPodcast } from "react-icons/cg"
import { FaSearch } from "react-icons/fa"
import "@/app/vidstack-theme.css"
import "@vidstack/react/player/styles/default/layouts/video.css"
import { HiOutlineCog6Tooth } from "react-icons/hi2"
import { LuSpeech } from "react-icons/lu"
import { MdOutlineSubtitles } from "react-icons/md"
import { toast } from "sonner"
import { PluginEpisodeGridItemMenuItems } from "../../_features/plugin/actions/plugin-actions"

type OnlinestreamPageProps = {
    animeEntry?: Anime_Entry
    animeEntryLoading?: boolean
    hideBackButton?: boolean
}

const log = logger("ONLINESTREAM")

// Episode view mode atom
export const __onlineStream_episodeViewModeAtom = atomWithStorage<"list" | "grid">("sea-onlinestream-episode-view-mode", "list")

function isValidVideoSourceType(type: string | null | undefined) {
    if (!type) return false
    return ["unknown", "mp4", "m3u8"].includes(type)
}

export function OnlinestreamPage({ animeEntry, animeEntryLoading, hideBackButton }: OnlinestreamPageProps) {
    const serverStatus = useAtomValue(serverStatusAtom)
    const router = useRouter()
    const pathname = usePathname()
    const searchParams = useSearchParams()
    const urlMediaId = searchParams.get("id")
    const urlEpNumber = searchParams.get("episode")
    const media = animeEntry?.media!
    const mediaId = media?.id!
    const progress = animeEntry?.listData?.progress ?? 0

    const [episodeViewMode, setEpisodeViewMode] = useAtom(__onlineStream_episodeViewModeAtom)

    const playerRef = React.useRef<HTMLVideoElement | null>(null)

    const [currentEpisodeNumber, setSelectedEpisodeNumber] = useAtom(__onlinestream_selectedEpisodeNumberAtom)
    const [server, setServer] = useAtom(__onlinestream_selectedServerAtom)
    const [quality, setQuality] = useAtom(__onlinestream_qualityAtom)
    const [dubbed, setDubbed] = useAtom(__onlinestream_selectedDubbedAtom)
    const [provider, setProvider] = useAtom(__onlinestream_selectedProviderAtom)

    const [overrideStreamType, setOverrideStreamType] = React.useState<VideoCore_VideoPlaybackInfo["streamType"] | null>(null)

    const [playbackError, setPlaybackError] = React.useState<string | null>(null)

    const { mutate: emptyCache, isPending: isEmptyingCache } = useOnlineStreamEmptyCache()

    // get extensions
    const { providerExtensions, providerExtensionOptions } = useHandleOnlinestreamProviderExtensions()
    const extension = React.useMemo(() => providerExtensions.find(p => p.id === provider), [providerExtensions, provider])

    // Nakama Watch Party
    const nakamaStatus = useNakamaStatus()
    const { isPeer: isWatchPartyPeer } = useNakamaWatchParty()
    const { streamToLoad, onLoadedStream, removeParamsFromUrl, redirectToStream } = useNakamaOnlineStreamWatchParty()
    const isLoadingFromWatchPartyRef = React.useRef(false)


    // Stream URL
    const [url, setUrl] = React.useState<string | null>(null)

    React.useEffect(() => {
        return () => {
            setUrl(null)
        }
    }, [])

    React.useLayoutEffect(() => {
        if (!streamToLoad || !providerExtensionOptions?.length) return
        log.info("Watch party stream to load", { streamToLoad })
        if (streamToLoad.mediaId !== mediaId) {
            // redirectToStream(streamToLoad)
            return
        }

        // Check if we have the provider
        if (!providerExtensionOptions.some(p => p.value === streamToLoad.provider)) {
            log.warning("Provider not found in options", { providerExtensionOptions, provider: streamToLoad.provider })
            toast.error("Watch Party: The provider used by the host is not installed.")
            return
        }

        // Set flag to prevent other effects from overriding
        isLoadingFromWatchPartyRef.current = true

        setUrl(null)

        // Remove query params from the URL
        removeParamsFromUrl()

        setProvider(streamToLoad.provider)
        setDubbed(streamToLoad.dubbed)
        setServer(streamToLoad.server)
        setQuality(streamToLoad.quality)
        setSelectedEpisodeNumber(streamToLoad.episodeNumber)

        onLoadedStream()

        const t = setTimeout(() => {
            isLoadingFromWatchPartyRef.current = false
        }, 1000)
        return () => clearTimeout(t)
    }, [streamToLoad, providerExtensionOptions])

    // get the list of episodes from the provider
    const {
        data: episodeListResponse,
        isFetching: isFetchingEpisodeList,
        isLoading: isLoadingEpisodeList,
        isSuccess: isEpisodeListFetched,
        isError: isEpisodeListError,
    } = useGetOnlineStreamEpisodeList(mediaId, provider, dubbed)

    const episodes = episodeListResponse?.episodes
    const currentEpisode = episodes?.find(e => e.number === currentEpisodeNumber)

    // AniSkip
    const { data: aniSkipData } = useSkipData(media.idMal, currentEpisode?.number)

    // get the current episode source from the provider
    const {
        data: episodeSource,
        isLoading: isLoadingEpisodeSource,
        isFetching: isFetchingEpisodeSource,
        isError: isErrorEpisodeSource,
        error: errorEpisodeSource,
    } = useGetOnlineStreamEpisodeSource(
        mediaId,
        provider,
        currentEpisodeNumber,
        (!!extension?.supportsDub) && dubbed,
        !!mediaId && currentEpisodeNumber !== null && isEpisodeListFetched,
    )

    // de-duplicate video sources
    const videoSources = React.useMemo(() => uniqBy(episodeSource?.videoSources?.filter(n => n.server === server),
        n => `${n.url}|${n.quality}|${n.server}`), [episodeSource?.number, server])
    const hasMultipleVideoSources = React.useMemo(() => !!videoSources?.length && videoSources?.length > 1, [videoSources])

    // list of servers
    const servers = React.useMemo(() => {
        if (!episodeSource) {
            log.info("Updating servers, no episode source", [])
            return []
        }
        const servers = episodeSource?.videoSources?.map((source) => source.server)
        log.info("Updating servers", servers)
        return uniq(servers)
    }, [episodeSource?.videoSources])

    // If the sources don't have the stored server, set it to the first one
    React.useLayoutEffect(() => {
        if (!!servers?.length && (!server || !servers.includes(server))) {
            setServer(servers[0])
        }
    }, [servers, server])

    // get the video source from the episode source
    // devnote: use videoSources instead of episodeSource.videoSources
    const videoSource = React.useMemo(() => {
        if (!episodeSource || !videoSources) return undefined

        let filtered = [...videoSources]
        let qualitySatinized = quality
        qualitySatinized = qualitySatinized?.includes("p") ? qualitySatinized?.split("p")?.[0]?.toLowerCase() + "p" : qualitySatinized

        log.info("Selecting video source", { qualitySatinized, server })
        // If server is set, filter sources by server
        if (server && filtered.some(n => n.server === server)) {
            filtered = filtered.filter(s => s.server === server)
        }

        const hasPreferredQuality = qualitySatinized && filtered.some(n => n.quality.toLowerCase().includes(qualitySatinized!))
        const hasAuto = filtered.some(n => n.quality === "auto")

        log.info("Filtering video sources by quality", {
            hasAuto,
            hasPreferredQuality,
        })

        // If quality is set, filter sources by quality
        // Only filter by quality if the quality is present in the sources
        if (qualitySatinized && hasPreferredQuality) {
            filtered = filtered.filter(n => n.quality.toLowerCase().includes(qualitySatinized!))
        } else if (hasAuto) {
            filtered = filtered.filter(n => n.quality.toLowerCase() === "auto" || n.quality.toLowerCase().includes("default"))
        } else {
            log.info("Choosing a quality")
            if (filtered.some(n => n.quality.includes("1080p"))) {
                filtered = filtered.filter(n => n.quality.includes("1080p"))
            } else if (filtered.some(n => n.quality.includes("720p"))) {
                filtered = filtered.filter(n => n.quality.includes("720p"))
            } else if (filtered.some(n => n.quality.includes("480p"))) {
                filtered = filtered.filter(n => n.quality.includes("480p"))
            } else if (filtered.some(n => n.quality.includes("360p"))) {
                filtered = filtered.filter(n => n.quality.includes("360p"))
            }

            if (filtered.some(n => n.quality.includes("default"))) {
                filtered = filtered.filter(n => n.quality.includes("default"))
            }
        }

        log.info("Selected video source", filtered[0])

        return filtered[0]
    }, [episodeSource, videoSources, server, quality])

    // Refs
    const currentProviderRef = React.useRef<string | null>(null)
    const [previousState, setPreviousState] = React.useState<{ currentTime: number, paused: boolean } | null>(null)

    React.useEffect(() => {
        setPreviousState(null)
        React.startTransition(() => {
            setPreviousState(null)
        })
    }, [currentEpisodeNumber, media])

    const { getHMACTokenQueryParam } = useServerHMACAuth()

    // update the stream URL when the video source changes
    React.useEffect(() => {
        (async () => {
            setPlaybackError(null)
            log.info("Changing stream URL using videoSource", { videoSource })
            setUrl(null)
            log.info("Setting stream URL to undefined")
            if (videoSource?.url) {
                setServer(videoSource.server)
                let _url = videoSource.url
                if (videoSource.headers && Object.keys(videoSource.headers).length > 0) {
                    _url = `${getServerBaseUrl()}/api/v1/proxy?url=${encodeURIComponent(videoSource?.url)}&headers=${encodeURIComponent(JSON.stringify(
                        videoSource?.headers))}` + (await getHMACTokenQueryParam("/api/v1/proxy", "&"))
                } else {
                    _url = videoSource.url
                }
                React.startTransition(async () => {
                    // If the video source is unknown or we can't determine if it's a native video from the url,
                    // send a HEAD request to determine the content type
                    if (videoSource.type === "unknown" || !isValidVideoSourceType(videoSource.type) || (videoSource.type === "mp4" && !isNativeVideoExtension(
                        _url)) || (videoSource.type === "m3u8" && !isHLSSrc(_url))) {
                        log.warning("Verifying original video source type", videoSource)
                        if (await isProbablyHls(_url) === "hls") {
                            log.info("Detected HLS source type")
                            setOverrideStreamType("hls")
                        } else {
                            setOverrideStreamType(!isValidVideoSourceType(videoSource.type) ? "native" : null)
                        }
                    }
                    React.startTransition(() => {
                        log.info("Setting stream URL", { url: _url, quality, server, dubbed, provider })
                        setUrl(_url)
                    })
                })
            }
        })()
    }, [videoSource, server, quality, dubbed, provider])

    const { currentPlaylist, playEpisode: playPlaylistEpisode, nextPlaylistEpisode, prevPlaylistEpisode } = usePlaylistManager()

    function savePreviousStateThen(cb: () => void) {
        setPreviousState({
            currentTime: playerRef.current?.currentTime ?? 0,
            paused: playerRef.current?.paused ?? true,
        })
        React.startTransition(() => {
            cb()
        })
    }

    const changeQuality = React.useCallback((source: VideoCore_VideoSource) => {
        savePreviousStateThen(() => {
            setQuality(source.resolution)
        })
    }, [videoSource])

    // Provider
    const changeProvider = React.useCallback((provider: string) => {
        savePreviousStateThen(() => {
            setProvider(provider)
        })
    }, [videoSource])

    // Server
    const changeServer = React.useCallback((server: string) => {
        savePreviousStateThen(() => {
            setServer(server)
        })
    }, [videoSource])

    // Dubbed
    const toggleDubbed = React.useCallback(() => {
        savePreviousStateThen(() => {
            setDubbed((prev) => !prev)
        })
    }, [videoSource])

    const episodeListLoading = isFetchingEpisodeList || isLoadingEpisodeList
    const episodeLoading = isLoadingEpisodeSource || isFetchingEpisodeSource

    /*
     * Set episode number on mount
     */
    const firstRenderRef = React.useRef(true)
    React.useEffect(() => {
        // Do not auto set the episode number if the user is in a watch party and is not the host
        if (isWatchPartyPeer) return

        // Do not auto set if we're loading from watch party
        if (isLoadingFromWatchPartyRef.current) {
            return
        }

        if (!!media && firstRenderRef.current && !!episodes) {
            const episodeNumberFromURL = urlEpNumber ? Number(urlEpNumber) : undefined
            const progress = animeEntry?.listData?.progress ?? 0
            let episodeNumber = 1
            const episodeToWatch = episodes.find(e => e.number === progress + 1)
            if (episodeToWatch) {
                episodeNumber = episodeToWatch.number
            }
            setSelectedEpisodeNumber(episodeNumberFromURL || episodeNumber || 1)
            log.info("Setting episode number to", episodeNumberFromURL || episodeNumber || 1)
            firstRenderRef.current = false
        }
    }, [episodes, media, animeEntry?.listData, urlEpNumber, currentPlaylist, isWatchPartyPeer])


    function onCanPlay() {
        if (urlEpNumber) {
            router.replace(pathname + `?id=${mediaId}`)
        }
    }

    const goToNextEpisode = useLatestFunction(() => {
        if (currentEpisodeNumber === null) return
        if (currentPlaylist) {
            playPlaylistEpisode("next", true)
            return
        }
        // check if the episode exists
        if (episodes?.find(e => e.number === currentEpisodeNumber + 1)) {
            setSelectedEpisodeNumber(currentEpisodeNumber + 1)
        }
    })

    const goToPreviousEpisode = useLatestFunction(() => {
        if (currentEpisodeNumber === null) return
        if (currentPlaylist) {
            playPlaylistEpisode("previous", true)
            return
        }
        if (currentEpisodeNumber > 1) {
            // check if the episode exists
            if (episodes?.find(e => e.number === currentEpisodeNumber - 1)) {
                setSelectedEpisodeNumber(currentEpisodeNumber - 1)
            }
        }
    })

    const handlePlayEpisode = useLatestFunction((which: "next" | "previous") => {
        setUrl(null)
        React.startTransition(() => {
            if (which === "next") {
                goToNextEpisode()
            } else {
                goToPreviousEpisode()
            }
        })
    })

    const useLibassRenderer = useAtomValue(vc_useLibassRendererAtom)

    // Store the errored servers, so we can switch to the next server
    const [erroredServers, setErroredServers] = React.useState<string[]>([])
    // Clear errored servers when the episode details change
    React.useEffect(() => {
        setErroredServers([])
    }, [currentEpisode])

    /*
     * Handle fatal errors
     * This function is called when the player encounters a fatal error
     * - Change the server if the server is errored
     * - Change the provider if all servers are errored
     */
    const onFatalError = (reason: string) => {
        log.error("onFatalError", {
            sameProvider: provider == currentProviderRef.current,
            reason: reason,
        })
        if (provider == currentProviderRef.current) {
            setUrl(null)
            log.error("Setting stream URL to undefined")
            toast.warning("Playback error, trying another server...")
            log.error("Player encountered a fatal error")
            setTimeout(() => {
                log.error("erroredServers", erroredServers)
                if (videoSource?.server) {
                    const otherServers = servers.filter((server) => server !== videoSource?.server && !erroredServers.includes(server))
                    if (otherServers.length > 0) {
                        setErroredServers((prev) => [...prev, videoSource?.server])
                        setServer(otherServers[0])
                    } else {
                        setProvider((prev) => providerExtensionOptions.find((p) => p.value !== prev)?.value ?? null)
                    }
                }
            }, 500)
        } else {
            setPlaybackError(reason)
        }
    }

    const parameters = (
        <>
            <Select
                value={provider || ""}
                options={[
                    ...providerExtensionOptions,
                    {
                        value: "add-provider",
                        label: "Find other providers",
                    },
                ]}
                onValueChange={(v) => {
                    if (v === "add-provider") {
                        router.push(`/extensions?tab=marketplace&type=onlinestream-provider`)
                        return
                    }
                    changeProvider(v)
                }}
                placeholder="Select provider"
                size="sm"
                leftAddon={<CgMediaPodcast />}
                fieldClass="w-fit"
                className="rounded-full rounded-l-none w-fit"
                addonClass="rounded-full rounded-r-none"
            />
            {!!servers.length && <Select
                size="sm"
                value={server}
                options={servers.map((server) => ({ label: server, value: server }))}
                onValueChange={(v) => {
                    changeServer(v)
                }}
                disabled={servers.length <= 1}
                fieldClass="w-fit"
                className="rounded-full w-fit !px-4"
                addonClass="rounded-full rounded-r-none"
            />}
            <IsomorphicPopover
                title="Stream"
                trigger={<Button
                    intent="gray-basic"
                    size="sm"
                    className="rounded-full"
                    leftIcon={<HiOutlineCog6Tooth className="text-xl" />}
                >
                    Cache
                </Button>}
            >
                <p className="text-sm text-[--muted]">
                    Empty the cache if you are experiencing issues with the stream.
                </p>
                <Button
                    size="sm"
                    intent="alert-subtle"
                    onClick={() => emptyCache({ mediaId: (mediaId!) })}
                    loading={isEmptyingCache}
                >
                    Empty stream cache
                </Button>
            </IsomorphicPopover>
        </>
    )

    if (!media || animeEntryLoading) return <div data-onlinestream-page-loading-container className="space-y-4">
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
            <VideoCoreInlineHelpers
                playerRef={playerRef}
                currentEpisodeNumber={currentEpisodeNumber}
                currentProgress={progress}
                media={media}
                url={url}
            />
            <VideoCoreInlineLayout
                mediaId={mediaId ? Number(mediaId) : undefined}
                currentEpisodeNumber={currentEpisodeNumber}
                title={media?.title?.userPreferred}
                hideBackButton={hideBackButton}
                episodes={episodes}
                loadingEpisodeList={episodeListLoading}
                leftHeaderActions={<>
                    {parameters}
                    {(animeEntry && !!provider) && <OnlinestreamManualMappingModal entry={animeEntry}>
                        <Button
                            size="sm"
                            intent="gray-basic"
                            className="rounded-full"
                            leftIcon={<FaSearch className="" />}
                        >
                            Manual match
                        </Button>
                    </OnlinestreamManualMappingModal>}
                    <Button
                        className=""
                        rounded
                        intent="gray-basic"
                        size="sm"
                        leftIcon={!dubbed ? <LuSpeech className="text-xl" /> : <MdOutlineSubtitles className="text-xl" />}
                        onClick={() => toggleDubbed()}
                    >
                        {dubbed ? "Switch to subs" : "Switch to dub"}
                    </Button>
                    <div className="hidden lg:flex flex-1"></div>
                </>}
                rightHeaderActions={<>
                    <VideoCoreInlineHelperUpdateProgressButton />
                    <IconButton
                        size="sm"
                        intent={episodeViewMode === "list" ? "gray-basic" : "white-subtle"}
                        icon={<BsFillGrid3X3GapFill />}
                        onClick={() => setEpisodeViewMode(prev => prev === "list" ? "grid" : "list")}
                        title={episodeViewMode === "list" ? "Switch to grid view" : "Switch to list view"}
                    />
                </>}
                mediaPlayer={!provider ? (
                    <div className="flex items-center flex-col justify-center w-full h-full">
                        <LuffyError title="No provider selected" />
                        <div className="flex gap-2">
                            {parameters}
                        </div>
                    </div>
                ) : isEpisodeListError ? <LuffyError title="Provider error">Could not fetch episode list from provider.</LuffyError> : (
                    <>
                        <VideoCoreProvider id="onlinestream">
                            <div data-onlinestream-video-container className="w-full aspect-video mx-auto border rounded-lg overflow-hidden">
                                <VideoCore
                                    id="onlinestream"
                                    mRef={playerRef}
                                    state={{
                                        active: true,
                                        playbackInfo: !!url ? {
                                            id: url,
                                            playbackType: "onlinestream",
                                            streamUrl: url!,
                                            media: media,
                                            episode: currentEpisode?.metadata,
                                            playlistExternalEpisodeNumbers: episodes?.map(e => e.number),
                                            streamType: overrideStreamType
                                                ? overrideStreamType
                                                : ((url && isHLSSrc(url)) || videoSource?.type === "m3u8") ? "hls" : "native",
                                            subtitleTracks: episodeSource?.subtitles?.map((sub, index) => ({
                                                index: index,
                                                label: sub.language,
                                                src: sub.url,
                                                language: sub.language,
                                                default: index === 0,
                                                useLibassRenderer: useLibassRenderer,
                                            })),
                                            videoSources: hasMultipleVideoSources ? videoSources?.map((source, index) => ({
                                                index: index,
                                                label: source.label,
                                                src: source.url,
                                                resolution: source.quality,
                                            })) : undefined,
                                            selectedVideoSource: videoSources?.findIndex(source => source.quality === videoSource?.quality) ?? undefined,
                                            initialState: previousState ?? undefined,
                                            onlinestreamParams: {
                                                mediaId: mediaId!,
                                                episodeNumber: currentEpisodeNumber!,
                                                provider: provider,
                                                dubbed: dubbed,
                                                server: server || "",
                                                quality: quality || "",
                                            },
                                            disableRestoreFromContinuity: !!nakamaStatus?.currentWatchPartySession,
                                        } : null,
                                        playbackError: isErrorEpisodeSource
                                            ? (errorEpisodeSource as AxiosError<{ error: string }>)?.response?.data?.error ?? null
                                            : playbackError,
                                        loadingState: !url ? "Loading stream" : null,
                                    }}
                                    inline
                                    aniSkipData={aniSkipData}
                                    onLoadedMetadata={onCanPlay}
                                    onError={v => onFatalError(v)}
                                    onPlayEpisode={handlePlayEpisode}
                                    onVideoSourceChange={changeQuality}
                                    onHlsFatalError={(err) => onFatalError(`HLS error: ${err.error.message}`)}
                                    onTerminateStream={() => {
                                        setUrl(null)
                                        setPlaybackError("Stream terminated")
                                    }}
                                />
                            </div>
                        </VideoCoreProvider>
                    </>
                )}
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
                                {episodes?.filter(Boolean)?.sort((a, b) => a!.number - b!.number)?.map((episode, idx) => {
                                    return (
                                        <EpisodeGridItem
                                            key={idx + (episode.title || "") + episode.number}
                                            id={`episode-${String(episode.number)}`}
                                            onClick={() => setSelectedEpisodeNumber(episode.number)}
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
                                            episodeNumber={episode.number}
                                            progressNumber={episode.number}
                                            action={<>
                                                <MediaEpisodeInfoModal
                                                    title={media.format === "MOVIE" ? "Complete movie" : `Episode ${episode.number}`}
                                                    image={episode?.image}
                                                    episodeTitle={episode.title}
                                                    summary={episode?.description}
                                                />

                                                <PluginEpisodeGridItemMenuItems isDropdownMenu={true} type="onlinestream" episode={episode} />
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
                                    number: ep.number,
                                    title: ep.title,
                                    isFiller: ep.isFiller,
                                })) || []}
                                currentEpisodeNumber={currentEpisodeNumber}
                                onEpisodeSelect={setSelectedEpisodeNumber}
                                progress={progress}
                                disabled={episodeLoading}
                                getEpisodeId={(ep) => `episode-${ep.number}`}
                            />
                        )}
                    </AnimatePresence>
                </>}
            />
        </>
    )
}


function IsomorphicPopover
(
    props: PopoverProps & ModalProps) {
    const { title, children, ...rest } = props
    const { width } = useWindowSize()

    if (width && width > 1024) {
        return <Popover
            {...rest}
            className="max-w-xl !w-full overflow-hidden space-y-2"
        >
            {children}
        </Popover>
    }

    return <Modal
        {...rest}
        title={title}
    >
        {children}
    </Modal>
}
