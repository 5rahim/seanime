"use client"
import { useAnilistListRecentAiringAnime } from "@/api/hooks/anilist.hooks"
import { MediaEntryCard } from "@/app/(main)/_features/media/_components/media-entry-card"
import { MediaEntryCardSkeleton } from "@/app/(main)/_features/media/_components/media-entry-card-skeleton"
import { useMediaPreviewModal } from "@/app/(main)/_features/media/_containers/media-preview-modal"
import { AppLayoutStack } from "@/components/ui/app-layout"
import { Carousel, CarouselContent, CarouselDotButtons } from "@/components/ui/carousel"
import { addSeconds, formatDistanceToNow, subDays } from "date-fns"
import { useRouter } from "next/navigation"
import React from "react"

export function RecentReleases() {

    const router = useRouter()

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

    const { setPreviewModalMediaId } = useMediaPreviewModal()

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
                                showLibraryBadge={false}
                                containerClassName="basis-[200px] md:basis-[250px] mx-2 mt-8 mb-0"
                                showTrailer
                                type="anime"
                                // overlay={<div className="flex flex-col w-fit bg-gray-950 pt-1 pb-2 px-2 !bg-opacity-80 rounded-br-lg">
                                //     <div
                                //         className="font-semibold text-white z-[1] w-fit text-center !bg-opacity-90 text-sm lg:text-base
                                // rounded-none rounded-br-lg" >{item?.media?.format === "MOVIE" ? "Complete Movie" : <span>Episode
                                // {item.episode}{item.media?.episodes && <span
                                // className="text-[--muted]">/{item.media?.episodes}</span>}</span>}</div> <div className="text-xs font-semibold
                                // opacity-70 z-[1] w-fit h-fit pt-1 text-center rounded-none rounded-br-lg"> {item.airingAt ?
                                // formatDistanceToNow(addSeconds(new Date(), item.timeUntilAiring), { addSuffix: true }) ?.replace("about ", "") :
                                // undefined} </div> </div>}
                                overlay={<div className="flex flex-col w-fit">
                                    <div
                                        className="font-semibold text-white bg-gray-950 z-[1] w-fit px-3 py-1.5 text-center !bg-opacity-90 text-sm lg:text-base rounded-none rounded-br-lg"
                                    >{item?.media?.format === "MOVIE" ? "Complete Movie" :
                                        <span>Episode {item.episode}{item.media?.episodes &&
                                            <span className="text-[--muted]">/{item.media?.episodes}</span>}</span>}</div>
                                    <div className="text-xs font-semibold z-[-1] w-fit h-fit px-2 py-1 text-center bg-gray-700 !bg-opacity-70 rounded-none rounded-br-lg">
                                        {item.airingAt
                                            ? formatDistanceToNow(addSeconds(new Date(), item.timeUntilAiring), { addSuffix: true })
                                                ?.replace("about ", "")
                                            : undefined}
                                    </div>
                                </div>}
                            />
                        )
                    }) : [...Array(10).keys()].map((v, idx) => <MediaEntryCardSkeleton key={idx} />)}
                    {/*{isLoading && ([1, 2, 3, 4, 5, 6, 7, 8])?.map((_, idx) => {*/}
                    {/*    return <CarouselItem*/}
                    {/*        key={idx}*/}
                    {/*        className="md:basis-1/2 lg:basis-1/3 2xl:basis-1/4 min-[2000px]:basis-1/5 relative h-[220px] px-2"*/}
                    {/*    ><Skeleton*/}
                    {/*        key={idx} className={cn(*/}
                    {/*        "w-full h-full absolute",*/}
                    {/*    )}*/}
                    {/*    /></CarouselItem>*/}
                    {/*})}*/}
                    {/*{media?.map(item => {*/}
                    {/*    return (*/}
                    {/*        <CarouselItem*/}
                    {/*            key={item.id}*/}
                    {/*            className="md:basis-1/2 lg:basis-1/3 2xl:basis-1/4 min-[2000px]:basis-1/5"*/}
                    {/*        >*/}
                    {/*            <SeaContextMenu*/}
                    {/*                content={<ContextMenuGroup>*/}
                    {/*                    <ContextMenuLabel className="text-[--muted] line-clamp-2 py-0 my-2">*/}
                    {/*                        {item.media?.title?.userPreferred}*/}
                    {/*                    </ContextMenuLabel>*/}
                    {/*                    <ContextMenuItem*/}
                    {/*                        onClick={() => {*/}
                    {/*                            setPreviewModalMediaId(item.media?.id || 0, "anime")*/}
                    {/*                        }}*/}
                    {/*                    >*/}
                    {/*                        <LuEye /> Preview*/}
                    {/*                    </ContextMenuItem>*/}
                    {/*                </ContextMenuGroup>}*/}
                    {/*            >*/}
                    {/*                <ContextMenuTrigger>*/}
                    {/*                    <EpisodeCard*/}
                    {/*                        key={item.id}*/}
                    {/*                        title={`Episode ${item.episode}`}*/}
                    {/*                        image={item.media?.bannerImage || item.media?.coverImage?.large}*/}
                    {/*                        topTitle={item.media?.title?.userPreferred}*/}
                    {/*                        progressTotal={item.media?.episodes}*/}
                    {/*                        meta={item.airingAt*/}
                    {/*                            ? formatDistanceToNow(addSeconds(new Date(), item.timeUntilAiring), { addSuffix: true })*/}
                    {/*                            : undefined}*/}
                    {/*                        onClick={() => router.push(`/entry?id=${item.media?.id}`)}*/}
                    {/*                        actionIcon={null}*/}
                    {/*                        anime={{*/}
                    {/*                            id: item.media?.id,*/}
                    {/*                            image: item.media?.coverImage?.medium,*/}
                    {/*                            title: item.media?.title?.userPreferred,*/}
                    {/*                        }}*/}
                    {/*                    />*/}
                    {/*                </ContextMenuTrigger>*/}
                    {/*            </SeaContextMenu>*/}

                    {/*        </CarouselItem>*/}
                    {/*    )*/}
                    {/*})}*/}
                </CarouselContent>
            </Carousel>
        </AppLayoutStack>
    )
}
