import React from "react"
import { AnimeListItem } from "@/components/shared/anime-list-item"
import { Slider } from "@/components/shared/slider"
import { useDiscoverUpcomingAnime } from "@/app/(main)/discover/_containers/discover-sections/_lib/queries"
import { AnimeSliderSkeletonItem } from "@/app/(main)/discover/_components/anime-slider-skeleton-item"

export function DiscoverUpcoming() {

    const { data, isLoading } = useDiscoverUpcomingAnime()

    return (
        <Slider>
            {!isLoading ? data?.Page?.media?.filter(Boolean).map(media => {
                return (
                    <AnimeListItem
                        key={media.id}
                        media={media}
                        showLibraryBadge
                        containerClassName={"min-w-[250px] max-w-[250px] mt-8"}
                    />
                )
            }) : [...Array(10).keys()].map((v, idx) => <AnimeSliderSkeletonItem key={idx}/>)}
        </Slider>
    )

}