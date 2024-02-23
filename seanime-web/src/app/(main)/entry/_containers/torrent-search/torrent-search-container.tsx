import { TorrentTable } from "@/app/(main)/entry/_containers/torrent-search/_components/torrent-table"
import {
    TorrentConfirmationContinueButton,
    TorrentConfirmationModal,
} from "@/app/(main)/entry/_containers/torrent-search/torrent-confirmation-modal"
import { TorrentPreviewList } from "@/app/(main)/entry/_containers/torrent-search/torrent-preview-list"
import { torrentSearchDrawerEpisodeAtom } from "@/app/(main)/entry/_containers/torrent-search/torrent-search-drawer"
import { cn } from "@/components/ui/core"
import { DataGridSearchInput } from "@/components/ui/datagrid"
import { NumberInput } from "@/components/ui/number-input"
import { Select } from "@/components/ui/select"
import { Switch } from "@/components/ui/switch"
import { useDebounceWithSet } from "@/hooks/use-debounce"
import { SeaEndpoints } from "@/lib/server/endpoints"
import { buildSeaQuery, useSeaQuery } from "@/lib/server/query"
import { AnimeTorrent, MediaEntry, TorrentSearchData } from "@/lib/server/types"
import { atom } from "jotai"
import { useAtom } from "jotai/react"
import React, { startTransition, useCallback, useEffect, useLayoutEffect, useMemo, useState } from "react"

export const __torrentSearch_selectedTorrentsAtom = atom<AnimeTorrent[]>([])

export function TorrentSearchContainer({ entry }: { entry: MediaEntry }) {

    const downloadInfo = entry.downloadInfo

    const hasEpisodesToDownload = !!downloadInfo?.episodesToDownload?.length

    const [soughtEpisode, setSoughtEpisode] = useAtom(torrentSearchDrawerEpisodeAtom)

    const [globalFilter, setGlobalFilter] = useState<string>(hasEpisodesToDownload ? "" : (entry.media?.title?.romaji || ""))
    const [selectedTorrents, setSelectedTorrents] = useAtom(__torrentSearch_selectedTorrentsAtom)
    const [quickSearch, setQuickSearch] = useState(hasEpisodesToDownload)
    const [quickSearchBatch, setQuickSearchBatch] = useState<boolean>(downloadInfo?.canBatch || false)
    const [quickSearchEpisode, setQuickSearchEpisode] = useState<number>(downloadInfo?.episodesToDownload?.[0]?.episode?.episodeNumber || 1)
    const [quickSearchResolution, setQuickSearchResolution] = useState("")
    const [dQuickSearchEpisode, setDQuickSearchEpisode] = useDebounceWithSet(quickSearchEpisode, 500)

    useLayoutEffect(() => {
        if (soughtEpisode !== undefined) {
            setQuickSearchEpisode(soughtEpisode)
            setDQuickSearchEpisode(soughtEpisode)
            startTransition(() => {
                setSoughtEpisode(undefined)
            })
        }
    }, [soughtEpisode])

    useEffect(() => {
        setSelectedTorrents([])
    }, [])

    useLayoutEffect(() => {
        if (quickSearch) {
            setGlobalFilter("")
        } else {
            setGlobalFilter(entry.media?.title?.romaji || "")
        }
    }, [quickSearch])

    const { data, isLoading, isFetching } = useSeaQuery<TorrentSearchData | undefined>({
        endpoint: SeaEndpoints.TORRENT_SEARCH,
        queryKey: ["nyaa-search", entry.mediaId, dQuickSearchEpisode, globalFilter, quickSearchBatch, quickSearchResolution, quickSearch, downloadInfo?.absoluteOffset],
        queryFn: async () => {
            return buildSeaQuery({
                endpoint: SeaEndpoints.TORRENT_SEARCH,
                method: "post",
                data: {
                    query: globalFilter,
                    episodeNumber: dQuickSearchEpisode,
                    batch: quickSearchBatch,
                    media: entry.media,
                    absoluteOffset: downloadInfo?.absoluteOffset || 0,
                    resolution: quickSearchResolution,
                    quickSearch: quickSearch,
                },
            })
        },
        refetchOnWindowFocus: false,
        retry: 0,
        retryDelay: 1000,
        enabled: !(quickSearchEpisode === undefined && globalFilter.length === 0),
    })

    const torrents = useMemo(() => data?.torrents ?? [], [data?.torrents])
    const previews = useMemo(() => data?.previews ?? [], [data?.previews])

    const EpisodeNumberInput = useCallback(() => {
        return <NumberInput
            label={"Episode number"}
            value={quickSearchEpisode}
            onChange={(value) => {
                startTransition(() => {
                    setQuickSearchEpisode(value)
                })
            }}
            discrete
            size="sm"
            fieldClassName={cn(
                "flex items-center justify-end gap-3 space-y-0",
                { "opacity-50 cursor-not-allowed pointer-events-none": (quickSearchBatch || !quickSearch) },
            )}
            fieldLabelClassName={"flex-none self-center font-normal !text-md sm:text-md lg:text-md"}
            inputClassName="max-w-[6rem]"
        />
    }, [quickSearch, quickSearchBatch, downloadInfo, soughtEpisode])

    const handleToggleTorrent = useCallback((t: AnimeTorrent) => {
        setSelectedTorrents(prev => {
            const idx = prev.findIndex(n => n.link === t.link)
            if (idx !== -1) {
                return prev.filter(n => n.link !== t.link)
            }
            return [...prev, t]
        })
    }, [setSelectedTorrents])

    return (
        <>
            <div>
                <div className="pb-4 flex w-full justify-between">
                    <Switch
                        label="Smart search"
                        help="Builds a search query automatically, based on parameters"
                        checked={quickSearch}
                        onChange={setQuickSearch}
                    />

                    <TorrentConfirmationContinueButton/>
                </div>

                {quickSearch && <div>
                    <div className="space-y-2">
                        <div className="flex gap-4 justify-between w-full">
                            <Switch
                                label="Batches"
                                help={!downloadInfo?.canBatch ? "Cannot look for batches for this media" : undefined}
                                checked={quickSearchBatch}
                                onChange={setQuickSearchBatch}
                                fieldClassName={cn(
                                    { "opacity-50 cursor-not-allowed pointer-events-none": !downloadInfo?.canBatch },
                                )}
                            />

                            <EpisodeNumberInput/>

                            <Select
                                label={"Resolution"}
                                value={quickSearchResolution}
                                onChange={e => setQuickSearchResolution(e.target.value ?? "")}
                                options={[
                                    { value: "", label: "Any" },
                                    { value: "1080", label: "1080p" },
                                    { value: "720", label: "720p" },
                                    { value: "480", label: "480p" },
                                ]}
                                size="sm"
                                fieldClassName={cn(
                                    "flex items-center justify-end gap-3 space-y-0",
                                    { "opacity-50 cursor-not-allowed pointer-events-none": !quickSearch },
                                )}
                                fieldLabelClassName={"flex-none self-center font-normal !text-md sm:text-md lg:text-md"}
                                className="w-[6rem]"
                            />
                        </div>

                        <DataGridSearchInput
                            value={globalFilter ?? ""}
                            onChange={v => setGlobalFilter(v)}
                            placeholder={quickSearch ? `Refine the title (${entry.media?.title?.romaji})` : "Search"}
                            fieldClassName="md:max-w-full w-full"
                        />

                        <div className="pb-1"/>

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
                    quickSearch={quickSearch}
                    isLoading={isLoading}
                    isFetching={isFetching}
                    selectedTorrents={selectedTorrents}
                    onToggleTorrent={handleToggleTorrent}
                />
            </div>
            <TorrentConfirmationModal
                onToggleTorrent={handleToggleTorrent}
                media={entry.media!!}
                entry={entry}
            />
        </>
    )

}
