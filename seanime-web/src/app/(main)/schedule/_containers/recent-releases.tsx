"use client"
import { useAnilistListRecentAiringAnime } from "@/api/hooks/anilist.hooks"
import { EpisodeCard } from "@/app/(main)/_features/anime/_components/episode-card"
import { AppLayoutStack } from "@/components/ui/app-layout"
import { Carousel, CarouselContent, CarouselDotButtons, CarouselItem } from "@/components/ui/carousel"
import { addSeconds, formatDistanceToNow, subDays } from "date-fns"
import { useRouter } from "next/navigation"
import React from "react"

export function RecentReleases() {

    const router = useRouter()

    const { data } = useAnilistListRecentAiringAnime({
        page: 1,
        perPage: 50,
        airingAt_lesser: Math.floor(new Date().getTime() / 1000),
        airingAt_greater: Math.floor(subDays(new Date(), 14).getTime() / 1000),
    })

    const media = data?.Page?.airingSchedules?.filter(item => item?.media?.isAdult === false
        && item?.media?.type === "ANIME"
        && item?.media?.countryOfOrigin === "JP"
        && item?.media?.format !== "TV_SHORT",
    ).filter(Boolean)

    if (!media?.length) return null

    return (
        <AppLayoutStack className="pb-6">
            <h2>Recent episodes</h2>
            <Carousel
                className="w-full max-w-full"
                gap="md"
                opts={{
                    align: "start",
                }}
                autoScroll
            >
                <CarouselDotButtons />
                <CarouselContent>
                    {media.map(item => {
                        return (
                            <CarouselItem
                                key={item.id}
                                className="md:basis-1/2 lg:basis-1/3 2xl:basis-1/4 min-[2000px]:basis-1/5"
                            >
                                <EpisodeCard
                                    key={item.id}
                                    title={`Episode ${item.episode}`}
                                    image={item.media?.bannerImage || item.media?.coverImage?.large}
                                    topTitle={item.media?.title?.userPreferred}
                                    progressTotal={item.media?.episodes}
                                    meta={item.airingAt
                                        ? formatDistanceToNow(addSeconds(new Date(), item.timeUntilAiring), { addSuffix: true })
                                        : undefined}
                                    onClick={() => router.push(`/entry?id=${item.media?.id}`)}
                                    actionIcon={null}
                                />
                            </CarouselItem>
                        )
                    })}
                </CarouselContent>
            </Carousel>
        </AppLayoutStack>
    )
}
