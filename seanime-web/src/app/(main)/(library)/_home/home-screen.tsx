import { AL_ListRecentAnime_Page_AiringSchedules, Anime_ScheduleItem, Models_HomeItem } from "@/api/generated/types"
import { useAnilistListAnime, useAnilistListRecentAiringAnime } from "@/api/hooks/anilist.hooks"
import { useGetLibraryCollection } from "@/api/hooks/anime_collection.hooks"
import { useAnilistListManga } from "@/api/hooks/manga.hooks"
import { useGetHomeItems } from "@/api/hooks/status.hooks"
import { LibraryHeader } from "@/app/(main)/(library)/_components/library-header"
import { BulkActionModal } from "@/app/(main)/(library)/_containers/bulk-action-modal"
import { ContinueWatching } from "@/app/(main)/(library)/_containers/continue-watching"
import { CustomLibraryBanner } from "@/app/(main)/(library)/_containers/custom-library-banner"
import { IgnoredFileManager } from "@/app/(main)/(library)/_containers/ignored-file-manager"
import { __scanner_modalIsOpen } from "@/app/(main)/(library)/_containers/scanner-modal"
import { UnknownMediaManager } from "@/app/(main)/(library)/_containers/unknown-media-manager"
import { UnmatchedFileManager } from "@/app/(main)/(library)/_containers/unmatched-file-manager"
import { DEFAULT_HOME_ITEMS, HOME_ITEMS, isAnimeLibraryItemsOnly } from "@/app/(main)/(library)/_home/home-items.utils"
import { __home_settingsModalOpen, HomeSettingsModal } from "@/app/(main)/(library)/_home/home-settings-modal"
import { HomeToolbar } from "@/app/(main)/(library)/_home/home-toolbar"
import { HandleLibraryCollectionProps, useHandleLibraryCollection } from "@/app/(main)/(library)/_lib/handle-library-collection"
import { DetailedLibraryView } from "@/app/(main)/(library)/_screens/detailed-library-view"
import { LibraryView } from "@/app/(main)/(library)/_screens/library-view"
import { __anilist_userAnimeMediaAtom } from "@/app/(main)/_atoms/anilist.atoms"
import { MangaLibraryHeader } from "@/app/(main)/manga/_components/library-header"
import { PageWrapper } from "@/components/shared/page-wrapper"
import { SeaLink } from "@/components/shared/sea-link"
import { Button } from "@/components/ui/button"
import { Carousel, CarouselContent, CarouselDotButtons } from "@/components/ui/carousel"
import { cn } from "@/components/ui/core/styling"
import { Skeleton } from "@/components/ui/skeleton"
import { ThemeLibraryScreenBannerType, useThemeSettings } from "@/lib/theme/hooks"
import { HIDE_IMAGES } from "@/types/constants"
import { addDays } from "date-fns/addDays"
import { useAtomValue } from "jotai"
import { atom, useSetAtom } from "jotai/index"
import { useAtom } from "jotai/react"
import { atomWithStorage } from "jotai/utils"
import { AnimatePresence, useInView } from "motion/react"
import React from "react"
import { FiSearch } from "react-icons/fi"
import { LiaPlayCircle } from "react-icons/lia"
import { LuPlus } from "react-icons/lu"
import { useWindowSize } from "react-use"
import { MediaEntryCard } from "../../_features/media/_components/media-entry-card"
import { MediaEntryCardSkeleton } from "../../_features/media/_components/media-entry-card-skeleton"
import { MediaEntryPageLoadingDisplay } from "../../_features/media/_components/media-entry-page-loading-display"
import { useServerStatus } from "../../_hooks/use-server-status"
import { DiscoverPageHeader } from "../../discover/_components/discover-page-header"
import { DiscoverTrending } from "../../discover/_containers/discover-trending"
import { DiscoverTrendingCountry } from "../../discover/_containers/discover-trending-country"
import { __discord_pageTypeAtom } from "../../discover/_lib/discover.atoms"
import { useHandleMangaCollection } from "../../manga/_lib/handle-manga-collection"
import { MangaLibraryView } from "../../manga/_screens/manga-library-view"
import { ScheduleCalendar } from "../../schedule/_components/schedule-calendar"
import { ComingUpNext } from "../../schedule/_containers/coming-up-next"
import { RecentReleases } from "../../schedule/_containers/recent-releases"
import { ContinueWatchingHeader } from "../_containers/continue-watching-header"

export const __home_currentView = atom<"base" | "detailed">("base")

export const __home_discoverHeaderType = atomWithStorage<"anime" | "manga">("sea-home-discover-header-type", "anime", undefined, { getOnInit: true })

export function HomeScreen() {
    const serverStatus = useServerStatus()
    const { data: _homeItems, isLoading: isLoadingItems } = useGetHomeItems()

    const { width } = useWindowSize()

    const allUserMedia = useAtomValue(__anilist_userAnimeMediaAtom)
    const noMediaInCollection = !allUserMedia?.length

    const {
        libraryGenres,
        libraryCollectionList,
        filteredLibraryCollectionList,
        isLoading,
        continueWatchingList,
        unmatchedLocalFiles,
        ignoredLocalFiles,
        unmatchedGroups,
        unknownGroups,
        streamingMediaIds,
        hasEntries,
        isStreamingOnly,
        isNakamaLibrary,
    } = useHandleLibraryCollection()

    const {
        mangaCollection,
    } = useHandleMangaCollection()

    const ts = useThemeSettings()

    // const homeItems = !isNakamaLibrary ? (!!_homeItems?.length ? _homeItems : DEFAULT_HOME_ITEMS) : DEFAULT_HOME_ITEMS
    const [view, setView] = useAtom(__home_currentView)
    const [discoverHeaderType, setDiscoverHeaderType] = useAtom(__home_discoverHeaderType)
    const [discoverPageType, setDiscoverPageType] = useAtom(__discord_pageTypeAtom)
    const setHomeSettingsModalOpen = useSetAtom(__home_settingsModalOpen)

    const homeItems = React.useMemo(() => {
        let ret = !!_homeItems?.length ? _homeItems : DEFAULT_HOME_ITEMS
        // replace anime-continue-watching-header with anime-continue-watching on mobile
        if (width < 1024 && ret[0]?.type === "anime-continue-watching-header") {
            if (ret.find(n => n.type === "anime-continue-watching")) {
                // remove any other anime continue watching
                ret = ret.filter(n => n.type !== "anime-continue-watching")
            }
            return ret.map(item => {
                if (item.type === "anime-continue-watching-header") {
                    return {
                        ...item,
                        type: "anime-continue-watching",
                    }
                }
                return item
            })
        }
        return ret
    }, [_homeItems, width < 1024])

    React.useEffect(() => {
        setDiscoverPageType(discoverPageType)
    }, [discoverPageType])

    const setScannerModalOpen = useSetAtom(__scanner_modalIsOpen)

    const animeLibraryType = (serverStatus?.torrentstreamSettings?.includeInLibrary || serverStatus?.debridSettings?.includeDebridStreamInLibrary || serverStatus?.settings?.library?.includeOnlineStreamingInLibrary)
        ?
        "stream"
        : "local"


    if (isLoading || isLoadingItems) return <React.Fragment>
        <div className="p-4 space-y-4 relative z-[4]">
            <Skeleton className="h-12 w-full max-w-lg relative" />
            <div
                className={cn(
                    "grid h-[22rem] min-[2000px]:h-[24rem] grid-cols-2 md:grid-cols-3 lg:grid-cols-4 xl:grid-cols-5 2xl:grid-cols-7 min-[2000px]:grid-cols-8 gap-4",
                )}
            >
                {[1, 2, 3, 4, 5, 6, 7, 8]?.map((_, idx) => {
                    return <Skeleton
                        key={idx} className={cn(
                        "h-[22rem] min-[2000px]:h-[24rem] col-span-1 aspect-[6/7] flex-none rounded-[--radius-md] relative overflow-hidden",
                        "[&:nth-child(8)]:hidden min-[2000px]:[&:nth-child(8)]:block",
                        "[&:nth-child(7)]:hidden 2xl:[&:nth-child(7)]:block",
                        "[&:nth-child(6)]:hidden xl:[&:nth-child(6)]:block",
                        "[&:nth-child(5)]:hidden xl:[&:nth-child(5)]:block",
                        "[&:nth-child(4)]:hidden lg:[&:nth-child(4)]:block",
                        "[&:nth-child(3)]:hidden md:[&:nth-child(3)]:block",
                    )}
                    />
                })}
            </div>
        </div>
    </React.Fragment>

    if (!hasEntries && isAnimeLibraryItemsOnly(homeItems) && !isLoading) {
        return (
            <div data-home-screen="no-entries" className="contents">
                <React.Fragment>
                    <DiscoverPageHeader playTrailer={!HIDE_IMAGES} />
                    <div className="h-0 visibility-hidden pointer-events-none opacity-0">
                        {/*{discoverHeaderType === "anime" && <DiscoverTrending />}*/}
                        {discoverHeaderType === "manga" && <DiscoverTrendingCountry country="JP" forDiscoverHeader />}
                    </div>
                </React.Fragment>

                <HomeToolbar
                    collectionList={libraryCollectionList}
                    unmatchedLocalFiles={unmatchedLocalFiles}
                    ignoredLocalFiles={ignoredLocalFiles}
                    unknownGroups={unknownGroups}
                    isLoading={isLoading}
                    hasEntries={hasEntries}
                    isStreamingOnly={isStreamingOnly}
                    isNakamaLibrary={isNakamaLibrary}
                    className={cn(
                        (homeItems[0]?.type === "discover-header" || homeItems[0]?.type === "anime-continue-watching-header") && "!mt-[-4rem] !mb-[-1rem]",
                    )}
                />

                <div className="text-center space-y-6 py-10">
                    <h2>Your home screen is empty</h2>

                    {!!serverStatus?.settings?.library?.libraryPath && <>
                        <Button
                            intent="primary-glass"
                            leftIcon={<FiSearch />}
                            size="xl"
                            rounded
                            onClick={() => setScannerModalOpen(true)}
                        >
                            Scan your anime library
                        </Button>
                    </>}

                    {!serverStatus?.settings?.library?.libraryPath && noMediaInCollection && <>
                        <SeaLink href="/discover" className="block">
                            <Button
                                intent="gray-glass"
                                leftIcon={<LuPlus />}
                                size="lg"
                                rounded
                            >
                                Add series to your collection
                            </Button>
                        </SeaLink>
                    </>}

                    {!serverStatus?.settings?.library?.libraryPath && !noMediaInCollection && <>
                        {animeLibraryType === "local" && <Button
                            intent="gray-glass"
                            leftIcon={<LiaPlayCircle className="text-2xl" />}
                            size="lg"
                            rounded
                            onClick={() => {
                                setHomeSettingsModalOpen(true)
                            }}
                        >
                            Add currently watched series to the library
                        </Button>}

                        {animeLibraryType === "stream" && <div className="p-4 border w-fit mx-auto border-dashed rounded-xl">
                            <p>
                                No series are currently being watched
                            </p>
                            <p className="text-[--muted]">
                                Add series to your 'Currently watching' list to get started
                            </p>
                        </div>}
                    </>}


                </div>

                <h3>Trending Right Now</h3>
                <DiscoverTrending />

                <div data-home-screen-item-divider className="h-8" />

                <HomeSettingsModal emptyLibrary isNakamaLibrary={isNakamaLibrary} />

                <UnmatchedFileManager
                    unmatchedGroups={unmatchedGroups}
                />
                <UnknownMediaManager
                    unknownGroups={unknownGroups}
                />
                <IgnoredFileManager
                    files={ignoredLocalFiles}
                />
                <BulkActionModal />
            </div>
        )
    }

    return (
        <div data-home-screen className="contents">

            {/*Discover Page Header*/}
            {homeItems[0]?.type === "discover-header" && <React.Fragment>
                <DiscoverPageHeader />
                <div className="h-0 visibility-hidden pointer-events-none opacity-0">
                    {discoverHeaderType === "anime" && <DiscoverTrending />}
                    {discoverHeaderType === "manga" && <DiscoverTrendingCountry country="JP" forDiscoverHeader />}
                </div>
            </React.Fragment>}

            {/*Continue Watching Header*/}
            {homeItems[0]?.type === "anime-continue-watching-header" && <React.Fragment>
                <ContinueWatchingHeader episodes={continueWatchingList} />
            </React.Fragment>}

            {/*Manga Library Header*/}
            {(ts.libraryScreenBannerType === ThemeLibraryScreenBannerType.Dynamic && homeItems[0]?.type === "manga-library") && (
                <>
                    <MangaLibraryHeader manga={mangaCollection?.lists?.flatMap(l => l.entries)?.flatMap(e => e?.media)?.filter(Boolean) || []} />
                </>
            )}

            <div
                className={cn(
                    "h-12 lg:hidden",
                )}
                data-library-toolbar-top-padding
            ></div>
            {(homeItems[0]?.type !== "anime-continue-watching-header" && homeItems[0]?.type === "anime-continue-watching") && <div
                className={cn(
                    "lg:h-16 hidden",
                )}
                data-library-toolbar-top-padding
            ></div>}

            {(
                (ts.libraryScreenBannerType === ThemeLibraryScreenBannerType.Dynamic && hasEntries) &&
                (homeItems[0]?.type === "anime-continue-watching" || homeItems[0]?.type === "manga-library")
            ) && <div
                className={cn(
                    "h-28",
                    ts.hideTopNavbar && "h-40",
                )}
                data-library-toolbar-top-padding
            ></div>}

            <HomeToolbar
                collectionList={libraryCollectionList}
                unmatchedLocalFiles={unmatchedLocalFiles}
                ignoredLocalFiles={ignoredLocalFiles}
                unknownGroups={unknownGroups}
                isLoading={isLoading}
                hasEntries={hasEntries}
                isStreamingOnly={isStreamingOnly}
                isNakamaLibrary={isNakamaLibrary}
                className={cn(
                    (homeItems[0]?.type === "discover-header" || homeItems[0]?.type === "anime-continue-watching-header") && "!mt-[-4rem] !mb-[-1rem]",
                )}
            />

            {/*Custom Library Banner*/}
            {(!!ts.libraryScreenCustomBannerImage
                && ts.libraryScreenBannerType === ThemeLibraryScreenBannerType.Custom
                && homeItems[0]?.type !== "discover-header"
                && homeItems[0]?.type !== "anime-continue-watching-header"
            ) && <div
                data-custom-library-banner-top-spacer
                className={cn(
                    "py-20",
                    ts.hideTopNavbar && "py-28",
                )}
            ></div>}
            {(
                hasEntries &&
                ts.libraryScreenBannerType === ThemeLibraryScreenBannerType.Custom
                && homeItems[0]?.type !== "discover-header"
                && homeItems[0]?.type !== "anime-continue-watching-header"
            ) && <CustomLibraryBanner isLibraryScreen isHomeScreen />}

            {(hasEntries && homeItems.findIndex(n => n.type === "anime-continue-watching") !== -1) && ts.libraryScreenBannerType === ThemeLibraryScreenBannerType.Dynamic &&
                <div
                    className={cn(
                        homeItems[0]?.type !== "anime-continue-watching" ? "visibility-hidden pointer-events-none opacity-0 !mt-0" : "contents !mt-0",
                    )}
                >
                    <LibraryHeader list={continueWatchingList} />
                </div>}


            {!isLoading && <AnimatePresence mode="wait">
                {view === "base" && <PageWrapper
                    key="base"
                    className="relative 2xl:order-first pb-10 pt-4"
                    {...{
                        initial: { opacity: 0, y: 5 },
                        animate: { opacity: 1, y: 0 },
                        exit: { opacity: 0, scale: 0.99 },
                        transition: {
                            duration: 0.25,
                        },
                    }}
                >
                    {homeItems.filter(n => n.type !== "discover-header" && n.type !== "anime-continue-watching-header").map((item, index) => {
                        return (
                            <React.Fragment key={item.id}>
                                {(index !== 0 &&
                                    !(item?.type === "manga-library" || item?.type === "anime-library" || item?.type === "anime-continue-watching" || item.type === "anime-library-stats")
                                ) && <div data-home-screen-item-divider={index} className="h-8" />}
                                <HomeScreenItem
                                    item={item}
                                    index={homeItems.findIndex(n => n.id === item.id)}
                                    libraryCollectionProps={{
                                        libraryGenres,
                                        libraryCollectionList,
                                        filteredLibraryCollectionList,
                                        isLoading,
                                        continueWatchingList,
                                        unmatchedLocalFiles,
                                        ignoredLocalFiles,
                                        unmatchedGroups,
                                        unknownGroups,
                                        streamingMediaIds,
                                        hasEntries,
                                        isStreamingOnly,
                                        isNakamaLibrary,
                                    }}
                                />
                            </React.Fragment>
                        )
                    })}

                    <div data-home-screen-item-divider className="h-8" />
                </PageWrapper>}

                {view === "detailed" && <PageWrapper
                    key="detailed"
                    className="relative 2xl:order-first pb-10 pt-4"
                    {...{
                        initial: { opacity: 0, y: 5 },
                        animate: { opacity: 1, y: 0 },
                        exit: { opacity: 0, scale: 0.99 },
                        transition: {
                            duration: 0.25,
                        },
                    }}
                >
                    <DetailedLibraryView
                        collectionList={libraryCollectionList}
                        continueWatchingList={continueWatchingList}
                        isLoading={isLoading}
                        hasEntries={hasEntries}
                        streamingMediaIds={streamingMediaIds}
                        isNakamaLibrary={isNakamaLibrary}
                    />
                </PageWrapper>}
            </AnimatePresence>}

            <HomeSettingsModal isNakamaLibrary={isNakamaLibrary} />

            <UnmatchedFileManager
                unmatchedGroups={unmatchedGroups}
            />
            <UnknownMediaManager
                unknownGroups={unknownGroups}
            />
            <IgnoredFileManager
                files={ignoredLocalFiles}
            />
            <BulkActionModal />
        </div>
    )
}

export type HomeScreenItemProps = {
    item: Models_HomeItem
    libraryCollectionProps: HandleLibraryCollectionProps
    index: number
}

export function HomeScreenItem(props: HomeScreenItemProps) {
    const { item: _item, index } = props
    const {
        libraryGenres,
        libraryCollectionList,
        filteredLibraryCollectionList,
        isLoading,
        continueWatchingList,
        unmatchedLocalFiles,
        ignoredLocalFiles,
        unmatchedGroups,
        unknownGroups,
        streamingMediaIds,
        hasEntries,
        isStreamingOnly,
        isNakamaLibrary,
    } = props.libraryCollectionProps


    const ts = useThemeSettings()

    const schema = HOME_ITEMS[_item.type]

    // remove item options if schema version has changed
    const item = React.useMemo(() => {
        if (!schema || !_item) return undefined
        if (!_item.schemaVersion || _item.schemaVersion !== schema.schemaVersion) {
            return {
                ..._item,
                schemaVersion: schema.schemaVersion,
                options: undefined,
            }
        }
        return _item
    }, [_item, schema])

    const { data } = useGetLibraryCollection({ enabled: item?.type === "local-anime-library-stats" })


    if (!schema || !item) return <div>
        Item not found
    </div>


    if (item.type === "centered-title") {
        return (
            <>
                <h2 data-home-screen-centered-title={item.options?.text} className="text-center text-3xl lg:text-4xl font-bold py-4">
                    {item.options?.text}
                </h2>
            </>
        )
    }

    if (item.type === "anime-continue-watching") {
        return (
            <>
                <ContinueWatching
                    episodes={continueWatchingList}
                    isLoading={isLoading}
                    withTitle={index === 0}
                />
            </>
        )
    }

    if (item.type === "anime-library") {
        return (
            <>
                <LibraryView
                    genres={libraryGenres}
                    collectionList={libraryCollectionList}
                    filteredCollectionList={filteredLibraryCollectionList}
                    continueWatchingList={continueWatchingList}
                    isLoading={isLoading}
                    hasEntries={hasEntries}
                    streamingMediaIds={streamingMediaIds}
                    showStatuses={item.options?.statuses}
                    type={item.options?.layout || "grid"}
                />
            </>
        )
    }

    if (item.type === "anime-carousel") {
        return (
            <>
                <AnimeCarousel
                    libraryCollectionProps={props.libraryCollectionProps}
                    item={item}
                />
            </>
        )
    }

    if (item.type === "manga-carousel") {
        return (
            <>
                <MangaCarousel libraryCollectionProps={props.libraryCollectionProps} item={item} />
            </>
        )
    }

    if (item.type === "anime-schedule-calendar") {
        return (
            <>
                {item.options?.type !== "global" && <AnimeScheduleCalendar libraryCollectionProps={props.libraryCollectionProps} item={item} />}
                {item.options?.type === "global" && <GlobalAnimeScheduleCalendar libraryCollectionProps={props.libraryCollectionProps} item={item} />}
            </>
        )
    }

    if (item.type === "library-upcoming-episodes") {
        return (
            <>
                <LibraryUpcomingEpisodes libraryCollectionProps={props.libraryCollectionProps} item={item} />
            </>
        )
    }

    if (item.type === "discover-header" || item.type === "anime-continue-watching-header") {
        return null
    }

    if (item.type === "aired-recently") {
        return (
            <PageWrapper className="px-4">
                <RecentReleases />
            </PageWrapper>
        )
    }

    if (item.type === "manga-library") {
        return (
            <>
                <MangaLibrary libraryCollectionProps={props.libraryCollectionProps} item={item} index={index} />
            </>
        )
    }

    if (item.type === "local-anime-library") {
        return (
            <>
                <LocalAnimeLibrary libraryCollectionProps={props.libraryCollectionProps} item={item} index={index} />
            </>
        )
    }

    if (item.type === "local-anime-library-stats") {
        return (
            <PageWrapper>
                <div
                    className={cn(
                        "grid grid-cols-3 lg:grid-cols-6 gap-4 [&>div]:text-center [&>div>p]:text-[--muted] py-4",
                        isNakamaLibrary && "lg:grid-cols-5",
                    )}
                    data-detailed-library-view-stats-container
                >
                    {!isNakamaLibrary && <div>
                        <h3>{data?.stats?.totalSize ?? "-"}</h3>
                        <p>Library</p>
                    </div>}
                    <div>
                        <h3>{data?.stats?.totalFiles ?? "-"}</h3>
                        <p>Files</p>
                    </div>
                    <div>
                        <h3>{data?.stats?.totalEntries ?? "-"}</h3>
                        <p>Entries</p>
                    </div>
                    <div>
                        <h3>{data?.stats?.totalShows ?? "-"}</h3>
                        <p>TV Shows</p>
                    </div>
                    <div>
                        <h3>{data?.stats?.totalMovies ?? "-"}</h3>
                        <p>Movies</p>
                    </div>
                    <div>
                        <h3>{data?.stats?.totalSpecials ?? "-"}</h3>
                        <p>Specials</p>
                    </div>
                </div>
            </PageWrapper>
        )
    }

    return <div>
        Item not found ({item.type})
    </div>
}

function LocalAnimeLibrary(props: { libraryCollectionProps: HandleLibraryCollectionProps, item: Models_HomeItem, index: number }) {
    return (
        <>
            <DetailedLibraryView
                isHomeItem={true}
                collectionList={props.libraryCollectionProps.libraryCollectionList}
                continueWatchingList={props.libraryCollectionProps.continueWatchingList}
                isLoading={props.libraryCollectionProps.isLoading}
                hasEntries={props.libraryCollectionProps.hasEntries}
                streamingMediaIds={props.libraryCollectionProps.streamingMediaIds}
                isNakamaLibrary={props.libraryCollectionProps.isNakamaLibrary}
                type={props.item?.options?.layout || "grid"}
            />
        </>
    )

}

function MangaLibrary(props: { libraryCollectionProps: HandleLibraryCollectionProps, item: Models_HomeItem, index: number }) {
    const { libraryCollectionProps, item, index } = props
    const {} = libraryCollectionProps
    const ts = useThemeSettings()

    const {
        mangaCollection,
        filteredMangaCollection,
        mangaCollectionLoading,
        storedFilters,
        storedProviders,
        mangaCollectionGenres,
        hasManga,
    } = useHandleMangaCollection()

    if (!mangaCollection || mangaCollectionLoading) return <MediaEntryPageLoadingDisplay />

    return <>

        <>
            <MangaLibraryView
                collection={mangaCollection}
                filteredCollection={filteredMangaCollection}
                // genres={mangaCollectionGenres}
                genres={[]}
                storedProviders={storedProviders}
                hasManga={hasManga}
                showStatuses={item.options?.statuses}
                type={item.options?.layout || "grid"}
                withTitle={index === 0}
            />
        </>
    </>
}

function LibraryUpcomingEpisodes(props: { libraryCollectionProps: HandleLibraryCollectionProps, item: Models_HomeItem }) {
    const { libraryCollectionProps, item } = props
    const { hasEntries } = libraryCollectionProps

    if (!hasEntries) return null

    return <PageWrapper className="space-y-0 px-4">
        <ComingUpNext />
    </PageWrapper>
}


function AnimeScheduleCalendar(props: { libraryCollectionProps: HandleLibraryCollectionProps, item: Models_HomeItem }) {
    const { libraryCollectionProps, item } = props
    const {} = libraryCollectionProps

    return <PageWrapper className="space-y-0 px-4 py-4">
        <ScheduleCalendar />
    </PageWrapper>
}

function GlobalAnimeScheduleCalendar(props: { libraryCollectionProps: HandleLibraryCollectionProps, item: Models_HomeItem }) {
    const { libraryCollectionProps, item } = props
    const {} = libraryCollectionProps

    const now = new Date()
    const twoWeeksBefore = addDays(now, -14)
    const weekBefore = addDays(now, -7)
    const weekAfter = addDays(now, 7)
    const twoWeeksAfter = addDays(now, 14)

    const { data: previous1, isLoading: loadingPrevious1 } = useAnilistListRecentAiringAnime({
        page: 1,
        perPage: 100,
        airingAt_lesser: Math.floor(weekBefore.getTime() / 1000),
        airingAt_greater: Math.floor(twoWeeksBefore.getTime() / 1000),
        notYetAired: false,
        sort: ["TIME"],
    })

    const { data: previous2, isLoading: loadingPrevious2 } = useAnilistListRecentAiringAnime({
        page: 2,
        perPage: 100,
        airingAt_lesser: Math.floor(weekBefore.getTime() / 1000),
        airingAt_greater: Math.floor(twoWeeksBefore.getTime() / 1000),
        notYetAired: false,
        sort: ["TIME"],
    })

    const { data: previous3, isLoading: loadingPrevious3 } = useAnilistListRecentAiringAnime({
        page: 1,
        perPage: 200,
        airingAt_lesser: Math.floor(now.getTime() / 1000),
        airingAt_greater: Math.floor(weekBefore.getTime() / 1000),
        notYetAired: false,
        sort: ["TIME"],
    })

    const { data: previous4, isLoading: loadingPrevious4 } = useAnilistListRecentAiringAnime({
        page: 2,
        perPage: 200,
        airingAt_lesser: Math.floor(now.getTime() / 1000),
        airingAt_greater: Math.floor(weekBefore.getTime() / 1000),
        notYetAired: false,
        sort: ["TIME"],
    })

    const { data: next1, isLoading: loadingNext1 } = useAnilistListRecentAiringAnime({
        page: 1,
        perPage: 100,
        airingAt_lesser: Math.floor(weekAfter.getTime() / 1000),
        airingAt_greater: Math.floor(now.getTime() / 1000),
        notYetAired: true,
        sort: ["TIME"],
    })

    const { data: next2, isLoading: loadingNext2 } = useAnilistListRecentAiringAnime({
        page: 2,
        perPage: 100,
        airingAt_lesser: Math.floor(weekAfter.getTime() / 1000),
        airingAt_greater: Math.floor(now.getTime() / 1000),
        notYetAired: true,
        sort: ["TIME"],
    })

    const { data: next3, isLoading: loadingNext3 } = useAnilistListRecentAiringAnime({
        page: 1,
        perPage: 100,
        airingAt_lesser: Math.floor(twoWeeksAfter.getTime() / 1000),
        airingAt_greater: Math.floor(weekAfter.getTime() / 1000),
        notYetAired: true,
        sort: ["TIME"],
    })

    const { data: next4, isLoading: loadingNext4 } = useAnilistListRecentAiringAnime({
        page: 2,
        perPage: 100,
        airingAt_lesser: Math.floor(twoWeeksAfter.getTime() / 1000),
        airingAt_greater: Math.floor(weekAfter.getTime() / 1000),
        notYetAired: true,
        sort: ["TIME"],
    })

    const items = React.useMemo<Anime_ScheduleItem[]>(() => {
        let airingSchedules: AL_ListRecentAnime_Page_AiringSchedules[] = []

        // Combine all results
        if (previous1?.Page?.airingSchedules) {
            airingSchedules = [...previous1.Page.airingSchedules]
        }
        if (previous2?.Page?.airingSchedules) {
            airingSchedules = [...airingSchedules, ...previous2.Page.airingSchedules]
        }
        if (previous3?.Page?.airingSchedules) {
            airingSchedules = [...airingSchedules, ...previous3.Page.airingSchedules]
        }
        if (previous4?.Page?.airingSchedules) {
            airingSchedules = [...airingSchedules, ...previous4.Page.airingSchedules]
        }
        if (next1?.Page?.airingSchedules) {
            airingSchedules = [...airingSchedules, ...next1.Page.airingSchedules]
        }
        if (next2?.Page?.airingSchedules) {
            airingSchedules = [...airingSchedules, ...next2.Page.airingSchedules]
        }
        if (next3?.Page?.airingSchedules) {
            airingSchedules = [...airingSchedules, ...next3.Page.airingSchedules]
        }
        if (next4?.Page?.airingSchedules) {
            airingSchedules = [...airingSchedules, ...next4.Page.airingSchedules]
        }

        return airingSchedules.map(schedule => {
            const airDate = new Date((schedule.airingAt || 0) * 1000)
            const dateTimeStr = airDate.toISOString()
            const timeStr = `${airDate.getHours().toString().padStart(2, "0")}:${airDate.getMinutes().toString().padStart(2, "0")}`
            return {
                mediaId: schedule.media!.id,
                title: schedule.media!.title!.userPreferred!,
                time: timeStr,
                dateTime: dateTimeStr,
                image: schedule.media?.bannerImage || schedule.media?.coverImage?.medium!,
                episodeNumber: schedule.episode,
                isMovie: schedule.media?.format === "MOVIE",
                isSeasonFinale: !!schedule.media?.episodes && schedule.media.episodes === schedule.episode,
            }
        })
    }, [previous1, previous2, previous3, previous4, next1, next2, next3, next4])

    return <PageWrapper className="space-y-0 px-4 py-4">
        <ScheduleCalendar key="home-screen" items={items} />
    </PageWrapper>
}

function AnimeCarousel(props: { libraryCollectionProps: HandleLibraryCollectionProps, item: Models_HomeItem }) {
    const { libraryCollectionProps, item } = props
    const {} = libraryCollectionProps
    const ref = React.useRef(null)

    const options = item.options as Record<string, any> | undefined

    const isInView = useInView(ref, { once: true })

    const { data, isLoading } = useAnilistListAnime({
        page: 1,
        perPage: 20,
        sort: options?.sorting ? [options.sorting] : ["SCORE_DESC"],
        season: options?.season || undefined,
        seasonYear: !!options?.year ? options.year : undefined,
        genres: !!options?.genres?.length ? options?.genres : undefined,
        format: options?.format || undefined,
        status: (!!options?.status?.length && Array.isArray(options.status)) ? options.status as any : ["RELEASING", "FINISHED"],
        isAdult: options?.isAdult || undefined,
        countryOfOrigin: options?.countryOfOrigin || undefined,
    }, !!options?.name && isInView)

    // if (!isLoading && !data && !isInView) return

    return (
        <PageWrapper className="space-y-0 px-4" ref={ref}>
            <h2>{options?.name || "Anime Carousel"}</h2>
            {(!isLoading && !data && isInView) ? <InvalidHomeItem item={item} /> : <Carousel
                className="w-full max-w-full"
                gap="xl"
                opts={{
                    align: "start",
                    dragFree: true,
                }}
                autoScroll
            >
                {/*<CarouselMasks />*/}
                <CarouselDotButtons className="-top-2" />
                <CarouselContent className="px-6">
                    {!!data ? data?.Page?.media?.filter(Boolean)?.sort((a, b) => b.meanScore! - a.meanScore!).map(media => {
                        return (
                            <MediaEntryCard
                                key={media.id}
                                media={media}
                                showLibraryBadge
                                containerClassName="basis-[200px] md:basis-[250px] mx-2 mt-8 mb-0"
                                showTrailer
                                type="anime"
                            />
                        )
                    }) : [...Array(10).keys()].map((v, idx) => <MediaEntryCardSkeleton key={idx} />)}
                </CarouselContent>
            </Carousel>}
        </PageWrapper>
    )
}

function MangaCarousel(props: { libraryCollectionProps: HandleLibraryCollectionProps, item: Models_HomeItem }) {
    const { libraryCollectionProps, item } = props
    const {} = libraryCollectionProps
    const ref = React.useRef(null)

    const options = item.options as Record<string, any> | undefined

    const isInView = useInView(ref, { once: true })

    const { data, isLoading } = useAnilistListManga({
        page: 1,
        perPage: 20,
        sort: options?.sorting ? [options.sorting] : ["SCORE_DESC"],
        year: !!options?.year ? options.year : undefined,
        genres: !!options?.genres?.length ? options?.genres : undefined,
        format: options?.format || undefined,
        status: (!!options?.status?.length && Array.isArray(options.status)) ? options.status as any : ["RELEASING", "FINISHED"],
        isAdult: options?.isAdult || undefined,
        countryOfOrigin: options?.countryOfOrigin || undefined,
    }, !!options?.name && isInView)

    // if (!isLoading && !data && !isInView) return <InvalidHomeItem item={item} />

    return (
        <PageWrapper className="space-y-0 px-4" ref={ref}>
            <h2>{options?.name || "Manga Carousel"}</h2>
            {(!isLoading && !data && isInView) ? <InvalidHomeItem item={item} /> : <Carousel
                className="w-full max-w-full"
                gap="xl"
                opts={{
                    align: "start",
                    dragFree: true,
                }}
                autoScroll
            >
                {/*<CarouselMasks />*/}
                <CarouselDotButtons className="-top-2" />
                <CarouselContent className="px-6">
                    {!!data ? data?.Page?.media?.filter(Boolean)?.sort((a, b) => b.meanScore! - a.meanScore!).map(media => {
                        return (
                            <MediaEntryCard
                                key={media.id}
                                media={media}
                                containerClassName="basis-[200px] md:basis-[250px] mx-2 mt-8 mb-0"
                                type="manga"
                            />
                        )
                    }) : [...Array(10).keys()].map((v, idx) => <MediaEntryCardSkeleton key={idx} />)}
                </CarouselContent>
            </Carousel>}
        </PageWrapper>
    )
}

function InvalidHomeItem(props: { item: Models_HomeItem }) {
    const { item } = props

    const schema = HOME_ITEMS[item.type]

    return (
        <PageWrapper className="rounded-xl bg-gray-900 border-2 border-dashed border-orange-400 p-4 !my-4">
            <p className="text-sm font-medium text-gray-400">
                Item "{schema?.name}" cannot be displayed because it is missing some required options.
            </p>
            {/* <pre>
             {JSON.stringify(item, null, 2)}
             </pre> */}
        </PageWrapper>
    )
}
