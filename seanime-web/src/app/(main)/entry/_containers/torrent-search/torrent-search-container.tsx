import { Anime_Entry, HibikeTorrent_AnimeTorrent } from "@/api/generated/types"
import { useGetTorrentstreamBatchHistory } from "@/api/hooks/torrentstream.hooks"
import { DebridStreamFileSelectionModal } from "@/app/(main)/entry/_containers/debrid-stream/debrid-stream-file-selection-modal"
import { TorrentResolutionBadge, TorrentSeedersBadge } from "@/app/(main)/entry/_containers/torrent-search/_components/torrent-item-badges"
import { TorrentPreviewItem } from "@/app/(main)/entry/_containers/torrent-search/_components/torrent-preview-item"
import { TorrentPreviewList } from "@/app/(main)/entry/_containers/torrent-search/_components/torrent-preview-list"
import { TorrentTable } from "@/app/(main)/entry/_containers/torrent-search/_components/torrent-table"
import { Torrent_SearchType, useHandleTorrentSearch } from "@/app/(main)/entry/_containers/torrent-search/_lib/handle-torrent-search"
import {
    TorrentConfirmationContinueButton,
    TorrentConfirmationModal,
} from "@/app/(main)/entry/_containers/torrent-search/torrent-confirmation-modal"
import { __torrentSearch_drawerIsOpenAtom, TorrentSelectionType } from "@/app/(main)/entry/_containers/torrent-search/torrent-search-drawer"
import { useHandleStartTorrentStream } from "@/app/(main)/entry/_containers/torrent-stream/_lib/handle-torrent-stream"
import {
    __torrentSearch_torrentstreamSelectedTorrentAtom,
    TorrentstreamFileSelectionModal,
} from "@/app/(main)/entry/_containers/torrent-stream/torrent-stream-file-selection-modal"
import { useTorrentStreamingSelectedEpisode } from "@/app/(main)/entry/_lib/torrent-streaming.atoms"
import { LuffyError } from "@/components/shared/luffy-error"
import { Alert } from "@/components/ui/alert"
import { AppLayoutStack } from "@/components/ui/app-layout"
import { Badge } from "@/components/ui/badge"
import { IconButton } from "@/components/ui/button"
import { cn } from "@/components/ui/core/styling"
import { DataGridSearchInput } from "@/components/ui/datagrid"
import { NumberInput } from "@/components/ui/number-input"
import { Select } from "@/components/ui/select"
import { Skeleton } from "@/components/ui/skeleton"
import { Switch } from "@/components/ui/switch"
import { Tooltip } from "@/components/ui/tooltip"
import { formatDistanceToNowSafe } from "@/lib/helpers/date"
import { TORRENT_PROVIDER } from "@/lib/server/settings"
import { atom, useSetAtom } from "jotai"
import { useAtom } from "jotai/react"
import React, { startTransition } from "react"
import { BiCalendarAlt, BiFile, BiLinkExternal } from "react-icons/bi"
import { RiFolderDownloadFill } from "react-icons/ri"

export const __torrentSearch_selectedTorrentsAtom = atom<HibikeTorrent_AnimeTorrent[]>([])

export function TorrentSearchContainer({ type, entry }: { type: TorrentSelectionType, entry: Anime_Entry }) {
    const downloadInfo = React.useMemo(() => entry.downloadInfo, [entry.downloadInfo])

    const shouldLookForBatches = React.useMemo(() => {
        if (type === "download") {
            return !!downloadInfo?.canBatch && !!downloadInfo?.episodesToDownload?.length
        } else {
            return !!downloadInfo?.canBatch
        }
    }, [downloadInfo?.canBatch, downloadInfo?.episodesToDownload?.length, type])

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
        isAdult,
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
    const debridInstantAvailability = React.useMemo(() => data?.debridInstantAvailability ?? {}, [data?.debridInstantAvailability])

    const EpisodeNumberInput = React.useCallback(() => {
        return <NumberInput
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
            className="max-w-[6rem]"
        />
    }, [searchType, smartSearchBatch, smartSearchBest, downloadInfo, soughtEpisode])

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
     * This function is called only when the type is 'select'
     * Meaning, the user has selected a torrent and wants to start streaming
     */
    const { handleManualTorrentStreamSelection } = useHandleStartTorrentStream()
    const { torrentStreamingSelectedEpisode } = useTorrentStreamingSelectedEpisode()
    const setTorrentstreamSelectedTorrent = useSetAtom(__torrentSearch_torrentstreamSelectedTorrentAtom)
    const [, setter] = useAtom(__torrentSearch_drawerIsOpenAtom)
    const onTorrentValidated = () => {
        if (type === "select") {
            if (selectedTorrents.length && !!torrentStreamingSelectedEpisode?.aniDBEpisode) {
                handleManualTorrentStreamSelection({
                    torrent: selectedTorrents[0],
                    entry,
                    aniDBEpisode: torrentStreamingSelectedEpisode.aniDBEpisode,
                    episodeNumber: torrentStreamingSelectedEpisode.episodeNumber,
                    chosenFileIndex: undefined,
                })
                setter(undefined)
                React.startTransition(() => {
                    setSelectedTorrents([])
                })
            }
        } else if (type === "select-file") {
            // Open the drawer to select the file
            if (selectedTorrents.length && !!torrentStreamingSelectedEpisode?.aniDBEpisode) {
                // This opens the file selection drawer
                setTorrentstreamSelectedTorrent(selectedTorrents[0])
                React.startTransition(() => {
                    setSelectedTorrents([])
                })
            }
        } else if (type === "debrid-stream") {
            // Open the drawer to select the file
            if (selectedTorrents.length && !!torrentStreamingSelectedEpisode?.aniDBEpisode) {
                // This opens the file selection drawer
                setTorrentstreamSelectedTorrent(selectedTorrents[0])
                React.startTransition(() => {
                    setSelectedTorrents([])
                })
            }
        }
    }

    return (
        <>
            {(type === "select" || type === "select-file") &&
                <TorrentSearchTorrentStreamBatchHistory type={type} entry={entry} />}

            <div className="py-4 space-y-4">
                <div className="max-w-[400px]">
                    <Select
                        name="torrentProvider"
                        leftAddon="Torrent Provider"
                        value={selectedProviderExtension?.id ?? TORRENT_PROVIDER.NONE}
                        onValueChange={setSelectedProviderExtensionId}
                        leftIcon={<RiFolderDownloadFill />}
                        options={[
                            ...(providerExtensions?.map(ext => ({
                                label: ext.name,
                                value: ext.id,
                            })) ?? []).sort((a, b) => a?.label?.localeCompare(b?.label) ?? 0),
                            { label: "None", value: TORRENT_PROVIDER.NONE },
                        ]}
                    />
                </div>

                {(selectedProviderExtensionId !== "none" && selectedProviderExtensionId !== "") ? (
                    <>

                        {Object.keys(warnings)?.map((key) => {
                            if ((warnings as any)[key]) {
                                return <Alert
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

                        {entry.media?.isAdult === true && <div className="">
                            <Switch
                                label="Adult"
                                help="If enabled, this media is considered adult content. Some extensions may not support adult content."
                                value={isAdult}
                                onValueChange={setIsAdult}
                            />
                        </div>}

                        <div className="flex w-full justify-between">
                            <Switch
                                label="Smart search"
                                help={selectedProviderExtension?.settings?.canSmartSearch
                                    ? "Builds a search query automatically, based on parameters"
                                    : "This provider does not support smart search"}
                                value={searchType === Torrent_SearchType.SMART}
                                onValueChange={v => setSearchType(v ? Torrent_SearchType.SMART : Torrent_SearchType.SIMPLE)}
                                disabled={!selectedProviderExtension?.settings?.canSmartSearch}
                            />
                            <div className="flex flex-1"></div>
                            <TorrentConfirmationContinueButton type={type} onTorrentValidated={onTorrentValidated} />
                        </div>

                        {(searchType === Torrent_SearchType.SMART) && <div>
                            <div className="space-y-2">
                                <div className="flex flex-col justify-around gap-3 md:flex-row w-full">

                                    {selectedProviderExtension?.settings?.smartSearchFilters?.includes("episodeNumber") && <EpisodeNumberInput />}

                                    {selectedProviderExtension?.settings?.smartSearchFilters?.includes("resolution") && <Select
                                        label="Resolution"
                                        value={smartSearchResolution || "-"}
                                        onValueChange={v => setSmartSearchResolution(v != "-" ? v : "")}
                                        options={[
                                            { value: "-", label: "Any" },
                                            { value: "1080", label: "1080p" },
                                            { value: "720", label: "720p" },
                                            { value: "540", label: "540p" },
                                            { value: "480", label: "480p" },
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

                                    {selectedProviderExtension?.settings?.smartSearchFilters?.includes("batch") && <Switch
                                        label="Batches"
                                        value={smartSearchBatch}
                                        onValueChange={setSmartSearchBatch}
                                        disabled={smartSearchBest || !downloadInfo?.canBatch}
                                        fieldClass={cn(
                                            "flex flex-none w-fit",
                                            { "opacity-50 cursor-not-allowed pointer-events-none": !downloadInfo?.canBatch || smartSearchBest },
                                        )}
                                        size="sm"
                                    />}

                                    {selectedProviderExtension?.settings?.smartSearchFilters?.includes("bestReleases") && <Switch
                                        label="Best releases"
                                        value={smartSearchBest}
                                        onValueChange={setSmartSearchBest}
                                        fieldClass={cn(
                                            "flex flex-none w-fit",
                                            { "opacity-50 cursor-not-allowed pointer-events-none": !downloadInfo?.canBatch },
                                        )}
                                        size="sm"
                                    />}

                                </div>

                                {!hasOneWarning && (
                                    <>
                                        <div className="pb-1" />
                                        {selectedProviderExtension?.settings?.smartSearchFilters?.includes("query") && <DataGridSearchInput
                                            value={globalFilter ?? ""}
                                            onChange={v => setGlobalFilter(v)}
                                            placeholder={searchType === Torrent_SearchType.SMART
                                                ? `Refine the title (${entry.media?.title?.romaji})`
                                                : "Search"}
                                            fieldClass="md:max-w-full w-full"
                                        />}
                                        <div className="pb-1" />

                                        <TorrentPreviewList
                                            entry={entry}
                                            previews={previews}
                                            isLoading={isLoading}
                                            selectedTorrents={selectedTorrents}
                                            onToggleTorrent={handleToggleTorrent}
                                            debridInstantAvailability={debridInstantAvailability}
                                        />
                                    </>
                                )}
                            </div>
                        </div>}

                        {hasOneWarning && <LuffyError />}

                        {!hasOneWarning && (
                            <>
                                <TorrentTable
                                    torrents={torrents}
                                    globalFilter={globalFilter}
                                    setGlobalFilter={setGlobalFilter}
                                    smartSearch={searchType == Torrent_SearchType.SMART}
                                    isLoading={isLoading}
                                    isFetching={isFetching}
                                    selectedTorrents={selectedTorrents}
                                    onToggleTorrent={handleToggleTorrent}
                                />
                            </>
                        )}

                    </>
                ) : (!!providerExtensions) ? <LuffyError title="No extension selected" /> : <div className="space-y-2">
                    <Skeleton className="h-[96px]" />
                    <Skeleton className="h-[96px]" />
                    <Skeleton className="h-[96px]" />
                    <Skeleton className="h-[96px]" />
                </div>}
            </div>

            {type === "download" && <TorrentConfirmationModal
                onToggleTorrent={handleToggleTorrent}
                media={entry.media!!}
                entry={entry}
            />}

            {type === "select-file" && <TorrentstreamFileSelectionModal entry={entry} />}
            {type === "debrid-stream" && <DebridStreamFileSelectionModal entry={entry} />}
        </>
    )

}

function TorrentSearchTorrentStreamBatchHistory({ entry, type }: {
    entry: Anime_Entry | undefined,
    type: TorrentSelectionType,
}) {

    const { data: batchHistory } = useGetTorrentstreamBatchHistory(entry?.mediaId, true)

    const { handleManualTorrentStreamSelection } = useHandleStartTorrentStream()
    const { torrentStreamingSelectedEpisode } = useTorrentStreamingSelectedEpisode()
    const setTorrentstreamSelectedTorrent = useSetAtom(__torrentSearch_torrentstreamSelectedTorrentAtom)
    const [, setter] = useAtom(__torrentSearch_drawerIsOpenAtom)

    if (!batchHistory?.torrent || !entry) return null

    return (
        <AppLayoutStack>
            <h4>Previous selection</h4>

            <TorrentPreviewItem
                confirmed={batchHistory?.torrent?.confirmed}
                key={batchHistory?.torrent.link}
                title={""}
                releaseGroup={batchHistory?.torrent.releaseGroup || ""}
                filename={batchHistory?.torrent.name}
                isBatch={batchHistory?.torrent.isBatch ?? false}
                image={entry?.media?.coverImage?.large || entry?.media?.bannerImage}
                fallbackImage={entry?.media?.coverImage?.large || entry?.media?.bannerImage}
                isBestRelease={batchHistory?.torrent.isBestRelease}
                onClick={() => {
                    if (type === "select") {
                        if (batchHistory?.torrent && !!torrentStreamingSelectedEpisode?.aniDBEpisode) {
                            handleManualTorrentStreamSelection({
                                torrent: batchHistory?.torrent,
                                entry,
                                aniDBEpisode: torrentStreamingSelectedEpisode.aniDBEpisode,
                                episodeNumber: torrentStreamingSelectedEpisode.episodeNumber,
                                chosenFileIndex: undefined,
                            })
                            setter(undefined)
                        }
                    } else if (type === "select-file") {
                        // Open the drawer to select the file
                        if (!!torrentStreamingSelectedEpisode?.aniDBEpisode) {
                            // This opens the file selection drawer
                            setTorrentstreamSelectedTorrent(batchHistory?.torrent)
                        }
                    }
                }}
                action={<Tooltip
                    side="left"
                    trigger={<IconButton
                        icon={<BiLinkExternal />}
                        intent="primary-basic"
                        size="sm"
                        onClick={() => window.open(batchHistory?.torrent?.link, "_blank")}
                    />}
                >Open in browser</Tooltip>}
            >
                <div className="flex flex-wrap gap-2 items-center">
                    {batchHistory?.torrent.isBestRelease && (
                        <Badge
                            className="rounded-md text-[0.8rem] bg-pink-800 border-pink-600 border"
                            intent="success-solid"
                        >
                            Best release
                        </Badge>
                    )}
                    <TorrentResolutionBadge resolution={batchHistory?.torrent.resolution} />
                    <TorrentSeedersBadge seeders={batchHistory?.torrent.seeders} />
                    {!!batchHistory?.torrent.size && <p className="text-gray-300 text-sm flex items-center gap-1">
                        <BiFile /> {batchHistory?.torrent.formattedSize}</p>}
                    <p className="text-[--muted] text-sm flex items-center gap-1">
                        <BiCalendarAlt /> {formatDistanceToNowSafe(batchHistory?.torrent.date)}
                    </p>
                </div>
            </TorrentPreviewItem>
        </AppLayoutStack>
    )
}
