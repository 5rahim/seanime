import { useGetAnilistMediaDetails } from "@/api/hooks/anilist.hooks"
import { useGetAnimeEntry } from "@/api/hooks/anime_entries.hooks"
import { MediaEntryPageLoadingDisplay } from "@/app/(main)/_features/media/_components/media-entry-page-loading-display"
import { MetaSection } from "@/app/(main)/entry/_components/meta-section"
import { EpisodeSection } from "@/app/(main)/entry/_containers/episode-list/episode-section"
import { TorrentSearchDrawer } from "@/app/(main)/entry/_containers/torrent-search/torrent-search-drawer"
import { TorrentStreamPage } from "@/app/(main)/entry/_containers/torrent-stream/torrent-stream-page"
import { PageWrapper } from "@/components/shared/page-wrapper"
import { AnimatePresence } from "framer-motion"
import { atom } from "jotai"
import { useAtom } from "jotai/react"
import { useRouter, useSearchParams } from "next/navigation"
import React from "react"
import { useUnmount } from "react-use"

export const __anime_torrentStreamingViewActiveAtom = atom(false)

export function AnimeEntryPage() {

    const router = useRouter()
    const searchParams = useSearchParams()
    const mediaId = searchParams.get("id")
    const { data: mediaEntry, isLoading: mediaEntryLoading } = useGetAnimeEntry(mediaId)
    const { data: mediaDetails, isLoading: mediaDetailsLoading } = useGetAnilistMediaDetails(mediaId)

    const [isTorrentStreamingView, setIsTorrentStreamingView] = useAtom(__anime_torrentStreamingViewActiveAtom)

    React.useEffect(() => {
        if (!mediaId) {
            router.push("/")
        } else if (!mediaEntryLoading && !mediaEntry) {
            router.push("/")
        }
    }, [mediaEntry, mediaEntryLoading])

    useUnmount(() => {
        setIsTorrentStreamingView(false)
    })

    if (mediaEntryLoading || mediaDetailsLoading) return <MediaEntryPageLoadingDisplay />
    if (!mediaEntry) return null

    return (
        <div>
            <MetaSection entry={mediaEntry} details={mediaDetails} />

            <div className="px-4 md:px-8 relative z-[8]">
                <PageWrapper
                    className="relative 2xl:order-first pb-10 pt-4"
                    {...{
                        initial: { opacity: 0, y: 60 },
                        animate: { opacity: 1, y: 0 },
                        exit: { opacity: 0, y: 60 },
                        transition: {
                            type: "spring",
                            damping: 10,
                            stiffness: 80,
                            delay: 0.6,
                        },
                    }}
                >
                    <AnimatePresence mode="wait" initial={false}>
                        {!isTorrentStreamingView && <PageWrapper
                            key="episode-list"
                            className="relative 2xl:order-first pb-10 pt-4"
                            {...{
                                initial: { opacity: 0, y: 60 },
                                animate: { opacity: 1, y: 0 },
                                exit: { opacity: 0, scale: 0.99 },
                                transition: {
                                    duration: 0.35,
                                },
                            }}
                        >
                            <EpisodeSection entry={mediaEntry} details={mediaDetails} />
                        </PageWrapper>}

                        {isTorrentStreamingView && <PageWrapper
                            key="torrent-streaming-episodes"
                            className="relative 2xl:order-first pb-10 pt-4"
                            {...{
                                initial: { opacity: 0, y: 60 },
                                animate: { opacity: 1, y: 0 },
                                exit: { opacity: 0, scale: 0.99 },
                                transition: {
                                    duration: 0.35,
                                },
                            }}
                        >
                            <TorrentStreamPage entry={mediaEntry} />
                        </PageWrapper>}

                    </AnimatePresence>
                </PageWrapper>
            </div>

            <TorrentSearchDrawer entry={mediaEntry} />
        </div>
    )
}

