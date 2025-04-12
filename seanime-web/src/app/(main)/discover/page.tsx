"use client"
import { useServerStatus } from "@/app/(main)/_hooks/use-server-status"
import { DiscoverPageHeader } from "@/app/(main)/discover/_components/discover-page-header"
import { DiscoverAiringSchedule } from "@/app/(main)/discover/_containers/discover-airing-schedule"
import { DiscoverMissedSequelsSection } from "@/app/(main)/discover/_containers/discover-missed-sequels"
import { DiscoverPastSeason } from "@/app/(main)/discover/_containers/discover-popular"
import { DiscoverTrending } from "@/app/(main)/discover/_containers/discover-trending"
import { DiscoverTrendingCountry } from "@/app/(main)/discover/_containers/discover-trending-country"
import { DiscoverTrendingMovies } from "@/app/(main)/discover/_containers/discover-trending-movies"
import { DiscoverUpcoming } from "@/app/(main)/discover/_containers/discover-upcoming"
import { __discord_pageTypeAtom } from "@/app/(main)/discover/_lib/discover.atoms"
import { RecentReleases } from "@/app/(main)/schedule/_containers/recent-releases"
import { PageWrapper } from "@/components/shared/page-wrapper"
import { Button } from "@/components/ui/button"
import { StaticTabs } from "@/components/ui/tabs"
import { AnimatePresence, motion } from "framer-motion"
import { useAtom } from "jotai/react"
import { useRouter, useSearchParams } from "next/navigation"
import React from "react"
import { FaSearch } from "react-icons/fa"

export const dynamic = "force-static"


export default function Page() {

    const serverStatus = useServerStatus()
    const router = useRouter()
    const [pageType, setPageType] = useAtom(__discord_pageTypeAtom)
    const searchParams = useSearchParams()
    const searchType = searchParams.get("type")

    React.useEffect(() => {
        if (searchType) {
            setPageType(searchType as any)
        }
    }, [searchParams])

    return (
        <>
            <DiscoverPageHeader />
            <motion.div
                initial={{ opacity: 0 }}
                animate={{ opacity: 1 }}
                transition={{ duration: 0.5, delay: 0.6 }}
                className="p-4 sm:p-8 space-y-10 pb-10 relative z-[4]"
                data-discover-page-container
            >
                <div
                    className="lg:absolute w-full lg:-top-10 left-0 flex gap-4 p-4 items-center justify-center flex-wrap"
                    data-discover-page-header-tabs-container
                >
                    <div className="max-w-fit border rounded-full" data-discover-page-header-tabs-inner-container>
                        <StaticTabs
                            className="h-10 overflow-hidden"
                            triggerClass="px-4 py-1"
                            items={[
                                { name: "Anime", isCurrent: pageType === "anime", onClick: () => setPageType("anime") },
                                { name: "Schedule", isCurrent: pageType === "schedule", onClick: () => setPageType("schedule") },
                                ...(serverStatus?.settings?.library?.enableManga ? [{
                                    name: "Manga",
                                    isCurrent: pageType === "manga",
                                    onClick: () => setPageType("manga"),
                                }] : []),
                            ]}
                        />
                    </div>
                    <div data-discover-page-header-advanced-search-container>
                        <Button
                            leftIcon={<FaSearch />}
                            intent="gray-outline"
                            // size="lg"
                            className="rounded-full"
                            onClick={() => router.push("/search")}
                        >
                            Advanced search
                        </Button>
                    </div>
                </div>
                <AnimatePresence mode="wait" initial={false}>
                    {pageType === "anime" && <PageWrapper
                        key="anime"
                        className="relative 2xl:order-first pb-10 pt-4"
                        {...{
                            initial: { opacity: 0, y: 60 },
                            animate: { opacity: 1, y: 0 },
                            exit: { opacity: 0, scale: 0.99 },
                            transition: {
                                duration: 0.35,
                            },
                        }}
                        data-discover-page-anime-container
                    >
                        <div className="space-y-2 z-[5] relative" data-discover-page-anime-trending-container>
                            <h2>Trending this season</h2>
                            <DiscoverTrending />
                        </div>
                        <RecentReleases />
                        <div className="space-y-2 z-[5] relative" data-discover-page-anime-highest-rated-container>
                            <h2>Highest rated last season</h2>
                            <DiscoverPastSeason />
                        </div>
                        <DiscoverMissedSequelsSection />
                        <div className="space-y-2 z-[5] relative" data-discover-page-anime-upcoming-container>
                            <h2>Upcoming</h2>
                            <DiscoverUpcoming />
                        </div>
                        <div className="space-y-2 z-[5] relative" data-discover-page-anime-trending-movies-container>
                            <h2>Trending movies</h2>
                            <DiscoverTrendingMovies />
                        </div>
                        {/*<div className="space-y-2 z-[5] relative">*/}
                        {/*    <h2>Popular shows</h2>*/}
                        {/*    <DiscoverPopular />*/}
                        {/*</div>*/}
                    </PageWrapper>}
                    {pageType === "schedule" && <PageWrapper
                        key="schedule"
                        className="relative 2xl:order-first pb-10 pt-4"
                        data-discover-page-schedule-container
                        {...{
                            initial: { opacity: 0, y: 60 },
                            animate: { opacity: 1, y: 0 },
                            exit: { opacity: 0, scale: 0.99 },
                            transition: {
                                duration: 0.35,
                            },
                        }}
                    >
                        <DiscoverAiringSchedule />
                    </PageWrapper>}
                    {pageType === "manga" && <PageWrapper
                        key="manga"
                        className="relative 2xl:order-first pb-10 pt-4"
                        data-discover-page-manga-container
                        {...{
                            initial: { opacity: 0, y: 60 },
                            animate: { opacity: 1, y: 0 },
                            exit: { opacity: 0, scale: 0.99 },
                            transition: {
                                duration: 0.35,
                            },
                        }}
                    >
                        {/*<div className="space-y-2 z-[5] relative">*/}
                        {/*    <h2>Trending right now</h2>*/}
                        {/*    <DiscoverTrendingMangaAll />*/}
                        {/*</div>*/}
                        <div className="space-y-2 z-[5] relative" data-discover-page-manga-trending-container>
                            <h2>Trending Manga</h2>
                            <DiscoverTrendingCountry country="JP" />
                        </div>
                        <div className="space-y-2 z-[5] relative" data-discover-page-manga-trending-manhwa-container>
                            <h2>Trending Manhwa</h2>
                            <DiscoverTrendingCountry country="KR" />
                        </div>
                        <div className="space-y-2 z-[5] relative" data-discover-page-manga-trending-manhua-container>
                            <h2>Trending Manhua</h2>
                            <DiscoverTrendingCountry country="CN" />
                        </div>
                        {/*<div className="space-y-2 z-[5] relative">*/}
                        {/*    <DiscoverMangaSearchBar />*/}
                        {/*</div>*/}
                    </PageWrapper>}
                </AnimatePresence>

            </motion.div>
        </>
    )
}
