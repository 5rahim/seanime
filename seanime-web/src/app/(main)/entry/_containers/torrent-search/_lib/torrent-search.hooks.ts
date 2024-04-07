import { MediaEntry, MediaEntryDownloadInfo } from "@/app/(main)/(library)/_lib/anime-library.types"
import { TorrentSearchData } from "@/app/(main)/entry/_containers/torrent-search/_lib/torrent.types"
import { __torrentSearch_selectedTorrentsAtom } from "@/app/(main)/entry/_containers/torrent-search/torrent-search-container"
import { torrentSearchDrawerEpisodeAtom } from "@/app/(main)/entry/_containers/torrent-search/torrent-search-drawer"
import { useDebounceWithSet } from "@/hooks/use-debounce"
import { SeaEndpoints } from "@/lib/server/endpoints"
import { useSeaQuery } from "@/lib/server/query"
import { useAtom } from "jotai/react"
import React, { startTransition } from "react"

type TorrentSearchHookProps = {
    hasEpisodesToDownload: boolean
    shouldLookForBatches: boolean
    downloadInfo: MediaEntryDownloadInfo | undefined
    entry: MediaEntry | undefined
    isAdult: boolean
}

export function useTorrentSearch(props: TorrentSearchHookProps) {

    const {
        hasEpisodesToDownload,
        shouldLookForBatches,
        downloadInfo,
        entry,
        isAdult,
    } = props

    const [soughtEpisode, setSoughtEpisode] = useAtom(torrentSearchDrawerEpisodeAtom)

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
    const { data: _data, isLoading: _isLoading, isFetching: _isFetching } = useSeaQuery<TorrentSearchData | undefined>({
        endpoint: SeaEndpoints.TORRENT_SEARCH,
        queryKey: ["torrent-search", entry?.mediaId, dSmartSearchEpisode, globalFilter, smartSearchBatch, smartSearchResolution, smartSearch,
            downloadInfo?.absoluteOffset, smartSearchBest],
        method: "post",
        data: {
            query: globalFilter,
            episodeNumber: dSmartSearchEpisode,
            batch: smartSearchBatch,
            media: entry?.media,
            absoluteOffset: downloadInfo?.absoluteOffset || 0,
            resolution: smartSearchResolution,
            smartSearch: smartSearch,
            best: smartSearch && smartSearchBest,
        },
        refetchOnWindowFocus: false,
        retry: 0,
        retryDelay: 1000,
        enabled: !(smartSearchEpisode === undefined && globalFilter.length === 0) && !isAdult,
    })

    /**
     * Fetch NSFW torrent search data
     */
    const { data: _nsfw_data, isLoading: _nsfw_isLoading, isFetching: _nsfw_isFetching } = useSeaQuery<TorrentSearchData | undefined>({
        endpoint: SeaEndpoints.TORRENT_NSFW_SEARCH,
        queryKey: ["torrent-nsfw-search", globalFilter],
        method: "post",
        data: {
            query: globalFilter,
        },
        refetchOnWindowFocus: false,
        retry: 0,
        retryDelay: 1000,
        enabled: isAdult,
    })

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
