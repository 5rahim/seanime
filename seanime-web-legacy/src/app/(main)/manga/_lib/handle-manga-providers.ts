import { useListMangaProviderExtensions } from "@/api/hooks/extensions.hooks"
import { __manga_entryProviderAtom } from "@/app/(main)/manga/_lib/handle-manga-selected-provider"
import { useAtom } from "jotai/react"
import React from "react"

export function useHandleMangaProviderExtensions(mId: string | null) {

    const { data: providerExtensions, isLoading: providersLoading } = useListMangaProviderExtensions()

    const [selectedProvider, setSelectedProvider] = useAtom(__manga_entryProviderAtom)


    React.useLayoutEffect(() => {
        if (!!providerExtensions?.length && !!selectedProvider && !!mId) {

            // Check if the selected provider is still available
            // The provider should default to "DEFAULT_MANGA_PROVIDER" if it's the first time loading the entry
            const isProviderAvailable = providerExtensions.some(provider => provider.id === selectedProvider[mId])

            // Fall back to the first provider if the selected provider is not available
            if (!isProviderAvailable && providerExtensions.length > 0) {
                setSelectedProvider({
                    ...selectedProvider,
                    [mId]: providerExtensions[0].id,
                })
            }
        }
    }, [providerExtensions, selectedProvider, mId])

    return {
        providerExtensions: providerExtensions,
        providerExtensionsLoading: providersLoading,
        providerOptions: (providerExtensions ?? []).map(provider => ({
            label: provider.name,
            value: provider.id,
        })).sort((a, b) => a.label.localeCompare(b.label)),
        // selectedProvider: !!mId ? selectedProvider[mId] : null,
    }

}
