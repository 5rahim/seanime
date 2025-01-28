import { Anime_Entry, Debrid_TorrentItemInstantAvailability, HibikeTorrent_AnimeTorrent } from "@/api/generated/types"
import {
    TorrentDebridInstantAvailabilityBadge,
    TorrentResolutionBadge,
    TorrentSeedersBadge,
} from "@/app/(main)/entry/_containers/torrent-search/_components/torrent-item-badges"
import { LuffyError } from "@/components/shared/luffy-error"
import { Badge } from "@/components/ui/badge"
import { IconButton } from "@/components/ui/button"
import { cn } from "@/components/ui/core/styling"
import { defineDataGridColumns } from "@/components/ui/datagrid"
import { ScrollArea } from "@/components/ui/scroll-area"
import { Skeleton } from "@/components/ui/skeleton"
import { TextInput } from "@/components/ui/text-input"
import { Tooltip } from "@/components/ui/tooltip"
import { openTab } from "@/lib/helpers/browser"
import { formatDistanceToNowSafe } from "@/lib/helpers/date"
import React, { memo, useMemo } from "react"
import { BiCalendarAlt, BiLinkExternal } from "react-icons/bi"
import { TorrentPreviewItem } from "./torrent-preview-item"

type TorrentTable = {
    entry?: Anime_Entry
    torrents: HibikeTorrent_AnimeTorrent[]
    selectedTorrents: HibikeTorrent_AnimeTorrent[]
    globalFilter: string,
    setGlobalFilter: React.Dispatch<React.SetStateAction<string>>
    smartSearch: boolean
    supportsQuery: boolean
    isLoading: boolean
    isFetching: boolean
    onToggleTorrent: (t: HibikeTorrent_AnimeTorrent) => void
    debridInstantAvailability: Record<string, Debrid_TorrentItemInstantAvailability>
}

export const TorrentTable = memo((
    {
        entry,
        torrents,
        selectedTorrents,
        globalFilter,
        supportsQuery,
        setGlobalFilter,
        smartSearch,
        isFetching,
        isLoading,
        onToggleTorrent,
        debridInstantAvailability,
    }: TorrentTable) => {

    const columns = useMemo(() => defineDataGridColumns<HibikeTorrent_AnimeTorrent>(() => [
        {
            accessorKey: "name",
            header: "Name",
            cell: info => <div className="flex items-center gap-2">
                <Tooltip
                    trigger={<IconButton
                        icon={<BiLinkExternal />}
                        intent="primary-basic"
                        size="sm"
                        onClick={() => openTab(info.row.original.link)}
                    />}
                >Open in browser</Tooltip>
                <Tooltip
                    trigger={
                        <div
                            className={cn(
                                "text-[.95rem] line-clamp-1 cursor-pointer max-w-[90%] overflow-hidden",
                                {
                                    "text-brand-300 font-semibold": selectedTorrents.some(torrent => torrent.link === info.row.original.link),
                                },
                            )}
                            onClick={() => onToggleTorrent(info.row.original)}
                        >
                            {info.getValue<string>()}
                        </div>}
                >
                    {info.getValue<string>()}
                </Tooltip>
            </div>,
            size: 350,
        },
        {
            accessorKey: "resolution",
            header: "Resolution",
            cell: info => <div className="text-center">
                <TorrentResolutionBadge resolution={info.getValue<string>()} />
            </div>,
            size: 70,
        },
        {
            accessorKey: "seeders",
            header: "Seeders",
            cell: info => (
                <div className="flex items-center gap-2 ">
                    <TorrentSeedersBadge seeders={info.getValue<number>()} />
                    {(!!info.row.original.infoHash && debridInstantAvailability[info.row.original.infoHash]) && (
                        <TorrentDebridInstantAvailabilityBadge />
                    )}
                </div>
            ),
            size: 80,
        },
        {
            accessorKey: "date",
            header: "Date",
            cell: info => formatDistanceToNowSafe(info.getValue<string>()),
            // size: 80,
        },
    ]), [torrents, selectedTorrents, debridInstantAvailability])

    return (
        <>
            <TextInput
                value={globalFilter}
                onValueChange={setGlobalFilter}
            />

            {(isLoading || isFetching) ? <div className="space-y-2">
                <Skeleton className="h-[96px]" />
                <Skeleton className="h-[96px]" />
                <Skeleton className="h-[96px]" />
                <Skeleton className="h-[96px]" />
            </div> : !torrents?.length ? <div>
                <LuffyError title="Nothing found" />
            </div> : (
                <>
                    <p className="text-sm text-[--muted]">{torrents?.length} results</p>
                    <ScrollArea className="h-[calc(100dvh_-_25rem)] relative border rounded-[--radius]">
                        <div
                            className="z-[5] absolute bottom-0 w-full h-8 bg-gradient-to-t from-[--background] to-transparent"
                        />
                        <div
                            className="z-[5] absolute top-0 w-full h-8 bg-gradient-to-b from-[--background] to-transparent"
                        />
                        <div className="space-y-2 p-6">
                            {torrents.map(torrent => {
                                return (
                                    <TorrentPreviewItem
                                        isBasic
                                        key={torrent.link}
                                        title={torrent.name}
                                        releaseGroup={torrent.releaseGroup || ""}
                                        subtitle={torrent.isBatch ? torrent.name : (torrent?.episodeNumber || -1) >= 0
                                            ? `Episode ${torrent?.episodeNumber ?? "N/A"}`
                                            : ""}
                                        isBatch={torrent.isBatch ?? false}
                                        isBestRelease={torrent.isBestRelease}
                                        // image={item.episode?.episodeMetadata?.image || item.episode?.baseAnime?.coverImage?.large ||
                                        //     (torrent.confirmed ? (entry.media?.coverImage?.large || entry.media?.bannerImage) : null)}
                                        // fallbackImage={entry.media?.coverImage?.large || entry.media?.bannerImage}
                                        isSelected={selectedTorrents.findIndex(n => n.link === torrent!.link) !== -1}
                                        onClick={() => onToggleTorrent(torrent!)}
                                        action={<Tooltip
                                            side="left"
                                            trigger={<IconButton
                                                icon={<BiLinkExternal />}
                                                intent="primary-basic"
                                                size="sm"
                                                onClick={() => openTab(torrent!.link)}
                                            />}
                                        >Open in browser</Tooltip>}
                                    >
                                        <div className="flex flex-wrap gap-3 items-center">
                                            {torrent.isBestRelease && (
                                                <Badge
                                                    className="rounded-[--radius-md] text-[0.8rem] bg-pink-800 border-transparent border"
                                                    intent="success-solid"
                                                >
                                                    Best release
                                                </Badge>
                                            )}
                                            <TorrentResolutionBadge resolution={torrent.resolution} />
                                            {(!!torrent.infoHash && debridInstantAvailability[torrent.infoHash]) && (
                                                <TorrentDebridInstantAvailabilityBadge />
                                            )}
                                            <TorrentSeedersBadge seeders={torrent.seeders} />
                                            {!!torrent.size && <p className="text-gray-300 text-sm flex items-center gap-1">
                                                {torrent.formattedSize}</p>}
                                            <p className="text-[--muted] text-sm flex items-center gap-1">
                                                <BiCalendarAlt /> {formatDistanceToNowSafe(torrent.date)}
                                            </p>
                                        </div>
                                    </TorrentPreviewItem>
                                )
                            })}
                        </div>
                    </ScrollArea>
                </>
            )}


            {/* <DataGrid<HibikeTorrent_AnimeTorrent>*/}
            {/*     columns={columns}*/}
            {/*     data={torrents?.slice(0, 20)}*/}
            {/*     rowCount={torrents?.length ?? 0}*/}
            {/*     initialState={{*/}
            {/*         pagination: {*/}
            {/*             pageSize: 20,*/}
            {/*             pageIndex: 0,*/}
            {/*         },*/}
            {/*     }}*/}
            {/*     tdClass="py-4 data-[row-selected=true]:bg-gray-900"*/}
            {/*     tableBodyClass="bg-transparent"*/}
            {/*     footerClass="hidden"*/}
            {/*     state={{*/}
            {/*         globalFilter,*/}
            {/*     }}*/}
            {/*     enableManualFiltering={true}*/}
            {/*     onGlobalFilterChange={setGlobalFilter}*/}
            {/*     isLoading={isLoading || isFetching}*/}
            {/*     isDataMutating={isFetching}*/}
            {/*     hideGlobalSearchInput={smartSearch}*/}
            {/* />*/}
        </>
    )

})

