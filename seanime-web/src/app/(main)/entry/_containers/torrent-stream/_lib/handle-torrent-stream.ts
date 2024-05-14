import { Anime_MediaEntry, Torrent_AnimeTorrent } from "@/api/generated/types"
import React from "react"

type ManualTorrentStreamSelectionProps = {
    torrent: Torrent_AnimeTorrent
    entry: Anime_MediaEntry
    episodeNumber: number
}

export function useHandleStartTorrentStream() {


    const handleManualTorrentStreamSelection = React.useCallback((params: ManualTorrentStreamSelectionProps) => {

    }, [])

    return {
        handleManualTorrentStreamSelection,
    }
}
