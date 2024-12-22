import { Anime_Entry, Debrid_TorrentItemInstantAvailability, HibikeTorrent_AnimeTorrent, Torrent_Preview } from "@/api/generated/types"
import {
    TorrentDebridInstantAvailabilityBadge,
    TorrentResolutionBadge,
    TorrentSeedersBadge,
} from "@/app/(main)/entry/_containers/torrent-search/_components/torrent-item-badges"
import { TorrentPreviewItem } from "@/app/(main)/entry/_containers/torrent-search/_components/torrent-preview-item"
import { Badge } from "@/components/ui/badge"
import { IconButton } from "@/components/ui/button"
import { Skeleton } from "@/components/ui/skeleton"
import { Tooltip } from "@/components/ui/tooltip"
import { openTab } from "@/lib/helpers/browser"
import { formatDistanceToNowSafe } from "@/lib/helpers/date"
import React from "react"
import { BiCalendarAlt, BiFile, BiLinkExternal } from "react-icons/bi"

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


    return (
        <div className="space-y-2">
            {previews.filter(Boolean).map(item => {
                if (!item.torrent) return null
                // const isReleasedBeforeMedia = differenceInCalendarYears(mediaReleaseDate, item.torrent.date) > 2
                return (
                    <TorrentPreviewItem
                        confirmed={item.torrent?.confirmed}
                        key={item.torrent.link}
                        title={item.episode?.displayTitle || item.episode?.baseAnime?.title?.userPreferred || ""}
                        releaseGroup={item.torrent.releaseGroup || ""}
                        filename={item.torrent.name}
                        isBatch={item.torrent.isBatch ?? false}
                        isBestRelease={item.torrent.isBestRelease}
                        image={item.episode?.episodeMetadata?.image || item.episode?.baseAnime?.coverImage?.large ||
                            (item.torrent.confirmed ? (entry.media?.coverImage?.large || entry.media?.bannerImage) : null)}
                        fallbackImage={entry.media?.coverImage?.large || entry.media?.bannerImage}
                        isSelected={selectedTorrents.findIndex(n => n.link === item.torrent!.link) !== -1}
                        onClick={() => onToggleTorrent(item.torrent!)}
                        action={<Tooltip
                            side="left"
                            trigger={<IconButton
                                icon={<BiLinkExternal />}
                                intent="primary-basic"
                                size="sm"
                                onClick={() => openTab(item.torrent!.link)}
                            />}
                        >Open in browser</Tooltip>}
                    >
                        <div className="flex flex-wrap gap-2 items-center">
                            {item.torrent.isBestRelease && (
                                <Badge
                                    className="rounded-md text-[0.8rem] bg-pink-800 border-pink-600 border"
                                    intent="success-solid"
                                >
                                    Best release
                                </Badge>
                            )}
                            <TorrentResolutionBadge resolution={item.torrent.resolution} />
                            <TorrentSeedersBadge seeders={item.torrent.seeders} />
                            {(!!item.torrent.infoHash && debridInstantAvailability[item.torrent.infoHash]) && (
                                <TorrentDebridInstantAvailabilityBadge />
                            )}
                            {!!item.torrent.size && <p className="text-gray-300 text-sm flex items-center gap-1">
                                <BiFile /> {item.torrent.formattedSize}</p>}
                            <p className="text-[--muted] text-sm flex items-center gap-1">
                                <BiCalendarAlt /> {formatDistanceToNowSafe(item.torrent.date)}
                            </p>
                        </div>
                    </TorrentPreviewItem>
                )
            })}
        </div>
    )

})
