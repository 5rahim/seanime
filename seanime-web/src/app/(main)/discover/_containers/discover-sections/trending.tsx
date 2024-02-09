import { AnimeSliderSkeletonItem } from "@/app/(main)/discover/_components/anime-slider-skeleton-item"
import { useDiscoverTrendingAnime } from "@/app/(main)/discover/_containers/discover-sections/_lib/queries"
import { AnimeListItem } from "@/components/shared/anime-list-item"
import { Slider } from "@/components/shared/slider"
import { BaseMediaFragment } from "@/lib/anilist/gql/graphql"
import { atom } from "jotai"
import { useSetAtom } from "jotai/react"
import React, { useEffect, useMemo } from "react"

export const __discover_randomTrendingAtom = atom<BaseMediaFragment | undefined>(undefined)

export function DiscoverTrending() {

    const { data, isLoading, fetchNextPage } = useDiscoverTrendingAnime()
    const setRandomTrendingAtom = useSetAtom(__discover_randomTrendingAtom)

    const randomNumber = useMemo(() => Math.floor(Math.random() * 6), [])

    useEffect(() => {
        setRandomTrendingAtom(data?.pages?.filter(Boolean).flatMap(n => n.Page?.media).filter(Boolean)[0])
    }, [data])

    return (
        <Slider
            onSlideEnd={() => fetchNextPage()}
        >
            {!isLoading ? data?.pages?.filter(Boolean).flatMap(n => n.Page?.media).filter(Boolean).map(media => {
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
