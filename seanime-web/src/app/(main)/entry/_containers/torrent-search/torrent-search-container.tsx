import { Anime_Entry, Debrid_TorrentItemInstantAvailability, HibikeTorrent_AnimeTorrent } from "@/api/generated/types"
import { useGetTorrentstreamBatchHistory } from "@/api/hooks/torrentstream.hooks"
import { EpisodeCard } from "@/app/(main)/_features/anime/_components/episode-card"
import { useAutoPlaySelectedTorrent } from "@/app/(main)/_features/autoplay/autoplay"
import { useServerStatus } from "@/app/(main)/_hooks/use-server-status"
import { useHandleStartDebridStream } from "@/app/(main)/entry/_containers/debrid-stream/_lib/handle-debrid-stream"
import { DebridStreamFileSelectionModal } from "@/app/(main)/entry/_containers/debrid-stream/debrid-stream-file-selection-modal"
import { TorrentListItem } from "@/app/(main)/entry/_containers/torrent-search/_components/torrent-preview-item"
import { TorrentPreviewList } from "@/app/(main)/entry/_containers/torrent-search/_components/torrent-preview-list"
import { TorrentTable } from "@/app/(main)/entry/_containers/torrent-search/_components/torrent-table"
import { Torrent_SearchType, useHandleTorrentSearch } from "@/app/(main)/entry/_containers/torrent-search/_lib/handle-torrent-search"
import { useTorrentSearchSelectedStreamEpisode } from "@/app/(main)/entry/_containers/torrent-search/_lib/handle-torrent-selection"
import { TorrentDownloadFileSelection } from "@/app/(main)/entry/_containers/torrent-search/torrent-download-file-selection"
import { TorrentDownloadModal } from "@/app/(main)/entry/_containers/torrent-search/torrent-download-modal"
import { __torrentSearch_selectionAtom, TorrentSelectionType } from "@/app/(main)/entry/_containers/torrent-search/torrent-search-drawer"
import { useHandleStartTorrentStream } from "@/app/(main)/entry/_containers/torrent-stream/_lib/handle-torrent-stream"
import {
    __torrentSearch_fileSelectionTorrentAtom,
    TorrentstreamFileSelectionModal,
} from "@/app/(main)/entry/_containers/torrent-stream/torrent-stream-file-selection-modal"
import { LuffyError } from "@/components/shared/luffy-error"
import { SeaLink } from "@/components/shared/sea-link"
import { Alert } from "@/components/ui/alert"
import { AppLayoutStack } from "@/components/ui/app-layout"
import { Button } from "@/components/ui/button"
import { cn } from "@/components/ui/core/styling"
import { DataGridSearchInput } from "@/components/ui/datagrid"
import { NumberInput } from "@/components/ui/number-input"
import { Select } from "@/components/ui/select"
import { Skeleton } from "@/components/ui/skeleton"
import { Switch } from "@/components/ui/switch"
import { TextInput } from "@/components/ui/text-input"
import { TORRENT_PROVIDER } from "@/lib/server/settings"
import { subDays, subMonths } from "date-fns"
import { atom, useSetAtom } from "jotai"
import React, { startTransition } from "react"
import { FiSearch } from "react-icons/fi"
import { LuCornerLeftDown, LuFileSearch, LuPlus } from "react-icons/lu"

export const __torrentSearch_selectedTorrentsAtom = atom<HibikeTorrent_AnimeTorrent[]>([])

export function TorrentSearchContainer({ type, entry }: { type: TorrentSelectionType, entry: Anime_Entry }) {
    const downloadInfo = React.useMemo(() => entry.downloadInfo, [entry.downloadInfo])
    const serverStatus = useServerStatus()

    const shouldLookForBatches = React.useMemo(() => {
        const endedDate = entry.media?.endDate?.year ? new Date(entry.media?.endDate?.year,
            entry.media?.endDate?.month ? entry.media?.endDate?.month - 1 : 0,
            entry.media?.endDate?.day || 0) : null
        const now = new Date()
        let flag = true

        if (type === "download") {
            if (endedDate && subDays(now, 6) < endedDate) {
                flag = false
            }
            return !!downloadInfo?.canBatch && !!downloadInfo?.episodesToDownload?.length && flag
        } else {
            if (endedDate && subMonths(now, 1) < endedDate) {
                flag = false
            }
            return !!downloadInfo?.canBatch && flag
        }
    }, [downloadInfo?.canBatch, downloadInfo?.episodesToDownload?.length, type, entry.media?.endDate])

    const hasEpisodesToDownload = React.useMemo(() => !!downloadInfo?.episodesToDownload?.length, [downloadInfo?.episodesToDownload?.length])
    const [isAdult, setIsAdult] = React.useState(entry.media?.isAdult === true)

    const {
        warnings,
        hasOneWarning,
        selectedProviderExtension,
        selectedProviderExtensionId,
        setSelectedProviderExtensionId,
        providerExtensions,
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
        data,
        isLoading,
        isFetching,
        soughtEpisode,
    } = useHandleTorrentSearch({
        isAdult: false,
        hasEpisodesToDownload,
        shouldLookForBatches,
        downloadInfo,
        entry,
        type,
    })

    React.useEffect(() => {
        setSelectedTorrents([])
    }, [])

    React.useLayoutEffect(() => {
        if (searchType === Torrent_SearchType.SMART) {
            setGlobalFilter("")
        } else if (searchType === Torrent_SearchType.SIMPLE) {
            const title = entry.media?.title?.romaji || entry.media?.title?.english || entry.media?.title?.userPreferred
            setGlobalFilter(title?.replaceAll(":", "").replaceAll("-", "") || "")
        }
    }, [searchType, entry.media?.title])

    const torrents = React.useMemo(() => data?.torrents ?? [], [data?.torrents])
    const previews = React.useMemo(() => data?.previews ?? [], [data?.previews])
    const debridInstantAvailability = React.useMemo(() => serverStatus?.debridSettings?.enabled ? data?.debridInstantAvailability ?? {} : {},
        [data?.debridInstantAvailability, serverStatus?.debridSettings?.enabled])

    /**
     * Select torrent
     * - Download: Select multiple torrents
     * - Select: Select only one torrent
     */
    const handleToggleTorrent = React.useCallback((t: HibikeTorrent_AnimeTorrent) => {
        if (type === "download") {
            setSelectedTorrents(prev => {
                const idx = prev.findIndex(n => n.link === t.link)
                if (idx !== -1) {
                    return prev.filter(n => n.link !== t.link)
                }
                return [...prev, t]
            })
        } else {
            setSelectedTorrents(prev => {
                const idx = prev.findIndex(n => n.link === t.link)
                if (idx !== -1) {
                    return []
                }
                return [t]
            })
        }
    }, [setSelectedTorrents, smartSearchBest, type])

    /**
     * Handle streams
     */

    const { torrentSearchStreamEpisode } = useTorrentSearchSelectedStreamEpisode()


    return (
        <>
            <AppLayoutStack
                className={cn(
                    "space-y-4",
                    type !== "download" && "xl:space-y-0 xl:grid xl:grid-cols-[35%,1fr] xl:gap-4",
                )} data-torrent-search-container
            >

                <div
                    className="space-y-3"
                    data-torrent-search-container-param-container
                >

                    {(type !== "download" && torrentSearchStreamEpisode) &&
                        <div className="hidden xl:block space-y-3" data-torrent-search-container-stream-episode>
                            <h4 className="!mb-4">
                                Select a torrent to stream
                            </h4>
                            <EpisodeCard
                                image={torrentSearchStreamEpisode.episodeMetadata?.image || torrentSearchStreamEpisode.baseAnime?.bannerImage || torrentSearchStreamEpisode.baseAnime?.coverImage?.extraLarge}
                                topTitle={torrentSearchStreamEpisode.episodeTitle || torrentSearchStreamEpisode?.baseAnime?.title?.userPreferred}
                                title={torrentSearchStreamEpisode.displayTitle}
                                isInvalid={torrentSearchStreamEpisode.isInvalid}
                                progressTotal={torrentSearchStreamEpisode.baseAnime?.episodes}
                                progressNumber={torrentSearchStreamEpisode.progressNumber}
                                episodeNumber={torrentSearchStreamEpisode.episodeNumber}
                                length={torrentSearchStreamEpisode.episodeMetadata?.length}
                                actionIcon={null}
                                anime={{
                                    id: torrentSearchStreamEpisode.baseAnime?.id,
                                    image: torrentSearchStreamEpisode.baseAnime?.coverImage?.large,
                                    title: torrentSearchStreamEpisode.baseAnime?.title?.userPreferred,
                                }}
                            />
                        </div>}

                    <div className="flex flex-wrap gap-3 items-center w-full" data-torrent-search-main-params>
                        <div className="w-[200px]" data-torrent-search-container-param-container-provider-select-container>
                            <Select
                                name="torrentProvider"
                                // leftAddon="Torrent Provider"
                                value={selectedProviderExtension?.id ?? TORRENT_PROVIDER.NONE}
                                onValueChange={setSelectedProviderExtensionId}
                                leftIcon={<LuFileSearch />}
                                options={[
                                    ...(providerExtensions?.map(ext => ({
                                        label: ext.name,
                                        value: ext.id,
                                    })) ?? []).sort((a, b) => a?.label?.localeCompare(b?.label) ?? 0),
                                    { label: "None", value: TORRENT_PROVIDER.NONE },
                                ]}
                            />
                        </div>

                        <div
                            className="h-10 rounded-[--radius] px-2 flex items-center"
                            data-torrent-search-container-param-container-smart-search-switch-container
                        >
                            <Switch
                                // side="right"
                                label="Smart search"
                                moreHelp={selectedProviderExtension?.settings?.canSmartSearch
                                    ? "Automatically search based on given parameters"
                                    : "This provider does not support smart search"}
                                value={searchType === Torrent_SearchType.SMART}
                                onValueChange={v => setSearchType(v ? Torrent_SearchType.SMART : Torrent_SearchType.SIMPLE)}
                                disabled={!selectedProviderExtension?.settings?.canSmartSearch}
                                containerClass="flex-row-reverse gap-1"
                            />
                        </div>

                        {/*{<div*/}
                        {/*    className="h-10 rounded-[--radius] px-2 flex items-center"*/}
                        {/*    data-torrent-search-container-param-container-adult-switch-container*/}
                        {/*>*/}
                        {/*    <Switch*/}
                        {/*        // side="right"*/}
                        {/*        label="Adult"*/}
                        {/*        moreHelp="If enabled, the adult content flag will be passed to the provider."*/}
                        {/*        value={isAdult}*/}
                        {/*        onValueChange={setIsAdult}*/}
                        {/*        containerClass="flex-row-reverse gap-1"*/}
                        {/*    />*/}
                        {/*</div>}*/}
                    </div>

                    {(selectedProviderExtensionId !== "none" && selectedProviderExtensionId !== "") && <>

                        {(searchType === Torrent_SearchType.SMART) &&
                            <AppLayoutStack className="w-full" data-torrent-search-smart-search-container>
                                <div
                                    data-torrent-search-smart-search-provider-param-container
                                    className={cn(
                                        "Sea-TorrentSearchContainer__providerParamContainer flex flex-col items-center flex-wrap justify-around gap-3 md:flex-row w-full border rounded-xl p-3",
                                        {
                                            "hidden": !selectedProviderExtension?.settings?.smartSearchFilters?.includes("episodeNumber") &&
                                                !selectedProviderExtension?.settings?.smartSearchFilters?.includes("resolution")
                                                && !selectedProviderExtension?.settings?.smartSearchFilters?.includes("batch")
                                                && !selectedProviderExtension?.settings?.smartSearchFilters?.includes("bestReleases")
                                                && !selectedProviderExtension?.settings?.smartSearchFilters?.includes("search"),
                                        },
                                    )}
                                >

                                    {selectedProviderExtension?.settings?.smartSearchFilters?.includes("episodeNumber") && <NumberInput
                                        data-torrent-search-smart-search-episode-number-input
                                        label="Episode number"
                                        value={smartSearchEpisode}
                                        disabled={entry?.media?.format === "MOVIE" || smartSearchBest}
                                        onValueChange={(value) => {
                                            startTransition(() => {
                                                setSmartSearchEpisode(value)
                                            })
                                        }}
                                        formatOptions={{ useGrouping: false }}
                                        // hideControls
                                        size="sm"
                                        fieldClass={cn(
                                            "flex flex-none w-fit md:justify-end gap-3 space-y-0",
                                            { "opacity-50 cursor-not-allowed pointer-events-none": (smartSearchBatch || searchType != Torrent_SearchType.SMART) },
                                        )}
                                        fieldLabelClass={cn(
                                            "flex-none self-center font-normal !text-md sm:text-md lg:text-md",
                                        )}
                                        className="max-w-[4rem]"
                                    />}

                                    {selectedProviderExtension?.settings?.smartSearchFilters?.includes("batch") && <Switch
                                        data-torrent-search-smart-search-batch-switch
                                        label="Batches"
                                        value={smartSearchBatch}
                                        onValueChange={setSmartSearchBatch}
                                        disabled={smartSearchBest || !downloadInfo?.canBatch}
                                        fieldClass={cn(
                                            "flex flex-none w-fit",
                                            { "opacity-50 cursor-not-allowed pointer-events-none": !downloadInfo?.canBatch || smartSearchBest },
                                        )}
                                        size="sm"
                                        containerClass="flex-row-reverse gap-1"
                                    />}

                                    {selectedProviderExtension?.settings?.smartSearchFilters?.includes("resolution") && <Select
                                        data-torrent-search-smart-search-resolution-select
                                        label="Resolution"
                                        value={smartSearchResolution || "-"}
                                        onValueChange={v => setSmartSearchResolution(v != "-" ? v : "")}
                                        options={[
                                            { value: "-", label: "Any" },
                                            { value: "1080", label: "1080p" },
                                            { value: "720", label: "720p" },
                                            { value: "540", label: "540p" },
                                            { value: "480", label: "480p" },
                                            { value: "2160", label: "2160p" },
                                        ]}
                                        disabled={smartSearchBest || searchType != Torrent_SearchType.SMART}
                                        size="sm"
                                        fieldClass={cn(
                                            "flex flex-none w-fit md:justify-center gap-3 space-y-0",
                                            { "opacity-50 cursor-not-allowed pointer-events-none": searchType != Torrent_SearchType.SMART || smartSearchBest },
                                        )}
                                        fieldLabelClass="flex-none self-center font-normal !text-md sm:text-md lg:text-md"
                                        className="w-[6rem]"
                                    />}

                                    {selectedProviderExtension?.settings?.smartSearchFilters?.includes("bestReleases") && <Switch
                                        data-torrent-search-smart-search-best-releases-switch
                                        label="Best releases"
                                        value={smartSearchBest}
                                        onValueChange={setSmartSearchBest}
                                        fieldClass={cn(
                                            "flex flex-none w-fit",
                                            { "opacity-50 cursor-not-allowed pointer-events-none": !downloadInfo?.canBatch },
                                        )}
                                        size="sm"
                                        containerClass="flex-row-reverse gap-1"
                                    />}

                                </div>

                                {!hasOneWarning && selectedProviderExtension?.settings?.smartSearchFilters?.includes("query") &&
                                    <div className="py-1" data-torrent-search-smart-search-query-input-container>
                                        <DataGridSearchInput
                                            value={globalFilter ?? ""}
                                            onChange={v => setGlobalFilter(v)}
                                            placeholder={searchType === Torrent_SearchType.SMART
                                                ? `Refine the title (${entry.media?.title?.romaji})`
                                                : "Search"}
                                            fieldClass="md:max-w-full w-full"
                                        />
                                    </div>}

                            </AppLayoutStack>}

                        {searchType === Torrent_SearchType.SIMPLE && (
                            <TextInput
                                value={globalFilter}
                                onValueChange={setGlobalFilter}
                                leftIcon={<FiSearch className="text-lg" />}
                            />
                        )}
                    </>}

                    {(selectedProviderExtensionId !== "none" && selectedProviderExtensionId !== "") && Object.keys(warnings)?.map((key) => {
                        if ((warnings as any)[key]) {
                            return <Alert
                                data-torrent-search-container-warning
                                key={key}
                                intent="warning"
                                description={<>
                                    {key === "extensionDoesNotSupportAdult" && "This provider does not support adult content"}
                                    {key === "extensionDoesNotSupportSmartSearch" && "This provider does not support smart search"}
                                    {key === "extensionDoesNotSupportBestRelease" && "This provider does not support best release search"}
                                </>}
                            />
                        }
                        return null
                    })}

                </div>


                {(selectedProviderExtensionId !== "none" && selectedProviderExtensionId !== "") ? (
                    <>

                        <div className="space-y-3" data-torrent-search-container-torrents-container>

                            {(type === "torrentstream-select" || type === "torrentstream-select-file" || type === "debridstream-select-file" || type === "debridstream-select") &&
                                <TorrentSearchTorrentStreamBatchHistory
                                    type={type}
                                    entry={entry}
                                    debridInstantAvailability={debridInstantAvailability}
                                />}

                            {hasOneWarning && <LuffyError />}
                            {(searchType === Torrent_SearchType.SMART) && !hasOneWarning && (
                                <>
                                    <TorrentPreviewList
                                        entry={entry}
                                        previews={previews}
                                        isLoading={isLoading || isFetching}
                                        selectedTorrents={selectedTorrents}
                                        onToggleTorrent={handleToggleTorrent}
                                        debridInstantAvailability={debridInstantAvailability}
                                        type={type}
                                        torrentMetadata={data?.torrentMetadata}
                                        // animeMetadata={data?.animeMetadata}
                                    />
                                </>
                            )}

                            {((searchType !== Torrent_SearchType.SMART) && !hasOneWarning && !previews?.length) && (
                                <>
                                    <TorrentTable
                                        entry={entry}
                                        type={type}
                                        torrents={torrents}
                                        globalFilter={globalFilter}
                                        setGlobalFilter={setGlobalFilter}
                                        smartSearch={false}
                                        supportsQuery
                                        isLoading={isLoading}
                                        isFetching={isFetching}
                                        selectedTorrents={selectedTorrents}
                                        onToggleTorrent={handleToggleTorrent}
                                        debridInstantAvailability={debridInstantAvailability}
                                        animeMetadata={data?.animeMetadata}
                                        torrentMetadata={data?.torrentMetadata}
                                    />
                                </>
                            )}
                        </div>

                    </>
                ) : (!!providerExtensions) ? <div className="space-y-2">
                    <LuffyError title="No extension selected" />
                    {!providerExtensions.length && <div className="flex justify-center">
                        <SeaLink href="/extensions">
                            <Button intent="white" leftIcon={<LuPlus />}>
                                Add extensions
                            </Button>
                        </SeaLink>
                    </div>}
                </div> : <div className="space-y-2">
                    <Skeleton className="h-[96px]" />
                    <Skeleton className="h-[96px]" />
                    <Skeleton className="h-[96px]" />
                    <Skeleton className="h-[96px]" />
                </div>}
            </AppLayoutStack>

            {type === "download" && <TorrentDownloadModal
                onToggleTorrent={handleToggleTorrent}
                media={entry.media!!}
                entry={entry}
            />}

            {type === "download" && <TorrentDownloadFileSelection entry={entry} />}

            {type === "torrentstream-select-file" && <TorrentstreamFileSelectionModal entry={entry} />}
            {type === "debridstream-select-file" && <DebridStreamFileSelectionModal entry={entry} />}
        </>
    )

}

function TorrentSearchTorrentStreamBatchHistory({ entry, type, debridInstantAvailability }: {
    entry: Anime_Entry | undefined,
    type: TorrentSelectionType,
    debridInstantAvailability: Record<string, Debrid_TorrentItemInstantAvailability>
}) {

    const { data: batchHistory } = useGetTorrentstreamBatchHistory(entry?.mediaId, true)

    const { handleStreamSelection: handleTorrentstreamSelection } = useHandleStartTorrentStream()
    const { handleStreamSelection: handleDebridstreamSelection } = useHandleStartDebridStream()
    const { torrentSearchStreamEpisode } = useTorrentSearchSelectedStreamEpisode()
    const setTorrentFileSelection = useSetAtom(__torrentSearch_fileSelectionTorrentAtom)
    const setTorrentSearch = useSetAtom(__torrentSearch_selectionAtom)
    const { setAutoPlayTorrent } = useAutoPlaySelectedTorrent()

    if (!batchHistory?.torrent || !entry) return null

    return (
        <AppLayoutStack>
            <h5 className="text-center flex gap-2 items-center"><LuCornerLeftDown className="mt-1" /> Previous selection</h5>

            <TorrentListItem
                torrent={batchHistory?.torrent}
                metadata={batchHistory?.metadata}
                media={entry.media}
                episode={undefined}
                debridCached={((type === "download" || type === "debridstream-select" || type === "debridstream-select-file") && !!batchHistory.torrent.infoHash && !!debridInstantAvailability[batchHistory.torrent.infoHash])}
                isSelected={false}
                onClick={() => {
                    if (!batchHistory?.torrent || !torrentSearchStreamEpisode?.aniDBEpisode) return
                    if (type === "torrentstream-select") {
                        setAutoPlayTorrent(batchHistory.torrent, entry!)
                        handleTorrentstreamSelection({
                            torrent: batchHistory?.torrent,
                            mediaId: entry!.mediaId,
                            aniDBEpisode: torrentSearchStreamEpisode.aniDBEpisode,
                            episodeNumber: torrentSearchStreamEpisode.episodeNumber,
                            chosenFileIndex: undefined,
                            batchEpisodeFiles: undefined,
                        })
                        setTorrentSearch(undefined)
                    } else if (type === "debridstream-select") {
                        setAutoPlayTorrent(batchHistory.torrent, entry!)
                        handleDebridstreamSelection({
                            torrent: batchHistory?.torrent,
                            mediaId: entry!.mediaId,
                            aniDBEpisode: torrentSearchStreamEpisode.aniDBEpisode,
                            episodeNumber: torrentSearchStreamEpisode.episodeNumber,
                            chosenFileId: "",
                            batchEpisodeFiles: undefined,
                        })
                        setTorrentSearch(undefined)
                    } else if (type === "torrentstream-select-file" || type === "debridstream-select-file") {
                        // Open the drawer to select the file
                        // This opens the file selection drawer
                        setTorrentFileSelection(batchHistory?.torrent)
                    }
                }}
            />
        </AppLayoutStack>
    )
}
