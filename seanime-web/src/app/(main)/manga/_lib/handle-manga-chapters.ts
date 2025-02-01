import { useGetMangaEntryChapters } from "@/api/hooks/manga.hooks"
import { useHandleMangaProviderExtensions } from "@/app/(main)/manga/_lib/handle-manga-providers"
import { useSelectedMangaFilters, useSelectedMangaProvider } from "@/app/(main)/manga/_lib/handle-manga-selected-provider"
import { LANGUAGES_LIST } from "@/app/(main)/manga/_lib/language-map"
import { uniq } from "lodash"
import React from "react"

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
    const {
        selectedExtension,
        selectedProvider,
        setSelectedProvider,
    } = useSelectedMangaProvider(mediaId)


    /**
     * 3. Fetch the chapters for this entry
     */
    const {
        data: chapterContainer,
        isLoading: chapterContainerLoading,
        isError: chapterContainerError,
    } = useGetMangaEntryChapters({
        mediaId: Number(mediaId),
        provider: selectedProvider || undefined,
    })

    const languageOptions = React.useMemo(() => {
        if (!selectedExtension) return []
        if (!selectedExtension.settings?.supportsMultiLanguage) return []

        const languages = uniq(chapterContainer?.chapters?.map(chapter => chapter.language)?.filter(Boolean) || [])
        return languages.map(lang => ({ value: lang, label: ((LANGUAGES_LIST as any)[lang as any] as any)?.nativeName || lang }))
    }, [selectedExtension, chapterContainer])

    const scanlatorOptions = React.useMemo(() => {
        if (!selectedExtension) return []
        if (!selectedExtension.settings?.supportsMultiScanlator) return []

        const scanlators = uniq(chapterContainer?.chapters?.map(chapter => chapter.scanlator)?.filter(Boolean) || [])
        return scanlators.map(scanlator => ({ value: scanlator, label: scanlator }))
    }, [selectedExtension, chapterContainer])


    /**
     * 4. Filters
     */
    const { setSelectedScanlator, setSelectedLanguage, selectedFilters } = useSelectedMangaFilters(
        mediaId,
        selectedExtension,
        selectedProvider,
        languageOptions.map(n => n.value),
        scanlatorOptions.map(n => n.value),
        !chapterContainerLoading,
    )

    /**
     * 5. Filter chapters based on language and scanlator
     */
    const filteredChapterContainer = React.useMemo(() => {
        if (!chapterContainer) return chapterContainer

        const filteredChapters = chapterContainer.chapters?.filter(ch => {
            if (selectedExtension?.settings?.supportsMultiLanguage && selectedFilters.language) {
                if (ch.language !== selectedFilters.language) return false
            }
            if (selectedExtension?.settings?.supportsMultiScanlator && selectedFilters.scanlators[0]) {
                if (ch.scanlator !== selectedFilters.scanlators[0]) return false
            }
            return true
        })

        return {
            ...chapterContainer,
            chapters: filteredChapters,
        }
    }, [chapterContainer, selectedExtension, selectedFilters])

    return {
        selectedExtension,
        providerExtensions,
        providerExtensionsLoading,
        // Selected provider
        providerOptions, // For dropdown
        selectedProvider, // Current provider
        setSelectedProvider,
        // Filters
        selectedFilters,
        setSelectedLanguage,
        setSelectedScanlator,
        languageOptions,
        scanlatorOptions,
        // Chapters
        chapterContainer: filteredChapterContainer,
        chapterContainerLoading,
        chapterContainerError,
    }
}
