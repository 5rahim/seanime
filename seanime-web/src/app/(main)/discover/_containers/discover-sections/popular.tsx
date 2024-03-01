import { AnimeSliderSkeletonItem } from "@/app/(main)/discover/_components/anime-slider-skeleton-item"
import { useDiscoverPopularAnime } from "@/app/(main)/discover/_containers/discover-sections/_lib/queries"
import { AnimeListItem } from "@/components/shared/anime-list-item"
import { Carousel, CarouselContent, CarouselDotButtons, CarouselMasks } from "@/components/ui/carousel"
import React from "react"

export function DiscoverPopular() {

    const { data, isLoading } = useDiscoverPopularAnime()

    return (
        <Carousel
            className="w-full max-w-full"
            gap="xl"
            opts={{
                align: "start",
            }}
            autoScroll
        >
            <CarouselMasks />
            <CarouselDotButtons />
            <CarouselContent className="px-6">
                {!isLoading ? data?.Page?.media?.filter(Boolean).map(media => {
                    return (

                        <AnimeListItem
                            key={media.id}
                            media={media}
                            showLibraryBadge
                            containerClassName="basis-[250px] mx-2 my-8"
                        />
                    )
                }) : [...Array(10).keys()].map((v, idx) => <AnimeSliderSkeletonItem key={idx} />)}
            </CarouselContent>
        </Carousel>
    )
}
