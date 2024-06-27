"use client"
import { useServerStatus } from "@/app/(main)/_hooks/use-server-status"
import { DiscoverPageHeader } from "@/app/(main)/discover/_components/discover-page-header"
import { DiscoverPastSeason, DiscoverPopular } from "@/app/(main)/discover/_containers/discover-popular"
import { DiscoverTrending } from "@/app/(main)/discover/_containers/discover-trending"
import { DiscoverMangaSearchBar, DiscoverTrendingManga } from "@/app/(main)/discover/_containers/discover-trending-manga"
import { DiscoverTrendingMovies } from "@/app/(main)/discover/_containers/discover-trending-movies"
import { DiscoverUpcoming } from "@/app/(main)/discover/_containers/discover-upcoming"
import { __discord_pageTypeAtom } from "@/app/(main)/discover/_lib/discover.atoms"
import { RecentReleases } from "@/app/(main)/schedule/_containers/recent-releases"
import { PageWrapper } from "@/components/shared/page-wrapper"
import { Button } from "@/components/ui/button"
import { StaticTabs } from "@/components/ui/tabs"
import { AnimatePresence, motion } from "framer-motion"
import { useAtom } from "jotai/react"
import { useRouter } from "next/navigation"
import React from "react"
import { FaSearch } from "react-icons/fa"

export const dynamic = "force-static"


export default function Page() {

    const serverStatus = useServerStatus()
    const router = useRouter()
    const [pageType, setPageType] = useAtom(__discord_pageTypeAtom)

    return (
        <>
            <DiscoverPageHeader />
            <motion.div
                initial={{ opacity: 0 }}
                animate={{ opacity: 1 }}
                transition={{ duration: 0.5, delay: 0.6 }}
                className="p-4 sm:p-8 space-y-10 pb-10 relative z-[4]"
            >
                <div className="lg:absolute w-full lg:-top-10 left-0 flex gap-4 p-4 items-center">
                    {serverStatus?.settings?.library?.enableManga && <div className="max-w-fit border rounded-full">
                        <StaticTabs
                            className="h-10"
                            triggerClass="px-4 py-1"
                            items={[
                                { name: "Anime", isCurrent: pageType === "anime", onClick: () => setPageType("anime") },
                                { name: "Manga", isCurrent: pageType === "manga", onClick: () => setPageType("manga") },
                            ]}
                        />
                    </div>}
                    <div>
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
                    >
                        <div className="space-y-2 z-[5] relative">
                            <h2>Trending this season</h2>
                            <DiscoverTrending />
                        </div>
                        <RecentReleases />
                        <div className="space-y-2 z-[5] relative">
                            <h2>Highest rated last season</h2>
                            <DiscoverPastSeason />
                        </div>
                        <div className="space-y-2 z-[5] relative">
                            <h2>Upcoming</h2>
                            <DiscoverUpcoming />
                        </div>
                        <div className="space-y-2 z-[5] relative">
                            <h2>Trending movies</h2>
                            <DiscoverTrendingMovies />
                        </div>
                        <div className="space-y-2 z-[5] relative">
                            <h2>Popular shows</h2>
                            <DiscoverPopular />
                        </div>
                    </PageWrapper>}
                    {pageType === "manga" && <PageWrapper
                        key="manga"
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
                        <div className="space-y-2 z-[5] relative">
                            <h2>Trending right now</h2>
                            <DiscoverTrendingManga />
                        </div>
                        <div className="space-y-2 z-[5] relative">
                            <DiscoverMangaSearchBar />
                        </div>
                    </PageWrapper>}
                </AnimatePresence>

            </motion.div>
        </>
    )
}
