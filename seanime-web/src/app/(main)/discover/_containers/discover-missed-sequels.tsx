import { useAnilistListMissedSequels } from "@/api/hooks/anilist.hooks"
import { MediaEntryCard } from "@/app/(main)/_features/media/_components/media-entry-card"
import { MediaEntryCardSkeleton } from "@/app/(main)/_features/media/_components/media-entry-card-skeleton"
import { Carousel, CarouselContent, CarouselDotButtons } from "@/components/ui/carousel"
import { useInView } from "framer-motion"
import React from "react"


export function DiscoverMissedSequelsSection() {
    const ref = React.useRef(null)
    const isInView = useInView(ref, { once: true })
    const { data, isLoading } = useAnilistListMissedSequels(isInView)

    if (!isInView && !data) return <div ref={ref} />

    if (!data?.length) return null

    return (
        <div className="space-y-2 z-[5] relative" data-discover-missed-sequels-container>
            <h2>You might have missed</h2>
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
                    {!isLoading ? data?.filter(Boolean).map(media => {
                        return (
                            <MediaEntryCard
                                key={media.id}
                                media={media}
                                showLibraryBadge
                                containerClassName="basis-[200px] md:basis-[250px] mx-2 my-8"
                                showTrailer
                                type="anime"
                            />
                        )
                    }) : [...Array(10).keys()].map((v, idx) => <MediaEntryCardSkeleton key={idx} />)}
                </CarouselContent>
            </Carousel>
        </div>
    )

}
