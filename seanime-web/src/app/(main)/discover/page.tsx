"use client"
import React from "react"
import { DiscoverTrending } from "@/app/(main)/discover/_containers/discover-sections/trending"
import { DiscoverUpcoming } from "@/app/(main)/discover/_containers/discover-sections/upcoming"
import { DiscoverPopular } from "@/app/(main)/discover/_containers/discover-sections/popular"
import { DiscoverTrendingMovies } from "@/app/(main)/discover/_containers/discover-sections/trending-movies"
import { DiscoverPageHeader } from "@/app/(main)/discover/_containers/discover-sections/header"


export default function Page() {

    return (
        <>
            <DiscoverPageHeader/>
            <div className={"px-4 pt-8 space-y-10 pb-10"}>
                <div className={"space-y-2"}>
                    <h2>Popular this season</h2>
                    <DiscoverTrending/>
                </div>
                <div className={"space-y-2"}>
                    <h2>Popular shows</h2>
                    <DiscoverPopular/>
                </div>
                <div className={"space-y-2"}>
                    <h2>Upcoming</h2>
                    <DiscoverUpcoming/>
                </div>
                <div className={"space-y-2"}>
                    <h2>Trending movies</h2>
                    <DiscoverTrendingMovies/>
                </div>
            </div>
        </>
    )
}
