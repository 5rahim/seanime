import { EpisodeListGrid } from "@/app/(main)/entry/_components/episode-list-grid"
import { torrentSearchDrawerEpisodeAtom, torrentSearchDrawerIsOpenAtom } from "@/app/(main)/entry/_containers/torrent-search/torrent-search-drawer"
import { EpisodeListItem } from "@/components/shared/episode-list-item"
import { BaseMediaFragment } from "@/lib/anilist/gql/graphql"

import { MediaEntryDownloadInfo } from "@/app/(main)/(library)/_lib/anime-library.types"
import { useSetAtom } from "jotai"
import React, { startTransition } from "react"
import { BiCalendarAlt, BiDownload } from "react-icons/bi"

export function UndownloadedEpisodeList({ downloadInfo, media }: {
    downloadInfo: MediaEntryDownloadInfo | undefined,
    media: BaseMediaFragment
}) {

    const episodes = downloadInfo?.episodesToDownload

    const setTorrentSearchIsOpen = useSetAtom(torrentSearchDrawerIsOpenAtom)
    const setTorrentSearchEpisode = useSetAtom(torrentSearchDrawerEpisodeAtom)

    if (!episodes?.length) return null

    return (
        <div className="space-y-4">
            <p className={""}>
                {downloadInfo?.rewatch ? "You have not downloaded the following:" : "You have not watched nor downloaded the following:"}
            </p>
            <EpisodeListGrid>
                {episodes?.sort((a, b) => a.episodeNumber - b.episodeNumber).slice(0, 28).map((ep, idx) => {
                    if (!ep.episode) return null
                    const episode = ep.episode
                    return (
                        <EpisodeListItem
                            key={ep.episode.localFile?.path || idx}
                            media={media}
                            image={episode.episodeMetadata?.image}
                            isInvalid={episode.isInvalid}
                            title={episode.displayTitle}
                            episodeTitle={episode.episodeTitle}
                            action={<div className={""}>
                                <div
                                    onClick={() => {
                                        setTorrentSearchEpisode(episode.episodeNumber)
                                        startTransition(() => {
                                            setTorrentSearchIsOpen(true)
                                        })
                                    }}
                                    className="inline-block cursor-pointer text-orange-200 absolue top-1 right-1 text-3xl absolute animate-pulse cursor-pointer"
                                >
                                    <BiDownload/>
                                </div>
                            </div>}
                        >
                            <div className="mt-1">
                                <p className="flex gap-1 items-center text-sm text-[--muted]">
                                    <BiCalendarAlt/> {episode.episodeMetadata?.airDate ? `Aired on ${new Date(episode.episodeMetadata?.airDate).toLocaleDateString()}` : "Aired"}
                                </p>
                            </div>
                        </EpisodeListItem>
                    )
                })}
            </EpisodeListGrid>
            {episodes.length > 28 && <h3>And more...</h3>}
        </div>
    )

}
