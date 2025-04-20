import { useAnilistListManga } from "@/api/hooks/manga.hooks"
import { MediaEntryCard } from "@/app/(main)/_features/media/_components/media-entry-card"
import { MediaEntryCardSkeleton } from "@/app/(main)/_features/media/_components/media-entry-card-skeleton"
import { MediaGenreSelector } from "@/app/(main)/_features/media/_components/media-genre-selector"
import { __discover_hoveringHeaderAtom } from "@/app/(main)/discover/_components/discover-page-header"
import { __discover_headerIsTransitioningAtom, __discover_randomTrendingAtom } from "@/app/(main)/discover/_containers/discover-trending"
import { ADVANCED_SEARCH_MEDIA_GENRES } from "@/app/(main)/search/_lib/advanced-search-constants"
import { Carousel, CarouselContent, CarouselDotButtons } from "@/components/ui/carousel"
import { atom } from "jotai/index"
import { useAtom, useAtomValue, useSetAtom } from "jotai/react"
import React, { useEffect, useState } from "react"

const trendingGenresAtom = atom<string[]>([])

export const __discover_mangaRandomNumberAtom = atom<number>(0)
export const __discover_mangaTotalItemsAtom = atom<number>(0)
export const __discover_setMangaRandomNumberAtom = atom(
    null,
    (get, set, randomNumber: number) => {
        set(__discover_mangaRandomNumberAtom, randomNumber)
    },
)

export function DiscoverTrendingCountry({ country }: { country: string }) {
    const genres = useAtomValue(trendingGenresAtom)
    const { data, isLoading } = useAnilistListManga({
        page: 1,
        perPage: 20,
        sort: ["TRENDING_DESC"],
        countryOfOrigin: country || undefined,
        genres: genres.length > 0 ? genres : undefined,
    })

    const setRandomTrendingAtom = useSetAtom(__discover_randomTrendingAtom)
    const isHoveringHeader = useAtomValue(__discover_hoveringHeaderAtom)
    const setHeaderIsTransitioning = useSetAtom(__discover_headerIsTransitioningAtom)
    const setMangaTotalItems = useSetAtom(__discover_mangaTotalItemsAtom)
    const [mangaRandomNumber, setMangaRandomNumber] = useAtom(__discover_mangaRandomNumberAtom)

    const [randomNumber, setRandomNumber] = useState(0)

    // Update the atom when randomNumber changes
    useEffect(() => {
        setMangaRandomNumber(randomNumber)
    }, [randomNumber])

    // Update randomNumber when mangaRandomNumber changes from outside
    useEffect(() => {
        if (mangaRandomNumber !== randomNumber) {
            setHeaderIsTransitioning(true)
            setTimeout(() => {
                setRandomNumber(mangaRandomNumber)
                setHeaderIsTransitioning(false)
            }, 900)
        }
    }, [mangaRandomNumber])

    useEffect(() => {
        if (country !== "JP") return
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
    }, [isHoveringHeader, country])

    const firedRef = React.useRef(false)
    React.useEffect(() => {
        if (country !== "JP") return
        if (!firedRef.current && data) {
            const mediaItems = data?.Page?.media?.filter(Boolean) || []
            const random = mediaItems[randomNumber]
            if (random) {
                setMangaTotalItems(mediaItems.length)
                setRandomTrendingAtom(random)
                firedRef.current = true
            }
        }
    }, [data, randomNumber, country])

    React.useEffect(() => {
        if (country !== "JP") return
        if (firedRef.current) {
            const random = data?.Page?.media?.filter(Boolean)[randomNumber]
            if (random) {
                setRandomTrendingAtom(random)
            }
        }
    }, [randomNumber, country])

    return (
        <Carousel
            className="w-full max-w-full"
            gap="md"
            opts={{
                align: "start",
                dragFree: true,
            }}
            autoScroll
        >
            <GenreSelector />
            <CarouselDotButtons />
            <CarouselContent className="px-6">
                {!isLoading ? data?.Page?.media?.filter(Boolean).map(media => {
                    return (
                        <MediaEntryCard
                            key={media.id}
                            media={media}
                            containerClassName="basis-[200px] md:basis-[250px] mx-2 my-8"
                            type="manga"
                        />
                    )
                }) : [...Array(10).keys()].map((v, idx) => <MediaEntryCardSkeleton key={idx} />)}
            </CarouselContent>
        </Carousel>
    )
}

function GenreSelector() {

    const [selectedGenre, setSelectedGenre] = useAtom(trendingGenresAtom)

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
