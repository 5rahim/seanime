import { Anime_MediaEntry, Anime_MediaEntryDownloadInfo } from "@/api/generated/types"
import { useSearchNsfwTorrent, useSearchTorrent } from "@/api/hooks/torrent_search.hooks"
import { __torrentSearch_selectedTorrentsAtom } from "@/app/(main)/entry/_containers/torrent-search/torrent-search-container"
import { __torrentSearch_drawerEpisodeAtom } from "@/app/(main)/entry/_containers/torrent-search/torrent-search-drawer"
import { useDebounceWithSet } from "@/hooks/use-debounce"
import { useAtom } from "jotai/react"
import React, { startTransition } from "react"

type TorrentSearchHookProps = {
    hasEpisodesToDownload: boolean
    shouldLookForBatches: boolean
    downloadInfo: Anime_MediaEntryDownloadInfo | undefined
    entry: Anime_MediaEntry | undefined
    isAdult: boolean
}

export function useHandleTorrentSearch(props: TorrentSearchHookProps) {

    const {
        hasEpisodesToDownload,
        shouldLookForBatches,
        downloadInfo,
        entry,
        isAdult,
    } = props

    const [soughtEpisode, setSoughtEpisode] = useAtom(__torrentSearch_drawerEpisodeAtom)

    // Smart search is not enabled for adult content
    const [smartSearch, setSmartSearch] = React.useState(!isAdult)

    const [globalFilter, setGlobalFilter] = React.useState<string>(hasEpisodesToDownload ? "" : (entry?.media?.title?.romaji || ""))
    const [selectedTorrents, setSelectedTorrents] = useAtom(__torrentSearch_selectedTorrentsAtom)
    const [smartSearchBatch, setSmartSearchBatch] = React.useState<boolean>(shouldLookForBatches || false)
    const [smartSearchEpisode, setSmartSearchEpisode] = React.useState<number>(downloadInfo?.episodesToDownload?.[0]?.episode?.episodeNumber || 1)
    const [smartSearchResolution, setSmartSearchResolution] = React.useState("1080")
    const [smartSearchBest, setSmartSearchBest] = React.useState(false)
    const [dSmartSearchEpisode, setDSmartSearchEpisode] = useDebounceWithSet(smartSearchEpisode, 500)

    /**
     * Fetch torrent search data
     */
    const { data: _data, isLoading: _isLoading, isFetching: _isFetching } = useSearchTorrent({
        query: globalFilter,
        episodeNumber: dSmartSearchEpisode,
        batch: smartSearchBatch,
        media: entry?.media,
        absoluteOffset: downloadInfo?.absoluteOffset || 0,
        resolution: smartSearchResolution,
        smartSearch: smartSearch,
        best: smartSearch && smartSearchBest,
    }, !(smartSearchEpisode === undefined && globalFilter.length === 0) && !isAdult)

    /**
     * Fetch NSFW torrent search data
     */
    const { data: _nsfw_data, isLoading: _nsfw_isLoading, isFetching: _nsfw_isFetching } = useSearchNsfwTorrent({
        query: globalFilter,
    }, isAdult)

    React.useLayoutEffect(() => {
        if (soughtEpisode !== undefined) {
            setSmartSearchEpisode(soughtEpisode)
            setDSmartSearchEpisode(soughtEpisode)
            startTransition(() => {
                setSoughtEpisode(undefined)
            })
        }
    }, [soughtEpisode])

    const data = React.useMemo(() => isAdult ? _nsfw_data : _data, [_data, _nsfw_data])
    const isLoading = React.useMemo(() => isAdult ? _nsfw_isLoading : _isLoading, [_isLoading, _nsfw_isLoading])
    const isFetching = React.useMemo(() => isAdult ? _nsfw_isFetching : _isFetching, [_isFetching, _nsfw_isFetching])

    return {
        globalFilter,
        setGlobalFilter,
        selectedTorrents,
        setSelectedTorrents,
        smartSearch,
        setSmartSearch,
        smartSearchBatch,
        setSmartSearchBatch,
        smartSearchEpisode,
        setSmartSearchEpisode,
        smartSearchResolution,
        setSmartSearchResolution,
        smartSearchBest,
        setSmartSearchBest,
        dSmartSearchEpisode,
        setDSmartSearchEpisode,
        soughtEpisode,
        data,
        isLoading,
        isFetching,
    }

}
