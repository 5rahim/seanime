import { Anime_LibraryCollectionList, Anime_MediaEntryEpisode } from "@/api/generated/types"
import { LibraryCollectionLists } from "@/app/(main)/(library)/_containers/library-collection"
import { __library_viewAtom } from "@/app/(main)/(library)/_lib/library-view.atoms"
import { IconButton } from "@/components/ui/button"
import { LoadingSpinner } from "@/components/ui/loading-spinner"
import { useThemeSettings } from "@/lib/theme/hooks"
import { useSetAtom } from "jotai"
import React from "react"
import { AiOutlineArrowLeft } from "react-icons/ai"

type LibraryViewProps = {
    collectionList: Anime_LibraryCollectionList[]
    continueWatchingList: Anime_MediaEntryEpisode[]
    isLoading: boolean
    hasScanned: boolean
}

export function DetailedLibraryView(props: LibraryViewProps) {

    const {
        collectionList,
        continueWatchingList,
        isLoading,
        hasScanned,
        ...rest
    } = props

    const ts = useThemeSettings()
    const setView = useSetAtom(__library_viewAtom)

    if (isLoading) return <LoadingSpinner />

    if (!hasScanned) return null

    return (
        <>
            <div className="flex p-4 gap-4 items-center relative w-full">
                <IconButton
                    icon={<AiOutlineArrowLeft />}
                    rounded
                    intent="white-outline"
                    size="sm"
                    onClick={() => setView("base")}
                />
                <h3 className="max-w-full lg:max-w-[50%] text-ellipsis truncate">Library</h3>
            </div>


            <LibraryCollectionLists
                collectionList={collectionList}
                isLoading={isLoading}
            />
        </>
    )
}
