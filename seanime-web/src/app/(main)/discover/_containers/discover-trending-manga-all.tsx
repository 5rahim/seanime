import { useAnilistListManga } from "@/api/hooks/manga.hooks"
import { MediaEntryCard } from "@/app/(main)/_features/media/_components/media-entry-card"
import { MediaEntryCardSkeleton } from "@/app/(main)/_features/media/_components/media-entry-card-skeleton"
import { MediaGenreSelector } from "@/app/(main)/_features/media/_components/media-genre-selector"
import { __discover_hoveringHeaderAtom } from "@/app/(main)/discover/_components/discover-page-header"
import { __discover_headerIsTransitioningAtom, __discover_randomTrendingAtom } from "@/app/(main)/discover/_containers/discover-trending"
import { ADVANCED_SEARCH_MEDIA_GENRES } from "@/app/(main)/search/_lib/advanced-search-constants"
import { Carousel, CarouselContent, CarouselDotButtons } from "@/components/ui/carousel"
import { TextInput } from "@/components/ui/text-input"
import { useDebounce } from "@/hooks/use-debounce"
import { atom } from "jotai/index"
import { useAtom, useAtomValue, useSetAtom } from "jotai/react"
import React from "react"
import { FiSearch } from "react-icons/fi"

const trendingGenresAtom = atom<string[]>([])

export function DiscoverTrendingMangaAll() {
    const genres = useAtomValue(trendingGenresAtom)
    const { data, isLoading } = useAnilistListManga({
        page: 1,
        perPage: 20,
        sort: ["TRENDING_DESC"],
        genres: genres.length > 0 ? genres : undefined,
    })

    const setRandomTrendingAtom = useSetAtom(__discover_randomTrendingAtom)
    const isHoveringHeader = useAtomValue(__discover_hoveringHeaderAtom)
    const setHeaderIsTransitioning = useSetAtom(__discover_headerIsTransitioningAtom)


    // useEffect(() => {
    //     const t = setInterval(() => {
    //         setHeaderIsTransitioning(true)
    //         setTimeout(() => {
    //             setRandomNumber(p => {
    //                 return p < 10 ? p + 1 : 0
    //             })
    //             setHeaderIsTransitioning(false)
    //         }, 900)
    //     }, 6000)
    //     if (isHoveringHeader) {
    //         clearInterval(t)
    //     }
    //     return () => clearInterval(t)
    // }, [isHoveringHeader])
    //
    // // Update randomNumber when mangaRandomNumber changes from outside
    // useEffect(() => {
    //     if (mangaRandomNumber !== randomNumber) {
    //         setHeaderIsTransitioning(true)
    //         setTimeout(() => {
    //             setRandomNumber(mangaRandomNumber)
    //             setHeaderIsTransitioning(false)
    //         }, 900)
    //     }
    // }, [mangaRandomNumber])
    //
    // const firedRef = React.useRef(false)
    // React.useEffect(() => {
    //     console.log("firedRef.current", firedRef.current, data?.Page?.media?.length)
    //     if (!firedRef.current && data) {
    //         const mediaItems = data?.Page?.media?.filter(Boolean) || []
    //         setMangaTotalItems(mediaItems.length)
    //         const random = mediaItems[randomNumber]
    //         if (random) {
    //             setRandomTrendingAtom(random)
    //             firedRef.current = true
    //         }
    //     }
    // }, [data, randomNumber])
    //
    // React.useEffect(() => {
    //     if (firedRef.current) {
    //         const mediaItems = data?.Page?.media?.filter(Boolean) || []
    //         const random = mediaItems[randomNumber]
    //         if (random) {
    //             setRandomTrendingAtom(random)
    //         }
    //     }
    // }, [randomNumber, data])

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

const mangaSearchInputAtom = atom<string>("")

export function DiscoverMangaSearchBar() {
    const [searchInput, setSearchInput] = useAtom(mangaSearchInputAtom)
    const search = useDebounce(searchInput, 500)

    const { data, isLoading, isFetching } = useAnilistListManga({
        page: 1,
        perPage: 10,
        search: search,
    })

    return (
        <div className="space-y-4" data-discover-manga-search-bar-container>
            <div className="container" data-discover-manga-search-bar-inner-container>
                <TextInput
                    leftIcon={<FiSearch />}
                    value={searchInput}
                    onValueChange={v => {
                        setSearchInput(v)
                    }}
                    className="rounded-full"
                    placeholder="Search manga"
                />
            </div>

            {!!search && <Carousel
                className="w-full max-w-full"
                gap="md"
                opts={{
                    align: "start",
                    dragFree: true,
                }}
                autoScroll
            >
                <CarouselContent className="px-6">
                    {!(isLoading || isFetching) ? data?.Page?.media?.filter(Boolean).map(media => {
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
            </Carousel>}
        </div>
    )
}
