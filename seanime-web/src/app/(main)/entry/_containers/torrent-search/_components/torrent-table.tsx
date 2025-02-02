import { Anime_Entry, Debrid_TorrentItemInstantAvailability, HibikeTorrent_AnimeTorrent } from "@/api/generated/types"
import {
    TorrentDebridInstantAvailabilityBadge,
    TorrentResolutionBadge,
    TorrentSeedersBadge,
} from "@/app/(main)/entry/_containers/torrent-search/_components/torrent-item-badges"
import { LuffyError } from "@/components/shared/luffy-error"
import { ScrollAreaBox } from "@/components/shared/scroll-area-box"
import { Badge } from "@/components/ui/badge"
import { Skeleton } from "@/components/ui/skeleton"
import { TextInput } from "@/components/ui/text-input"
import { formatDistanceToNowSafe } from "@/lib/helpers/date"
import React, { memo } from "react"
import { BiCalendarAlt } from "react-icons/bi"
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
                    <ScrollAreaBox className="h-[calc(100dvh_-_25rem)]">
                        {torrents.map(torrent => {
                            return (
                                <TorrentPreviewItem
                                    isBasic
                                    link={torrent.link}
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
                    </ScrollAreaBox>
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

