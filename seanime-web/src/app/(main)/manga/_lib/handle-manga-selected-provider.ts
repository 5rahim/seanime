import { Nullish } from "@/api/generated/types"
import { useListMangaProviderExtensions } from "@/api/hooks/extensions.hooks"
import { useServerStatus } from "@/app/(main)/_hooks/use-server-status"
import { useAtom } from "jotai/react"
import { atomWithStorage } from "jotai/utils"
import React from "react"

/**
 * Stores the selected provider for each manga entry
 */
export const __manga_entryProviderAtom = atomWithStorage<Record<string, string>>("sea-manga-entry-provider", {})

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


        if (!defaultProvider || extensions.length === 0) {
            setStoredProvider(prev => {
                delete prev[String(mId)]
                return prev
            })
            return
        }

        // Set the default provider if it's not already set
        if (!storedProvider?.[String(mId)]) {
            setStoredProvider(prev => {
                return {
                    ...prev,
                    [String(mId)]: defaultProvider,
                }
            })
        } else {
            // Check if the selected provider is still available
            const isProviderAvailable = extensions.some(provider => provider.id === storedProvider?.[String(mId)])

            // Fall back to the first provider if the selected provider is not available
            if (!isProviderAvailable && extensions.length > 0) {
                setStoredProvider(prev => {
                    return {
                        ...prev,
                        [String(mId)]: defaultProvider,
                    }
                })
            }
        }

        if (!extensions.some(provider => provider.id === storedProvider?.[String(mId)])) {
            setStoredProvider(prev => {
                return {
                    ...prev,
                    [String(mId)]: defaultProvider,
                }
            })
        }

    }, [mId, storedProvider, extensions, serverStatus])

    return {
        selectedProvider: storedProvider?.[String(mId)] || null,
        setSelectedProvider: ({ mId, provider }: { mId: Nullish<string | number>, provider: string }) => {
            if (!mId) return
            setStoredProvider(prev => {
                return {
                    ...prev,
                    [String(mId)]: provider,
                }
            })
        },
    }
}
