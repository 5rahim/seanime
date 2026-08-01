import { useSeaQuery } from "@/api/client/requests"
import { API_ENDPOINTS } from "@/api/generated/endpoints"
import { ExtensionRepo_MangaProviderExtensionItem, Manga_MangaEntryPreference, Manga_MangaPreferences, Nullish, Status } from "@/api/generated/types"
import { useListMangaProviderExtensions } from "@/api/hooks/extensions.hooks"
import { useServerStatus } from "@/app/(main)/_hooks/use-server-status"
import { useQueryClient } from "@tanstack/react-query"
import { atom } from "jotai"
import { withImmer } from "jotai-immer"
import { useAtom, useAtomValue } from "jotai/react"
import { atomWithStorage } from "jotai/utils"
import React from "react"
import { toast } from "sonner"
import { getActiveMangaFilters, MangaEntryFilters } from "./manga-preferences"

export type { MangaEntryFilters } from "./manga-preferences"

/**
 * Stores the selected provider for each manga entry
 */
export const __manga_entryProviderAtom = atomWithStorage<Record<string, string>>("sea-manga-entry-provider", {}, undefined, { getOnInit: true })
export const __manga_preferencesHydratedAtom = atom(false)

type MangaPreferencePatch = {
    provider?: string
    filter?: {
        provider: string
        scanlators: string[]
        language: string
    }
}

function useSaveMangaPreference() {
    const { seaFetch } = useSeaQuery()
    const queryClient = useQueryClient()

    return React.useCallback(async (mediaId: string | number, patch: MangaPreferencePatch) => {
        try {
            const preference = await seaFetch<Manga_MangaEntryPreference>(
                API_ENDPOINTS.MANGA.PatchMangaPreference.endpoint.replace("{mediaId}", String(mediaId)),
                API_ENDPOINTS.MANGA.PatchMangaPreference.methods[0],
                patch,
            )
            if (preference) {
                queryClient.setQueryData<Manga_MangaPreferences>([API_ENDPOINTS.MANGA.GetMangaPreferences.key], current => ({
                    entries: {
                        ...(current?.entries ?? {}),
                        [Number(mediaId)]: preference,
                    },
                }))
            }
        }
        catch {
            toast.error("Could not save manga preference")
            await queryClient.invalidateQueries({ queryKey: [API_ENDPOINTS.MANGA.GetMangaPreferences.key] })
        }
    }, [queryClient, seaFetch])
}

// Key: "{mediaId}${providerId}"
// Value: { [filter]: string }
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
): string | null => {
    const firstExt = ((!!extensions?.length && extensions?.length > 1) ? extensions?.filter(n => n.id !== "local-manga")?.[0] : extensions?.[0])
    const defaultExt = !!serverStatus?.settings?.manga?.defaultMangaProvider
        ? extensions?.find(n => n.id === serverStatus?.settings?.manga?.defaultMangaProvider)
        : null
    return defaultExt?.id || firstExt?.id || null
}

/**
 * Returns a record of all stored manga providers
 */
export function useStoredMangaProviders(_extensions: ExtensionRepo_MangaProviderExtensionItem[] | undefined) {
    const serverStatus = useServerStatus()
    const hydrated = useAtomValue(__manga_preferencesHydratedAtom)
    const savePreference = useSaveMangaPreference()

    const extensions = React.useMemo(() => {
        return _extensions?.toSorted((a, b) => a.name.localeCompare(b.name))
    }, [_extensions])

    const [storedProvider, setStoredProvider] = useAtom(__manga_entryProviderAtom)

    React.useLayoutEffect(() => {
        if (!extensions || !serverStatus || !hydrated) return
        const defaultProvider = getDefaultMangaProvider(serverStatus, extensions)

        // Keep preferences when no provider is currently available
        if (!defaultProvider || extensions.length === 0) {
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
            for (const [mediaId, providerId] of Object.entries(validatedProviders)) {
                if (storedProvider[mediaId] === providerId) continue
                savePreference(mediaId, { provider: providerId })
            }
        }
    }, [storedProvider, extensions, serverStatus, hydrated, savePreference, setStoredProvider])

    return {
        storedProviders: storedProvider,
        setStoredProvider: ({ mediaId, providerId }: { mediaId: string | number, providerId: string }) => {
            if (!mediaId) return
            setStoredProvider(prev => ({
                ...prev,
                [String(mediaId)]: providerId,
            }))
            savePreference(mediaId, { provider: providerId })
        },
        overwriteStoredProvidersWith: (providerId: string) => {
            const ext = extensions?.find(p => p.id === providerId)
            if (!ext) return
            const next = { ...storedProvider }
            const changedMediaIds: string[] = []
            for (const [mediaId, currentProvider] of Object.entries(next)) {
                if (currentProvider === providerId) continue
                next[mediaId] = providerId
                changedMediaIds.push(mediaId)
            }
            setStoredProvider(next)
            for (const mediaId of changedMediaIds) {
                void savePreference(mediaId, { provider: providerId })
            }
        },
        overwriteStoredProviders: (rec: Record<string, string>) => {
            setStoredProvider(rec)
            for (const [mediaId, providerId] of Object.entries(rec)) {
                void savePreference(mediaId, { provider: providerId })
            }
        },
    }
}

/**
 * - Get the manga provider for a specific manga entry
 * - Set the manga provider for a specific manga entry
 */
export function useSelectedMangaProvider(mId: Nullish<string | number>) {
    const serverStatus = useServerStatus()
    const { data: _extensions } = useListMangaProviderExtensions()
    const hydrated = useAtomValue(__manga_preferencesHydratedAtom)
    const savePreference = useSaveMangaPreference()

    const extensions = React.useMemo(() => {
        return _extensions?.toSorted((a, b) => a.name.localeCompare(b.name))
    }, [_extensions])

    const [storedProvider, setStoredProvider] = useAtom(__manga_entryProviderAtom)

    React.useLayoutEffect(() => {
        if (!extensions || !serverStatus || !hydrated) return
        const defaultProvider = getDefaultMangaProvider(serverStatus, extensions)

        // Keep preferences when no provider is currently available
        if (!defaultProvider || extensions.length === 0) {
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
            if (mId) {
                void savePreference(mId, { provider: defaultProvider })
            }
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
                if (mId) {
                    void savePreference(mId, { provider: defaultProvider })
                }
            }
        }

    }, [mId, storedProvider, extensions, serverStatus, hydrated, savePreference, setStoredProvider])

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
            void savePreference(mId, { provider })
        },
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
    const hydrated = useAtomValue(__manga_preferencesHydratedAtom)
    const savePreference = useSaveMangaPreference()

    const key = `${String(mId)}$${selectedProvider}`

    React.useLayoutEffect(() => {
        if (!isLoaded || !hydrated) return

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

    }, [isLoaded, selectedExtension, selectedProvider, hydrated, key, setStoredFilters, storedFilters])


    return {
        selectedFilters: storedFilters[key] || { scanlators: [], language: "" },
        setSelectedScanlator: ({ mId, scanlators }: { mId: Nullish<string | number>, scanlators: string[] }) => {
            if (!mId || !selectedProvider) return
            const language = storedFilters[key]?.language ?? ""
            setStoredFilters(draft => {
                draft[key] ??= { scanlators: [], language: "" }
                draft[key]["scanlators"] = scanlators
                return
            })
            savePreference(mId, {
                filter: { provider: selectedProvider, scanlators, language },
            })
        },
        setSelectedLanguage: ({ mId, language }: { mId: Nullish<string | number>, language: string }) => {
            if (!mId || !selectedProvider) return
            const scanlators = storedFilters[key]?.scanlators ?? []
            setStoredFilters(draft => {
                draft[key] ??= { scanlators: [], language: "" }
                draft[key]["language"] = language
                return
            })
            savePreference(mId, {
                filter: { provider: selectedProvider, scanlators, language },
            })
        },
    }
}

export function useStoredMangaFilters(_extensions: ExtensionRepo_MangaProviderExtensionItem[] | undefined,
    selectedProviders: Record<string, string>,
) {
    const [_storedFilters] = useAtom(withImmer(__manga_entryFiltersAtom))

    const storedFilters = React.useMemo(() => {
        return getActiveMangaFilters(_storedFilters, selectedProviders, _extensions)
    }, [_storedFilters, _extensions, selectedProviders])

    return {
        storedFilters,
    }
}
