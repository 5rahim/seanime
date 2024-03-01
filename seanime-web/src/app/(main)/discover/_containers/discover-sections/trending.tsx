import { AnimeSliderSkeletonItem } from "@/app/(main)/discover/_components/anime-slider-skeleton-item"
import { useDiscoverTrendingAnime } from "@/app/(main)/discover/_containers/discover-sections/_lib/queries"
import { __discover_hoveringHeaderAtom } from "@/app/(main)/discover/_containers/discover-sections/header"
import { AnimeListItem } from "@/components/shared/anime-list-item"
import { Carousel, CarouselContent, CarouselDotButtons, CarouselMasks } from "@/components/ui/carousel"
import { BaseMediaFragment } from "@/lib/anilist/gql/graphql"
import { atom } from "jotai"
import { useAtomValue, useSetAtom } from "jotai/react"
import React, { useEffect, useState } from "react"

export const __discover_randomTrendingAtom = atom<BaseMediaFragment | undefined>(undefined)
export const __discover_headerIsTransitioningAtom = atom(false)

export function DiscoverTrending() {

    const { data, isLoading, fetchNextPage } = useDiscoverTrendingAnime()
    const setRandomTrendingAtom = useSetAtom(__discover_randomTrendingAtom)
    const isHoveringHeader = useAtomValue(__discover_hoveringHeaderAtom)
    const setHeaderIsTransitioning = useSetAtom(__discover_headerIsTransitioningAtom)

    const [randomNumber, setRandomNumber] = useState(Math.floor(Math.random() * 8))

    useEffect(() => {
        const t = setInterval(() => {
            setHeaderIsTransitioning(true)
            setTimeout(() => {
                setRandomNumber(p => {
                    return p < 10 ? p + 1 : 0
                })
                setHeaderIsTransitioning(false)
            }, 900)
        }, 6000)
        if (isHoveringHeader) {
            clearInterval(t)
        }
        return () => clearInterval(t)
    }, [isHoveringHeader])

    useEffect(() => {
        setRandomTrendingAtom(data?.pages?.filter(Boolean).flatMap(n => n.Page?.media).filter(Boolean)[randomNumber])
    }, [data, randomNumber])

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
                {!isLoading ? data?.pages?.filter(Boolean).flatMap(n => n.Page?.media).filter(Boolean).map(media => {
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

    // return (
    //     <HorizontalDraggableScroll
    //         onSlideEnd={() => fetchNextPage()}
    //     >
    //         {!isLoading ? data?.pages?.filter(Boolean).flatMap(n => n.Page?.media).filter(Boolean).map(media => {
    //             return (
    //                 <AnimeListItem
    //                     key={media.id}
    //                     media={media}
    //                     showLibraryBadge
    //                     containerClassName="min-w-[250px] max-w-[250px] mt-8"
    //                 />
    //             )
    //         }) : [...Array(10).keys()].map((v, idx) => <AnimeSliderSkeletonItem key={idx} />)}
    //     </HorizontalDraggableScroll>
    // )

}
