import { Anime_MediaEntryEpisode } from "@/api/generated/types"
import { atom } from "jotai/index"
import { useAtom } from "jotai/react"

const __torrentStreaming_selectedEpisodeAtom = atom<Anime_MediaEntryEpisode | null>(null)

export function useTorrentStreamingSelectedEpisode() {
    const [value, setter] = useAtom(__torrentStreaming_selectedEpisodeAtom)

    return {
        torrentStreamingSelectedEpisode: value,
        setTorrentStreamingSelectedEpisode: setter,
    }
}
