import { useGetAnilistAnimeDetails } from "@/api/hooks/anilist.hooks"
import { useGetAnimeEntry } from "@/api/hooks/anime_entries.hooks"
import { MediaEntryCharactersSection } from "@/app/(main)/_features/media/_components/media-entry-characters-section"
import { MediaEntryPageLoadingDisplay } from "@/app/(main)/_features/media/_components/media-entry-page-loading-display"
import { useServerStatus } from "@/app/(main)/_hooks/use-server-status"
import { MetaSection } from "@/app/(main)/entry/_components/meta-section"
import { RelationsRecommendationsSection } from "@/app/(main)/entry/_components/relations-recommendations-section"
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

    const serverStatus = useServerStatus()
    const router = useRouter()
    const searchParams = useSearchParams()
    const mediaId = searchParams.get("id")
    const { data: animeEntry, isLoading: animeEntryLoading } = useGetAnimeEntry(mediaId)
    const { data: animeDetails, isLoading: animeDetailsLoading } = useGetAnilistAnimeDetails(mediaId)

    const [isTorrentStreamingView, setIsTorrentStreamingView] = useAtom(__anime_torrentStreamingViewActiveAtom)

    const switchedView = React.useRef(false)
    React.useLayoutEffect(() => {
        if (!animeEntryLoading &&
            animeEntry?.media?.status !== "NOT_YET_RELEASED" &&
            !animeEntry?.libraryData &&
            !isTorrentStreamingView &&
            (serverStatus?.torrentstreamSettings?.enabled && serverStatus?.torrentstreamSettings?.fallbackToTorrentStreamingView) &&
            !switchedView.current
        ) {
            switchedView.current = true
            setIsTorrentStreamingView(true)
        }
    }, [animeEntryLoading, searchParams, serverStatus?.torrentstreamSettings?.fallbackToTorrentStreamingView, isTorrentStreamingView])

    React.useEffect(() => {
        if (!mediaId) {
            router.push("/")
        } else if (!animeEntryLoading && !animeEntry) {
            router.push("/")
        }
    }, [animeEntry, animeEntryLoading])

    useUnmount(() => {
        setIsTorrentStreamingView(false)
    })

    if (animeEntryLoading || animeDetailsLoading) return <MediaEntryPageLoadingDisplay />
    if (!animeEntry) return null

    return (
        <div>
            <MetaSection entry={animeEntry} details={animeDetails} />

            <div className="px-4 md:px-8 relative z-[8]">
                <PageWrapper
                    className="relative 2xl:order-first pb-10"
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
                            className="relative 2xl:order-first pb-10"
                            {...{
                                initial: { opacity: 0, y: 60 },
                                animate: { opacity: 1, y: 0 },
                                exit: { opacity: 0, scale: 0.99 },
                                transition: {
                                    duration: 0.35,
                                },
                            }}
                        >
                            <EpisodeSection
                                entry={animeEntry}
                                details={animeDetails}
                                bottomSection={<>
                                    <MediaEntryCharactersSection details={animeDetails} />
                                    <RelationsRecommendationsSection entry={animeEntry} details={animeDetails} />
                                </>}
                            />
                        </PageWrapper>}

                        {isTorrentStreamingView && <PageWrapper
                            key="torrent-streaming-episodes"
                            className="relative 2xl:order-first pb-10 lg:pt-0"
                            {...{
                                initial: { opacity: 0, y: 60 },
                                animate: { opacity: 1, y: 0 },
                                exit: { opacity: 0, scale: 0.99 },
                                transition: {
                                    duration: 0.35,
                                },
                            }}
                        >
                            <TorrentStreamPage
                                entry={animeEntry}
                                bottomSection={<>
                                    <MediaEntryCharactersSection details={animeDetails} />
                                    <RelationsRecommendationsSection entry={animeEntry} details={animeDetails} />
                                </>}
                            />
                        </PageWrapper>}

                    </AnimatePresence>
                </PageWrapper>
            </div>

            <TorrentSearchDrawer entry={animeEntry} />
        </div>
    )
}

