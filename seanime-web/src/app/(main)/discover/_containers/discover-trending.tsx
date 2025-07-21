import { AL_BaseAnime, AL_BaseManga } from "@/api/generated/types"
import { MediaEntryCard } from "@/app/(main)/_features/media/_components/media-entry-card"
import { MediaEntryCardSkeleton } from "@/app/(main)/_features/media/_components/media-entry-card-skeleton"
import { MediaGenreSelector } from "@/app/(main)/_features/media/_components/media-genre-selector"
import { __discover_clickedCarouselDotAtom, __discover_hoveringHeaderAtom } from "@/app/(main)/discover/_components/discover-page-header"
import { __discover_trendingGenresAtom, useDiscoverTrendingAnime } from "@/app/(main)/discover/_lib/handle-discover-queries"
import { ADVANCED_SEARCH_MEDIA_GENRES } from "@/app/(main)/search/_lib/advanced-search-constants"
import { Carousel, CarouselContent, CarouselDotButtons } from "@/components/ui/carousel"
import { atom } from "jotai"
import { useAtom, useAtomValue, useSetAtom } from "jotai/react"
import React, { useEffect, useState } from "react"

export const __discover_randomTrendingAtom = atom<AL_BaseAnime | AL_BaseManga | undefined>(undefined)
export const __discover_headerIsTransitioningAtom = atom(false)
export const __discover_animeRandomNumberAtom = atom<number>(0)
export const __discover_animeTotalItemsAtom = atom<number>(0)
export const __discover_setAnimeRandomNumberAtom = atom(
    null,
    (get, set, randomNumber: number) => {
        set(__discover_animeRandomNumberAtom, randomNumber)
    },
)

export function DiscoverTrending() {

    const { data, isLoading } = useDiscoverTrendingAnime()
    const setRandomTrendingAtom = useSetAtom(__discover_randomTrendingAtom)
    const isHoveringHeader = useAtomValue(__discover_hoveringHeaderAtom)
    const clickedHeaderDot = useAtomValue(__discover_clickedCarouselDotAtom) // clears interval
    const setHeaderIsTransitioning = useSetAtom(__discover_headerIsTransitioningAtom)
    const setAnimeTotalItems = useSetAtom(__discover_animeTotalItemsAtom)
    const [animeRandomNumber, setAnimeRandomNumber] = useAtom(__discover_animeRandomNumberAtom)

    // Random number between 0 and 12
    const [randomNumber, setRandomNumber] = useState(0)

    // Update the atom when randomNumber changes
    useEffect(() => {
        setAnimeRandomNumber(randomNumber)
    }, [randomNumber])

    useEffect(() => {
        const t = setInterval(() => {
            setHeaderIsTransitioning(true)
            setTimeout(() => {
                setRandomNumber(p => {
                    return p < 11 ? p + 1 : 0
                })
                setHeaderIsTransitioning(false)
            }, 900)
        }, 6000)
        if (isHoveringHeader) {
            clearInterval(t)
        }
        return () => clearInterval(t)
    }, [isHoveringHeader, clickedHeaderDot])

    // Update randomNumber when animeRandomNumber changes from outside
    useEffect(() => {
        if (animeRandomNumber !== randomNumber) {
            setHeaderIsTransitioning(true)
            setTimeout(() => {
                setRandomNumber(animeRandomNumber)
                setHeaderIsTransitioning(false)
            }, 900)
        }
    }, [animeRandomNumber])

    const firedRef = React.useRef(false)
    React.useEffect(() => {
        if (!firedRef.current && data) {
            const mediaItems = data?.Page?.media?.filter(Boolean) || []
            setAnimeTotalItems(mediaItems.length)
            const random = mediaItems[randomNumber]
            if (random) {
                setRandomTrendingAtom(random)
                firedRef.current = true
            }
        }
    }, [data, randomNumber])

    React.useEffect(() => {
        if (firedRef.current) {
            const mediaItems = data?.Page?.media?.filter(Boolean) || []
            const random = mediaItems[randomNumber]
            if (random) {
                setRandomTrendingAtom(random)
            }
        }
    }, [randomNumber, data])

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
            <GenreSelector />
            {/*<CarouselMasks />*/}
            <CarouselDotButtons />
            <CarouselContent className="px-6">
                {!isLoading ? data?.Page?.media?.filter(Boolean).map(media => {
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
    )

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
        <MediaGenreSelector
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
    )
}
