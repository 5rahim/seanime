"use client"
import { useAnilistListRecentAiringAnime } from "@/api/hooks/anilist.hooks"
import { MediaEntryCard } from "@/app/(main)/_features/media/_components/media-entry-card"
import { MediaEntryCardSkeleton } from "@/app/(main)/_features/media/_components/media-entry-card-skeleton"
import { AppLayoutStack } from "@/components/ui/app-layout"
import { Carousel, CarouselContent, CarouselDotButtons } from "@/components/ui/carousel"
import { addSeconds, formatDistanceToNow, subDays } from "date-fns"
import React from "react"

export function RecentReleases() {

    const { data, isLoading } = useAnilistListRecentAiringAnime({
        page: 1,
        perPage: 50,
        airingAt_lesser: Math.floor(new Date().getTime() / 1000),
        airingAt_greater: Math.floor(subDays(new Date(), 14).getTime() / 1000),
    })

    const aired = data?.Page?.airingSchedules?.filter(item => item?.media?.isAdult === false
        && item?.media?.type === "ANIME"
        && item?.media?.countryOfOrigin === "JP"
        && item?.media?.format !== "TV_SHORT",
    ).filter(Boolean)

    if (!aired?.length && !isLoading) return null

    return (
        <AppLayoutStack className="pb-6">
            <h2>Aired Recently</h2>
            <Carousel
                className="w-full max-w-full"
                gap="md"
                opts={{
                    align: "start",
                }}
                carouselButtonContainerClass="top-[-3.5rem]"
                autoScroll
            >
                <CarouselDotButtons />
                <CarouselContent className="px-6">
                    {!isLoading ? aired?.map(item => {
                        return (
                            <MediaEntryCard
                                key={item.id}
                                media={item?.media!}
                                showLibraryBadge
                                containerClassName="basis-[200px] md:basis-[250px] mx-2 mt-8 mb-0"
                                hideReleasingBadge
                                showTrailer
                                type="anime"
                                overlay={<div className="flex flex-col w-fit absolute right-0 items-end">
                                    <div
                                        className="font-semibold text-white bg-gray-950 z-[1] pl-3 pr-[0.2rem] w-full py-1.5 text-center !tracking-wider !bg-opacity-80 rounded-none rounded-bl-lg"
                                    >{item?.media?.format === "MOVIE" ? "Movie" :
                                        <span className="tracking-wider"><span className="!text-lg">{item.episode}</span><span className="text-[--muted] tracking-wider !text-md">/{item.media?.episodes ?? "-"}</span></span>}</div>
                                    <div className="text-xs font-semibold z-[-1] w-fit h-fit pl-2 pr-[0.3rem] py-1 ml-2 text-center bg-gray-700 !bg-opacity-70 rounded-none rounded-bl-lg">
                                        {item.airingAt
                                            ? formatDistanceToNow(addSeconds(new Date(), item.timeUntilAiring), { addSuffix: true })
                                                ?.replace("less than a", "1")
                                                ?.replace("about ", "")
                                                ?.replace(" minutes", "m")
                                                ?.replace(" minute", "m")
                                                ?.replace(" hours", "h")
                                                ?.replace(" hour", "h")
                                                ?.replace(" days", "d")

                                            : undefined}
                                    </div>
                                </div>}
                            />
                        )
                    }) : [...Array(10).keys()].map((v, idx) => <MediaEntryCardSkeleton key={idx} />)}
                </CarouselContent>
            </Carousel>
        </AppLayoutStack>
    )
}
