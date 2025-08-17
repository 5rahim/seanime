import { Anime_Episode } from "@/api/generated/types"
import { useGetAnimeEntry } from "@/api/hooks/anime_entries.hooks"
import { nativePlayer_stateAtom, NativePlayerState } from "@/app/(main)/_features/native-player/native-player.atoms"
import { atom, useAtomValue } from "jotai"
import { useAtom } from "jotai/react"
import React from "react"

type VideoCorePlaylistState = {
    type: NonNullable<NativePlayerState["playbackInfo"]>["streamType"]
    episodes: Anime_Episode[]
    previousEpisode: Anime_Episode | null
    nextEpisode: Anime_Episode | null
    currentEpisode: Anime_Episode
}

export const vc_playlistState = atom<VideoCorePlaylistState | null>(null)

// call once, maintains playlist state
export function useVideoCorePlaylistSetup() {
    const [playlistState, setPlaylistState] = useAtom(vc_playlistState)

    const state = useAtomValue(nativePlayer_stateAtom)
    const playbackInfo = state.playbackInfo
    const mediaId = state.playbackInfo?.media?.id
    const mediaType = state.playbackInfo?.streamType

    const currProgressNumber = playbackInfo?.episode?.progressNumber || 0

    const { data: animeEntry } = useGetAnimeEntry(!!mediaId ? mediaId : null)

    const episodes = React.useMemo(() => {
        if (!animeEntry?.episodes) return null

        return animeEntry.episodes.filter(ep => ep.type === "main")
    }, [animeEntry?.episodes, currProgressNumber])

    const currentEpisode = episodes?.find?.(ep => ep.progressNumber === currProgressNumber) ?? null
    const previousEpisode = episodes?.find?.(ep => ep.progressNumber === currProgressNumber - 1) ?? null
    const nextEpisode = episodes?.find?.(ep => ep.progressNumber === currProgressNumber + 1) ?? null

    React.useEffect(() => {
        if (!playbackInfo || !currentEpisode || !episodes?.length) {
            setPlaylistState(null)
            return
        }

        setPlaylistState({
            type: mediaType!,
            episodes: episodes ?? [],
            currentEpisode,
            previousEpisode,
            nextEpisode,
        })
    }, [playbackInfo, currentEpisode, previousEpisode, nextEpisode])
}

export function useVideoCorePlaylist() {
    const playlistState = useAtomValue(vc_playlistState)

    const playNextEpisode = () => {

    }
}
