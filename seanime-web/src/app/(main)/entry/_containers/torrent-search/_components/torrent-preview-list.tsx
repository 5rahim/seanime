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
import { TorrentList, TorrentListItem } from "@/app/(main)/entry/_containers/torrent-search/_components/torrent-preview-item"
import { TorrentSelectionType } from "@/app/(main)/entry/_containers/torrent-search/torrent-search-drawer"
import { LuffyError } from "@/components/shared/luffy-error"
import { ScrollAreaBox } from "@/components/shared/scroll-area-box"
import { Skeleton } from "@/components/ui/skeleton"
import React from "react"

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
            <ScrollAreaBox className="h-[calc(100dvh_-_25rem)] bg-gray-950/60">
                <TorrentList>
                    {sortedPreviews.filter(Boolean).map(item => {
                        if (!item.torrent) return null
                        // const isReleasedBeforeMedia = differenceInCalendarYears(mediaReleaseDate, item.torrent.date) > 2
                        return (
                            <TorrentListItem
                                key={item.torrent.link}
                                torrent={item.torrent}
                                media={entry.media}
                                episode={item.episode}
                                metadata={torrentMetadata?.[item.torrent.infoHash!]?.metadata}
                                debridCached={((type === "download" || type === "debridstream-select" || type === "debridstream-select-file") && !!item.torrent.infoHash && !!debridInstantAvailability[item.torrent.infoHash])}
                                isSelected={selectedTorrents.findIndex(n => n.link === item.torrent!.link) !== -1}
                                onClick={() => onToggleTorrent(item.torrent!)}
                            />
                        )
                    })}
                </TorrentList>
            </ScrollAreaBox>
        </div>
    )

})
