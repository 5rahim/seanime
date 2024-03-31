import { TorrentResolutionBadge, TorrentSeedersBadge } from "@/app/(main)/entry/_containers/torrent-search/_components/torrent-item-badges"
import { TorrentPreviewItem } from "@/app/(main)/entry/_containers/torrent-search/_components/torrent-preview-item"
import { IconButton } from "@/components/ui/button"
import { LoadingSpinner } from "@/components/ui/loading-spinner"
import { Tooltip } from "@/components/ui/tooltip"
import { formatDistanceToNowSafe } from "@/lib/helpers/date"
import { AnimeTorrent, TorrentPreview } from "@/lib/server/types"
import React from "react"
import { BiCalendarAlt, BiFile, BiLinkExternal } from "react-icons/bi"

type TorrentPreviewList = {
    previews: TorrentPreview[],
    isLoading: boolean
    selectedTorrents: AnimeTorrent[]
    onToggleTorrent: (t: AnimeTorrent) => void
}

export const TorrentPreviewList = React.memo((
    {
        previews,
        isLoading,
        selectedTorrents,
        onToggleTorrent,
    }: TorrentPreviewList) => {

    if (isLoading) return <LoadingSpinner />

    return (
        <div className="space-y-2">
            {previews.filter(Boolean).map(item => {
                return (
                    <TorrentPreviewItem
                        key={item.torrent.link}
                        title={item.episode?.displayTitle || item.episode?.basicMedia?.title?.userPreferred || ""}
                        releaseGroup={item.torrent.releaseGroup || ""}
                        filename={item.torrent.name}
                        isBatch={item.torrent.isBatch}
                        image={item.episode?.episodeMetadata?.image || item.episode?.basicMedia?.coverImage?.large}
                        isSelected={selectedTorrents.findIndex(n => n.link === item.torrent.link) !== -1}
                        onClick={() => onToggleTorrent(item.torrent)}
                        action={<Tooltip
                            trigger={<IconButton
                                icon={<BiLinkExternal />}
                                intent="primary-basic"
                                size="sm"
                                onClick={() => window.open(item.torrent.link, "_blank")}
                            />}
                        >Open in browser</Tooltip>}
                    >
                        <div className="flex flex-wrap gap-2 items-center">
                            <TorrentResolutionBadge resolution={item.torrent.resolution} />
                            <TorrentSeedersBadge seeders={item.torrent.seeders} />
                            <p className="text-gray-300 text-sm flex items-center gap-1">
                                <BiFile /> {item.torrent.formattedSize}</p>
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
