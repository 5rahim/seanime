import { Anime_LibraryCollectionList, Anime_MediaEntryEpisode } from "@/api/generated/types"
import { LibraryHeader } from "@/app/(main)/(library)/_components/library-header"
import { ContinueWatching } from "@/app/(main)/(library)/_containers/continue-watching"
import { LibraryCollectionLists } from "@/app/(main)/(library)/_containers/library-collection"
import { ThemeLibraryScreenBannerType, useThemeSettings } from "@/lib/theme/hooks"
import React from "react"

type LibraryViewProps = {
    collectionList: Anime_LibraryCollectionList[]
    continueWatchingList: Anime_MediaEntryEpisode[]
    isLoading: boolean
    hasScanned: boolean
}

export function LibraryView(props: LibraryViewProps) {

    const {
        collectionList,
        continueWatchingList,
        isLoading,
        hasScanned,
        ...rest
    } = props

    const ts = useThemeSettings()

    return (
        <>
            {hasScanned && ts.libraryScreenBannerType === ThemeLibraryScreenBannerType.Dynamic && <LibraryHeader list={continueWatchingList} />}

            <ContinueWatching
                episodes={continueWatchingList}
                isLoading={isLoading}
            />
            <LibraryCollectionLists
                collectionList={collectionList}
                isLoading={isLoading}
            />
        </>
    )
}
