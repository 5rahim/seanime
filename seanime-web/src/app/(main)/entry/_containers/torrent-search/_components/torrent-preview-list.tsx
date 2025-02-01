import { Anime_Entry, Debrid_TorrentItemInstantAvailability, HibikeTorrent_AnimeTorrent, Torrent_Preview } from "@/api/generated/types"
import {
    TorrentDebridInstantAvailabilityBadge,
    TorrentResolutionBadge,
    TorrentSeedersBadge,
} from "@/app/(main)/entry/_containers/torrent-search/_components/torrent-item-badges"
import { TorrentPreviewItem } from "@/app/(main)/entry/_containers/torrent-search/_components/torrent-preview-item"
import { LuffyError } from "@/components/shared/luffy-error"
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
}

export const TorrentPreviewList = React.memo((
    {
        entry,
        previews,
        isLoading,
        selectedTorrents,
        onToggleTorrent,
        debridInstantAvailability,
    }: TorrentPreviewList) => {

    if (isLoading) return <div className="space-y-2">
        <Skeleton className="h-[96px]" />
        <Skeleton className="h-[96px]" />
        <Skeleton className="h-[96px]" />
        <Skeleton className="h-[96px]" />
    </div>

    // const mediaReleaseDate = new Date(entry?.media?.startDate?.year || 0, entry?.media?.startDate?.month! - 1, entry?.media?.startDate?.day)

    if (!isLoading && !previews?.length) {
        return <LuffyError title="Nothing found" />
    }

    return (
        <div className="space-y-2">
            <p className="text-sm text-[--muted]">{previews?.length} results</p>
            {/*<ScrollAreaBox className="h-[calc(100dvh_-_25rem)]">*/}
            {previews.filter(Boolean).map(item => {
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
                            {(!!item.torrent.infoHash && debridInstantAvailability[item.torrent.infoHash]) && (
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
            {/*</ScrollAreaBox>*/}
        </div>
    )

})
