import { Anime_Entry, Anime_EntryDownloadInfo } from "@/api/generated/types"
import { useAnimeListTorrentProviderExtensions } from "@/api/hooks/extensions.hooks"
import { useSearchTorrent } from "@/api/hooks/torrent_search.hooks"
import { useServerStatus } from "@/app/(main)/_hooks/use-server-status"
import { __torrentSearch_selectedTorrentsAtom } from "@/app/(main)/entry/_containers/torrent-search/torrent-search-container"
import { __torrentSearch_selectionEpisodeAtom, TorrentSelectionType } from "@/app/(main)/entry/_containers/torrent-search/torrent-search-drawer"
import { useDebounceWithTrigger } from "@/hooks/use-debounce"
import { logger } from "@/lib/helpers/debug"
import { TORRENT_PROVIDER } from "@/lib/server/settings"
import { useAtom } from "jotai/react"
import { atomWithStorage } from "jotai/utils"
import React, { startTransition } from "react"

type TorrentSearchHookProps = {
    hasEpisodesToDownload: boolean
    shouldLookForBatches: boolean
    downloadInfo: Anime_EntryDownloadInfo | undefined
    entry: Anime_Entry | undefined
    isAdult: boolean
    type: TorrentSelectionType
}

export const enum Torrent_SearchType {
    SMART = "smart",
    SIMPLE = "simple",
}

export const __torrentSearch_searchAcrossProvidersAtom = atomWithStorage("sea-torrent-search-across-providers", false, undefined, { getOnInit: true })
export const __torrentSearch_extraProviderIdsAtom = atomWithStorage<string[]>("sea-torrent-search-extra-provider-ids",
    [],
    undefined,
    { getOnInit: true })

const retryDelays = [60_000, 120_000, 300_000]

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
        if (serverStatus?.settings?.library?.torrentProvider === TORRENT_PROVIDER.NONE) {
            return undefined
        }

        const defaultExt = providerExtensions?.find(ext => ext.id === serverStatus?.settings?.library?.torrentProvider)
        if (!defaultExt) {
            return providerExtensions?.[0]
        }
        return defaultExt
    }, [serverStatus?.settings?.library?.torrentProvider, providerExtensions])

    // Gives the ability to change the selected provider extension
    const [selectedProviderExtensionId, setSelectedProviderExtensionId] = React.useState(defaultProviderExtension?.id || "none")

    // Update the selected provider only when the default provider changes
    React.useLayoutEffect(() => {
        setSelectedProviderExtensionId(defaultProviderExtension?.id || "none")
    }, [defaultProviderExtension?.id])

    // Get the selected provider extension
    const selectedProviderExtension = React.useMemo(() => {
        return providerExtensions?.find(ext => ext.id === selectedProviderExtensionId)
    }, [selectedProviderExtensionId, providerExtensions])

    const [soughtEpisode, setSoughtEpisode] = useAtom(__torrentSearch_selectionEpisodeAtom)

    // Smart search is not enabled for adult content
    const [searchType, setSearchType] = React.useState(!isAdult ? Torrent_SearchType.SMART : Torrent_SearchType.SIMPLE)

    const {
        value: globalFilter,
        debouncedValue: debouncedGlobalFilter,
        setValue: setGlobalFilter,
        triggerImmediate: triggerImmediateSearch,
    } = useDebounceWithTrigger(hasEpisodesToDownload ? "" : (entry?.media?.title?.romaji || ""), 500)
    const [selectedTorrents, setSelectedTorrents] = useAtom(__torrentSearch_selectedTorrentsAtom)
    const [searchAcrossProviders, setSearchAcrossProviders] = useAtom(__torrentSearch_searchAcrossProvidersAtom)
    const [extraProviderIds, setExtraProviderIds] = useAtom(__torrentSearch_extraProviderIdsAtom)
    const [smartSearchBatch, setSmartSearchBatch] = React.useState<boolean>(shouldLookForBatches || false)
    // const [smartSearchEpisode, setSmartSearchEpisode] = React.useState<number>(downloadInfo?.episodesToDownload?.[0]?.episode?.episodeNumber || 1)
    const [smartSearchResolution, setSmartSearchResolution] = React.useState("")
    const [smartSearchBest, setSmartSearchBest] = React.useState(false)
    const {
        value: smartSearchEpisode,
        debouncedValue: debouncedSmartSearchEpisode,
        setValue: setSmartSearchEpisode,
        triggerImmediate: triggerImmediateEpisode,
    } = useDebounceWithTrigger(downloadInfo?.episodesToDownload?.[0]?.episode?.episodeNumber ?? 1, 500)

    const activeExtraProviderIds = React.useMemo(() => {
        const validProviderIds = new Set(providerExtensions?.map(ext => ext.id) ?? [])
        return extraProviderIds.filter((id, idx) => {
            return id !== selectedProviderExtensionId && validProviderIds.has(id) && extraProviderIds.indexOf(id) === idx
        })
    }, [extraProviderIds, providerExtensions, selectedProviderExtensionId])

    const searchProvider = React.useMemo(() => {
        if (!selectedProviderExtension?.id) return ""
        if (!searchAcrossProviders) return selectedProviderExtension.id
        return [selectedProviderExtension.id, ...activeExtraProviderIds].join(",")
    }, [activeExtraProviderIds, searchAcrossProviders, selectedProviderExtension?.id])

    const warnings = {
        noProvider: !selectedProviderExtension,
        extensionDoesNotSupportAdult: isAdult && selectedProviderExtension && !selectedProviderExtension?.settings?.supportsAdult,
        extensionDoesNotSupportSmartSearch: searchType === Torrent_SearchType.SMART && selectedProviderExtension && !selectedProviderExtension?.settings?.canSmartSearch,
        extensionDoesNotSupportBestRelease: smartSearchBest && selectedProviderExtension && !selectedProviderExtension?.settings?.smartSearchFilters?.includes(
            "bestReleases"),
        extensionDoesNotSupportBatchSearch: smartSearchBatch && selectedProviderExtension && !selectedProviderExtension?.settings?.smartSearchFilters?.includes(
            "batch"),
    }

    // Change fields when changing the selected provider - i.e. when [selectedProviderExtensionId] changes
    React.useLayoutEffect(() => {
        // If the selected provider supports smart search, enable it if it's not already enabled
        if (searchType === Torrent_SearchType.SIMPLE && selectedProviderExtension?.settings?.canSmartSearch) {
            setSearchType(Torrent_SearchType.SMART)
        }
    }, [searchType && warnings.extensionDoesNotSupportSmartSearch, selectedProviderExtension?.settings?.canSmartSearch, selectedProviderExtensionId])
    React.useLayoutEffect(() => {
        // If the selected provider does not support smart search, disable it
        if (searchType === Torrent_SearchType.SMART && warnings.extensionDoesNotSupportSmartSearch) {
            setSearchType(Torrent_SearchType.SIMPLE)
        }
    }, [warnings.extensionDoesNotSupportSmartSearch, selectedProviderExtensionId, searchType])
    React.useLayoutEffect(() => {
        // If the selected provider does not support best release, disable it
        if (smartSearchBest && warnings.extensionDoesNotSupportBestRelease) {
            setSmartSearchBest(false)
        }
    }, [warnings.extensionDoesNotSupportBestRelease, selectedProviderExtensionId, smartSearchBest])
    React.useLayoutEffect(() => {
        // If the selected provider does not support batch search, disable it
        if (smartSearchBatch && warnings.extensionDoesNotSupportBatchSearch) {
            setSmartSearchBatch(false)
        }
    }, [warnings.extensionDoesNotSupportBatchSearch, selectedProviderExtensionId, smartSearchBatch])

    React.useEffect(() => {
        console.log("globalFilter", globalFilter)
    }, [globalFilter])

    console.log("smartSearchResolution", smartSearchResolution)

    /**
     * Fetch torrent search data
     */
    const searchVariables = {
        query: debouncedGlobalFilter.trim().toLowerCase(),
        episodeNumber: debouncedSmartSearchEpisode,
        batch: smartSearchBatch,
        media: entry?.media,
        absoluteOffset: downloadInfo?.absoluteOffset || 0,
        resolution: smartSearchResolution,
        type: searchType,
        provider: searchProvider,
        bestRelease: searchType === Torrent_SearchType.SMART && smartSearchBest,
    }
    const searchEnabled = !(searchType === Torrent_SearchType.SIMPLE && debouncedGlobalFilter.length === 0) // If simple search, user input must not be empty
        && !warnings.noProvider
        && !warnings.extensionDoesNotSupportAdult
        && !warnings.extensionDoesNotSupportSmartSearch
        && !warnings.extensionDoesNotSupportBestRelease
        && !!providerExtensions // Provider extensions must be loaded
    const { data: _data, isLoading: _isLoading, isFetching: _isFetching, isError: _isError, refetch } = useSearchTorrent(
        searchVariables,
        searchEnabled,
    )

    const hasResults = searchType === Torrent_SearchType.SMART
        ? !!_data?.previews?.length
        : !!_data?.torrents?.length
    const canAutoRetry = searchEnabled
        && entry?.media?.status === "RELEASING"
        && !!_data
        && !hasResults
        && !_isError
    const searchKey = JSON.stringify(searchVariables)
    const [retryAttempt, setRetryAttempt] = React.useState(0)
    const [nextRetryAt, setNextRetryAt] = React.useState<number>()
    const [autoRetrySeconds, setAutoRetrySeconds] = React.useState<number>()

    React.useEffect(() => {
        setRetryAttempt(0)
        setNextRetryAt(undefined)
    }, [searchKey])

    React.useEffect(() => {
        if (!canAutoRetry || _isFetching) {
            setNextRetryAt(undefined)
            return
        }

        const delay = retryDelays[Math.min(retryAttempt, retryDelays.length - 1)]
        setNextRetryAt(Date.now() + delay)
        const timer = window.setTimeout(() => {
            setNextRetryAt(undefined)
            setRetryAttempt(current => Math.min(current + 1, retryDelays.length - 1))
            refetch()
        }, delay)

        return () => window.clearTimeout(timer)
    }, [canAutoRetry, _isFetching, refetch, retryAttempt, searchKey])

    React.useEffect(() => {
        if (!nextRetryAt) {
            setAutoRetrySeconds(undefined)
            return
        }

        const update = () => setAutoRetrySeconds(Math.max(0, Math.ceil((nextRetryAt - Date.now()) / 1000)))
        update()
        const timer = window.setInterval(update, 1000)
        return () => window.clearInterval(timer)
    }, [nextRetryAt])

    React.useLayoutEffect(() => {
        if (soughtEpisode !== undefined) {
            setSmartSearchEpisode(soughtEpisode)
            startTransition(() => {
                setSoughtEpisode(undefined)
            })
        }
    }, [soughtEpisode])

    // const data = React.useMemo(() => isAdult ? _nsfw_data : _data, [_data, _nsfw_data])
    // const isLoading = React.useMemo(() => isAdult ? _nsfw_isLoading : _isLoading, [_isLoading, _nsfw_isLoading])
    // const isFetching = React.useMemo(() => isAdult ? _nsfw_isFetching : _isFetching, [_isFetching, _nsfw_isFetching])

    React.useEffect(() => {
        logger("TORRENT SEARCH").info({ warnings })
    }, [warnings])
    React.useEffect(() => {
        logger("TORRENT SEARCH").info({ selectedProviderExtension })
    }, [warnings])
    React.useEffect(() => {
        logger("TORRENT SEARCH").info({
            globalFilter,
            searchType,
            smartSearchBatch,
            smartSearchEpisode,
            smartSearchResolution,
            smartSearchBest,
            debouncedSmartSearchEpisode,
            searchProvider,
        })
        },
        [globalFilter, searchType, smartSearchBatch, smartSearchEpisode, smartSearchResolution, smartSearchBest, debouncedSmartSearchEpisode,
            searchProvider])

    return {
        warnings,
        hasOneWarning: Object.values(warnings).some(w => w),
        providerExtensions,
        selectedProviderExtension,
        selectedProviderExtensionId,
        setSelectedProviderExtensionId,
        globalFilter,
        setGlobalFilter,
        debouncedGlobalFilter,
        triggerImmediateSearch,
        selectedTorrents,
        setSelectedTorrents,
        searchAcrossProviders,
        setSearchAcrossProviders,
        extraProviderIds,
        setExtraProviderIds,
        searchType,
        setSearchType,
        smartSearchBatch,
        setSmartSearchBatch,
        smartSearchEpisode,
        setSmartSearchEpisode,
        debouncedSmartSearchEpisode,
        triggerImmediateEpisode,
        smartSearchResolution,
        setSmartSearchResolution,
        smartSearchBest,
        setSmartSearchBest,
        soughtEpisode,
        data: _data,
        isLoading: _isLoading,
        isFetching: _isFetching,
        isError: _isError,
        isAutoRetrying: canAutoRetry && _isFetching && !_isLoading,
        autoRetrySeconds,
        refetch,
    }

}
