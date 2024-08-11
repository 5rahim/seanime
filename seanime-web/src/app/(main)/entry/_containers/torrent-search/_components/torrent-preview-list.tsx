import { Anime_AnimeEntry, HibikeTorrent_AnimeTorrent, Torrent_Preview } from "@/api/generated/types"
import { TorrentResolutionBadge, TorrentSeedersBadge } from "@/app/(main)/entry/_containers/torrent-search/_components/torrent-item-badges"
import { TorrentPreviewItem } from "@/app/(main)/entry/_containers/torrent-search/_components/torrent-preview-item"
import { Badge } from "@/components/ui/badge"
import { IconButton } from "@/components/ui/button"
import { Skeleton } from "@/components/ui/skeleton"
import { Tooltip } from "@/components/ui/tooltip"
import { formatDistanceToNowSafe } from "@/lib/helpers/date"
import React from "react"
import { BiCalendarAlt, BiFile, BiLinkExternal } from "react-icons/bi"

type TorrentPreviewList = {
    entry: Anime_AnimeEntry
    previews: Torrent_Preview[],
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
    }: TorrentPreviewList) => {

    if (isLoading) return <div className="space-y-2">
        <Skeleton className="h-[96px]" />
        <Skeleton className="h-[96px]" />
        <Skeleton className="h-[96px]" />
        <Skeleton className="h-[96px]" />
    </div>

    return (
        <div className="space-y-2">
            {previews.filter(Boolean).map(item => {
                if (!item.torrent) return null
                return (
                    <TorrentPreviewItem
                        confirmed={item.torrent?.confirmed}
                        key={item.torrent.link + item.episode?.displayTitle}
                        title={item.episode?.displayTitle || item.episode?.baseAnime?.title?.userPreferred || ""}
                        releaseGroup={item.torrent.releaseGroup || ""}
                        filename={item.torrent.name}
                        isBatch={item.torrent.isBatch ?? false}
                        image={item.episode?.episodeMetadata?.image || item.episode?.baseAnime?.coverImage?.large ||
                            (item.torrent.confirmed ? (entry.media?.coverImage?.large || entry.media?.bannerImage) : null)}
                        fallbackImage={entry.media?.coverImage?.large || entry.media?.bannerImage}
                        isSelected={selectedTorrents.findIndex(n => n.link === item.torrent!.link) !== -1}
                        onClick={() => onToggleTorrent(item.torrent!)}
                        action={<Tooltip
                            trigger={<IconButton
                                icon={<BiLinkExternal />}
                                intent="primary-basic"
                                size="sm"
                                onClick={() => window.open(item.torrent!.link, "_blank")}
                            />}
                        >Open in browser</Tooltip>}
                    >
                        <div className="flex flex-wrap gap-2 items-center">
                            {item.torrent.isBestRelease && (
                                <Badge
                                    className="rounded-md text-[0.8rem] bg-green-700 border-green-400 border"
                                    intent="success-solid"
                                >
                                    Best release
                                </Badge>
                            )}
                            <TorrentResolutionBadge resolution={item.torrent.resolution} />
                            <TorrentSeedersBadge seeders={item.torrent.seeders} />
                            {!!item.torrent.formattedSize && <p className="text-gray-300 text-sm flex items-center gap-1">
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
