"use client"
import { DiscoverPageHeader } from "@/app/(main)/discover/_components/discover-page-header"
import { DiscoverPastSeason, DiscoverPopular } from "@/app/(main)/discover/_containers/discover-popular"
import { DiscoverTrending } from "@/app/(main)/discover/_containers/discover-trending"
import { DiscoverTrendingMovies } from "@/app/(main)/discover/_containers/discover-trending-movies"
import { DiscoverUpcoming } from "@/app/(main)/discover/_containers/discover-upcoming"
import { motion } from "framer-motion"
import React from "react"

export const dynamic = "force-static"

export default function Page() {

    return (
        <>
            <DiscoverPageHeader />
            <motion.div
                initial={{ opacity: 0 }}
                animate={{ opacity: 1 }}
                transition={{ duration: 0.5, delay: 0.6 }}
                className="p-4 mt-8 sm:p-8 space-y-10 pb-10 relative z-[4]"
            >
                <div className="space-y-2 z-[5] relative">
                    <h2>Trending this season</h2>
                    <DiscoverTrending />
                </div>
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
            </motion.div>
        </>
    )
}
