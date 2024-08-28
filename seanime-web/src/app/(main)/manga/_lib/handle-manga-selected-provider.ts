import { ExtensionRepo_MangaProviderExtensionItem, Nullish } from "@/api/generated/types"
import { useListMangaProviderExtensions } from "@/api/hooks/extensions.hooks"
import { useServerStatus } from "@/app/(main)/_hooks/use-server-status"
import { withImmer } from "jotai-immer"
import { useAtom } from "jotai/react"
import { atomWithStorage } from "jotai/utils"
import React from "react"

/**
 * Stores the selected provider for each manga entry
 */
export const __manga_entryProviderAtom = atomWithStorage<Record<string, string>>("sea-manga-entry-provider", {}, undefined, { getOnInit: true })

// Key: "{mediaId}${providerId}"
// Value: { [filter]: string }
type MangaEntryFilters = {
    scanlators: string[]
    language: string
}
export const __manga_entryFiltersAtom = atomWithStorage<Record<string, MangaEntryFilters>>("sea-manga-entry-filters",
    {},
    undefined,
    { getOnInit: true })

/**
 * - Get the manga provider for a specific manga entry
 * - Set the manga provider for a specific manga entry
 */
export function useSelectedMangaProvider(mId: Nullish<string | number>) {
    const serverStatus = useServerStatus()
    const { data: _extensions } = useListMangaProviderExtensions()

    const extensions = React.useMemo(() => {
        return _extensions?.toSorted((a, b) => a.name.localeCompare(b.name))
    }, [_extensions])

    const [storedProvider, setStoredProvider] = useAtom(__manga_entryProviderAtom)

    React.useLayoutEffect(() => {
        if (!extensions || !serverStatus) return
        const defaultProvider = serverStatus?.settings?.manga?.defaultMangaProvider || extensions[0]?.id || null

        // Remove the stored provider if there are no providers available
        if (!defaultProvider || extensions.length === 0) {
            setStoredProvider(prev => {
                delete prev[String(mId)]
                return prev
            })
            return
        }

        // (Case 1) No provider has been chosen yet for this manga
        // -> Set the default provider & filters
        if (!storedProvider?.[String(mId)]) {
            setStoredProvider(prev => {
                return {
                    ...prev,
                    [String(mId)]: defaultProvider,
                }
            })
        } else {
            // (Case 2) There is a selected provider for this manga, but it's not available anymore in the extensions
            const isProviderAvailable = extensions.some(provider => provider.id === storedProvider?.[String(mId)])
            // -> Fall back to the default provider
            if (!isProviderAvailable && extensions.length > 0) {
                setStoredProvider(prev => {
                    return {
                        ...prev,
                        [String(mId)]: defaultProvider,
                    }
                })
            }
        }

    }, [mId, storedProvider, extensions, serverStatus])

    return {
        selectedExtension: extensions?.find(provider => provider.id === storedProvider?.[String(mId)]),
        selectedProvider: storedProvider?.[String(mId)] || null,
        setSelectedProvider: ({ mId, provider }: { mId: Nullish<string | number>, provider: string }) => {
            if (!mId) return
            setStoredProvider(prev => {
                return {
                    ...prev,
                    [String(mId)]: provider,
                }
            })
        }
    }
}

export function useSelectedMangaFilters(
    mId: Nullish<string | number>,
    selectedExtension: Nullish<ExtensionRepo_MangaProviderExtensionItem>,
    selectedProvider: Nullish<string>,
    languages: string[],
    scanlators: string[],
    isLoaded: boolean,
) {

    const [storedFilters, setStoredFilters] = useAtom(withImmer(__manga_entryFiltersAtom))

    const key = `${String(mId)}$${selectedProvider}`

    React.useLayoutEffect(() => {
        if (!isLoaded) return

        const defaultFilters: MangaEntryFilters = {
            scanlators: [],
            language: "",
        }

        if (!selectedProvider) {
            setStoredFilters(draft => {
                delete draft[key]
                return draft
            })
            return
        }

        // (Case 1) No filters have been chosen yet for this manga
        // -> Set the default filters
        if (!storedFilters[key] && (selectedExtension?.settings?.supportsMultiScanlator || selectedExtension?.settings?.supportsMultiLanguage)) {
            setStoredFilters(draft => {
                draft[key] = defaultFilters
                return
            })
        }

    }, [isLoaded, languages, scanlators, selectedExtension])


    return {
        selectedFilters: storedFilters[key] || { scanlators: [], language: "" },
        setSelectedScanlator: ({ mId, scanlators }: { mId: Nullish<string | number>, scanlators: string[] }) => {
            if (!mId) return
            setStoredFilters(draft => {
                draft[key]["scanlators"] = scanlators
                return
            })
        },
        setSelectedLanguage: ({ mId, language }: { mId: Nullish<string | number>, language: string }) => {
            if (!mId) return
            setStoredFilters(draft => {
                draft[key]["language"] = language
                return
            })
        },
    }
}
