import { useGetMangaEntryChapters } from "@/api/hooks/manga.hooks"
import { useHandleMangaProviderExtensions } from "@/app/(main)/manga/_lib/handle-manga-providers"
import { useSelectedMangaProvider } from "@/app/(main)/manga/_lib/handle-manga-selected-provider"

export function useHandleMangaChapters(
    mediaId: string | null,
) {

    /**
     * 1. Fetch the provider extensions
     */
    const { providerExtensions, providerOptions, providerExtensionsLoading } = useHandleMangaProviderExtensions(mediaId)

    /**
     * 2. Get the selected provider for this entry
     */
    const { selectedProvider, setSelectedProvider } = useSelectedMangaProvider(mediaId)


    /**
     * 3. Fetch the chapters for this entry
     */
    const {
        data: chapterContainer,
        isLoading: chapterContainerLoading,
        isError: chapterContainerError,
    } = useGetMangaEntryChapters({
        mediaId: Number(mediaId),
        provider: selectedProvider,
    })


    return {
        providerExtensions,
        providerExtensionsLoading,

        providerOptions, // For dropdown
        selectedProvider, // Current provider
        setSelectedProvider,

        chapterContainer,
        chapterContainerLoading,
        chapterContainerError,
    }
}
