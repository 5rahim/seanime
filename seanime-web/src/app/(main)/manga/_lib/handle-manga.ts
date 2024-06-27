import { Manga_Collection, Manga_Provider, Nullish } from "@/api/generated/types"
import { useGetMangaCollection } from "@/api/hooks/manga.hooks"
import { atom } from "jotai"
import { useAtom, useSetAtom } from "jotai/react"
import { atomWithStorage } from "jotai/utils"
import { useRouter } from "next/navigation"
import React, { useMemo } from "react"

/**
 * All available manga providers
 */
export const MANGA_PROVIDER_OPTIONS = [
    { value: "mangasee", label: "Mangasee" },
    { value: "mangadex", label: "Mangadex" },
    { value: "mangapill", label: "Mangapill" },
    { value: "manganato", label: "Manganato" },
    { value: "comick", label: "ComicK" },
]

const DEFAULT_MANGA_PROVIDER = "mangapill"

/**
 * Stores the selected provider for each manga entry
 */
export const __manga_entryProviderAtom = atomWithStorage<Record<string, Manga_Provider>>("sea-manga-entry-provider", {})

// Atom to retrieve the provider for a specific manga entry
export const __manga_getEntryProviderAtom = atom(get => get(__manga_entryProviderAtom), (get, set, mId: Nullish<string | number>): Manga_Provider => {
    if (!mId) return DEFAULT_MANGA_PROVIDER
    return get(__manga_entryProviderAtom)[String(mId)] || DEFAULT_MANGA_PROVIDER
})
// Atom to set the provider for a specific manga entry
export const __manga_setEntryProviderAtom = atom(null, (get, set, payload: { mId: Nullish<string | number>, provider: Manga_Provider }) => {
    if (!payload.mId) return
    set(__manga_entryProviderAtom, {
        ...get(__manga_entryProviderAtom),
        [String(payload.mId)]: payload.provider,
    })
})

/**
 * - Get the manga provider for a specific manga entry
 * - Set the manga provider for a specific manga entry
 */
export function useMangaProvider(mId: Nullish<string | number>) {
    const [_, getProvider] = useAtom(__manga_getEntryProviderAtom)
    const provider = React.useMemo(() => {
        return getProvider(mId)
    }, [mId, _])

    const setProvider = useSetAtom(__manga_setEntryProviderAtom)

    return {
        provider,
        setProvider,
    }
}

/**
 * Get the manga collection
 */
export function useMangaCollection() {
    const router = useRouter()
    const { data, isLoading, isError } = useGetMangaCollection()

    React.useEffect(() => {
        if (isError) {
            router.push("/")
        }
    }, [isError])

    const sortedCollection = useMemo(() => {
        if (!data || !data.lists) return data
        return {
            ...data,
            lists: [
                data.lists.find(n => n.type === "current"),
                data.lists.find(n => n.type === "paused"),
                data.lists.find(n => n.type === "planned"),
                // data.lists.find(n => n.type === "completed"), // DO NOT SHOW THIS LIST IN MANGA VIEW
                // data.lists.find(n => n.type === "dropped"), // DO NOT SHOW THIS LIST IN MANGA VIEW
            ].filter(Boolean),
        } as Manga_Collection
    }, [data])

    return {
        mangaCollection: sortedCollection,
        mangaCollectionLoading: isLoading,
    }
}

