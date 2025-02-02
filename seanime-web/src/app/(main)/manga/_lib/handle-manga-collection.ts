import { Manga_Collection } from "@/api/generated/types"
import { useGetMangaChapterCountMap, useGetMangaCollection } from "@/api/hooks/manga.hooks"
import { CollectionParams, DEFAULT_COLLECTION_PARAMS, filterCollectionEntries, filterMangaCollectionEntries } from "@/lib/helpers/filtering"
import { useThemeSettings } from "@/lib/theme/hooks"
import { atomWithImmer } from "jotai-immer"
import { useAtom } from "jotai/react"
import { useRouter } from "next/navigation"
import React from "react"

export const MANGA_LIBRARY_DEFAULT_PARAMS: CollectionParams<"manga"> = {
    ...DEFAULT_COLLECTION_PARAMS,
    sorting: "TITLE",
    unreadOnly: false,
}

export const __mangaLibrary_paramsAtom = atomWithImmer<CollectionParams<"manga">>(MANGA_LIBRARY_DEFAULT_PARAMS)

export const __mangaLibrary_paramsInputAtom = atomWithImmer<CollectionParams<"manga">>(MANGA_LIBRARY_DEFAULT_PARAMS)

export const __mangaLibrary_chapterCountsAtom = atomWithImmer<Record<number, number>>({})

/**
 * Get the manga collection
 */
export function useHandleMangaCollection() {
    const router = useRouter()
    const { data, isLoading, isError } = useGetMangaCollection()

    const { data: chapterCounts } = useGetMangaChapterCountMap()

    const { mangaLibraryCollectionDefaultSorting } = useThemeSettings()

    React.useEffect(() => {
        if (isError) {
            router.push("/")
        }
    }, [isError])

    const [, setChapterCounts] = useAtom(__mangaLibrary_chapterCountsAtom)
    React.useEffect(() => {
        if (chapterCounts) {
            setChapterCounts(chapterCounts)
        }
    }, [chapterCounts])

    const [params, setParams] = useAtom(__mangaLibrary_paramsAtom)

    // Reset params when data changes
    React.useEffect(() => {
        if (!!data) {
            setParams(MANGA_LIBRARY_DEFAULT_PARAMS)
        }
    }, [data])

    const genres = React.useMemo(() => {
        const genresSet = new Set<string>()
        data?.lists?.forEach(l => {
            l.entries?.forEach(e => {
                e.media?.genres?.forEach(g => {
                    genresSet.add(g)
                })
            })
        })
        return Array.from(genresSet)?.sort((a, b) => a.localeCompare(b))
    }, [data])

    const sortedCollection = React.useMemo(() => {
        if (!data || !data.lists) return data

        let _lists = data.lists.map(obj => {
            if (!obj) return obj

            const newParams = { ...params, sorting: mangaLibraryCollectionDefaultSorting as any }
            let arr = filterMangaCollectionEntries(obj.entries, newParams, true, chapterCounts)

            // Reset `unreadOnly` if it's about to make the list disappear
            if (arr.length === 0 && newParams.unreadOnly) {
                const newParams = { ...params, unreadOnly: false, sorting: mangaLibraryCollectionDefaultSorting as any }
                arr = filterMangaCollectionEntries(obj.entries, newParams, true, chapterCounts)
            }

            return {
                type: obj.type,
                status: obj.status,
                entries: arr,
            }
        })

        return {
            lists: [
                _lists.find(n => n.type === "CURRENT"),
                _lists.find(n => n.type === "PAUSED"),
                _lists.find(n => n.type === "PLANNING"),
                // data.lists.find(n => n.type === "COMPLETED"), // DO NOT SHOW THIS LIST IN MANGA VIEW
                // data.lists.find(n => n.type === "DROPPED"), // DO NOT SHOW THIS LIST IN MANGA VIEW
            ].filter(Boolean),
        } as Manga_Collection
    }, [data, params, chapterCounts])

    const filteredCollection = React.useMemo(() => {
        if (!data || !data.lists) return data

        let _lists = data.lists.map(obj => {
            if (!obj) return obj

            const newParams = { ...params, sorting: mangaLibraryCollectionDefaultSorting as any }
            const arr = filterCollectionEntries("manga", obj.entries, newParams, true)
            return {
                type: obj.type,
                status: obj.status,
                entries: arr,
            }
        })
        return {
            lists: [
                _lists.find(n => n.type === "CURRENT"),
                _lists.find(n => n.type === "PAUSED"),
                _lists.find(n => n.type === "PLANNING"),
                // data.lists.find(n => n.type === "COMPLETED"), // DO NOT SHOW THIS LIST IN MANGA VIEW
                // data.lists.find(n => n.type === "DROPPED"), // DO NOT SHOW THIS LIST IN MANGA VIEW
            ].filter(Boolean),
        } as Manga_Collection
    }, [data, params])

    return {
        genres,
        mangaCollection: sortedCollection,
        filteredMangaCollection: filteredCollection,
        mangaCollectionLoading: isLoading,
    }
}
