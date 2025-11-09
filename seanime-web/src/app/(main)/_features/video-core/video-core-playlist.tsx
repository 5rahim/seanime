import { Anime_Entry, Anime_Episode, HibikeTorrent_AnimeTorrent } from "@/api/generated/types"
import { useGetAnimeEpisodeCollection } from "@/api/hooks/anime.hooks"
import { useGetAnimeEntry } from "@/api/hooks/anime_entries.hooks"
import { EpisodeGridItem } from "@/app/(main)/_features/anime/_components/episode-grid-item"
import { useAutoPlaySelectedTorrent } from "@/app/(main)/_features/autoplay/autoplay"
import { nativePlayer_stateAtom, NativePlayerState } from "@/app/(main)/_features/native-player/native-player.atoms"
import { usePlaylistManager } from "@/app/(main)/_features/playlists/_containers/global-playlist-manager"
import { VideoCoreNextButton, VideoCorePreviousButton } from "@/app/(main)/_features/video-core/video-core-control-bar"
import { useHandleStartDebridStream } from "@/app/(main)/entry/_containers/debrid-stream/_lib/handle-debrid-stream"
import {
    __debridStream_autoSelectFileAtom,
    __debridStream_currentSessionAutoSelectAtom,
} from "@/app/(main)/entry/_containers/debrid-stream/debrid-stream-page"
import { useTorrentSearchSelectedStreamEpisode } from "@/app/(main)/entry/_containers/torrent-search/_lib/handle-torrent-selection"
import {
    __torrentSearch_selectionAtom,
    __torrentSearch_selectionEpisodeAtom,
    TorrentSearchDrawer,
} from "@/app/(main)/entry/_containers/torrent-search/torrent-search-drawer"
import { useHandleStartTorrentStream } from "@/app/(main)/entry/_containers/torrent-stream/_lib/handle-torrent-stream"
import {
    __torrentStream_autoSelectFileAtom,
    __torrentStream_currentSessionAutoSelectAtom,
} from "@/app/(main)/entry/_containers/torrent-stream/torrent-stream-page"
import { useHandlePlayMedia } from "@/app/(main)/entry/_lib/handle-play-media"
import { HoverCard } from "@/components/ui/hover-card"
import { logger } from "@/lib/helpers/debug"
import { atom, useAtomValue } from "jotai"
import { useAtom, useSetAtom } from "jotai/react"
import React from "react"
import { useUpdateEffect } from "react-use"
import { toast } from "sonner"

type VideoCorePlaylistState = {
    type: NonNullable<NativePlayerState["playbackInfo"]>["streamType"]
    episodes: Anime_Episode[]
    previousEpisode: Anime_Episode | null
    nextEpisode: Anime_Episode | null
    currentEpisode: Anime_Episode
    currentTorrent?: HibikeTorrent_AnimeTorrent // for torrent and debrid stream type
    animeEntry: Anime_Entry | null
}

const log = logger("VIDEO CORE PLAYLIST")

export const vc_playlistState = atom<VideoCorePlaylistState | null>(null)

// call once, maintains playlist state
export function useVideoCorePlaylistSetup() {
    const [playlistState, setPlaylistState] = useAtom(vc_playlistState)

    const state = useAtomValue(nativePlayer_stateAtom)
    const playbackInfo = state.playbackInfo
    const streamType = state.playbackInfo?.streamType
    const mediaId = state.playbackInfo?.media?.id
    const mediaType = state.playbackInfo?.streamType

    const currProgressNumber = playbackInfo?.episode?.progressNumber || 0

    // Fetch anime entry and episode collection
    // episode collection will be used for non-localfile streams
    const { data: animeEntry } = useGetAnimeEntry(mediaId)
    const { data: episodeCollection, isLoading, refetch } = useGetAnimeEpisodeCollection(mediaId)

    useUpdateEffect(() => {
        if (mediaId) {
            refetch()
        }
    }, [state.playbackInfo, mediaId])

    // Get the episodes depending on the stream type
    const episodes = React.useMemo(() => {
        if (!episodeCollection) return []

        if (streamType === "localfile") {
            return animeEntry?.episodes?.filter(ep => ep.type === "main") ?? []
        }

        return episodeCollection?.episodes ?? []
    }, [animeEntry?.episodes, episodeCollection?.episodes, currProgressNumber, streamType])

    const currentEpisode = episodes.find?.(ep => ep.progressNumber === currProgressNumber) ?? null
    const previousEpisode = episodes.find?.(ep => ep.progressNumber === currProgressNumber - 1) ?? null
    const nextEpisode = episodes.find?.(ep => ep.progressNumber === currProgressNumber + 1) ?? null

    React.useEffect(() => {
        if (!playbackInfo || !currentEpisode || !episodes.length || !animeEntry) {
            setPlaylistState(null)
            return
        }

        setPlaylistState({
            type: mediaType!,
            episodes: episodes ?? [],
            currentEpisode,
            previousEpisode,
            nextEpisode,
            animeEntry,
        })
    }, [animeEntry, playbackInfo, currentEpisode, previousEpisode, nextEpisode])
}

export function useVideoCorePlaylist() {
    const playlistState = useAtomValue(vc_playlistState)
    const streamType = useAtomValue(nativePlayer_stateAtom)?.playbackInfo?.streamType
    const animeEntry = playlistState?.animeEntry

    const setTorrentSearch = useSetAtom(__torrentSearch_selectionAtom)
    const setTorrentSearchEpisode = useSetAtom(__torrentSearch_selectionEpisodeAtom)
    const { setTorrentSearchStreamEpisode } = useTorrentSearchSelectedStreamEpisode()
    const {
        handleStreamSelection: handleTorrentstreamSelection,
        handleAutoSelectStream: handleTorrentstreamAutoSelect,
    } = useHandleStartTorrentStream()
    const { handleStreamSelection: handleDebridstreamSelection, handleAutoSelectStream: handleDebridstreamAutoSelect } = useHandleStartDebridStream()
    const { playMediaFile } = useHandlePlayMedia()

    // If user is auto-selecting the torrent
    const [torrentStream_currentSessionAutoSelect] = useAtom(__torrentStream_currentSessionAutoSelectAtom)
    const [debridStream_currentSessionAutoSelect] = useAtom(__debridStream_currentSessionAutoSelectAtom)
    // If user is auto-selecting the file
    const [torrentStream_autoSelectFile] = useAtom(__torrentStream_autoSelectFileAtom)
    const [debridStream_autoSelectFile] = useAtom(__debridStream_autoSelectFileAtom)

    // The torrent to continue playing from
    const { autoPlayTorrent } = useAutoPlaySelectedTorrent()

    // Global playlist
    const {
        nextPlaylistEpisode: globalPlaylistNextEpisode,
        prevPlaylistEpisode: globalPlaylistPreviousEpisode,
        currentPlaylist: globalCurrentPlaylist,
        playEpisode: playGlobalPlaylistEpisode,
    } = usePlaylistManager()

    function startStream(episode: Anime_Episode) {
        if (!playlistState?.animeEntry || !episode.aniDBEpisode) return
        log.info("Stream requested for ", episode.episodeNumber)

        if (streamType === "torrent" || streamType === "debrid") {
            log.info("Auto selecting torrent for ", episode.episodeNumber)
            if (streamType === "torrent" && torrentStream_currentSessionAutoSelect) {
                handleTorrentstreamAutoSelect({
                    mediaId: playlistState.animeEntry.mediaId,
                    episodeNumber: episode.episodeNumber,
                    aniDBEpisode: episode.aniDBEpisode,
                })
                return
            } else if (streamType === "debrid" && debridStream_currentSessionAutoSelect) {

                handleDebridstreamAutoSelect({
                    mediaId: playlistState.animeEntry.mediaId,
                    episodeNumber: episode.episodeNumber,
                    aniDBEpisode: episode.aniDBEpisode,
                })
                return
            }
        }

        // If a torrent was selected for auto play (i.e. user manually select torrent with auto select file)
        if (autoPlayTorrent?.torrent?.isBatch) {
            log.info("Previous torrent selected for auto play", autoPlayTorrent)
            let fileIndex: number | undefined = undefined
            if (autoPlayTorrent?.batchFiles) {
                const file = autoPlayTorrent.batchFiles.files?.find(n => n.index === autoPlayTorrent.batchFiles!.current + 1)
                if (file) {
                    fileIndex = file.index
                }
            }
            if (streamType === "torrent") {
                handleTorrentstreamSelection({
                    mediaId: playlistState.animeEntry.mediaId,
                    episodeNumber: episode.episodeNumber,
                    aniDBEpisode: episode.aniDBEpisode,
                    torrent: autoPlayTorrent.torrent,
                    chosenFileIndex: fileIndex,
                    batchEpisodeFiles: (autoPlayTorrent?.batchFiles && fileIndex !== undefined) ? {
                        ...autoPlayTorrent.batchFiles,
                        current: fileIndex,
                        currentEpisodeNumber: episode.episodeNumber,
                        currentAniDBEpisode: episode.aniDBEpisode,
                    } : undefined,
                })
            } else if (streamType === "debrid") {
                handleDebridstreamSelection({
                    mediaId: playlistState.animeEntry.mediaId,
                    episodeNumber: episode.episodeNumber,
                    aniDBEpisode: episode.aniDBEpisode,
                    torrent: autoPlayTorrent.torrent,
                    chosenFileId: fileIndex !== undefined ? String(fileIndex) : "",
                    batchEpisodeFiles: (autoPlayTorrent?.batchFiles && fileIndex !== undefined) ? {
                        ...autoPlayTorrent.batchFiles,
                        current: fileIndex,
                        currentEpisodeNumber: episode.episodeNumber,
                        currentAniDBEpisode: episode.aniDBEpisode,
                    } : undefined,
                })
            }
        } else {
            setTorrentSearchEpisode(episode.episodeNumber)
            setTorrentSearchStreamEpisode(episode)
            log.info("Torrent search for ", episode.episodeNumber)
            React.startTransition(() => {
                if (streamType === "torrent") {
                    setTorrentSearch(torrentStream_autoSelectFile ? "torrentstream-select" : "torrentstream-select-file")
                } else if (streamType === "debrid") {
                    setTorrentSearch(debridStream_autoSelectFile ? "debridstream-select" : "debridstream-select-file")
                }
            })
        }
    }

    const playEpisode = (which: "previous" | "next" | string) => {
        if (!playlistState) {
            toast.error("Unexpected error: No playlist state")
            return
        }
        if (!animeEntry) {
            toast.error("Unexpected error: No entry")
            return
        }

        log.info("Requesting episode", which)

        if (globalCurrentPlaylist) {
            log.info("Playing global playlist episode", which)
            switch (which) {
                case "previous":
                    if (globalPlaylistPreviousEpisode) {
                        playGlobalPlaylistEpisode("previous", true)
                    }
                    break
                case "next":
                    if (globalPlaylistNextEpisode) {
                        playGlobalPlaylistEpisode("next", true)
                    }
                    break
            }

            return
        }

        let episode: Anime_Episode | null = null
        switch (which) {
            case "previous":
                if (playlistState?.previousEpisode) {
                    episode = playlistState.previousEpisode
                }
                break
            case "next":
                if (playlistState?.nextEpisode) {
                    episode = playlistState.nextEpisode
                }
                break
            default:
                episode = playlistState?.episodes?.find(n => n.aniDBEpisode === which) ?? null
        }

        if (!episode) {
            log.info("Episode not found for", which)
            return
        }

        log.info("Playing episode", episode)

        switch (streamType) {
            case "localfile":
                if (!episode?.localFile?.path) {
                    toast.error("Local file not found")
                    return
                }
                playMediaFile({
                    path: episode?.localFile?.path,
                    episode: episode,
                    mediaId: animeEntry?.mediaId,
                })
                break
            case "torrent":
            case "debrid":
                startStream(episode)
                break
        }
    }

    return {
        playlistState,
        animeEntry: playlistState?.animeEntry,
        hasPreviousEpisode: !!playlistState?.previousEpisode,
        hasNextEpisode: !!playlistState?.nextEpisode,
        playEpisode,

    }
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

function PlaylistEpisodeHoverCard({ episode, children }: { episode?: Anime_Episode, children: React.ReactNode }) {
    return (
        <HoverCard
            className="max-w-xl w-fit py-2 px-4 ml-4"
            sideOffset={38}
            closeDelay={200}
            trigger={<span>
                {children}
            </span>}
        >
            <EpisodeGridItem
                key={JSON.stringify(episode)}
                media={episode?.baseAnime as any}
                title={episode?.displayTitle || episode?.baseAnime?.title?.userPreferred || ""}
                image={episode?.episodeMetadata?.image || episode?.baseAnime?.coverImage?.large}
                episodeTitle={episode?.episodeTitle}
                fileName={episode?.localFile?.parsedInfo?.original}
                description={episode?.episodeMetadata?.summary || episode?.episodeMetadata?.overview}
                isFiller={episode?.episodeMetadata?.isFiller}
                length={episode?.episodeMetadata?.length}
                className="flex-none w-full"
                episodeNumber={episode?.episodeNumber}
                progressNumber={episode?.progressNumber}
            />
        </HoverCard>
    )
}

export function VideoCorePlaylistControl() {
    const { animeEntry, hasNextEpisode, hasPreviousEpisode, playEpisode } = useVideoCorePlaylist()

    // Global playlist
    const { nextPlaylistEpisode, prevPlaylistEpisode, currentPlaylist, playEpisode: playPlaylistEpisode } = usePlaylistManager()

    if (currentPlaylist) {
        return <>
            {!!prevPlaylistEpisode && <PlaylistEpisodeHoverCard episode={prevPlaylistEpisode?.episode}>
                <VideoCorePreviousButton
                    onClick={() => {
                        playPlaylistEpisode("previous", true)
                    }}
                />
            </PlaylistEpisodeHoverCard>}
            {!!nextPlaylistEpisode && <PlaylistEpisodeHoverCard episode={nextPlaylistEpisode?.episode}>
                <VideoCoreNextButton
                    onClick={() => {
                        playPlaylistEpisode("next", true)
                    }}
                />
            </PlaylistEpisodeHoverCard>}
        </>
    }

    return (
        <>
            {hasPreviousEpisode && <VideoCorePreviousButton
                onClick={() => {
                    playEpisode("previous")
                }}
            />}
            {hasNextEpisode && <VideoCoreNextButton
                onClick={() => {
                    playEpisode("next")
                }}
            />}
            {animeEntry && <TorrentSearchDrawer entry={animeEntry as Anime_Entry} />}
        </>
    )
}
