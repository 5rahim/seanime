"use client"
import { useAnilistListRecentAiringAnime } from "@/api/hooks/anilist.hooks"
import { EpisodeCard } from "@/app/(main)/_features/anime/_components/episode-card"
import { SeaContextMenu } from "@/app/(main)/_features/context-menu/sea-context-menu"
import { useMediaPreviewModal } from "@/app/(main)/_features/media/_containers/media-preview-modal"
import { AppLayoutStack } from "@/components/ui/app-layout"
import { Carousel, CarouselContent, CarouselDotButtons, CarouselItem } from "@/components/ui/carousel"
import { ContextMenuGroup, ContextMenuItem, ContextMenuLabel, ContextMenuTrigger } from "@/components/ui/context-menu"
import { cn } from "@/components/ui/core/styling"
import { Skeleton } from "@/components/ui/skeleton"
import { addSeconds, formatDistanceToNow, subDays } from "date-fns"
import { useRouter } from "next/navigation"
import React from "react"
import { LuEye } from "react-icons/lu"

export function RecentReleases() {

    const router = useRouter()

    const { data, isLoading } = useAnilistListRecentAiringAnime({
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

    const { setPreviewModalMediaId } = useMediaPreviewModal()

    if (!media?.length && !isLoading) return null

    return (
        <AppLayoutStack className="pb-6">
            <h2>Aired Recently</h2>
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
                    {isLoading && ([1, 2, 3, 4, 5, 6, 7, 8])?.map((_, idx) => {
                        return <CarouselItem
                            key={idx}
                            className="md:basis-1/2 lg:basis-1/3 2xl:basis-1/4 min-[2000px]:basis-1/5 relative h-[220px] px-2"
                        ><Skeleton
                            key={idx} className={cn(
                            "w-full h-full absolute",
                        )}
                        /></CarouselItem>
                    })}
                    {media?.map(item => {
                        return (
                            <CarouselItem
                                key={item.id}
                                className="md:basis-1/2 lg:basis-1/3 2xl:basis-1/4 min-[2000px]:basis-1/5"
                            >
                                <SeaContextMenu
                                    content={<ContextMenuGroup>
                                        <ContextMenuLabel className="text-[--muted] line-clamp-2 py-0 my-2">
                                            {item.media?.title?.userPreferred}
                                        </ContextMenuLabel>
                                        <ContextMenuItem
                                            onClick={() => {
                                                setPreviewModalMediaId(item.media?.id || 0, "anime")
                                            }}
                                        >
                                            <LuEye /> Preview
                                        </ContextMenuItem>
                                    </ContextMenuGroup>}
                                >
                                    <ContextMenuTrigger>
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
                                            anime={{
                                                id: item.media?.id,
                                                image: item.media?.coverImage?.medium,
                                                title: item.media?.title?.userPreferred,
                                            }}
                                        />
                                    </ContextMenuTrigger>
                                </SeaContextMenu>

                            </CarouselItem>
                        )
                    })}
                </CarouselContent>
            </Carousel>
        </AppLayoutStack>
    )
}
