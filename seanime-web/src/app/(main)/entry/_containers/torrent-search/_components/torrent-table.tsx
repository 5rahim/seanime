import {
    Anime_Entry,
    Debrid_TorrentItemInstantAvailability,
    HibikeTorrent_AnimeTorrent,
    Metadata_AnimeMetadata,
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
    animeMetadata: Metadata_AnimeMetadata | undefined
    torrentMetadata: Record<string, Torrent_TorrentMetadata> | undefined
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
        animeMetadata,
        torrentMetadata,
    }: TorrentTable) => {
    // Use hooks for sorting and filtering
    const { sortField, sortDirection, handleSortChange } = useTorrentSorting()
    const { filters, handleFilterChange } = useTorrentFiltering()

    // Apply filters using the generic helper
    const filteredTorrents = filterItems(torrents, torrentMetadata, filters)

    // Sort the torrents after filtering using the generic helper
    const sortedTorrents = sortItems(filteredTorrents, sortField, sortDirection)

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
                    <TorrentFilterSortControls
                        resultCount={sortedTorrents?.length || 0}
                        sortField={sortField}
                        sortDirection={sortDirection}
                        filters={filters}
                        onSortChange={handleSortChange}
                        onFilterChange={handleFilterChange}
                    />
                    <ScrollAreaBox className="h-[calc(100dvh_-_25rem)]">
                        {sortedTorrents.map(torrent => {
                            const parsedEpisodeNumberStr = torrentMetadata?.[torrent.infoHash!]?.metadata?.episode_number?.[0]
                            const parsedEpisodeNumber = parsedEpisodeNumberStr ? parseInt(parsedEpisodeNumberStr) : undefined
                            const releaseGroup = torrent.releaseGroup || torrentMetadata?.[torrent.infoHash!]?.metadata?.release_group || ""
                            let episodeNumber = torrent.episodeNumber || parsedEpisodeNumber || -1
                            let totalEpisodes = entry?.media?.episodes || (entry?.media?.nextAiringEpisode?.episode ? entry?.media?.nextAiringEpisode?.episode : 0)
                            if (episodeNumber > totalEpisodes) {
                                // normalize episode number
                                for (const epKey in animeMetadata?.episodes) {
                                    const ep = animeMetadata?.episodes?.[epKey]
                                    if (ep?.absoluteEpisodeNumber === episodeNumber) {
                                        episodeNumber = ep.episodeNumber
                                    }
                                }
                            }

                            let episodeImage: string | undefined
                            if (!!animeMetadata && (episodeNumber ?? -1) >= 0) {
                                const episode = animeMetadata.episodes?.[episodeNumber!.toString()]
                                if (episode) {
                                    episodeImage = episode.image
                                }
                            }
                            let distance = 9999
                            if (!!torrentMetadata && !!torrent.infoHash) {
                                const metadata = torrentMetadata[torrent.infoHash!]
                                if (metadata) {
                                    distance = metadata.distance
                                }
                            }
                            if (distance > 20) {
                                episodeImage = undefined
                            }
                            return (
                                <TorrentPreviewItem
                                    // isBasic
                                    link={torrent.link}
                                    key={torrent.link}
                                    title={torrent.name}
                                    releaseGroup={releaseGroup}
                                    subtitle={torrent.isBatch ? torrent.name : (episodeNumber ?? -1) >= 0
                                        ? `Episode ${episodeNumber}`
                                        : ""}
                                    isBatch={torrent.isBatch ?? false}
                                    isBestRelease={torrent.isBestRelease}
                                    image={distance <= 20 ? episodeImage : undefined}
                                    fallbackImage={(entry?.media?.coverImage?.large || entry?.media?.bannerImage)}
                                    isSelected={selectedTorrents.findIndex(n => n.link === torrent!.link) !== -1}
                                    onClick={() => onToggleTorrent(torrent!)}
                                    // confirmed={distance === 0}
                                >
                                    <div className="flex flex-wrap gap-2 items-center">
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
                                    <TorrentParsedMetadata metadata={torrentMetadata?.[torrent.infoHash!]} />
                                </TorrentPreviewItem>
                            )
                        })}
                    </ScrollAreaBox>
                </>
            )}
        </>
    )

})
