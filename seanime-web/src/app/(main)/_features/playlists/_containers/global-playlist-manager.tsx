import { Anime_Entry, Anime_Playlist, Anime_PlaylistEpisode, HibikeTorrent_AnimeTorrent } from "@/api/generated/types"
import { useGetAnimeEntry } from "@/api/hooks/anime_entries.hooks"
import { useCurrentDevicePlaybackSettings } from "@/app/(main)/_atoms/playback.atoms"
import { useAutoPlaySelectedTorrent } from "@/app/(main)/_features/autoplay/autoplay"
import { nativePlayer_stateAtom } from "@/app/(main)/_features/native-player/native-player.atoms"
import { PlaylistManagerPopup } from "@/app/(main)/_features/playlists/_components/global-playlist-popup"
import { playlist_getEpisodeKey, playlist_isSameEpisode } from "@/app/(main)/_features/playlists/_components/playlist-editor"
import { useWebsocketMessageListener, useWebsocketSender } from "@/app/(main)/_hooks/handle-websockets"
import { useServerStatus } from "@/app/(main)/_hooks/use-server-status"
import { useHandleStartDebridStream } from "@/app/(main)/entry/_containers/debrid-stream/_lib/handle-debrid-stream"
import { __debridStream_autoSelectFileAtom } from "@/app/(main)/entry/_containers/debrid-stream/debrid-stream-page"
import { useTorrentSearchSelectedStreamEpisode } from "@/app/(main)/entry/_containers/torrent-search/_lib/handle-torrent-selection"
import {
    __torrentSearch_selectionAtom,
    __torrentSearch_selectionEpisodeAtom,
    TorrentSearchDrawer,
} from "@/app/(main)/entry/_containers/torrent-search/torrent-search-drawer"
import { useHandleStartTorrentStream } from "@/app/(main)/entry/_containers/torrent-stream/_lib/handle-torrent-stream"
import { __torrentStream_autoSelectFileAtom } from "@/app/(main)/entry/_containers/torrent-stream/torrent-stream-page"
import { useHandlePlayMedia } from "@/app/(main)/entry/_lib/handle-play-media"
import { useMediastreamActiveOnDevice } from "@/app/(main)/mediastream/_lib/mediastream.atoms"
import { clientIdAtom, websocketConnectedAtom } from "@/app/websocket-provider"
import { imageShimmer } from "@/components/shared/image-helpers"
import { SeaImage } from "@/components/shared/sea-image"
import { Button, IconButton } from "@/components/ui/button"
import { cn } from "@/components/ui/core/styling"
import { Modal } from "@/components/ui/modal"
import { Tooltip } from "@/components/ui/tooltip"
import { logger } from "@/lib/helpers/debug"
import { getImageUrl } from "@/lib/server/assets"
import { WSEvents } from "@/lib/server/ws-events"
import { __isElectronDesktop__ } from "@/types/constants"
import { atom, useAtomValue } from "jotai"
import { useAtom, useSetAtom } from "jotai/react"
import { useRouter } from "next/navigation"
import React from "react"
import { BiX } from "react-icons/bi"
import { LuRefreshCw } from "react-icons/lu"
import { MdSkipNext, MdSkipPrevious } from "react-icons/md"
import { toast } from "sonner"

const log = logger("PLAYLIST MANAGER")

type ServerEvents =
    "current-playlist" |
    "play-episode" |
    "playing-episode"

const pm_currentPlaylist = atom<Anime_Playlist | null>(null)
const pm_currentPlaylistEpisode = atom<Anime_PlaylistEpisode | null>(null)
const pm_confirmProgressUpdateModalOpen = atom<"next" | "previous" | null>(null)

export function usePlaylistManager() {
    const { sendMessage } = useWebsocketSender()
    const clientId = useAtomValue(clientIdAtom)

    const [currentPlaylist, setCurrentPlaylist] = useAtom(pm_currentPlaylist)
    const [currentPlaylistEpisode, setCurrentPlaylistEpisode] = useAtom(pm_currentPlaylistEpisode)
    const [confirmOpen, setConfirmOpen] = useAtom(pm_confirmProgressUpdateModalOpen)

    const { downloadedMediaPlayback, torrentStreamingPlayback, electronPlaybackMethod } = useCurrentDevicePlaybackSettings()
    const { activeOnDevice } = useMediastreamActiveOnDevice()

    function startPlaylist(playlist: Anime_Playlist) {
        toast.info("Starting playlist...")
        sendMessage({
            type: WSEvents.PLAYLIST,
            payload: {
                type: "start-playlist",
                payload: {
                    clientId: clientId,
                    dbId: playlist.dbId,
                    localFilePlaybackMethod: __isElectronDesktop__ && electronPlaybackMethod !== "default"
                        ? electronPlaybackMethod
                        : activeOnDevice ? "transcode" : downloadedMediaPlayback,
                    streamPlaybackMethod: __isElectronDesktop__ && electronPlaybackMethod !== "default"
                        ? electronPlaybackMethod
                        : torrentStreamingPlayback,
                },
            },
        })
    }

    function stopPlaylist() {
        log.info("Sending stop playlist event")
        sendMessage({
            type: WSEvents.PLAYLIST,
            payload: {
                type: "stop-playlist",
            },
        })
    }

    function reopenEpisode() {
        log.info("Sending reopen episode event")
        sendMessage({
            type: WSEvents.PLAYLIST,
            payload: {
                type: "reopen-episode",
            },
        })
    }

    function playEpisode(which: "next" | "previous", isCurrentCompleted: boolean) {
        log.info("Sending play episode event", which, isCurrentCompleted)
        if (isCurrentCompleted) {
            sendMessage({
                type: WSEvents.PLAYLIST,
                payload: {
                    type: "play-episode",
                    payload: {
                        which,
                        isCurrentCompleted: false, // server doesn't need to update progress
                    },
                },
            })
        } else {
            log.info("Awaiting confirmation to update progress for", which)
            setConfirmOpen(which)
        }
    }

    function onConfirmedProgress(shouldUpdate: boolean) {
        if (confirmOpen) {
            setConfirmOpen(null)
            log.info("Sending play episode event", confirmOpen, shouldUpdate)
            sendMessage({
                type: WSEvents.PLAYLIST,
                payload: {
                    type: "play-episode",
                    payload: {
                        which: confirmOpen,
                        isCurrentCompleted: shouldUpdate,
                    },
                },
            })
        }
    }

    const currentPlaylistEpisodeIndex = currentPlaylist?.episodes?.findIndex(n => playlist_isSameEpisode(n, currentPlaylistEpisode)) ?? -1

    const nextPlaylistEpisode = currentPlaylist?.episodes?.[currentPlaylistEpisodeIndex + 1]
    const prevPlaylistEpisode = currentPlaylist?.episodes?.[currentPlaylistEpisodeIndex - 1]

    return {
        currentPlaylist,
        startPlaylist,
        stopPlaylist,
        reopenEpisode,
        playEpisode,
        nextPlaylistEpisode,
        prevPlaylistEpisode,
        onConfirmedProgress,
    }
}

export function GlobalPlaylistManager() {
    const { sendMessage } = useWebsocketSender()
    const websocketConnected = useAtomValue(websocketConnectedAtom)
    const serverStatus = useServerStatus()
    const router = useRouter()

    const nativePlayerState = useAtomValue(nativePlayer_stateAtom)

    const { stopPlaylist, reopenEpisode, playEpisode, nextPlaylistEpisode, prevPlaylistEpisode, onConfirmedProgress } = usePlaylistManager()

    // state
    const [currentPlaylist, setCurrentPlaylist] = useAtom(pm_currentPlaylist)
    const [currentPlaylistEpisode, setCurrentPlaylistEpisode] = useAtom(pm_currentPlaylistEpisode)

    React.useEffect(() => {
        sendMessage({ type: WSEvents.PLAYLIST, payload: { type: "current-playlist" } })
    }, [websocketConnected])

    //------------------------------------------------------------------------------------------------------------------------------------------------

    const {
        handleStreamSelection: handleTorrentstreamSelection,
        handleAutoSelectStream: handleTorrentstreamAutoSelect,
    } = useHandleStartTorrentStream()
    const { handleStreamSelection: handleDebridstreamSelection, handleAutoSelectStream: handleDebridstreamAutoSelect } = useHandleStartDebridStream()
    const { playMediaFile } = useHandlePlayMedia()

    // If user is auto-selecting the torrent
    // const [torrentStream_currentSessionAutoSelect] = useAtom(__torrentStream_currentSessionAutoSelectAtom)
    // const [debridStream_currentSessionAutoSelect] = useAtom(__debridStream_currentSessionAutoSelectAtom)
    const torrentStream_currentSessionAutoSelect = serverStatus?.torrentstreamSettings?.autoSelect
    const debridStream_currentSessionAutoSelect = serverStatus?.debridSettings?.streamAutoSelect
    // If user is auto-selecting the file
    const [torrentStream_autoSelectFile] = useAtom(__torrentStream_autoSelectFileAtom)
    const [debridStream_autoSelectFile] = useAtom(__debridStream_autoSelectFileAtom)
    const { torrentSearchStreamEpisode, setTorrentSearchStreamEpisode } = useTorrentSearchSelectedStreamEpisode()

    const setTorrentSearch = useSetAtom(__torrentSearch_selectionAtom)
    const setTorrentSearchEpisode = useSetAtom(__torrentSearch_selectionEpisodeAtom)

    const { data: animeEntry } = useGetAnimeEntry(torrentSearchStreamEpisode?.baseAnime?.id)

    // The torrent to continue playing from
    const { autoPlayTorrent } = useAutoPlaySelectedTorrent()

    function sameTorrent(autoPlayTorrent: { entry: Anime_Entry, torrent: HibikeTorrent_AnimeTorrent } | null, episode: Anime_PlaylistEpisode) {
        if (!autoPlayTorrent) return false

        return autoPlayTorrent.entry.mediaId == episode.episode?.baseAnime?.id
    }

    useWebsocketMessageListener({
        type: WSEvents.PLAYLIST,
        onMessage: (data: { type: ServerEvents, payload: unknown }) => {
            switch (data.type) {
                case "current-playlist":
                    let payload = data.payload as { playlist: Anime_Playlist | null, playlistEpisode: Anime_PlaylistEpisode | null }
                    setCurrentPlaylist(payload.playlist ?? null)
                    setCurrentPlaylistEpisode(payload.playlistEpisode ?? null)
                    break

                case "playing-episode":
                    setTorrentSearchStreamEpisode(null)
                    setTorrentSearch(undefined)
                    break

                case "play-episode":
                    log.info("Received play episode event", data.payload)
                    const payload2 = data.payload as { playlistEpisode: Anime_PlaylistEpisode }
                    const episode = payload2.playlistEpisode

                    toast.info(`Playing episode ${episode.episode?.aniDBEpisode} of ${episode.episode?.baseAnime?.title?.userPreferred}`)

                    switch (payload2.playlistEpisode.watchType) {
                        case "nakama":
                        case "localfile":
                            log.info("Playing local file. Is Nakama: ", payload2.playlistEpisode.episode?._isNakamaEpisode)
                            if (payload2.playlistEpisode.episode?.localFile?.path) {
                                playMediaFile({
                                    path: payload2.playlistEpisode.episode?.localFile?.path,
                                    mediaId: payload2.playlistEpisode.episode?.baseAnime?.id!,
                                    episode: payload2.playlistEpisode.episode,
                                })
                            }
                            break
                        case "torrent":
                            if (torrentStream_currentSessionAutoSelect) {
                                log.info("Auto select is enabled, auto-selecting torrent for streaming")
                                handleTorrentstreamAutoSelect({
                                    mediaId: episode.episode?.baseAnime?.id!,
                                    episodeNumber: episode.episode?.episodeNumber!,
                                    aniDBEpisode: episode.episode?.aniDBEpisode!,
                                })
                                return
                            } else {
                                if (autoPlayTorrent?.torrent?.isBatch && torrentStream_autoSelectFile && sameTorrent(autoPlayTorrent, episode)) {
                                    log.info("Previous selection matches, auto-selecting file for torrent stream")
                                    handleTorrentstreamSelection({
                                        mediaId: episode.episode?.baseAnime?.id!,
                                        episodeNumber: episode.episode?.episodeNumber!,
                                        aniDBEpisode: episode.episode?.aniDBEpisode!,
                                        torrent: autoPlayTorrent.torrent,
                                        chosenFileIndex: undefined,
                                        batchEpisodeFiles: undefined,
                                    })
                                    return
                                } else {
                                    log.info("No previous torrent found, opening torrent search")
                                    setTorrentSearchEpisode(episode.episode?.episodeNumber)
                                    setTorrentSearchStreamEpisode(episode.episode!)
                                    setTorrentSearch(torrentStream_autoSelectFile ? "torrentstream-select" : "torrentstream-select-file")
                                    return
                                }
                            }
                            break
                        case "debrid":
                            if (debridStream_currentSessionAutoSelect) {
                                log.info("Auto select is enabled, auto-selecting torrent for debrid stream")
                                handleDebridstreamAutoSelect({
                                    mediaId: episode.episode?.baseAnime?.id!,
                                    episodeNumber: episode.episode?.episodeNumber!,
                                    aniDBEpisode: episode.episode?.aniDBEpisode!,
                                })
                                return
                            } else {
                                if (autoPlayTorrent?.torrent?.isBatch && debridStream_autoSelectFile && sameTorrent(autoPlayTorrent, episode)) {
                                    log.info("Previous selection matches, auto-selecting file for debrid stream")
                                    handleDebridstreamSelection({
                                        mediaId: episode.episode?.baseAnime?.id!,
                                        episodeNumber: episode.episode?.episodeNumber!,
                                        aniDBEpisode: episode.episode?.aniDBEpisode!,
                                        torrent: autoPlayTorrent.torrent,
                                        chosenFileId: "",
                                        batchEpisodeFiles: undefined,
                                    })
                                    return
                                } else {
                                    log.info("No previous debrid found, opening debrid search")
                                    setTorrentSearchEpisode(episode.episode?.episodeNumber)
                                    setTorrentSearchStreamEpisode(episode.episode!)
                                    setTorrentSearch(debridStream_autoSelectFile ? "debridstream-select" : "debridstream-select-file")
                                    return
                                }
                            }
                            break
                        case "online":
                            const params = {
                                mediaId: episode.episode?.baseAnime?.id!,
                                episodeNumber: episode.episode?.episodeNumber!,
                            }
                            router.push("/entry?id=" + params.mediaId + "&tab=onlinestream&episode=" + params.episodeNumber)
                            break
                    }
                    break
            }
        },
    })

    const [confirmProgress, setConfirmProgress] = useAtom(pm_confirmProgressUpdateModalOpen)

    if (!currentPlaylist) return null

    return <>
        {animeEntry && <TorrentSearchDrawer entry={animeEntry} isPlaylistDrawer />}

        <Modal
            open={confirmProgress !== null}
            onOpenChange={open => {
                if (!open) {
                    onConfirmedProgress(false)
                    setConfirmProgress(null)
                }
            }}
            title="Update progress?"
        >
            <p>
                Do you want to update the progress of the current episode?
            </p>

            <div className="flex gap-2 mt-4 justify-end">
                <Button intent="primary" onClick={() => onConfirmedProgress(true)}>
                    Yes
                </Button>
                <Button intent="white-subtle" onClick={() => onConfirmedProgress(false)}>
                    No
                </Button>
            </div>

        </Modal>

        {!nativePlayerState.active && <PlaylistManagerPopup position="bottom-right">
            <p className="p-3 text-sm font-semibold">
                Playlist: {currentPlaylist.name}
            </p>
            <div className="p-3 space-y-2 overflow-auto">
                <div className="space-y-2 relative">
                    {currentPlaylist.episodes?.map(ep => <EpisodeItem key={playlist_getEpisodeKey(ep)} episode={ep} />)}
                </div>
            </div>
            <div className="p-3 bg-[--paper] relative">
                <div className="absolute top-[-2rem] h-[2rem] left-0 right-0 bg-gradient-to-t from-[--paper] to-transparent">
                </div>
                <div className="flex gap-2">
                    <IconButton
                        icon={<MdSkipPrevious />}
                        intent="white-subtle"
                        className="rounded-full"
                        disabled={!prevPlaylistEpisode}
                        onClick={() => playEpisode("previous", false)}
                    />
                    <div className="flex flex-1"></div>
                    <Tooltip
                        className="z-[99999]" trigger={<span>
                        <IconButton
                            icon={<LuRefreshCw />}
                            intent="gray-basic"
                            className="rounded-full"
                            onClick={() => reopenEpisode()}
                        />
                    </span>}
                    >
                        Reopen episode
                    </Tooltip>
                    <Tooltip
                        className="z-[99999]" trigger={<span>
                        <IconButton
                            icon={<BiX />}
                            intent="alert-basic"
                            className="rounded-full"
                            onClick={() => stopPlaylist()}
                        />
                    </span>}
                    >
                        Stop playlist
                    </Tooltip>
                    <div className="flex flex-1"></div>
                    <IconButton
                        icon={<MdSkipNext />}
                        intent="white-subtle"
                        className="rounded-full"
                        disabled={!nextPlaylistEpisode}
                        onClick={() => playEpisode("next", false)}
                    />
                </div>
            </div>
        </PlaylistManagerPopup>}
    </>
}

function EpisodeItem({ episode }: { episode: Anime_PlaylistEpisode }) {

    const currentPlaylistEpisode = useAtomValue(pm_currentPlaylistEpisode)

    return (
        <div
            className={cn(
                "px-2.5 py-2 bg-[--background] rounded-md border flex gap-3 relative",
                "opacity-50 hover:opacity-70",
                playlist_isSameEpisode(currentPlaylistEpisode,
                    episode) && "opacity-100 hover:opacity-100 border-[rgba(255,255,255,0.5)] sticky top-0 bottom-0 z-10 shadow-xl",
            )}
        >
            <div className="size-20 aspect-square flex-none rounded-md overflow-hidden relative transition bg-[--background]">
                {episode.episode!.episodeMetadata?.image && <SeaImage
                    data-episode-card-image
                    src={getImageUrl(episode.episode!.episodeMetadata?.image)}
                    alt={""}
                    fill
                    quality={100}
                    placeholder={imageShimmer(700, 475)}
                    sizes="20rem"
                    className={cn(
                        "object-cover rounded-lg object-center transition lg:group-hover/episode-card:scale-105 duration-200",
                        episode.isCompleted && "opacity-10",
                    )}
                />}
            </div>
            <div className="max-w-full space-y-1">
                <p className="text-sm text-[--muted]">{episode.episode?.baseAnime?.title?.userPreferred}</p>
                <p className="">{episode.episode?.baseAnime?.format !== "MOVIE" ? `Episode ${episode.episode!.episodeNumber}` : "Movie"}</p>

                <div>
                    <div className="text-xs text-[--muted] line-clamp-1 tracking-wide">
                        {episode.watchType === "torrent" ? "Torrent streaming" : episode.watchType === "debrid" ? "Debrid streaming" :
                            episode.watchType === "online" ? "Online streaming" :
                                episode.episode?.localFile?.name}
                    </div>
                </div>
            </div>
        </div>
    )
}
