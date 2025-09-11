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
import { TorrentSelectionType } from "@/app/(main)/entry/_containers/torrent-search/torrent-search-drawer"
import { LuffyError } from "@/components/shared/luffy-error"
import { ScrollAreaBox } from "@/components/shared/scroll-area-box"
import { Skeleton } from "@/components/ui/skeleton"
import { anilist_animeIsSingleEpisode } from "@/lib/helpers/media"
import React, { memo } from "react"
import { TorrentList, TorrentListItem } from "./torrent-preview-item"

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
    type: TorrentSelectionType
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
        type,
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
                        <TorrentList>
                            {sortedTorrents.map(torrent => {
                                const metadata = torrentMetadata?.[torrent.infoHash!]
                                const parsedEpisodeNumberStr = metadata?.metadata?.episode_number?.[0]
                                const parsedEpisodeNumber = parsedEpisodeNumberStr ? parseInt(parsedEpisodeNumberStr) : undefined
                                const releaseGroup = torrent.releaseGroup || metadata?.metadata?.release_group || ""
                                let episodeNumber = torrent.episodeNumber ?? parsedEpisodeNumber ?? -1
                                let totalEpisodes = entry?.media?.episodes || (entry?.media?.nextAiringEpisode?.episode
                                    ? entry?.media?.nextAiringEpisode?.episode
                                    : 0)
                                if (episodeNumber > totalEpisodes) {
                                    // normalize episode number
                                    for (const epKey in animeMetadata?.episodes) {
                                        const ep = animeMetadata?.episodes?.[epKey]
                                        if (ep?.absoluteEpisodeNumber === episodeNumber) {
                                            episodeNumber = ep.episodeNumber
                                        }
                                    }
                                }

                                const isBatch = torrent.isBatch ?? (!anilist_animeIsSingleEpisode(entry?.media) && (metadata?.metadata?.episode_number?.length ?? 0) > 1 || (metadata?.metadata?.episode_number?.length ?? 0) == 0)

                                let episodeImage: string | undefined
                                if (!!animeMetadata && (episodeNumber ?? -1) >= 0) {
                                    const episode = animeMetadata.episodes?.[episodeNumber!.toString()]
                                    if (episode) {
                                        episodeImage = episode.image
                                    }
                                }
                                let distance = 9999
                                if (!!torrentMetadata && !!torrent.infoHash) {
                                    if (metadata) {
                                        distance = metadata.distance
                                    }
                                }
                                if (distance > 20) {
                                    episodeImage = undefined
                                }
                                return (
                                    <TorrentListItem
                                        key={torrent.link}
                                        torrent={torrent}
                                        metadata={torrentMetadata?.[torrent.infoHash!]?.metadata}
                                        media={entry?.media}
                                        episode={undefined}
                                        debridCached={((type === "download" || type === "debridstream-select" || type === "debridstream-select-file") && !!torrent.infoHash && !!debridInstantAvailability[torrent.infoHash])}
                                        isSelected={selectedTorrents.findIndex(n => n.link === torrent!.link) !== -1}
                                        onClick={() => onToggleTorrent(torrent!)}
                                        overrideProps={{
                                            releaseGroup: releaseGroup,
                                            displayName: (episodeNumber ?? -1) >= 0
                                                ? `Episode ${episodeNumber}`
                                                : "",
                                            isBatch: torrent.isBestRelease ? true : isBatch,
                                            image: distance <= 20 ? episodeImage : undefined,
                                            fallbackImage: (entry?.media?.coverImage?.large || entry?.media?.bannerImage),
                                            confirmed: distance === 0,
                                        }}
                                    />
                                )
                            })}
                        </TorrentList>
                    </ScrollAreaBox>
                </>
            )}
        </>
    )

})
