"use client"
import { Manga_CollectionList } from "@/api/generated/types"
import { CustomLibraryBanner } from "@/app/(main)/(library)/_containers/custom-library-banner"
import { MediaCardGrid } from "@/app/(main)/_features/media/_components/media-card-grid"
import { MediaEntryCard } from "@/app/(main)/_features/media/_components/media-entry-card"
import { __mangaLibraryHeaderImageAtom, __mangaLibraryHeaderMangaAtom, LibraryHeader } from "@/app/(main)/manga/_components/library-header"
import { useMangaCollection } from "@/app/(main)/manga/_lib/handle-manga"
import { PageWrapper } from "@/components/shared/page-wrapper"
import { TextGenerateEffect } from "@/components/shared/text-generate-effect"
import { Skeleton } from "@/components/ui/skeleton"
import { getMangaCollectionTitle } from "@/lib/server/utils"
import { ThemeLibraryScreenBannerType, useThemeSettings } from "@/lib/theme/hooks"
import { useAtom, useAtomValue } from "jotai/react"
import React, { memo } from "react"

export const dynamic = "force-static"

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
            {ts.libraryScreenBannerType === ThemeLibraryScreenBannerType.Dynamic && (
                <>
                    <LibraryHeader manga={mangaCollection?.lists?.flatMap(l => l.entries)?.flatMap(e => e?.media)?.filter(Boolean) || []} />
                    <div className="h-10"></div>
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
                        },
                    }}
                >

                    <div className="space-y-8">
                        {mangaCollection.lists?.map(list => {
                            return <CollectionListItem key={list.type} list={list} />
                        })}
                    </div>

                </PageWrapper>
            </div>
        </div>
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

            <MediaCardGrid>
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
            </MediaCardGrid>
        </React.Fragment>
    )
})



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
