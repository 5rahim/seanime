import { Manga_Collection } from "@/api/generated/types"
import { useGetMangaCollection } from "@/api/hooks/manga.hooks"
import { CollectionParams, DEFAULT_COLLECTION_PARAMS, filterCollectionEntries } from "@/lib/helpers/filtering"
import { atomWithImmer } from "jotai-immer"
import { useAtom } from "jotai/react"
import { useRouter } from "next/navigation"
import React, { useMemo } from "react"

export const MANGA_LIBRARY_DEFAULT_PARAMS: CollectionParams = {
    ...DEFAULT_COLLECTION_PARAMS,
    sorting: "PROGRESS_DESC",
}

export const __mangaLibrary_paramsAtom = atomWithImmer<CollectionParams>(MANGA_LIBRARY_DEFAULT_PARAMS)

export const __mangaLibrary_paramsInputAtom = atomWithImmer<CollectionParams>(MANGA_LIBRARY_DEFAULT_PARAMS)

/**
 * Get the manga collection
 */
export function useHandleMangaCollection() {
    const router = useRouter()
    const { data, isLoading, isError } = useGetMangaCollection()

    React.useEffect(() => {
        if (isError) {
            router.push("/")
        }
    }, [isError])

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

    const sortedCollection = useMemo(() => {
        if (!data || !data.lists) return data
        return {
            lists: [
                data.lists.find(n => n.type === "CURRENT"),
                data.lists.find(n => n.type === "PAUSED"),
                data.lists.find(n => n.type === "PLANNING"),
                // data.lists.find(n => n.type === "COMPLETED"), // DO NOT SHOW THIS LIST IN MANGA VIEW
                // data.lists.find(n => n.type === "DROPPED"), // DO NOT SHOW THIS LIST IN MANGA VIEW
            ].filter(Boolean),
        } as Manga_Collection
    }, [data])

    const filteredCollection = React.useMemo(() => {
        if (!data || !data.lists) return data

        let _lists = data.lists.map(obj => {
            if (!obj) return obj
            const arr = filterCollectionEntries(obj.entries, params, true)
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
