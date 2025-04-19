import { Anime_Episode, Anime_LibraryCollectionList } from "@/api/generated/types"
import { ContinueWatching } from "@/app/(main)/(library)/_containers/continue-watching"
import { LibraryCollectionFilteredLists, LibraryCollectionLists } from "@/app/(main)/(library)/_containers/library-collection"
import { __mainLibrary_paramsAtom, __mainLibrary_paramsInputAtom } from "@/app/(main)/(library)/_lib/handle-library-collection"
import { MediaGenreSelector } from "@/app/(main)/_features/media/_components/media-genre-selector"
import { PageWrapper } from "@/components/shared/page-wrapper"
import { cn } from "@/components/ui/core/styling"
import { Skeleton } from "@/components/ui/skeleton"
import { useDebounce } from "@/hooks/use-debounce"
import { useThemeSettings } from "@/lib/theme/hooks"
import { AnimatePresence } from "framer-motion"
import { useSetAtom } from "jotai/index"
import { useAtom } from "jotai/react"
import React from "react"


type LibraryViewProps = {
    genres: string[]
    collectionList: Anime_LibraryCollectionList[]
    filteredCollectionList: Anime_LibraryCollectionList[]
    continueWatchingList: Anime_Episode[]
    isLoading: boolean
    hasEntries: boolean
}

export function LibraryView(props: LibraryViewProps) {

    const {
        genres,
        collectionList,
        continueWatchingList,
        filteredCollectionList,
        isLoading,
        hasEntries,
        ...rest
    } = props

    const ts = useThemeSettings()

    const [params, setParams] = useAtom(__mainLibrary_paramsAtom)

    if (isLoading) return <React.Fragment>
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

    return (
        <>
            <ContinueWatching
                episodes={continueWatchingList}
                isLoading={isLoading}
            />

            {(
                !ts.disableLibraryScreenGenreSelector &&
                collectionList.flatMap(n => n.entries)?.length > 2
            ) && <GenreSelector genres={genres} />}

            <PageWrapper key="library-collection-lists" className="p-4 space-y-8 relative z-[4]" data-library-collection-lists-container>
                <AnimatePresence mode="wait" initial={false}>
                    {!params.genre?.length ?
                        <LibraryCollectionLists
                            key="library-collection-lists"
                            collectionList={collectionList}
                            isLoading={isLoading}
                        />
                        : <LibraryCollectionFilteredLists
                            key="library-filtered-lists"
                            collectionList={filteredCollectionList}
                            isLoading={isLoading}
                        />
                    }
                </AnimatePresence>
            </PageWrapper>
        </>
    )
}

function GenreSelector({
    genres,
}: { genres: string[] }) {
    const [params, setParams] = useAtom(__mainLibrary_paramsInputAtom)
    const setActualParams = useSetAtom(__mainLibrary_paramsAtom)
    const debouncedParams = useDebounce(params, 200)

    React.useEffect(() => {
        setActualParams(params)
    }, [debouncedParams])

    if (!genres.length) return null

    return (
        <PageWrapper className="space-y-3 lg:space-y-6 relative z-[4]" data-library-genre-selector-container>
            <MediaGenreSelector
                items={[
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
        </PageWrapper>
    )
}
