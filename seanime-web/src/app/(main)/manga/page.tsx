"use client"
import { CustomLibraryBanner } from "@/app/(main)/(library)/_containers/custom-library-banner"
import { AnimeSliderSkeletonItem } from "@/app/(main)/discover/_components/anime-slider-skeleton-item"
import { ADVANCED_SEARCH_MEDIA_GENRES } from "@/app/(main)/discover/_containers/advanced-search/_lib/constants"
import { useMangaCollection } from "@/app/(main)/manga/_lib/queries"
import { MangaCollectionList } from "@/app/(main)/manga/_lib/types"
import { AnimeListItem } from "@/components/shared/anime-list-item"
import { PageWrapper } from "@/components/shared/styling/page-wrapper"
import { Carousel, CarouselContent, CarouselDotButtons } from "@/components/ui/carousel"
import { HorizontalDraggableScroll } from "@/components/ui/horizontal-draggable-scroll"
import { Skeleton } from "@/components/ui/skeleton"
import { StaticTabs } from "@/components/ui/tabs"
import { TextInput } from "@/components/ui/text-input"
import { useDebounce } from "@/hooks/use-debounce"
import { ListMangaQuery } from "@/lib/anilist/gql/graphql"
import { SeaEndpoints } from "@/lib/server/endpoints"
import { useSeaQuery } from "@/lib/server/query"
import { getMangaCollectionTitle } from "@/lib/server/utils"
import { ThemeLibraryScreenBannerType, useThemeSettings } from "@/lib/theme/hooks"
import { atom } from "jotai/index"
import { useAtom, useAtomValue } from "jotai/react"
import React, { memo } from "react"
import { FiSearch } from "react-icons/fi"

export default function Page() {
    const { mangaCollection, mangaCollectionLoading } = useMangaCollection()

    const ts = useThemeSettings()

    if (!mangaCollection || mangaCollectionLoading) return <LoadingDisplay />

    return (
        <div>
            {ts.libraryScreenBannerType === ThemeLibraryScreenBannerType.Custom && (
                <>
                    <CustomLibraryBanner />
                    <div className="h-32"></div>
                </>
            )}

            <div className="px-4 md:px-8 relative z-[8]">

                <PageWrapper
                    className="relative 2xl:order-first pb-10 pt-4"
                    {...{
                        initial: { opacity: 0, y: 60 },
                        animate: { opacity: 1, y: 0 },
                        exit: { opacity: 0, y: 60 },
                        transition: {
                            type: "spring",
                            damping: 10,
                            stiffness: 80,
                            delay: 0.6,
                        },
                    }}
                >

                    <div className="space-y-8">
                        {mangaCollection.lists.map(list => {
                            return <CollectionListItem key={list.type} list={list} />
                        })}

                        <h2>
                            Trending
                        </h2>

                        <TrendingManga />

                        <SearchManga />
                    </div>

                </PageWrapper>
            </div>
        </div>
    )
}

const CollectionListItem = memo(({ list }: { list: MangaCollectionList }) => {
    return (
        <React.Fragment key={list.type}>
            <h2>{getMangaCollectionTitle(list.type)}</h2>
            <div
                className="grid grid-cols-2 md:grid-cols-3 lg:grid-cols-4 xl:grid-cols-5 2xl:grid-cols-7 min-[2000px]:grid-cols-8 gap-4"
            >
                {list.entries?.map(entry => {
                    return <AnimeListItem
                        key={entry.media.id}
                        media={entry.media!}
                        listData={entry.listData}
                        showListDataButton
                        withAudienceScore={false}
                        isManga
                    />
                })}
            </div>
        </React.Fragment>
    )
})

const trendingGenresAtom = atom<string[]>([])

function TrendingManga() {
    const genres = useAtomValue(trendingGenresAtom)
    const { data, isLoading } = useSeaQuery<ListMangaQuery>({
        queryKey: ["discover-trending-manga", genres],
        endpoint: SeaEndpoints.MANGA_ANILIST_LIST_MANGA,
        method: "post",
        data: {
            page: 1,
            perPage: 20,
            sort: ["TRENDING_DESC"],
            genres: genres.length > 0 ? genres : undefined,
        },
    })

    return (
        <Carousel
            className="w-full max-w-full"
            gap="md"
            opts={{
                align: "start",
            }}
            autoScroll
        >
            <GenreSelector />
            <CarouselDotButtons />
            <CarouselContent className="px-6">
                {!isLoading ? data?.Page?.media?.filter(Boolean).map(media => {
                    return (
                        <AnimeListItem
                            key={media.id}
                            media={media}
                            containerClassName="basis-[200px] md:basis-[250px] mx-2 my-8"
                            isManga
                        />
                    )
                }) : [...Array(10).keys()].map((v, idx) => <AnimeSliderSkeletonItem key={idx} />)}
            </CarouselContent>
        </Carousel>
    )
}

const mangaSearchInputAtom = atom<string>("")

function SearchManga() {
    const [searchInput, setSearchInput] = useAtom(mangaSearchInputAtom)
    const search = useDebounce(searchInput, 500)

    const { data, isLoading, isFetching } = useSeaQuery<ListMangaQuery>({
        queryKey: ["search-manga", search],
        endpoint: SeaEndpoints.MANGA_ANILIST_LIST_MANGA,
        method: "post",
        data: {
            page: 1,
            perPage: 10,
            search: search,
        },
    })

    return (
        <div className="space-y-4">
            <div className="container">
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
                }}
                autoScroll
            >
                <CarouselContent className="px-6">
                    {!(isLoading || isFetching) ? data?.Page?.media?.filter(Boolean).map(media => {
                        return (
                            <AnimeListItem
                                key={media.id}
                                media={media}
                                containerClassName="basis-[200px] md:basis-[250px] mx-2 my-8"
                                isManga
                            />
                        )
                    }) : [...Array(10).keys()].map((v, idx) => <AnimeSliderSkeletonItem key={idx} />)}
                </CarouselContent>
            </Carousel>}
        </div>
    )
}


function GenreSelector() {

    const [selectedGenre, setSelectedGenre] = useAtom(trendingGenresAtom)

    return (
        <HorizontalDraggableScroll className="w-full scroll-pb-1 pt-0">
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


function LoadingDisplay() {
    return (
        <div className="__header h-[30rem]">
            <div
                className="h-[30rem] w-full flex-none object-cover object-center absolute top-0 overflow-hidden"
            >
                <div
                    className="w-full absolute z-[1] top-0 h-[15rem] bg-gradient-to-b from-[--background] to-transparent via"
                />
                <Skeleton className="h-full absolute w-full" />
                <div
                    className="w-full absolute bottom-0 h-[20rem] bg-gradient-to-t from-[--background] via-transparent to-transparent"
                />
            </div>
        </div>
    )
}
