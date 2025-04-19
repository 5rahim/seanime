import {
    Anime_Entry,
    Debrid_TorrentItemInstantAvailability,
    HibikeTorrent_AnimeTorrent,
    Torrent_Preview,
    Torrent_TorrentMetadata,
} from "@/api/generated/types"
import {
    filterItems,
    sortItems,
    TorrentFilterSortControls,
    useTorrentFiltering,
    useTorrentSorting,
} from "@/app/(main)/entry/_containers/torrent-search/_components/torrent-common-helpers"
import {
    TorrentDebridInstantAvailabilityBadge,
    TorrentParsedMetadata,
    TorrentResolutionBadge,
    TorrentSeedersBadge,
} from "@/app/(main)/entry/_containers/torrent-search/_components/torrent-item-badges"
import { TorrentPreviewItem } from "@/app/(main)/entry/_containers/torrent-search/_components/torrent-preview-item"
import { TorrentSelectionType } from "@/app/(main)/entry/_containers/torrent-search/torrent-search-drawer"
import { LuffyError } from "@/components/shared/luffy-error"
import { ScrollAreaBox } from "@/components/shared/scroll-area-box"
import { Badge } from "@/components/ui/badge"
import { Skeleton } from "@/components/ui/skeleton"
import { formatDistanceToNowSafe } from "@/lib/helpers/date"
import React from "react"
import { BiCalendarAlt } from "react-icons/bi"

type TorrentPreviewList = {
    entry: Anime_Entry
    previews: Torrent_Preview[]
    debridInstantAvailability: Record<string, Debrid_TorrentItemInstantAvailability>
    isLoading: boolean
    selectedTorrents: HibikeTorrent_AnimeTorrent[]
    onToggleTorrent: (t: HibikeTorrent_AnimeTorrent) => void
    type: TorrentSelectionType
    torrentMetadata: Record<string, Torrent_TorrentMetadata> | undefined
}

export const TorrentPreviewList = React.memo((
    {
        entry,
        previews,
        isLoading,
        selectedTorrents,
        onToggleTorrent,
        debridInstantAvailability,
        type,
        torrentMetadata,
    }: TorrentPreviewList) => {
    // Use hooks for sorting and filtering
    const { sortField, sortDirection, handleSortChange } = useTorrentSorting()
    const { filters, handleFilterChange } = useTorrentFiltering()

    if (isLoading) return <div className="space-y-2">
        <Skeleton className="h-[96px]" />
        <Skeleton className="h-[96px]" />
        <Skeleton className="h-[96px]" />
        <Skeleton className="h-[96px]" />
    </div>

    if (!isLoading && !previews?.length) {
        return <LuffyError title="Nothing found" />
    }

    // Apply filters using the generic helper
    const filteredPreviews = filterItems(previews, torrentMetadata, filters)

    // Sort the previews based on current sort settings using the generic helper
    const sortedPreviews = sortItems(filteredPreviews, sortField, sortDirection)

    return (
        <div className="space-y-2" data-torrent-preview-list>

            <TorrentFilterSortControls
                resultCount={sortedPreviews?.length || 0}
                sortField={sortField}
                sortDirection={sortDirection}
                filters={filters}
                onSortChange={handleSortChange}
                onFilterChange={handleFilterChange}
            />
            <ScrollAreaBox className="h-[calc(100dvh_-_25rem)]">
                {/*<div className="grid grid-cols-1 lg:grid-cols-2 gap-2">*/}
                {sortedPreviews.filter(Boolean).map(item => {
                    if (!item.torrent) return null
                    // const isReleasedBeforeMedia = differenceInCalendarYears(mediaReleaseDate, item.torrent.date) > 2
                    return (
                        <TorrentPreviewItem
                            link={item.torrent?.link}
                            confirmed={item.torrent?.confirmed}
                            key={item.torrent.link}
                            title={item.episode?.displayTitle || item.episode?.baseAnime?.title?.userPreferred || ""}
                            releaseGroup={item.torrent.releaseGroup || ""}
                            subtitle={item.torrent.name}
                            isBatch={item.torrent.isBatch ?? false}
                            isBestRelease={item.torrent.isBestRelease}
                            image={item.episode?.episodeMetadata?.image || item.episode?.baseAnime?.coverImage?.large ||
                                (item.torrent.confirmed ? (entry.media?.coverImage?.large || entry.media?.bannerImage) : null)}
                            fallbackImage={entry.media?.coverImage?.large || entry.media?.bannerImage}
                            isSelected={selectedTorrents.findIndex(n => n.link === item.torrent!.link) !== -1}
                            onClick={() => onToggleTorrent(item.torrent!)}
                        >
                            <div className="flex flex-wrap gap-2 items-center">
                                {item.torrent.isBestRelease && (
                                    <Badge
                                        className="rounded-[--radius-md] text-[0.8rem] bg-pink-800 border-transparent border"
                                        intent="success-solid"
                                    >
                                        Best release
                                    </Badge>
                                )}
                                <TorrentResolutionBadge resolution={item.torrent.resolution} />
                                {((type === "download" || type === "debrid-stream-select" || type === "debrid-stream-select-file") && !!item.torrent.infoHash && debridInstantAvailability[item.torrent.infoHash]) && (
                                    <TorrentDebridInstantAvailabilityBadge />
                                )}
                                <TorrentSeedersBadge seeders={item.torrent.seeders} />
                                {!!item.torrent.size && <p className="text-gray-300 text-sm flex items-center gap-1">
                                    {item.torrent.formattedSize}</p>}
                                <p className="text-[--muted] text-sm flex items-center gap-1">
                                    <BiCalendarAlt /> {formatDistanceToNowSafe(item.torrent.date)}
                                </p>
                            </div>
                            <TorrentParsedMetadata metadata={torrentMetadata?.[item.torrent.infoHash!]} />
                        </TorrentPreviewItem>
                    )
                })}
                {/*</div>*/}
            </ScrollAreaBox>
        </div>
    )

})
