import { SearchTorrent, TorrentPreview } from "@/lib/server/types"
import { TorrentPreviewItem } from "@/app/(main)/entry/_containers/torrent-search/_components/torrent-preview-item"
import {
    TorrentResolutionBadge,
    TorrentSeedersBadge,
} from "@/app/(main)/entry/_containers/torrent-search/_components/torrent-item-badges"
import React, { memo } from "react"
import { LoadingSpinner } from "@/components/ui/loading-spinner"
import { IconButton } from "@/components/ui/button"
import { BiLinkExternal } from "@react-icons/all-files/bi/BiLinkExternal"
import { Tooltip } from "@/components/ui/tooltip"
import formatDistanceToNow from "date-fns/formatDistanceToNow"
import { BiCalendarAlt } from "@react-icons/all-files/bi/BiCalendarAlt"
import { BiFile } from "@react-icons/all-files/bi/BiFile"

type TorrentPreviewList = {
    previews: TorrentPreview[],
    isLoading: boolean
    selectedTorrents: SearchTorrent[]
    onToggleTorrent: (t: SearchTorrent) => void
}

export const TorrentPreviewList = memo((
    {
        previews,
        isLoading,
        selectedTorrents,
        onToggleTorrent,
    }: TorrentPreviewList) => {

    if (isLoading) return <LoadingSpinner/>

    return (
        <div className="space-y-2">
            {previews.filter(Boolean).map(item => {
                return (
                    <TorrentPreviewItem
                        key={item.torrent.guid}
                        title={item.episode?.displayTitle || ""}
                        releaseGroup={item.releaseGroup}
                        filename={item.torrent.name}
                        isBatch={item.isBatch}
                        image={item.episode?.episodeMetadata?.image}
                        isSelected={selectedTorrents.findIndex(n => n.guid === item.torrent.guid) !== -1}
                        onClick={() => onToggleTorrent(item.torrent)}
                        action={<Tooltip trigger={<IconButton
                            icon={<BiLinkExternal/>}
                            intent={"primary-basic"}
                            size={"sm"}
                            onClick={() => window.open(item.torrent.guid, "_blank")}
                        />}>View on NYAA</Tooltip>}
                    >
                        <TorrentResolutionBadge resolution={item.resolution}/>
                        <TorrentSeedersBadge seeders={item.torrent.seeders}/>
                        <p className="text-gray-300 text-sm flex items-center gap-1">
                            <BiFile/> {item.torrent.size}</p>
                        <p className="text-[--muted] text-sm flex items-center gap-1">
                            - <BiCalendarAlt/> {formatDistanceToNow(new Date(item.torrent.date), { addSuffix: true })}
                        </p>
                    </TorrentPreviewItem>
                )
            })}
        </div>
    )

})