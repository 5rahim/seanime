import { useServerMutation, useServerQuery } from "@/api/client/requests"
import {
    AnilistListManga_Variables,
    EmptyMangaEntryCache_Variables,
    GetAnilistMangaCollection_Variables,
    GetMangaEntryChapters_Variables,
    GetMangaEntryPages_Variables,
    GetMangaMapping_Variables,
    MangaManualMapping_Variables,
    MangaManualSearch_Variables,
    RefetchMangaChapterContainers_Variables,
    RemoveMangaMapping_Variables,
    UpdateMangaProgress_Variables,
} from "@/api/generated/endpoint.types"
import { API_ENDPOINTS } from "@/api/generated/endpoints"
import {
    AL_ListManga,
    AL_MangaCollection,
    AL_MangaDetailsById_Media,
    HibikeManga_SearchResult,
    Manga_ChapterContainer,
    Manga_Collection,
    Manga_Entry,
    Manga_MangaLatestChapterNumberItem,
    Manga_MappingResponse,
    Manga_PageContainer,
    Nullish,
} from "@/api/generated/types"
import { useQueryClient } from "@tanstack/react-query"
import { toast } from "sonner"

export function useGetAnilistMangaCollection() {
    return useServerQuery<AL_MangaCollection, GetAnilistMangaCollection_Variables>({
        endpoint: API_ENDPOINTS.MANGA.GetAnilistMangaCollection.endpoint,
        method: API_ENDPOINTS.MANGA.GetAnilistMangaCollection.methods[0],
        queryKey: [API_ENDPOINTS.MANGA.GetAnilistMangaCollection.key],
        enabled: true,
    })
}


export function useGetRawAnilistMangaCollection() {
    return useServerQuery<AL_MangaCollection, GetAnilistMangaCollection_Variables>({
        endpoint: API_ENDPOINTS.MANGA.GetRawAnilistMangaCollection.endpoint,
        method: API_ENDPOINTS.MANGA.GetRawAnilistMangaCollection.methods[0],
        queryKey: [API_ENDPOINTS.MANGA.GetRawAnilistMangaCollection.key],
        enabled: true,
    })
}

export function useGetMangaCollection() {
    return useServerQuery<Manga_Collection>({
        endpoint: API_ENDPOINTS.MANGA.GetMangaCollection.endpoint,
        method: API_ENDPOINTS.MANGA.GetMangaCollection.methods[0],
        queryKey: [API_ENDPOINTS.MANGA.GetMangaCollection.key],
        enabled: true,
    })
}

export function useGetMangaEntry(id: Nullish<string | number>) {
    return useServerQuery<Manga_Entry>({
        endpoint: API_ENDPOINTS.MANGA.GetMangaEntry.endpoint.replace("{id}", String(id)),
        method: API_ENDPOINTS.MANGA.GetMangaEntry.methods[0],
        queryKey: [API_ENDPOINTS.MANGA.GetMangaEntry.key, String(id)],
        enabled: !!id,
    })
}

export function useGetMangaEntryDetails(id: Nullish<string | number>) {
    return useServerQuery<AL_MangaDetailsById_Media>({
        endpoint: API_ENDPOINTS.MANGA.GetMangaEntryDetails.endpoint.replace("{id}", String(id)),
        method: API_ENDPOINTS.MANGA.GetMangaEntryDetails.methods[0],
        queryKey: [API_ENDPOINTS.MANGA.GetMangaEntryDetails.key, String(id)],
        enabled: !!id,
    })
}

export function useEmptyMangaEntryCache() {
    const queryClient = useQueryClient()

    return useServerMutation<boolean, EmptyMangaEntryCache_Variables>({
        endpoint: API_ENDPOINTS.MANGA.EmptyMangaEntryCache.endpoint,
        method: API_ENDPOINTS.MANGA.EmptyMangaEntryCache.methods[0],
        mutationKey: [API_ENDPOINTS.MANGA.EmptyMangaEntryCache.key],
        onSuccess: async () => {
            await queryClient.invalidateQueries({ queryKey: [API_ENDPOINTS.MANGA.GetMangaEntryChapters.key] })
            await queryClient.invalidateQueries({ queryKey: [API_ENDPOINTS.MANGA.GetMangaEntryPages.key] })
        },
    })
}

export function useGetMangaEntryChapters(variables: Partial<GetMangaEntryChapters_Variables>) {
    return useServerQuery<Manga_ChapterContainer, GetMangaEntryChapters_Variables>({
        endpoint: API_ENDPOINTS.MANGA.GetMangaEntryChapters.endpoint,
        method: API_ENDPOINTS.MANGA.GetMangaEntryChapters.methods[0],
        queryKey: [API_ENDPOINTS.MANGA.GetMangaEntryChapters.key, String(variables.mediaId), variables.provider],
        data: variables as GetMangaEntryChapters_Variables,
        enabled: !!variables.mediaId && !!variables.provider,
        muteError: true,
    })
}

export function useGetMangaEntryPages(variables: Partial<GetMangaEntryPages_Variables>) {
    return useServerQuery<Manga_PageContainer, GetMangaEntryPages_Variables>({
        endpoint: API_ENDPOINTS.MANGA.GetMangaEntryPages.endpoint,
        method: API_ENDPOINTS.MANGA.GetMangaEntryPages.methods[0],
        queryKey: [API_ENDPOINTS.MANGA.GetMangaEntryPages.key, String(variables.mediaId), variables.provider, variables.chapterId,
            variables.doublePage],
        data: variables as GetMangaEntryPages_Variables,
        enabled: !!variables.mediaId && !!variables.provider && !!variables.chapterId,
    })
}

export function useAnilistListManga(variables: AnilistListManga_Variables, enabled?: boolean) {
    return useServerQuery<AL_ListManga, AnilistListManga_Variables>({
        endpoint: API_ENDPOINTS.MANGA.AnilistListManga.endpoint,
        method: API_ENDPOINTS.MANGA.AnilistListManga.methods[0],
        queryKey: [API_ENDPOINTS.MANGA.AnilistListManga.key, variables],
        data: variables,
        enabled: enabled ?? true,
    })
}

export function useUpdateMangaProgress(id: Nullish<string | number>) {
    const queryClient = useQueryClient()

    return useServerMutation<boolean, UpdateMangaProgress_Variables>({
        endpoint: API_ENDPOINTS.MANGA.UpdateMangaProgress.endpoint,
        method: API_ENDPOINTS.MANGA.UpdateMangaProgress.methods[0],
        mutationKey: [API_ENDPOINTS.MANGA.UpdateMangaProgress.key],
        onSuccess: async () => {
            await queryClient.invalidateQueries({ queryKey: [API_ENDPOINTS.MANGA.GetMangaEntry.key, String(id)] })
            await queryClient.invalidateQueries({ queryKey: [API_ENDPOINTS.MANGA.GetMangaCollection.key] })
        },
    })
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

export function useMangaManualSearch(mediaId: Nullish<number>, provider: Nullish<string>) {
    return useServerMutation<Array<HibikeManga_SearchResult>, MangaManualSearch_Variables>({
        endpoint: API_ENDPOINTS.MANGA.MangaManualSearch.endpoint,
        method: API_ENDPOINTS.MANGA.MangaManualSearch.methods[0],
        mutationKey: [API_ENDPOINTS.MANGA.MangaManualSearch.key, String(mediaId), provider],
        gcTime: 0,
    })
}

export function useMangaManualMapping() {
    const queryClient = useQueryClient()

    return useServerMutation<boolean, MangaManualMapping_Variables>({
        endpoint: API_ENDPOINTS.MANGA.MangaManualMapping.endpoint,
        method: API_ENDPOINTS.MANGA.MangaManualMapping.methods[0],
        mutationKey: [API_ENDPOINTS.MANGA.MangaManualMapping.key],
        onSuccess: async () => {
            toast.success("Mapping added")
            await queryClient.invalidateQueries({ queryKey: [API_ENDPOINTS.MANGA.GetMangaEntryChapters.key] })
            await queryClient.invalidateQueries({ queryKey: [API_ENDPOINTS.MANGA.GetMangaEntryPages.key] })
            await queryClient.invalidateQueries({ queryKey: [API_ENDPOINTS.MANGA.GetMangaMapping.key] })
        },
    })
}

export function useGetMangaMapping(variables: Partial<GetMangaMapping_Variables>) {
    return useServerQuery<Manga_MappingResponse, GetMangaMapping_Variables>({
        endpoint: API_ENDPOINTS.MANGA.GetMangaMapping.endpoint,
        method: API_ENDPOINTS.MANGA.GetMangaMapping.methods[0],
        queryKey: [API_ENDPOINTS.MANGA.GetMangaMapping.key, String(variables.mediaId), variables.provider],
        data: variables as GetMangaMapping_Variables,
        enabled: !!variables.provider && !!variables.mediaId,
    })
}

export function useRemoveMangaMapping() {
    const queryClient = useQueryClient()

    return useServerMutation<boolean, RemoveMangaMapping_Variables>({
        endpoint: API_ENDPOINTS.MANGA.RemoveMangaMapping.endpoint,
        method: API_ENDPOINTS.MANGA.RemoveMangaMapping.methods[0],
        mutationKey: [API_ENDPOINTS.MANGA.RemoveMangaMapping.key],
        onSuccess: async () => {
            toast.info("Mapping removed")
            await queryClient.invalidateQueries({ queryKey: [API_ENDPOINTS.MANGA.GetMangaEntryChapters.key] })
            await queryClient.invalidateQueries({ queryKey: [API_ENDPOINTS.MANGA.GetMangaEntryPages.key] })
            await queryClient.invalidateQueries({ queryKey: [API_ENDPOINTS.MANGA.GetMangaMapping.key] })
        },
    })
}

export function useGetMangaEntryDownloadedChapters(mId: Nullish<string | number>) {
    return useServerQuery<Array<Manga_ChapterContainer>>({
        endpoint: API_ENDPOINTS.MANGA.GetMangaEntryDownloadedChapters.endpoint.replace("{id}", String(mId)),
        method: API_ENDPOINTS.MANGA.GetMangaEntryDownloadedChapters.methods[0],
        queryKey: [API_ENDPOINTS.MANGA.GetMangaEntryDownloadedChapters.key, String(mId)],
    })
}

export function useGetMangaLatestChapterNumbersMap() {
    return useServerQuery<Record<number, Array<Manga_MangaLatestChapterNumberItem>>>({
        endpoint: API_ENDPOINTS.MANGA.GetMangaLatestChapterNumbersMap.endpoint,
        method: API_ENDPOINTS.MANGA.GetMangaLatestChapterNumbersMap.methods[0],
        queryKey: [API_ENDPOINTS.MANGA.GetMangaLatestChapterNumbersMap.key],
        enabled: true,
    })
}

export function useRefetchMangaChapterContainers() {
    const queryClient = useQueryClient()

    return useServerMutation<boolean, RefetchMangaChapterContainers_Variables>({
        endpoint: API_ENDPOINTS.MANGA.RefetchMangaChapterContainers.endpoint,
        method: API_ENDPOINTS.MANGA.RefetchMangaChapterContainers.methods[0],
        mutationKey: [API_ENDPOINTS.MANGA.RefetchMangaChapterContainers.key],
        onSuccess: async () => {
            await queryClient.invalidateQueries({ queryKey: [API_ENDPOINTS.MANGA.GetMangaLatestChapterNumbersMap.key] })
            await queryClient.invalidateQueries({ queryKey: [API_ENDPOINTS.MANGA.GetMangaEntryChapters.key] })
            await queryClient.invalidateQueries({ queryKey: [API_ENDPOINTS.MANGA.GetMangaEntryPages.key] })
            toast.success("Sources refreshed")
        },
    })
}
