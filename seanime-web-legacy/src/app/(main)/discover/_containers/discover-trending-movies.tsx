import { MediaEntryCard } from "@/app/(main)/_features/media/_components/media-entry-card"
import { MediaEntryCardSkeleton } from "@/app/(main)/_features/media/_components/media-entry-card-skeleton"
import { useDiscoverTrendingMovies } from "@/app/(main)/discover/_lib/handle-discover-queries"
import { Carousel, CarouselContent, CarouselDotButtons } from "@/components/ui/carousel"
import React from "react"

export function DiscoverTrendingMovies() {

    const ref = React.useRef<HTMLDivElement>(null)
    const { data, isLoading } = useDiscoverTrendingMovies(ref)

    return (
        <Carousel
            className="w-full max-w-full"
            gap="xl"
            opts={{
                align: "start",
                dragFree: true,
            }}
            autoScroll
        >
            {/*<CarouselMasks />*/}
            <CarouselDotButtons />
            <CarouselContent className="px-6" ref={ref}>
                {!!data ? data?.Page?.media?.filter(Boolean).map(media => {
                    return (

                        <MediaEntryCard
                            key={media.id}
                            media={media}
                            showLibraryBadge
                            containerClassName="basis-[200px] md:basis-[250px] mx-2 mt-8 mb-0"
                            showTrailer
                            type="anime"
                        />
                    )
                }) : [...Array(10).keys()].map((v, idx) => <MediaEntryCardSkeleton key={idx} />)}
            </CarouselContent>
        </Carousel>
    )
}
