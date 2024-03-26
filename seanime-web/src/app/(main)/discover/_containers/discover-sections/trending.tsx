import { AnimeSliderSkeletonItem } from "@/app/(main)/discover/_components/anime-slider-skeleton-item"
import { ADVANCED_SEARCH_MEDIA_GENRES } from "@/app/(main)/discover/_containers/advanced-search/_lib/constants"
import { __discover_trendingGenresAtom, useDiscoverTrendingAnime } from "@/app/(main)/discover/_containers/discover-sections/_lib/queries"
import { __discover_hoveringHeaderAtom } from "@/app/(main)/discover/_containers/discover-sections/header"
import { AnimeListItem } from "@/components/shared/anime-list-item"
import { Carousel, CarouselContent, CarouselDotButtons } from "@/components/ui/carousel"
import { HorizontalDraggableScroll } from "@/components/ui/horizontal-draggable-scroll"
import { StaticTabs } from "@/components/ui/tabs"
import { BaseMediaFragment } from "@/lib/anilist/gql/graphql"
import { atom } from "jotai"
import { useAtom, useAtomValue, useSetAtom } from "jotai/react"
import React, { useEffect, useState } from "react"

export const __discover_randomTrendingAtom = atom<BaseMediaFragment | undefined>(undefined)
export const __discover_headerIsTransitioningAtom = atom(false)

export function DiscoverTrending() {

    const { data, isLoading } = useDiscoverTrendingAnime()
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
        setRandomTrendingAtom(data?.Page?.media?.filter(Boolean)[randomNumber])
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
            <GenreSelector />
            {/*<CarouselMasks />*/}
            <CarouselDotButtons />
            <CarouselContent className="px-6">
                {!isLoading ? data?.Page?.media?.filter(Boolean).map(media => {
                    return (

                        <AnimeListItem
                            key={media.id}
                            media={media}
                            showLibraryBadge
                            containerClassName="basis-[200px] md:basis-[250px] mx-2 my-8"
                            showTrailer
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

type GenreSelectorProps = {
    children?: React.ReactNode
}

function GenreSelector(props: GenreSelectorProps) {

    const {
        children,
        ...rest
    } = props

    const [selectedGenre, setSelectedGenre] = useAtom(__discover_trendingGenresAtom)

    return (
        <HorizontalDraggableScroll className="w-full scroll-pb-1 pt-4">
            <StaticTabs
                className="px-2 overflow-visible py-4"
                triggerClass="text-base rounded-md ring-2 ring-transparent data-[current=true]:ring-brand-500 data-[current=true]:text-brand-300"
                items={[
                    {
                        name: "All",
                        isCurrent: selectedGenre.length === 0,
                        onClick: () => setSelectedGenre([]),
                    },
                    ...ADVANCED_SEARCH_MEDIA_GENRES.map(genre => ({
                        name: genre,
                        isCurrent: selectedGenre.includes(genre),
                        onClick: () => setSelectedGenre([genre]),
                    })),
                ]}
            />
        </HorizontalDraggableScroll>
    )
}

