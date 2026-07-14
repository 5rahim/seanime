import { API_ENDPOINTS } from "@/api/generated/endpoints"
import { Manga_MangaPreferences } from "@/api/generated/types"
import { useGetMangaPreferences, useImportMangaPreferences } from "@/api/hooks/manga.hooks"
import { useWebsocketMessageListener } from "@/app/(main)/_hooks/handle-websockets"
import { WSEvents } from "@/lib/server/ws-events"
import { useQueryClient } from "@tanstack/react-query"
import { useAtom, useSetAtom } from "jotai/react"
import React from "react"
import { __manga_entryFiltersAtom, __manga_entryProviderAtom, __manga_preferencesHydratedAtom } from "./handle-manga-selected-provider"
import { fromMangaPreferences, toMangaPreferences } from "./manga-preferences"

// As of v3.10.0 source preferences are stored in the server
// This helps keep them in sync with the client
export function MangaPreferencesSync() {
    const queryClient = useQueryClient()
    const { data: serverPreferences } = useGetMangaPreferences()
    const { mutate: importPreferences, isPending: importPending } = useImportMangaPreferences()
    const [providers, setProviders] = useAtom(__manga_entryProviderAtom)
    const [filters, setFilters] = useAtom(__manga_entryFiltersAtom)
    const setHydrated = useSetAtom(__manga_preferencesHydratedAtom)
    const initialized = React.useRef(false)
    const legacyPreferences = React.useRef(toMangaPreferences(providers, filters))

    const applyPreferences = React.useCallback((pref: Manga_MangaPreferences) => {
        const unpacked = fromMangaPreferences(pref)
        setProviders(unpacked.providers)
        setFilters(unpacked.filters)
        setHydrated(true)
    }, [setFilters, setHydrated, setProviders])

    React.useEffect(() => {
        if (!serverPreferences) return

        if (!initialized.current) {
            initialized.current = true
            importPreferences(legacyPreferences.current, {
                onSuccess: pref => {
                    const canonical = pref ?? serverPreferences
                    queryClient.setQueryData([API_ENDPOINTS.MANGA.GetMangaPreferences.key], canonical)
                    applyPreferences(canonical)
                },
                onError: () => setHydrated(false),
            })
            return
        }

        if (!importPending) {
            applyPreferences(serverPreferences)
        }
    }, [applyPreferences, importPending, importPreferences, queryClient, serverPreferences, setHydrated])

    const handlePreferencesUpdated = React.useCallback(() => {
        queryClient.invalidateQueries({ queryKey: [API_ENDPOINTS.MANGA.GetMangaPreferences.key] })
    }, [queryClient])

    useWebsocketMessageListener({
        type: WSEvents.MANGA_PREFERENCES_UPDATED,
        onMessage: handlePreferencesUpdated,
    })

    return null
}
