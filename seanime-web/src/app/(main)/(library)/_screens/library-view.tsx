import { Anime_LibraryCollectionList, Anime_MediaEntryEpisode } from "@/api/generated/types"
import { ContinueWatching } from "@/app/(main)/(library)/_containers/continue-watching"
import { LibraryCollectionFilteredLists, LibraryCollectionLists } from "@/app/(main)/(library)/_containers/library-collection"
import { __scanner_modalIsOpen } from "@/app/(main)/(library)/_containers/scanner-modal"
import { __mainLibrary_paramsAtom, __mainLibrary_paramsInputAtom } from "@/app/(main)/(library)/_lib/handle-library-collection"
import { DiscoverPageHeader } from "@/app/(main)/discover/_components/discover-page-header"
import { DiscoverTrending } from "@/app/(main)/discover/_containers/discover-trending"
import { PageWrapper } from "@/components/shared/page-wrapper"
import { Button } from "@/components/ui/button"
import { cn } from "@/components/ui/core/styling"
import { HorizontalDraggableScroll } from "@/components/ui/horizontal-draggable-scroll"
import { Skeleton } from "@/components/ui/skeleton"
import { StaticTabs } from "@/components/ui/tabs"
import { useDebounce } from "@/hooks/use-debounce"
import { useThemeSettings } from "@/lib/theme/hooks"
import { useSetAtom } from "jotai/index"
import { useAtom } from "jotai/react"
import React from "react"
import { FiSearch } from "react-icons/fi"

type LibraryViewProps = {
    genres: string[]
    collectionList: Anime_LibraryCollectionList[]
    filteredCollectionList: Anime_LibraryCollectionList[]
    continueWatchingList: Anime_MediaEntryEpisode[]
    isLoading: boolean
    hasScanned: boolean
}

export function LibraryView(props: LibraryViewProps) {

    const {
        genres,
        collectionList,
        continueWatchingList,
        filteredCollectionList,
        isLoading,
        hasScanned,
        ...rest
    } = props

    const ts = useThemeSettings()
    const setScannerModalOpen = useSetAtom(__scanner_modalIsOpen)

    const [params, setParams] = useAtom(__mainLibrary_paramsAtom)

    if (isLoading) return <React.Fragment>
        <div className="p-4 space-y-4 relative z-[4]">
            <Skeleton className="h-12 w-full max-w-lg relative" />
            <div
                className={cn(
                    "grid h-[22rem] min-[2000px]:h-[24rem] grid-cols-2 md:grid-cols-3 lg:grid-cols-4 xl:grid-cols-5 2xl:grid-cols-7 min-[2000px]:grid-cols-8 gap-4",
                    // "md:grid-cols-3 lg:grid-cols-4 xl:grid-cols-5 2xl:grid-cols-7 min-[2000px]:grid-cols-8"
                )}
            >
                {[1, 2, 3, 4, 5, 6, 7, 8]?.map((_, idx) => {
                    return <Skeleton
                        key={idx} className={cn(
                        "h-[22rem] min-[2000px]:h-[24rem] col-span-1 aspect-[6/7] flex-none rounded-md relative overflow-hidden",
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

    if (!hasScanned && !isLoading) return (
        <>
            <DiscoverPageHeader />
            <PageWrapper className="p-4 sm:p-8 pt-0 space-y-8 relative z-[4]">
                <div className="text-center space-y-4">
                    <div className="w-fit mx-auto space-y-4">
                        <h2>Empty library</h2>
                        <Button
                            intent="warning-subtle"
                            leftIcon={<FiSearch />}
                            size="xl"
                            rounded
                            onClick={() => setScannerModalOpen(true)}
                        >
                            Scan your library
                        </Button>
                    </div>
                </div>
                <div>
                    <h3>Trending this season</h3>
                    <DiscoverTrending />
                </div>
            </PageWrapper>
        </>
    )

    return (
        <>
            <ContinueWatching
                episodes={continueWatchingList}
                isLoading={isLoading}
            />

            <PageWrapper className="space-y-3 lg:space-y-6 relative z-[4]">
                <GenreSelector genres={genres} />
            </PageWrapper>

            {!params.genre?.length ?
                <LibraryCollectionLists
                    collectionList={collectionList}
                    isLoading={isLoading}
                />
                : <LibraryCollectionFilteredLists
                    collectionList={filteredCollectionList}
                    isLoading={isLoading}
                />
            }
        </>
    )
}

function GenreSelector({
    genres,
}: { genres: string[] }) {
    const [params, setParams] = useAtom(__mainLibrary_paramsInputAtom)
    const setActualParams = useSetAtom(__mainLibrary_paramsAtom)
    const debouncedParams = useDebounce(params, 500)

    React.useEffect(() => {
        setActualParams(params)
    }, [debouncedParams])

    if (!genres.length) return null

    return (
        <HorizontalDraggableScroll className="scroll-pb-1 pt-4 flex">
            <div className="flex flex-1"></div>
            <StaticTabs
                className="px-2 overflow-visible gap-2 py-4 w-fit"
                triggerClass="text-base rounded-md ring-2 ring-transparent data-[current=true]:ring-brand-500 data-[current=true]:text-brand-300"
                items={[
                    // {
                    //     name: "All",
                    //     isCurrent: !params!.genre?.length,
                    //     onClick: () => setParams(draft => {
                    //         draft.genre = []
                    //         return
                    //     }),
                    // },
                    ...genres.map(genre => ({
                        name: genre,
                        isCurrent: params!.genre?.includes(genre) ?? false,
                        onClick: () => setParams(draft => {
                            if (draft.genre?.includes(genre)) {
                                draft.genre = draft.genre?.filter(g => g !== genre)
                            } else {
                                draft.genre = [...(draft.genre || []), genre]
                            }
                            return
                        }),
                    })),
                ]}
            />
            <div className="flex flex-1"></div>
        </HorizontalDraggableScroll>
    )
}
