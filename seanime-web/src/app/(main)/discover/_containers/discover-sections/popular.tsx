import { AnimeSliderSkeletonItem } from "@/app/(main)/discover/_components/anime-slider-skeleton-item"
import { useDiscoverPopularAnime } from "@/app/(main)/discover/_containers/discover-sections/_lib/queries"
import { AnimeListItem } from "@/components/shared/anime-list-item"
import { HorizontalDraggableScroll } from "@/components/ui/horizontal-draggable-scroll"
import React from "react"

export function DiscoverPopular() {

    const { data, isLoading } = useDiscoverPopularAnime()

    return (
        <HorizontalDraggableScroll>
            {!isLoading ? data?.Page?.media?.filter(Boolean).map(media => {
                return (
                    <AnimeListItem
                        key={media.id}
                        media={media}
                        showLibraryBadge
                        containerClassName="min-w-[250px] max-w-[250px] mt-8"
                    />
                )
            }) : [...Array(10).keys()].map((v, idx) => <AnimeSliderSkeletonItem key={idx} />)}
        </HorizontalDraggableScroll>
    )
}
