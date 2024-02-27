import { MediaEntryDownloadInfo } from "@/lib/server/types"
import { EpisodeListItem } from "@/components/shared/episode-list-item"
import React, { startTransition } from "react"
import { BaseMediaFragment } from "@/lib/anilist/gql/graphql"
import { BiDownload } from "react-icons/bi"
import { BiCalendarAlt } from "react-icons/bi"
import {
    torrentSearchDrawerEpisodeAtom,
    torrentSearchDrawerIsOpenAtom,
} from "@/app/(main)/entry/_containers/torrent-search/torrent-search-drawer"
import { useSetAtom } from "jotai"

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
            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                {episodes?.sort((a, b) => a.episodeNumber - b.episodeNumber).map((ep, idx) => {
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
                                <a
                                    onClick={() => {
                                        setTorrentSearchEpisode(episode.episodeNumber)
                                        startTransition(() => {
                                            setTorrentSearchIsOpen(true)
                                        })
                                    }}
                                    className="text-orange-200 absolue top-1 right-1 text-3xl absolute animate-pulse cursor-pointer"
                                >
                                    <BiDownload/>
                                </a>
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
            </div>
        </div>
    )

}
