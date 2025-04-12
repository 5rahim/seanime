import { Anime_Entry, Debrid_TorrentItemInstantAvailability, HibikeTorrent_AnimeTorrent, Torrent_Preview } from "@/api/generated/types"
import {
    TorrentDebridInstantAvailabilityBadge,
    TorrentResolutionBadge,
    TorrentSeedersBadge,
} from "@/app/(main)/entry/_containers/torrent-search/_components/torrent-item-badges"
import { TorrentPreviewItem } from "@/app/(main)/entry/_containers/torrent-search/_components/torrent-preview-item"
import {
    getSortIcon,
    handleSort,
    SortDirection,
    SortField,
    sortPreviewTorrents,
} from "@/app/(main)/entry/_containers/torrent-search/_components/torrent-sorting-helpers"
import { TorrentSelectionType } from "@/app/(main)/entry/_containers/torrent-search/torrent-search-drawer"
import { LuffyError } from "@/components/shared/luffy-error"
import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import { Skeleton } from "@/components/ui/skeleton"
import { formatDistanceToNowSafe } from "@/lib/helpers/date"
import React, { useState } from "react"
import { BiCalendarAlt } from "react-icons/bi"

type TorrentPreviewList = {
    entry: Anime_Entry
    previews: Torrent_Preview[]
    debridInstantAvailability: Record<string, Debrid_TorrentItemInstantAvailability>
    isLoading: boolean
    selectedTorrents: HibikeTorrent_AnimeTorrent[]
    onToggleTorrent: (t: HibikeTorrent_AnimeTorrent) => void
    type: TorrentSelectionType
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
    }: TorrentPreviewList) => {
    // Add sorting state
    const [sortField, setSortField] = useState<SortField>("seeders")
    const [sortDirection, setSortDirection] = useState<SortDirection>("desc")

    if (isLoading) return <div className="space-y-2">
        <Skeleton className="h-[96px]" />
        <Skeleton className="h-[96px]" />
        <Skeleton className="h-[96px]" />
        <Skeleton className="h-[96px]" />
    </div>

    if (!isLoading && !previews?.length) {
        return <LuffyError title="Nothing found" />
    }

    // Sort the previews based on current sort settings
    const sortedPreviews = sortPreviewTorrents(previews, sortField, sortDirection)

    return (
        <div className="space-y-2" data-torrent-preview-list>
            <div className="flex items-center justify-between gap-4">
                <p className="text-sm text-[--muted] flex-none" data-torrent-preview-list-results-count>
                    {previews?.length} results
                </p>
                <div className="flex items-center gap-1 flex-wrap">
                    <Button
                        size="xs"
                        intent="gray-basic"
                        leftIcon={<>
                            {/* <RiSeedlingLine className="mr-1 text-lg" /> */}
                            {getSortIcon(sortField, "seeders", sortDirection)}
                        </>}
                        onClick={() => handleSort("seeders", sortField, sortDirection, setSortField, setSortDirection)}
                    >
                        Seeders
                    </Button>
                    <Button
                        size="xs"
                        intent="gray-basic"
                        leftIcon={<>
                            {/* <LuFile className="mr-1 text-lg" /> */}
                            {getSortIcon(sortField, "size", sortDirection)}
                        </>}
                        onClick={() => handleSort("size", sortField, sortDirection, setSortField, setSortDirection)}
                    >
                        Size
                    </Button>
                    <Button
                        size="xs"
                        intent="gray-basic"
                        leftIcon={<>
                            {/* <BiCalendarAlt className="mr-1 text-lg" /> */}
                            {getSortIcon(sortField, "date", sortDirection)}
                        </>}
                        onClick={() => handleSort("date", sortField, sortDirection, setSortField, setSortDirection)}
                    >
                        Date
                    </Button>
                    <Button
                        size="xs"
                        intent="gray-basic"
                        leftIcon={<>
                            {/* <HiOutlineVideoCamera className="mr-1 text-lg" /> */}
                            {getSortIcon(sortField, "resolution", sortDirection)}
                        </>}
                        onClick={() => handleSort("resolution", sortField, sortDirection, setSortField, setSortDirection)}
                    >
                        Resolution
                    </Button>
                </div>
            </div>
            {/*<ScrollAreaBox className="h-[calc(100dvh_-_25rem)]">*/}
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
                        <div className="flex flex-wrap gap-3 items-center">
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
                    </TorrentPreviewItem>
                )
            })}
            {/*</div>*/}
            {/*</ScrollAreaBox>*/}
        </div>
    )

})
