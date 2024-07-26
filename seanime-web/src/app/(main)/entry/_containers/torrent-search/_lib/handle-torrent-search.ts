import { Anime_AnimeEntry, Anime_AnimeEntryDownloadInfo } from "@/api/generated/types"
import { useAnimeListTorrentProviderExtensions } from "@/api/hooks/extensions.hooks"
import { useSearchTorrent } from "@/api/hooks/torrent_search.hooks"
import { useServerStatus } from "@/app/(main)/_hooks/use-server-status"
import { __torrentSearch_selectedTorrentsAtom } from "@/app/(main)/entry/_containers/torrent-search/torrent-search-container"
import { __torrentSearch_drawerEpisodeAtom, TorrentSelectionType } from "@/app/(main)/entry/_containers/torrent-search/torrent-search-drawer"
import { useDebounceWithSet } from "@/hooks/use-debounce"
import { logger } from "@/lib/helpers/debug"
import { useAtom } from "jotai/react"
import React, { startTransition } from "react"

type TorrentSearchHookProps = {
    hasEpisodesToDownload: boolean
    shouldLookForBatches: boolean
    downloadInfo: Anime_AnimeEntryDownloadInfo | undefined
    entry: Anime_AnimeEntry | undefined
    isAdult: boolean
    type: TorrentSelectionType
}

export const enum Torrent_SearchType {
    SMART = "smart",
    SIMPLE = "simple",
}

export function useHandleTorrentSearch(props: TorrentSearchHookProps) {

    const {
        hasEpisodesToDownload,
        shouldLookForBatches,
        downloadInfo,
        entry,
        isAdult,
    } = props

    const serverStatus = useServerStatus()

    const { data: providerExtensions } = useAnimeListTorrentProviderExtensions()

    // Get the selected provider extension
    const defaultProviderExtension = React.useMemo(() => {
        return providerExtensions?.find(ext => ext.id === serverStatus?.settings?.library?.torrentProvider)
    }, [serverStatus?.settings?.library?.torrentProvider, providerExtensions])

    // Gives the ability to change the selected provider extension
    const [selectedProviderExtensionId, setSelectedProviderExtensionId] = React.useState(defaultProviderExtension?.id || "none")

    // Update the selected provider only when the default provider changes
    React.useLayoutEffect(() => {
        setSelectedProviderExtensionId(defaultProviderExtension?.id || "none")
    }, [defaultProviderExtension])

    // Get the selected provider extension
    const selectedProviderExtension = React.useMemo(() => {
        return providerExtensions?.find(ext => ext.id === selectedProviderExtensionId)
    }, [selectedProviderExtensionId, providerExtensions])

    const [soughtEpisode, setSoughtEpisode] = useAtom(__torrentSearch_drawerEpisodeAtom)

    // Smart search is not enabled for adult content
    const [searchType, setSearchType] = React.useState(!isAdult ? Torrent_SearchType.SMART : Torrent_SearchType.SIMPLE)

    const [globalFilter, setGlobalFilter] = React.useState<string>(hasEpisodesToDownload ? "" : (entry?.media?.title?.romaji || ""))
    const [selectedTorrents, setSelectedTorrents] = useAtom(__torrentSearch_selectedTorrentsAtom)
    const [smartSearchBatch, setSmartSearchBatch] = React.useState<boolean>(shouldLookForBatches || false)
    const [smartSearchEpisode, setSmartSearchEpisode] = React.useState<number>(downloadInfo?.episodesToDownload?.[0]?.episode?.episodeNumber || 1)
    const [smartSearchResolution, setSmartSearchResolution] = React.useState("1080")
    const [smartSearchBest, setSmartSearchBest] = React.useState(false)
    const [dSmartSearchEpisode, setDSmartSearchEpisode] = useDebounceWithSet(smartSearchEpisode, 500)

    const warnings = {
        noProvider: !selectedProviderExtension,
        extensionDoesNotSupportAdult: isAdult && selectedProviderExtension && !selectedProviderExtension?.settings?.supportsAdult,
        extensionDoesNotSupportSmartSearch: searchType === Torrent_SearchType.SMART && selectedProviderExtension && !selectedProviderExtension?.settings?.canSmartSearch,
        extensionDoesNotSupportBestRelease: smartSearchBest && selectedProviderExtension && !selectedProviderExtension?.settings?.smartSearchFilters?.includes(
            "bestReleases"),
        extensionDoesNotSupportBatchSearch: smartSearchBatch && selectedProviderExtension && !selectedProviderExtension?.settings?.smartSearchFilters?.includes(
            "batch"),
    }

    // Change fields based on selected provider
    React.useLayoutEffect(() => {
        if (searchType === Torrent_SearchType.SMART && warnings.extensionDoesNotSupportSmartSearch) {
            setSearchType(Torrent_SearchType.SIMPLE)
        }
    }, [warnings.extensionDoesNotSupportSmartSearch, selectedProviderExtensionId, searchType])
    React.useLayoutEffect(() => {
        if (smartSearchBest && warnings.extensionDoesNotSupportBestRelease) {
            setSmartSearchBest(false)
        }
    }, [warnings.extensionDoesNotSupportBestRelease, selectedProviderExtensionId, smartSearchBest])
    React.useLayoutEffect(() => {
        if (smartSearchBatch && warnings.extensionDoesNotSupportBatchSearch) {
            setSmartSearchBatch(false)
        }
    }, [warnings.extensionDoesNotSupportBatchSearch, selectedProviderExtensionId, smartSearchBatch])

    /**
     * Fetch torrent search data
     */
    const { data: _data, isLoading: _isLoading, isFetching: _isFetching } = useSearchTorrent({
        query: globalFilter.trim().toLowerCase(),
            episodeNumber: dSmartSearchEpisode,
            batch: smartSearchBatch,
            media: entry?.media,
            absoluteOffset: downloadInfo?.absoluteOffset || 0,
            resolution: smartSearchResolution,
            type: searchType,
            provider: selectedProviderExtension?.id!,
            bestRelease: searchType === Torrent_SearchType.SMART && smartSearchBest,
        },
        !(searchType === Torrent_SearchType.SIMPLE && globalFilter.length === 0) // If simple search, user input must not be empty
        && !warnings.noProvider
        && !warnings.extensionDoesNotSupportAdult
        && !warnings.extensionDoesNotSupportSmartSearch
        && !warnings.extensionDoesNotSupportBestRelease,
    )

    React.useLayoutEffect(() => {
        if (soughtEpisode !== undefined) {
            setSmartSearchEpisode(soughtEpisode)
            setDSmartSearchEpisode(soughtEpisode)
            startTransition(() => {
                setSoughtEpisode(undefined)
            })
        }
    }, [soughtEpisode])

    // const data = React.useMemo(() => isAdult ? _nsfw_data : _data, [_data, _nsfw_data])
    // const isLoading = React.useMemo(() => isAdult ? _nsfw_isLoading : _isLoading, [_isLoading, _nsfw_isLoading])
    // const isFetching = React.useMemo(() => isAdult ? _nsfw_isFetching : _isFetching, [_isFetching, _nsfw_isFetching])

    React.useEffect(() => {
        logger("Torrent Provider").info(warnings)
    }, [warnings])

    return {
        warnings,
        hasOneWarning: Object.values(warnings).some(w => w),
        providerExtensions,
        selectedProviderExtension,
        selectedProviderExtensionId,
        setSelectedProviderExtensionId,
        globalFilter,
        setGlobalFilter,
        selectedTorrents,
        setSelectedTorrents,
        searchType,
        setSearchType,
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
        data: _data,
        isLoading: _isLoading,
        isFetching: _isFetching,
    }

}
