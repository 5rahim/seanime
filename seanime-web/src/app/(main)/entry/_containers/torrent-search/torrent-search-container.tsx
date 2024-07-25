import { Anime_AnimeEntry, HibikeTorrent_AnimeTorrent } from "@/api/generated/types"
import { TorrentPreviewList } from "@/app/(main)/entry/_containers/torrent-search/_components/torrent-preview-list"
import { TorrentTable } from "@/app/(main)/entry/_containers/torrent-search/_components/torrent-table"
import {
    __torrentSearch_torrentstreamSelectedTorrentAtom,
    TorrentstreamFileSelectionModal,
} from "@/app/(main)/entry/_containers/torrent-search/_components/torrentstream-file-section-modal"
import { Torrent_SearchType, useHandleTorrentSearch } from "@/app/(main)/entry/_containers/torrent-search/_lib/handle-torrent-search"
import {
    TorrentConfirmationContinueButton,
    TorrentConfirmationModal,
} from "@/app/(main)/entry/_containers/torrent-search/torrent-confirmation-modal"
import { __torrentSearch_drawerIsOpenAtom, TorrentSelectionType } from "@/app/(main)/entry/_containers/torrent-search/torrent-search-drawer"
import { useHandleStartTorrentStream } from "@/app/(main)/entry/_containers/torrent-stream/_lib/handle-torrent-stream"
import { useTorrentStreamingSelectedEpisode } from "@/app/(main)/entry/_lib/torrent-streaming.atoms"
import { LuffyError } from "@/components/shared/luffy-error"
import { Alert } from "@/components/ui/alert"
import { cn } from "@/components/ui/core/styling"
import { DataGridSearchInput } from "@/components/ui/datagrid"
import { NumberInput } from "@/components/ui/number-input"
import { Select } from "@/components/ui/select"
import { Switch } from "@/components/ui/switch"
import { TORRENT_PROVIDER } from "@/lib/server/settings"
import { atom, useSetAtom } from "jotai"
import { useAtom } from "jotai/react"
import React, { startTransition } from "react"
import { RiFolderDownloadFill } from "react-icons/ri"

export const __torrentSearch_selectedTorrentsAtom = atom<HibikeTorrent_AnimeTorrent[]>([])

export function TorrentSearchContainer({ type, entry }: { type: TorrentSelectionType, entry: Anime_AnimeEntry }) {
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
            hideControls
            size="sm"
            fieldClass={cn(
                "flex items-center md:justify-end gap-3 space-y-0",
                { "opacity-50 cursor-not-allowed pointer-events-none": (smartSearchBatch || searchType != Torrent_SearchType.SMART) },
            )}
            fieldLabelClass="flex-none self-center font-normal !text-md sm:text-md lg:text-md"
            className="max-w-[6rem]"
        />
    }, [searchType, smartSearchBatch, downloadInfo, soughtEpisode])

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
        }
    }

    return (
        <>
            <div className="py-4 space-y-4">
                <Select
                    name="torrentProvider"
                    leftAddon="Torrent Provider"
                    // label="Torrent Provider"
                    value={selectedProviderExtension?.id ?? TORRENT_PROVIDER.NONE}
                    onValueChange={setSelectedProviderExtensionId}
                    leftIcon={<RiFolderDownloadFill className="text-orange-500" />}
                    options={[
                        ...(providerExtensions?.map(ext => ({
                            label: ext.name,
                            value: ext.id,
                        })) ?? []).sort((a, b) => a?.label?.localeCompare(b?.label) ?? 0),
                        { label: "None", value: TORRENT_PROVIDER.NONE },
                    ]}
                />

                {selectedProviderExtensionId !== "none" && selectedProviderExtensionId !== "" ? (
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
                                <div className="flex flex-col justify-between gap-3 md:flex-row w-full">

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
                                            "flex items-center md:justify-center gap-3 space-y-0",
                                            { "opacity-50 cursor-not-allowed pointer-events-none": searchType != Torrent_SearchType.SMART || smartSearchBest },
                                        )}
                                        fieldLabelClass="flex-none self-center font-normal !text-md sm:text-md lg:text-md"
                                        className="w-[6rem]"
                                    />}

                                    {selectedProviderExtension?.settings?.smartSearchFilters?.includes("bestReleases") && <Switch
                                        label="Best releases"
                                        help={!downloadInfo?.canBatch ? "Cannot look for best releases yet" : "Look for the best releases"}
                                        value={smartSearchBest}
                                        onValueChange={setSmartSearchBest}
                                        fieldClass={cn(
                                            { "opacity-50 cursor-not-allowed pointer-events-none": !downloadInfo?.canBatch },
                                        )}
                                        size="sm"
                                    />}

                                    {selectedProviderExtension?.settings?.smartSearchFilters?.includes("batch") && <Switch
                                        label="Batches"
                                        help={!downloadInfo?.canBatch ? "Cannot look for batches yet" : "Look for batches"}
                                        value={smartSearchBatch}
                                        onValueChange={setSmartSearchBatch}
                                        disabled={smartSearchBest || !downloadInfo?.canBatch}
                                        fieldClass={cn(
                                            { "opacity-50 cursor-not-allowed pointer-events-none": !downloadInfo?.canBatch || smartSearchBest },
                                        )}
                                        size="sm"
                                    />}

                                </div>

                                {!hasOneWarning && (
                                    <>
                                        {selectedProviderExtension?.settings?.smartSearchFilters?.includes("query") && <DataGridSearchInput
                                            value={globalFilter ?? ""}
                                            onChange={v => setGlobalFilter(v)}
                                            placeholder={searchType == Torrent_SearchType.SMART
                                                ? `Refine the title (${entry.media?.title?.romaji})`
                                                : "Search"}
                                            fieldClass="md:max-w-full w-full"
                                        />}

                                        <div className="pb-1" />

                                        <TorrentPreviewList
                                            previews={previews}
                                            isLoading={isLoading}
                                            selectedTorrents={selectedTorrents}
                                            onToggleTorrent={handleToggleTorrent}
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
                ) : <LuffyError title="No extension selected" />}
            </div>

            {type === "download" && <TorrentConfirmationModal
                onToggleTorrent={handleToggleTorrent}
                media={entry.media!!}
                entry={entry}
            />}

            {type === "select-file" && <TorrentstreamFileSelectionModal entry={entry} />}
        </>
    )

}
