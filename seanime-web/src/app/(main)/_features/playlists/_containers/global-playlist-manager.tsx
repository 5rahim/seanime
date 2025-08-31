import { Anime_Playlist, Anime_PlaylistEpisode } from "@/api/generated/types"
import { useCurrentDevicePlaybackSettings } from "@/app/(main)/_atoms/playback.atoms"
import { PlaylistManagerPopup } from "@/app/(main)/_features/playlists/_components/global-playlist-popup"
import { playlist_getEpisodeKey, playlist_isSameEpisode } from "@/app/(main)/_features/playlists/_components/playlist-editor"
import { useWebsocketMessageListener, useWebsocketSender } from "@/app/(main)/_hooks/handle-websockets"
import { useMediastreamActiveOnDevice } from "@/app/(main)/mediastream/_lib/mediastream.atoms"
import { websocketConnectedAtom } from "@/app/websocket-provider"
import { imageShimmer } from "@/components/shared/image-helpers"
import { IconButton } from "@/components/ui/button"
import { cn } from "@/components/ui/core/styling"
import { Tooltip } from "@/components/ui/tooltip"
import { getImageUrl } from "@/lib/server/assets"
import { WSEvents } from "@/lib/server/ws-events"
import { __isElectronDesktop__ } from "@/types/constants"
import { atom, useAtomValue } from "jotai"
import { useAtom } from "jotai/react"
import Image from "next/image"
import React from "react"
import { BiX } from "react-icons/bi"
import { LuRefreshCw } from "react-icons/lu"
import { MdSkipNext, MdSkipPrevious } from "react-icons/md"
import { toast } from "sonner"

type ServerEvents =
    "current-playlist"

const pm_currentPlaylist = atom<Anime_Playlist | null>(null)
const pm_currentPlaylistEpisode = atom<Anime_PlaylistEpisode | null>(null)

export function usePlaylistManager() {
    const { sendMessage } = useWebsocketSender()

    const [currentPlaylist, setCurrentPlaylist] = useAtom(pm_currentPlaylist)

    const { downloadedMediaPlayback, torrentStreamingPlayback, electronPlaybackMethod } = useCurrentDevicePlaybackSettings()
    const { activeOnDevice } = useMediastreamActiveOnDevice()

    function startPlaylist(playlist: Anime_Playlist) {
        toast.info("Starting playlist...")
        sendMessage({
            type: WSEvents.PLAYLIST,
            payload: {
                type: "start-playlist",
                payload: {
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
        sendMessage({
            type: WSEvents.PLAYLIST,
            payload: {
                type: "stop-playlist",
            },
        })
    }

    function reopenEpisode() {
        sendMessage({
            type: WSEvents.PLAYLIST,
            payload: {
                type: "reopen-episode",
            },
        })
    }

    return {
        currentPlaylist,
        startPlaylist,
        stopPlaylist,
        reopenEpisode,
    }
}

export function GlobalPlaylistManager() {
    const { sendMessage } = useWebsocketSender()
    const websocketConnected = useAtomValue(websocketConnectedAtom)

    const { stopPlaylist, reopenEpisode } = usePlaylistManager()

    // state
    const [currentPlaylist, setCurrentPlaylist] = useAtom(pm_currentPlaylist)
    const [currentPlaylistEpisode, setCurrentPlaylistEpisode] = useAtom(pm_currentPlaylistEpisode)

    React.useEffect(() => {
        sendMessage({ type: WSEvents.PLAYLIST, payload: { type: "current-playlist" } })
    }, [websocketConnected])

    const currentPlaylistEpisodeIndex = currentPlaylist?.episodes?.findIndex(n => playlist_isSameEpisode(n, currentPlaylistEpisode)) ?? -1

    const nextPlaylistEpisode = currentPlaylist?.episodes?.[currentPlaylistEpisodeIndex + 1]
    const prevPlaylistEpisode = currentPlaylist?.episodes?.[currentPlaylistEpisodeIndex - 1]

    useWebsocketMessageListener({
        type: WSEvents.PLAYLIST,
        onMessage: (data: { type: ServerEvents, payload: unknown }) => {
            switch (data.type) {
                case "current-playlist":
                    const payload = data.payload as { playlist: Anime_Playlist | null, playlistEpisode: Anime_PlaylistEpisode | null }
                    setCurrentPlaylist(payload.playlist ?? null)
                    setCurrentPlaylistEpisode(payload.playlistEpisode ?? null)
                    break
            }
        },
    })

    if (!currentPlaylist) return null

    return <PlaylistManagerPopup position="bottom-right">
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
                />
            </div>
        </div>
    </PlaylistManagerPopup>
}

function EpisodeItem({ episode }: { episode: Anime_PlaylistEpisode }) {

    const currentPlaylistEpisode = useAtomValue(pm_currentPlaylistEpisode)

    return (
        <div
            className={cn(
                "px-2.5 py-2 bg-[--background] rounded-md border flex gap-3 relative",
                "opacity-50 hover:opacity-70",
                playlist_isSameEpisode(currentPlaylistEpisode,
                    episode) && "opacity-100 hover:opacity-100 border-[rgba(255,255,255,0.5)] sticky top-0 z-10 shadow-xl",
            )}
        >
            <div className="size-20 aspect-square flex-none rounded-md overflow-hidden relative transition bg-[--background]">
                {episode.episode!.episodeMetadata?.image && <Image
                    data-episode-card-image
                    src={getImageUrl(episode.episode!.episodeMetadata?.image)}
                    alt={""}
                    fill
                    quality={100}
                    placeholder={imageShimmer(700, 475)}
                    sizes="20rem"
                    className={cn(
                        "object-cover rounded-lg object-center transition lg:group-hover/episode-card:scale-105 duration-200",
                    )}
                />}
            </div>
            <div className="max-w-full space-y-1">
                <p className="text-sm text-[--muted]">{episode.episode?.baseAnime?.title?.userPreferred}</p>
                <p className="">{episode.episode?.baseAnime?.format !== "MOVIE" ? `Episode ${episode.episode!.episodeNumber}` : "Movie"}</p>

                <div>
                    <div className="text-xs text-[--muted] line-clamp-1 tracking-wide">
                        {episode.watchType === "torrent" ? "Torrent streaming" : episode.watchType === "debrid" ? "Debrid streaming " :
                            episode.episode?.localFile?.name}
                    </div>
                </div>
            </div>
        </div>
    )
}
