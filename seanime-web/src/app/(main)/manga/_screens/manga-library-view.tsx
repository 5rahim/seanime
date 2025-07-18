import { Manga_Collection, Manga_CollectionList } from "@/api/generated/types"
import { useRefetchMangaChapterContainers } from "@/api/hooks/manga.hooks"
import { MediaCardLazyGrid } from "@/app/(main)/_features/media/_components/media-card-grid"
import { MediaEntryCard } from "@/app/(main)/_features/media/_components/media-entry-card"
import { MediaGenreSelector } from "@/app/(main)/_features/media/_components/media-genre-selector"
import { SeaCommandInjectableItem, useSeaCommandInject } from "@/app/(main)/_features/sea-command/use-inject"
import { seaCommand_compareMediaTitles } from "@/app/(main)/_features/sea-command/utils"
import { __mangaLibraryHeaderImageAtom, __mangaLibraryHeaderMangaAtom } from "@/app/(main)/manga/_components/library-header"
import { __mangaLibrary_paramsAtom, __mangaLibrary_paramsInputAtom } from "@/app/(main)/manga/_lib/handle-manga-collection"
import { LuffyError } from "@/components/shared/luffy-error"
import { PageWrapper } from "@/components/shared/page-wrapper"
import { TextGenerateEffect } from "@/components/shared/text-generate-effect"
import { Button, IconButton } from "@/components/ui/button"
import { DropdownMenu, DropdownMenuItem } from "@/components/ui/dropdown-menu"
import { useDebounce } from "@/hooks/use-debounce"
import { getMangaCollectionTitle } from "@/lib/server/utils"
import { ThemeLibraryScreenBannerType, useThemeSettings } from "@/lib/theme/hooks"
import { useSetAtom } from "jotai/index"
import { useAtom, useAtomValue } from "jotai/react"
import { AnimatePresence } from "motion/react"
import Link from "next/link"
import { useRouter } from "next/navigation"
import React, { memo } from "react"
import { BiDotsVertical } from "react-icons/bi"
import { LuBookOpenCheck, LuRefreshCcw } from "react-icons/lu"
import { toast } from "sonner"
import { CommandItemMedia } from "../../_features/sea-command/_components/command-utils"

type MangaLibraryViewProps = {
    collection: Manga_Collection
    filteredCollection: Manga_Collection | undefined
    genres: string[]
    storedProviders: Record<string, string>
    hasManga: boolean
}

export function MangaLibraryView(props: MangaLibraryViewProps) {

    const {
        collection,
        filteredCollection,
        genres,
        storedProviders,
        hasManga,
        ...rest
    } = props

    const [params, setParams] = useAtom(__mangaLibrary_paramsAtom)

    return (
        <>
            <PageWrapper
                key="lists"
                className="relative 2xl:order-first pb-10 p-4"
                data-manga-library-view-container
            >
                <div className="w-full flex justify-end">
                </div>

                <AnimatePresence mode="wait" initial={false}>

                    {!!collection && !hasManga && <LuffyError
                        title="No manga found"
                    >
                        <div className="space-y-2">
                            <p>
                                No manga has been added to your library yet.
                            </p>

                            <div className="!mt-4">
                                <Link href="/discover?type=manga">
                                    <Button intent="white-outline" rounded>
                                        Browse manga
                                    </Button>
                                </Link>
                            </div>
                        </div>
                    </LuffyError>}

                    {!params.genre?.length ?
                        <CollectionLists key="lists" collectionList={collection} genres={genres} storedProviders={storedProviders} />
                        : <FilteredCollectionLists key="filtered-collection" collectionList={filteredCollection} genres={genres} />
                    }
                </AnimatePresence>
            </PageWrapper>
        </>
    )
}

export function CollectionLists({ collectionList, genres, storedProviders }: {
    collectionList: Manga_Collection | undefined
    genres: string[]
    storedProviders: Record<string, string>
}) {

    return (
        <PageWrapper
            className="p-4 space-y-8 relative z-[4]"
            data-manga-library-view-collection-lists-container
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
                return (
                    <React.Fragment key={collection.type}>
                        <CollectionListItem list={collection} storedProviders={storedProviders} />

                        {(collection.type === "CURRENT" && !!genres?.length) && <GenreSelector genres={genres} />}
                    </React.Fragment>
                )
            })}
        </PageWrapper>
    )

}

export function FilteredCollectionLists({ collectionList, genres }: {
    collectionList: Manga_Collection | undefined
    genres: string[]
}) {

    const entries = React.useMemo(() => {
        return collectionList?.lists?.flatMap(n => n.entries).filter(Boolean) ?? []
    }, [collectionList])

    return (
        <PageWrapper
            className="p-4 space-y-8 relative z-[4]"
            data-manga-library-view-filtered-collection-lists-container
            {...{
                initial: { opacity: 0, y: 60 },
                animate: { opacity: 1, y: 0 },
                exit: { opacity: 0, scale: 0.99 },
                transition: {
                    duration: 0.35,
                },
            }}
        >

            {!!genres?.length && <div className="mt-24">
                <GenreSelector genres={genres} />
            </div>}

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

const CollectionListItem = memo(({ list, storedProviders }: { list: Manga_CollectionList, storedProviders: Record<string, string> }) => {

    const ts = useThemeSettings()
    const [currentHeaderImage, setCurrentHeaderImage] = useAtom(__mangaLibraryHeaderImageAtom)
    const headerManga = useAtomValue(__mangaLibraryHeaderMangaAtom)
    const [params, setParams] = useAtom(__mangaLibrary_paramsAtom)
    const router = useRouter()

    const { mutate: refetchMangaChapterContainers, isPending: isRefetchingMangaChapterContainers } = useRefetchMangaChapterContainers()

    const { inject, remove } = useSeaCommandInject()

    React.useEffect(() => {
        if (list.type === "CURRENT") {
            if (currentHeaderImage === null && list.entries?.[0]?.media?.bannerImage) {
                setCurrentHeaderImage(list.entries?.[0]?.media?.bannerImage)
            }
        }
    }, [])

    // Inject command for currently reading manga
    React.useEffect(() => {
        if (list.type === "CURRENT" && list.entries?.length) {
            inject("currently-reading-manga", {
                items: list.entries.map(entry => ({
                    data: entry,
                    id: `manga-${entry.mediaId}`,
                    value: entry.media?.title?.userPreferred || "",
                    heading: "Currently Reading",
                    priority: 100,
                    render: () => (
                        <CommandItemMedia media={entry.media!} />
                    ),
                    onSelect: () => {
                        router.push(`/manga/entry?id=${entry.mediaId}`)
                    },
                })),
                filter: ({ item, input }: { item: SeaCommandInjectableItem, input: string }) => {
                    if (!input) return true
                    return seaCommand_compareMediaTitles((item.data as typeof list.entries[0])?.media?.title, input)
                },
                priority: 100,
            })
        }

        return () => remove("currently-reading-manga")
    }, [list.entries])

    return (
        <React.Fragment>

            <div className="flex gap-3 items-center" data-manga-library-view-collection-list-item-header-container>
                <h2 data-manga-library-view-collection-list-item-header-title>{list.type === "CURRENT" ? "Continue reading" : getMangaCollectionTitle(
                    list.type)}</h2>
                <div className="flex flex-1" data-manga-library-view-collection-list-item-header-spacer></div>

                {list.type === "CURRENT" && params.unreadOnly && (
                    <Button
                        intent="white-link"
                        size="xs"
                        className="!px-2 !py-1"
                        onClick={() => {
                            setParams(draft => {
                                draft.unreadOnly = false
                                return
                            })
                        }}
                    >
                        Show all
                    </Button>
                )}

                {list.type === "CURRENT" && <DropdownMenu
                    trigger={<div className="relative">
                        <IconButton
                            intent="white-basic"
                            size="xs"
                            className="mt-1"
                            icon={<BiDotsVertical />}
                            // loading={isRefetchingMangaChapterContainers}
                        />
                        {/*{params.unreadOnly && <div className="absolute -top-1 -right-1 bg-[--blue] size-2 rounded-full"></div>}*/}
                        {isRefetchingMangaChapterContainers &&
                            <div className="absolute -top-1 -right-1 bg-[--orange] size-3 rounded-full animate-ping"></div>}
                    </div>}
                >
                    <DropdownMenuItem
                        onClick={() => {
                            if (isRefetchingMangaChapterContainers) return

                            toast.info("Refetching from sources...")
                            refetchMangaChapterContainers({
                                selectedProviderMap: storedProviders,
                            })
                        }}
                    >
                        <LuRefreshCcw /> {isRefetchingMangaChapterContainers ? "Refetching..." : "Refresh sources"}
                    </DropdownMenuItem>
                    <DropdownMenuItem
                        onClick={() => {
                            setParams(draft => {
                                draft.unreadOnly = !draft.unreadOnly
                                return
                            })
                        }}
                    >
                        <LuBookOpenCheck /> {params.unreadOnly ? "Show all" : "Unread chapters only"}
                    </DropdownMenuItem>
                </DropdownMenu>}

            </div>

            {(list.type === "CURRENT" && ts.libraryScreenBannerType === ThemeLibraryScreenBannerType.Dynamic && headerManga) &&
                <TextGenerateEffect
                    data-manga-library-view-collection-list-item-header-media-title
                    words={headerManga?.title?.userPreferred || ""}
                    className="w-full text-xl lg:text-5xl lg:max-w-[50%] h-[3.2rem] !mt-1 line-clamp-1 truncate text-ellipsis hidden lg:block pb-1"
                />
            }

            <MediaCardLazyGrid itemCount={list.entries?.length ?? 0}>
                {list.entries?.map(entry => {
                    return <div
                        key={entry.media?.id}
                        onMouseEnter={() => {
                            if (list.type === "CURRENT" && entry.media?.bannerImage) {
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
