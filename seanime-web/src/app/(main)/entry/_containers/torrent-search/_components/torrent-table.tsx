import { Anime_Entry, Debrid_TorrentItemInstantAvailability, HibikeTorrent_AnimeTorrent, Metadata_AnimeMetadata, Torrent_TorrentMetadata } from "@/api/generated/types"
import {
    TorrentDebridInstantAvailabilityBadge,
    TorrentParsedMetadata,
    TorrentResolutionBadge,
    TorrentSeedersBadge,
} from "@/app/(main)/entry/_containers/torrent-search/_components/torrent-item-badges"
import {
    getSortIcon,
    handleSort,
    SortDirection,
    SortField,
    sortTorrents,
} from "@/app/(main)/entry/_containers/torrent-search/_components/torrent-sorting-helpers"
import { LuffyError } from "@/components/shared/luffy-error"
import { ScrollAreaBox } from "@/components/shared/scroll-area-box"
import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import { Skeleton } from "@/components/ui/skeleton"
import { TextInput } from "@/components/ui/text-input"
import { formatDistanceToNowSafe } from "@/lib/helpers/date"
import React, { memo, useState } from "react"
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
    // Add sorting state
    const [sortField, setSortField] = useState<SortField>("seeders")
    const [sortDirection, setSortDirection] = useState<SortDirection>("desc")

    // Sort the torrents
    const sortedTorrents = sortTorrents(torrents, sortField, sortDirection)

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
                    <div className="flex items-center justify-between gap-4">
                        <p className="text-sm text-[--muted] flex-none">{torrents?.length} results</p>
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
                    <ScrollAreaBox className="h-[calc(100dvh_-_25rem)]">
                        {sortedTorrents.map(torrent => {
                            let episodeNumber = torrent.episodeNumber || -1
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
                            if (distance > 10) {
                                episodeImage = undefined
                            }
                            return (
                                <TorrentPreviewItem
                                    // isBasic
                                    link={torrent.link}
                                    key={torrent.link}
                                    title={torrent.name}
                                    releaseGroup={torrent.releaseGroup || ""}
                                    subtitle={torrent.isBatch ? torrent.name : (episodeNumber ?? -1) >= 0
                                        ? `Episode ${episodeNumber}`
                                        : ""}
                                    isBatch={torrent.isBatch ?? false}
                                    isBestRelease={torrent.isBestRelease}
                                    image={distance <= 10 ? episodeImage : undefined}
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


            {/* <DataGrid<HibikeTorrent_AnimeTorrent>*/}
            {/*     columns={columns}*/}
            {/*     data={torrents?.slice(0, 20)}*/}
            {/*     rowCount={torrents?.length ?? 0}*/}
            {/*     initialState={{*/}
            {/*         pagination: {*/}
            {/*             pageSize: 20,*/}
            {/*             pageIndex: 0,*/}
            {/*         },*/}
            {/*     }}*/}
            {/*     tdClass="py-4 data-[row-selected=true]:bg-gray-900"*/}
            {/*     tableBodyClass="bg-transparent"*/}
            {/*     footerClass="hidden"*/}
            {/*     state={{*/}
            {/*         globalFilter,*/}
            {/*     }}*/}
            {/*     enableManualFiltering={true}*/}
            {/*     onGlobalFilterChange={setGlobalFilter}*/}
            {/*     isLoading={isLoading || isFetching}*/}
            {/*     isDataMutating={isFetching}*/}
            {/*     hideGlobalSearchInput={smartSearch}*/}
            {/* />*/}
        </>
    )

})

