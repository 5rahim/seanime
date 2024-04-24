import { MediaEntryCard } from "@/app/(main)/_components/features/media/media-entry-card"
import { MediaEntryCardSkeleton } from "@/app/(main)/_components/features/media/media-entry-card-skeleton"
import { useDiscoverUpcomingAnime } from "@/app/(main)/discover/_containers/discover-sections/_lib/queries"
import { Carousel, CarouselContent, CarouselDotButtons } from "@/components/ui/carousel"
import React from "react"

export function DiscoverUpcoming() {

    const ref = React.useRef<HTMLDivElement>(null)
    const { data, isLoading } = useDiscoverUpcomingAnime(ref)

    return (
        <Carousel
            className="w-full max-w-full"
            gap="xl"
            opts={{
                align: "start",
            }}
            autoScroll
        >
            {/*<CarouselMasks />*/}
            <CarouselDotButtons />
            <CarouselContent className="px-6" ref={ref}>
                {!isLoading ? data?.Page?.media?.filter(Boolean).map(media => {
                    return (

                        <MediaEntryCard
                            key={media.id}
                            media={media}
                            showLibraryBadge
                            containerClassName="basis-[200px] md:basis-[250px] mx-2 my-8"
                            showTrailer
                        />
                    )
                }) : [...Array(10).keys()].map((v, idx) => <MediaEntryCardSkeleton key={idx} />)}
            </CarouselContent>
        </Carousel>
    )

}
