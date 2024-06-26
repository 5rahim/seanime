import { Anime_MediaEntry, Torrent_AnimeTorrent } from "@/api/generated/types"
import { useServerStatus } from "@/app/(main)/_hooks/use-server-status"
import { TorrentPreviewList } from "@/app/(main)/entry/_containers/torrent-search/_components/torrent-preview-list"
import { TorrentTable } from "@/app/(main)/entry/_containers/torrent-search/_components/torrent-table"
import { useHandleTorrentSearch } from "@/app/(main)/entry/_containers/torrent-search/_lib/handle-torrent-search"
import {
    TorrentConfirmationContinueButton,
    TorrentConfirmationModal,
} from "@/app/(main)/entry/_containers/torrent-search/torrent-confirmation-modal"
import { __torrentSearch_drawerIsOpenAtom, TorrentSearchType } from "@/app/(main)/entry/_containers/torrent-search/torrent-search-drawer"
import { useHandleStartTorrentStream } from "@/app/(main)/entry/_containers/torrent-stream/_lib/handle-torrent-stream"
import { useTorrentStreamingSelectedEpisode } from "@/app/(main)/entry/_lib/torrent-streaming.atoms"
import { cn } from "@/components/ui/core/styling"
import { DataGridSearchInput } from "@/components/ui/datagrid"
import { NumberInput } from "@/components/ui/number-input"
import { Select } from "@/components/ui/select"
import { Switch } from "@/components/ui/switch"
import { atom } from "jotai"
import { useAtom } from "jotai/react"
import React, { startTransition } from "react"

export const __torrentSearch_selectedTorrentsAtom = atom<Torrent_AnimeTorrent[]>([])

export function TorrentSearchContainer({ type, entry }: { type: TorrentSearchType, entry: Anime_MediaEntry }) {
    const serverStatus = useServerStatus()
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
        if (smartSearch) {
            setGlobalFilter("")
        } else {
            setGlobalFilter(entry.media?.title?.romaji || "")
        }
    }, [smartSearch])

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
                { "opacity-50 cursor-not-allowed pointer-events-none": (smartSearchBatch || !smartSearch) },
            )}
            fieldLabelClass="flex-none self-center font-normal !text-md sm:text-md lg:text-md"
            className="max-w-[6rem]"
        />
    }, [smartSearch, smartSearchBatch, downloadInfo, soughtEpisode])

    /**
     * Select torrent
     * - Download: Select multiple torrents
     * - Select: Select only one torrent
     */
    const handleToggleTorrent = React.useCallback((t: Torrent_AnimeTorrent) => {
        if (type === "download") {
            setSelectedTorrents(prev => {
                const idx = prev.findIndex(n => n.link === t.link)
                if (idx !== -1) {
                    return prev.filter(n => n.link !== t.link)
                }
                return [...prev, t]
            })
        } else if (type === "select") {
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
    const [, setter] = useAtom(__torrentSearch_drawerIsOpenAtom)
    const onTorrentValidated = () => {
        if (type === "select") {
            if (selectedTorrents.length && !!torrentStreamingSelectedEpisode?.aniDBEpisode) {
                handleManualTorrentStreamSelection({
                    torrent: selectedTorrents[0],
                    entry,
                    aniDBEpisode: torrentStreamingSelectedEpisode.aniDBEpisode,
                    episodeNumber: smartSearchEpisode,
                })
                setter(undefined)
                React.startTransition(() => {
                    setSelectedTorrents([])
                })
            }
        }
    }

    return (
        <>
            <div>
                {entry.media?.isAdult === true && <div className="py-2">
                    <Switch
                        label="Adult"
                        help="If enabled, Seanime will switch providers"
                        value={isAdult}
                        onValueChange={setIsAdult}
                    />
                </div>}

                {!isAdult ? <div className="py-4 flex w-full justify-between">
                    <Switch
                        label="Smart search"
                        help="Builds a search query automatically, based on parameters"
                        value={smartSearch}
                        onValueChange={setSmartSearch}
                    />

                    <TorrentConfirmationContinueButton type={type} onTorrentValidated={onTorrentValidated} />
                </div> : <div className="py-4 flex items-center">
                    <div>
                        <div className="text-[--muted] italic">Smart search is not enabled for adult content</div>
                        <div className="">Provider: <strong>Nyaa Sukeibei</strong></div>
                    </div>
                    <div className="flex flex-1"></div>
                    <TorrentConfirmationContinueButton type={type} onTorrentValidated={onTorrentValidated} />
                </div>}

                {smartSearch && <div>
                    <div className="space-y-2">
                        <div className="flex flex-col md:flex-row gap-4 justify-between w-full">

                            <EpisodeNumberInput />

                            <Select
                                label="Resolution"
                                value={smartSearchResolution || "-"}
                                onValueChange={v => setSmartSearchResolution(v != "-" ? v : "")}
                                options={[
                                    { value: "-", label: "Any" },
                                    { value: "1080", label: "1080p" },
                                    { value: "720", label: "720p" },
                                    { value: "480", label: "480p" },
                                ]}
                                disabled={smartSearchBest || !smartSearch}
                                size="sm"
                                fieldClass={cn(
                                    "flex items-center md:justify-center gap-3 space-y-0",
                                    { "opacity-50 cursor-not-allowed pointer-events-none": !smartSearch || smartSearchBest },
                                )}
                                fieldLabelClass="flex-none self-center font-normal !text-md sm:text-md lg:text-md"
                                className="w-[6rem]"
                            />

                            <Switch
                                label="Best releases"
                                help={!downloadInfo?.canBatch ? "Cannot look for best releases yet" : "Look for the best releases"}
                                value={smartSearchBest}
                                onValueChange={setSmartSearchBest}
                                fieldClass={cn(
                                    { "opacity-50 cursor-not-allowed pointer-events-none": !downloadInfo?.canBatch },
                                )}
                                size="sm"
                            />

                            <Switch
                                label="Batches"
                                help={!downloadInfo?.canBatch ? "Cannot look for batches yet" : "Look for batches"}
                                value={smartSearchBatch}
                                onValueChange={setSmartSearchBatch}
                                disabled={smartSearchBest || !downloadInfo?.canBatch}
                                fieldClass={cn(
                                    { "opacity-50 cursor-not-allowed pointer-events-none": !downloadInfo?.canBatch || smartSearchBest },
                                )}
                                size="sm"
                            />

                        </div>

                        {serverStatus?.settings?.library?.torrentProvider != "animetosho" && <DataGridSearchInput
                            value={globalFilter ?? ""}
                            onChange={v => setGlobalFilter(v)}
                            placeholder={smartSearch ? `Refine the title (${entry.media?.title?.romaji})` : "Search"}
                            fieldClass="md:max-w-full w-full"
                        />}

                        <div className="pb-1" />

                        <TorrentPreviewList
                            previews={previews}
                            isLoading={isLoading}
                            selectedTorrents={selectedTorrents}
                            onToggleTorrent={handleToggleTorrent}
                        />
                    </div>
                </div>}

                <TorrentTable
                    torrents={torrents}
                    globalFilter={globalFilter}
                    setGlobalFilter={setGlobalFilter}
                    smartSearch={smartSearch}
                    isLoading={isLoading}
                    isFetching={isFetching}
                    selectedTorrents={selectedTorrents}
                    onToggleTorrent={handleToggleTorrent}
                />
            </div>
            {type === "download" && <TorrentConfirmationModal
                onToggleTorrent={handleToggleTorrent}
                media={entry.media!!}
                entry={entry}
            />}
        </>
    )

}
