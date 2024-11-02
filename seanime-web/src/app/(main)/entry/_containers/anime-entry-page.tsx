import { useGetAnilistAnimeDetails } from "@/api/hooks/anilist.hooks"
import { useGetAnimeEntry } from "@/api/hooks/anime_entries.hooks"
import { MediaEntryCharactersSection } from "@/app/(main)/_features/media/_components/media-entry-characters-section"
import { MediaEntryPageLoadingDisplay } from "@/app/(main)/_features/media/_components/media-entry-page-loading-display"
import { useServerStatus } from "@/app/(main)/_hooks/use-server-status"
import { MetaSection } from "@/app/(main)/entry/_components/meta-section"
import { RelationsRecommendationsSection } from "@/app/(main)/entry/_components/relations-recommendations-section"
import { DebridStreamPage } from "@/app/(main)/entry/_containers/debrid-stream/debrid-stream-page"
import { EpisodeSection } from "@/app/(main)/entry/_containers/episode-list/episode-section"
import { TorrentSearchDrawer } from "@/app/(main)/entry/_containers/torrent-search/torrent-search-drawer"
import { TorrentStreamPage } from "@/app/(main)/entry/_containers/torrent-stream/torrent-stream-page"
import { OnlinestreamPage } from "@/app/(main)/onlinestream/_containers/onlinestream-page"
import { PageWrapper } from "@/components/shared/page-wrapper"
import { AnimatePresence } from "framer-motion"
import { atom } from "jotai"
import { useAtom } from "jotai/react"
import { useRouter, useSearchParams } from "next/navigation"
import React from "react"
import { useUnmount } from "react-use"

export const __anime_entryPageViewAtom = atom<"library" | "torrentstream" | "debridstream" | "onlinestream">("library")

export function useAnimeEntryPageView() {
    const [currentView, setView] = useAtom(__anime_entryPageViewAtom)

    const isLibraryView = currentView === "library"
    const isTorrentStreamingView = currentView === "torrentstream"
    const isDebridStreamingView = currentView === "debridstream"
    const isOnlineStreamingView = currentView === "onlinestream"

    function toggleTorrentStreamingView() {
        setView(p => p === "torrentstream" ? "library" : "torrentstream")
    }

    function toggleDebridStreamingView() {
        setView(p => p === "debridstream" ? "library" : "debridstream")
    }

    function toggleOnlineStreamingView() {
        setView(p => p === "onlinestream" ? "library" : "onlinestream")
    }

    return {
        currentView,
        setView,
        isLibraryView,
        isTorrentStreamingView,
        isDebridStreamingView,
        isOnlineStreamingView,
        toggleTorrentStreamingView,
        toggleDebridStreamingView,
        toggleOnlineStreamingView,
    }
}

export function AnimeEntryPage() {

    const serverStatus = useServerStatus()
    const router = useRouter()
    const searchParams = useSearchParams()
    const mediaId = searchParams.get("id")
    const { data: animeEntry, isLoading: animeEntryLoading } = useGetAnimeEntry(mediaId)
    const { data: animeDetails, isLoading: animeDetailsLoading } = useGetAnilistAnimeDetails(mediaId)

    const { currentView, isLibraryView, setView } = useAnimeEntryPageView()

    const switchedView = React.useRef(false)
    React.useLayoutEffect(() => {
            if (!animeEntryLoading &&
                animeEntry?.media?.status !== "NOT_YET_RELEASED" && // Anime is not yet released
                !animeEntry?.libraryData && // Anime is not in library
                isLibraryView && // Current view is library
                (
                    // If any of the fallbacks are enabled and the view has not been switched yet
                    (serverStatus?.torrentstreamSettings?.enabled && serverStatus?.torrentstreamSettings?.includeInLibrary) ||
                    (serverStatus?.debridSettings?.enabled && serverStatus?.debridSettings?.includeDebridStreamInLibrary) ||
                    (serverStatus?.settings?.library?.enableOnlinestream && serverStatus?.settings?.library?.includeOnlineStreamingInLibrary)
                ) &&
                !switchedView.current // View has not been switched yet
            ) {
                switchedView.current = true
                if (serverStatus?.debridSettings?.enabled && serverStatus?.debridSettings?.includeDebridStreamInLibrary) {
                    setView("debridstream")
                } else if (serverStatus?.torrentstreamSettings?.enabled && serverStatus?.torrentstreamSettings?.includeInLibrary) {
                    setView("torrentstream")
                } else if (serverStatus?.settings?.library?.enableOnlinestream && serverStatus?.settings?.library?.includeOnlineStreamingInLibrary) {
                    setView("onlinestream")
                }
            }
        },
        [animeEntryLoading, searchParams, serverStatus?.torrentstreamSettings?.includeInLibrary, currentView])

    React.useEffect(() => {
        if (!mediaId || (!animeEntryLoading && !animeEntry)) {
            router.push("/")
        }
    }, [animeEntry, animeEntryLoading])

    // Reset view when unmounting
    useUnmount(() => {
        setView("library")
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
                        {(currentView === "library") && <PageWrapper
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

                        {currentView === "torrentstream" && <PageWrapper
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

                        {currentView === "debridstream" && <PageWrapper
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
                            <DebridStreamPage
                                entry={animeEntry}
                                bottomSection={<>
                                    <MediaEntryCharactersSection details={animeDetails} />
                                    <RelationsRecommendationsSection entry={animeEntry} details={animeDetails} />
                                </>}
                            />
                        </PageWrapper>}

                        {currentView === "onlinestream" && <PageWrapper
                            key="online-streaming-episodes"
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
                            <div className="space-y-4">
                                <div className="absolute right-0 top-[-3rem]">
                                    <h2 className="text-xl lg:text-3xl flex items-center gap-3">Online streaming</h2>
                                </div>
                                <OnlinestreamPage
                                    animeEntry={animeEntry}
                                    animeEntryLoading={animeEntryLoading}
                                    hideBackButton
                                />
                            </div>
                        </PageWrapper>}

                    </AnimatePresence>
                </PageWrapper>
            </div>

            <TorrentSearchDrawer entry={animeEntry} />
        </div>
    )
}

