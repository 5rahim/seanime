import { MangaChapterContainer, MangaCollection, MangaEntry, MangaPageContainer } from "@/app/(main)/manga/_lib/types"
import { SeaEndpoints } from "@/lib/server/endpoints"
import { useSeaQuery } from "@/lib/server/query"
import { useAtomValue } from "jotai/react"
import { atomWithStorage } from "jotai/utils"

const enum MangaProvider {
    COMICK = "comick",
    MANGASEE = "mangasee",
}

export const __manga_selectedProviderAtom = atomWithStorage<string>("sea-manga-provider", MangaProvider.COMICK)

export function useMangaCollection() {
    const { data, isLoading } = useSeaQuery<MangaCollection>({
        endpoint: SeaEndpoints.MANGA_COLLECTION,
        queryKey: ["get-manga-collection"],
    })

    return {
        mangaCollection: data,
        mangaCollectionLoading: isLoading,
    }
}

export function useMangaEntry(mediaId: string | undefined | null) {
    const { data, isLoading } = useSeaQuery<MangaEntry>({
        endpoint: SeaEndpoints.MANGA_ENTRY,
        queryKey: ["get-manga-entry", mediaId],
        enabled: !!mediaId,
    })

    return {
        mangaEntry: data,
        mangaEntryLoading: isLoading,
    }
}

export function useMangaChapterContainer(mediaId: string | undefined | null) {
    const provider = useAtomValue(__manga_selectedProviderAtom)

    const { data, isLoading } = useSeaQuery<MangaChapterContainer>({
        endpoint: SeaEndpoints.MANGA_CHAPTERS,
        queryKey: ["get-manga-chapters", mediaId, provider],
        enabled: !!mediaId,
    })

    return {
        chapterContainer: data,
        chapterContainerLoading: isLoading,
    }
}

export function useMangaPageContainer(mediaId: string | undefined | null, chapterId: string | undefined | null) {
    const provider = useAtomValue(__manga_selectedProviderAtom)

    const { data, isLoading } = useSeaQuery<MangaPageContainer>({
        endpoint: SeaEndpoints.MANGA_PAGES,
        queryKey: ["get-manga-pages", mediaId, provider, chapterId],
        enabled: !!mediaId && !!chapterId,
    })

    return {
        pageContainer: data,
        pageContainerLoading: isLoading,
    }
}
