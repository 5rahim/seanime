import { useListMangaProviderExtensions } from "@/api/hooks/extensions.hooks"

export function useHandleMangaProviders() {

    const { data: providerExtensions } = useListMangaProviderExtensions()

    return {
        providers: providerExtensions ?? [],
        providerOptions: (providerExtensions ?? []).map(provider => ({
            label: provider.name,
            value: provider.id,
        })).sort((a, b) => a.label.localeCompare(b.label)),
    }

}
