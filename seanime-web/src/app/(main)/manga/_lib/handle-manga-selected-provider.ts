import { ExtensionRepo_MangaProviderExtensionItem, Manga_MangaLatestChapterNumberItem, Nullish, Status } from "@/api/generated/types"
import { useListMangaProviderExtensions } from "@/api/hooks/extensions.hooks"
import { useServerStatus } from "@/app/(main)/_hooks/use-server-status"
import { withImmer } from "jotai-immer"
import { useAtom } from "jotai/react"
import { atomWithStorage } from "jotai/utils"
import { sortBy } from "lodash"
import React from "react"

/**
 * Stores the selected provider for each manga entry
 */
export const __manga_entryProviderAtom = atomWithStorage<Record<string, string>>("sea-manga-entry-provider", {}, undefined, { getOnInit: true })

// Key: "{mediaId}${providerId}"
// Value: { [filter]: string }
export type MangaEntryFilters = {
    scanlators: string[]
    language: string
}
export const __manga_entryFiltersAtom = atomWithStorage<Record<string, MangaEntryFilters>>("sea-manga-entry-filters",
    {},
    undefined,
    { getOnInit: true })

/**
 * Helper function to get the default provider from server status or available extensions
 */
const getDefaultMangaProvider = (
    serverStatus: Status | undefined,
    extensions: ExtensionRepo_MangaProviderExtensionItem[] | undefined,
) => {
    return serverStatus?.settings?.manga?.defaultMangaProvider || extensions?.[0]?.id || null
}

/**
 * Returns a record of all stored manga providers
 */
export function useStoredMangaProviders(_extensions: ExtensionRepo_MangaProviderExtensionItem[] | undefined) {
    const serverStatus = useServerStatus()

    const extensions = React.useMemo(() => {
        return _extensions?.toSorted((a, b) => a.name.localeCompare(b.name))
    }, [_extensions])

    const [storedProvider, setStoredProvider] = useAtom(__manga_entryProviderAtom)

    React.useLayoutEffect(() => {
        if (!extensions || !serverStatus) return
        const defaultProvider = getDefaultMangaProvider(serverStatus, extensions)

        // Remove invalid providers if there are no providers available
        if (!defaultProvider || extensions.length === 0) {
            setStoredProvider({})
            return
        }

        // Validate all stored providers and replace invalid ones with default
        const validatedProviders = { ...storedProvider }
        let hasChanges = false

        Object.entries(storedProvider).forEach(([mediaId, providerId]) => {
            const isProviderAvailable = extensions.some(provider => provider.id === providerId)
            if (!isProviderAvailable) {
                validatedProviders[mediaId] = defaultProvider
                hasChanges = true
            }
        })

        if (hasChanges) {
            setStoredProvider(validatedProviders)
        }
    }, [storedProvider, extensions, serverStatus])

    const setStoredProviderCallback = React.useCallback(({ mediaId, providerId }: { mediaId: string | number, providerId: string }) => {
        if (!mediaId) return
        setStoredProvider(prev => ({
            ...prev,
            [String(mediaId)]: providerId,
        }))
    }, [setStoredProvider])

    return {
        storedProviders: storedProvider,
        setStoredProvider: setStoredProviderCallback,
    }
}

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
        const defaultProvider = getDefaultMangaProvider(serverStatus, extensions)

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

    const selectedExtension = React.useMemo(() =>
        extensions?.find(provider => provider.id === storedProvider?.[String(mId)]),
        [extensions, storedProvider, mId]
    )

    const setSelectedProviderCallback = React.useCallback(({ mId, provider }: { mId: Nullish<string | number>, provider: string }) => {
        if (!mId) return
        setStoredProvider(prev => {
            return {
                ...prev,
                [String(mId)]: provider,
            }
        })
    }, [setStoredProvider])

    return {
        selectedExtension,
        selectedProvider: storedProvider?.[String(mId)] || null,
        setSelectedProvider: setSelectedProviderCallback,
    }
}

/**
 * This function takes in the manga id, the selected extension, the selected provider, the languages, the scanlators, and the isLoaded flag
 * It returns the stored filters for the manga entry
 * It also returns the functions to set the scanlators and the language
 */
export function useSelectedMangaFilters(
    mId: Nullish<string | number>,
    selectedExtension: Nullish<ExtensionRepo_MangaProviderExtensionItem>,
    selectedProvider: Nullish<string>,
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

    }, [isLoaded, selectedExtension])


    const selectedFilters = React.useMemo(() =>
        storedFilters[key] || { scanlators: [], language: "" },
        [storedFilters, key]
    )

    const setSelectedScanlatorCallback = React.useCallback(({ mId, scanlators }: { mId: Nullish<string | number>, scanlators: string[] }) => {
        if (!mId) return
        setStoredFilters(draft => {
            draft[key]["scanlators"] = scanlators
            return
        })
    }, [setStoredFilters, key])

    const setSelectedLanguageCallback = React.useCallback(({ mId, language }: { mId: Nullish<string | number>, language: string }) => {
        if (!mId) return
        setStoredFilters(draft => {
            draft[key]["language"] = language
            return
        })
    }, [setStoredFilters, key])

    return {
        selectedFilters,
        setSelectedScanlator: setSelectedScanlatorCallback,
        setSelectedLanguage: setSelectedLanguageCallback,
    }
}

export function useStoredMangaFilters(_extensions: ExtensionRepo_MangaProviderExtensionItem[] | undefined,
    selectedProviders: Record<string, string>,
) {
    const [_storedFilters] = useAtom(withImmer(__manga_entryFiltersAtom))
    const prevFiltersRef = React.useRef<Record<string, MangaEntryFilters>>({})

    const storedFilters = React.useMemo(() => {
        let filters: Record<string, MangaEntryFilters> = {}
        const entries = Object.entries(_storedFilters)

        for (const [key, value] of entries) {
            const [mangaId, providerId] = key.split("$")
            const mangaProvider = selectedProviders[mangaId]
            const extension = _extensions?.find(extension => extension.id === mangaProvider)

            if (extension?.settings?.supportsMultiScanlator || extension?.settings?.supportsMultiLanguage) {
                filters[mangaId] = {
                    scanlators: value.scanlators ?? [],
                    language: value.language ?? "",
                }
            }
        }

        // Deep comparison to avoid unnecessary re-renders
        const prevFilters = prevFiltersRef.current
        const filtersKeys = Object.keys(filters).sort()
        const prevFiltersKeys = Object.keys(prevFilters).sort()

        // Quick check: different number of keys
        if (filtersKeys.length !== prevFiltersKeys.length) {
            prevFiltersRef.current = filters
            return filters
        }

        // Quick check: different keys
        if (filtersKeys.join(',') !== prevFiltersKeys.join(',')) {
            prevFiltersRef.current = filters
            return filters
        }

        // Deep check: same keys but potentially different values
        for (const key of filtersKeys) {
            const current = filters[key]
            const previous = prevFilters[key]

            if (!previous ||
                current.language !== previous.language ||
                current.scanlators.length !== previous.scanlators.length ||
                !current.scanlators.every((s, i) => s === previous.scanlators[i])) {
                prevFiltersRef.current = filters
                return filters
            }
        }

        // No changes detected, return the previous reference
        return prevFilters
    }, [_storedFilters, _extensions, selectedProviders])

    return {
        storedFilters,
    }
}

export function getMangaEntryLatestChapterNumber(
    mangaId: string | number,
    latestChapterNumbers: Record<number, Manga_MangaLatestChapterNumberItem[]>,
    storedProviders: Record<string, string>,
    storedFilters: Record<string, MangaEntryFilters>,
) {
    const provider = storedProviders[String(mangaId)]
    const filters = storedFilters?.[String(mangaId)]

    if (!provider) return null

    const mangaLatestChapterNumbers = latestChapterNumbers[Number(mangaId)]?.filter(item => {
        return item.provider === provider
    })

    let found: Manga_MangaLatestChapterNumberItem | null | undefined = null

    // If filters are set for this manga
    if (!!filters) {
        // Find entry with matching scanlator & language
        found = mangaLatestChapterNumbers?.find(item => {
            return !!filters.scanlators[0] && !!filters.language &&
                filters.scanlators[0] === item.scanlator && filters.language === item.language
        })

        // If no entry with matching scanlator & language is found, find entry with matching language
        if (!found) {
            // Get all entries with matching language
            const entries = mangaLatestChapterNumbers?.filter(item => {
                return !!filters.language && filters.language === item.language
            }) ?? []

            // Get the highest chapter number from all entries with matching language
            found = sortBy(entries, "number").reverse()[0]
        }

        // If no entry with matching language is found, find entry with matching scanlator
        if (!found) {
            // Get all entries with matching scanlator
            const entries = mangaLatestChapterNumbers?.filter(item => {
                return !!filters.scanlators[0] && filters.scanlators[0] === item.scanlator
            }) ?? []

            // Get the highest chapter number from all entries with matching scanlator
            found = sortBy(entries, "number").reverse()[0]
        }
    }

    // If no filters are set or no entry is found for the filters, get the highest chapter number
    if (!found) {
        // Get the highest chapter number from any
        const highestChapterNumber = mangaLatestChapterNumbers?.reduce((max, item) => {
            return Math.max(max, item.number)
        }, 0)
        found = {
            provider: provider,
            language: "",
            scanlator: "",
            number: highestChapterNumber,
        }
    }

    return found?.number

}
