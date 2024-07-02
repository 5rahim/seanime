import { Manga_Collection, Manga_CollectionList } from "@/api/generated/types"
import { MediaCardLazyGrid } from "@/app/(main)/_features/media/_components/media-card-grid"
import { MediaEntryCard } from "@/app/(main)/_features/media/_components/media-entry-card"
import { MediaGenreSelector } from "@/app/(main)/_features/media/_components/media-genre-selector"
import { __mangaLibraryHeaderImageAtom, __mangaLibraryHeaderMangaAtom } from "@/app/(main)/manga/_components/library-header"
import { __mangaLibrary_paramsAtom, __mangaLibrary_paramsInputAtom } from "@/app/(main)/manga/_lib/handle-manga-collection"
import { PageWrapper } from "@/components/shared/page-wrapper"
import { TextGenerateEffect } from "@/components/shared/text-generate-effect"
import { IconButton } from "@/components/ui/button"
import { Disclosure, DisclosureContent, DisclosureItem, DisclosureTrigger } from "@/components/ui/disclosure"
import { Tooltip } from "@/components/ui/tooltip"
import { useDebounce } from "@/hooks/use-debounce"
import { getMangaCollectionTitle } from "@/lib/server/utils"
import { ThemeLibraryScreenBannerType, useThemeSettings } from "@/lib/theme/hooks"
import { AnimatePresence } from "framer-motion"
import { useSetAtom } from "jotai/index"
import { useAtom, useAtomValue } from "jotai/react"
import React, { memo } from "react"
import { PiBooksDuotone } from "react-icons/pi"

type MangaLibraryViewProps = {
    collection: Manga_Collection
    filteredCollection: Manga_Collection | undefined
    genres: string[]
}

export function MangaLibraryView(props: MangaLibraryViewProps) {

    const {
        collection,
        filteredCollection,
        genres,
        ...rest
    } = props

    const [params, setParams] = useAtom(__mangaLibrary_paramsAtom)

    return (
        <>
            <PageWrapper
                key="lists"
                className="relative 2xl:order-first pb-10 p-4"
            >
                {!!genres?.length && <div className="flex w-full">
                    <Disclosure type="single" collapsible className="w-full">
                        <DisclosureItem value="item-1" className="flex w-full flex-col gap-2">
                            <div className="w-full flex justify-end">
                                <Tooltip
                                    side="right"
                                    trigger={<DisclosureTrigger>
                                        <IconButton
                                            icon={<PiBooksDuotone />}
                                            intent="white-outline"
                                            rounded
                                        />
                                    </DisclosureTrigger>}
                                >Genres</Tooltip>
                            </div>
                            <DisclosureContent>
                                <div className="pb-4">
                                    <GenreSelector genres={genres} />
                                </div>
                            </DisclosureContent>
                        </DisclosureItem>
                    </Disclosure>
                </div>}

                <AnimatePresence mode="wait" initial={false}>
                    {!params.genre?.length ?
                        <CollectionLists key="lists" collectionList={collection} />
                        : <FilteredCollectionLists key="filtered-collection" collectionList={filteredCollection} />
                    }
                </AnimatePresence>
            </PageWrapper>
        </>
    )
}

export function CollectionLists({ collectionList }: {
    collectionList: Manga_Collection | undefined
}) {

    return (
        <PageWrapper
            className="p-4 space-y-8 relative z-[4]"
            {...{
                initial: { opacity: 0, y: 60 },
                animate: { opacity: 1, y: 0 },
                exit: { opacity: 0, scale: 0.99 },
                transition: {
                    duration: 0.35,
                },
            }}
        >
            {collectionList?.lists?.map(collection => {
                if (!collection.entries?.length) return null
                return <CollectionListItem key={collection.type} list={collection} />
            })}
        </PageWrapper>
    )

}

export function FilteredCollectionLists({ collectionList }: {
    collectionList: Manga_Collection | undefined
}) {

    const entries = React.useMemo(() => {
        return collectionList?.lists?.flatMap(n => n.entries).filter(Boolean) ?? []
    }, [collectionList])

    return (
        <PageWrapper
            className="p-4 space-y-8 relative z-[4]"
            {...{
                initial: { opacity: 0, y: 60 },
                animate: { opacity: 1, y: 0 },
                exit: { opacity: 0, scale: 0.99 },
                transition: {
                    duration: 0.35,
                },
            }}
        >
            <MediaCardLazyGrid itemCount={entries?.length || 0}>
                {entries.map(entry => {
                    return <div
                        key={entry.media?.id}
                    >
                        <MediaEntryCard
                            media={entry.media!}
                            listData={entry.listData}
                            showListDataButton
                            withAudienceScore={false}
                            type="manga"
                        />
                    </div>
                })}
            </MediaCardLazyGrid>
        </PageWrapper>
    )

}

const CollectionListItem = memo(({ list }: { list: Manga_CollectionList }) => {

    const ts = useThemeSettings()
    const [currentHeaderImage, setCurrentHeaderImage] = useAtom(__mangaLibraryHeaderImageAtom)
    const headerManga = useAtomValue(__mangaLibraryHeaderMangaAtom)

    React.useEffect(() => {
        if (list.type === "current") {
            if (currentHeaderImage === null && list.entries?.[0]?.media?.bannerImage) {
                setCurrentHeaderImage(list.entries?.[0]?.media?.bannerImage)
            }
        }
    }, [])

    return (
        <React.Fragment key={list.type}>
            <h2>{list.type === "current" ? "Continue reading" : getMangaCollectionTitle(list.type)}</h2>

            {(list.type === "current" && ts.libraryScreenBannerType === ThemeLibraryScreenBannerType.Dynamic && headerManga) &&
                <TextGenerateEffect
                    words={headerManga?.title?.userPreferred || ""}
                    className="w-full text-xl lg:text-5xl lg:max-w-[50%] h-[3.2rem] !mt-1 line-clamp-1 truncate text-ellipsis hidden lg:block pb-1"
                />
            }

            <MediaCardLazyGrid itemCount={list.entries?.length ?? 0}>
                {list.entries?.map(entry => {
                    return <div
                        key={entry.media?.id}
                        onMouseEnter={() => {
                            if (list.type === "current" && entry.media?.bannerImage) {
                                React.startTransition(() => {
                                    setCurrentHeaderImage(entry.media?.bannerImage!)
                                })
                            }
                        }}
                    >
                        <MediaEntryCard
                            media={entry.media!}
                            listData={entry.listData}
                            showListDataButton
                            withAudienceScore={false}
                            type="manga"
                        />
                    </div>
                })}
            </MediaCardLazyGrid>
        </React.Fragment>
    )
})

function GenreSelector({
    genres,
}: { genres: string[] }) {
    const [params, setParams] = useAtom(__mangaLibrary_paramsInputAtom)
    const setActualParams = useSetAtom(__mangaLibrary_paramsAtom)
    const debouncedParams = useDebounce(params, 200)

    React.useEffect(() => {
        setActualParams(params)
    }, [debouncedParams])

    if (!genres.length) return null

    return (
        <MediaGenreSelector
            // className="bg-gray-950 border p-0 rounded-xl mx-auto"
            staticTabsClass=""
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
    )
}
