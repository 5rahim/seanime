import { useServerMutation, useServerQuery } from "@/api/client/requests"
import {
    AnimeEntryBulkAction_Variables,
    AnimeEntryManualMatch_Variables,
    FetchAnimeEntrySuggestions_Variables,
    OpenAnimeEntryInExplorer_Variables,
    ToggleAnimeEntrySilenceStatus_Variables,
    UpdateAnimeEntryProgress_Variables,
    UpdateAnimeEntryRepeat_Variables,
} from "@/api/generated/endpoint.types"
import { API_ENDPOINTS } from "@/api/generated/endpoints"
import { AL_BaseAnime, Anime_Entry, Anime_LocalFile, Anime_MissingEpisodes, Nullish } from "@/api/generated/types"
import { useQueryClient } from "@tanstack/react-query"
import { toast } from "sonner"

export function useGetAnimeEntry(id: Nullish<string | number>) {
    return useServerQuery<Anime_Entry>({
        endpoint: API_ENDPOINTS.ANIME_ENTRIES.GetAnimeEntry.endpoint.replace("{id}", String(id)),
        method: API_ENDPOINTS.ANIME_ENTRIES.GetAnimeEntry.methods[0],
        queryKey: [API_ENDPOINTS.ANIME_ENTRIES.GetAnimeEntry.key, String(id)],
        enabled: !!id,
    })
}

export function useAnimeEntryBulkAction(id?: Nullish<number>, onSuccess?: () => void) {
    const queryClient = useQueryClient()

    return useServerMutation<Array<Anime_LocalFile>, AnimeEntryBulkAction_Variables>({
        endpoint: API_ENDPOINTS.ANIME_ENTRIES.AnimeEntryBulkAction.endpoint,
        method: API_ENDPOINTS.ANIME_ENTRIES.AnimeEntryBulkAction.methods[0],
        mutationKey: [API_ENDPOINTS.ANIME_ENTRIES.AnimeEntryBulkAction.key, String(id)],
        onSuccess: async () => {
            await queryClient.invalidateQueries({ queryKey: [API_ENDPOINTS.ANIME_COLLECTION.GetLibraryCollection.key] })
            queryClient.invalidateQueries({ queryKey: [API_ENDPOINTS.ANIME_ENTRIES.GetAnimeEntry.key, String(id)] })
            queryClient.invalidateQueries({ queryKey: [API_ENDPOINTS.LIBRARY_EXPLORER.GetLibraryExplorerFileTree.key] })
            onSuccess?.()
        },
    })
}

export function useOpenAnimeEntryInExplorer() {
    return useServerMutation<boolean, OpenAnimeEntryInExplorer_Variables>({
        endpoint: API_ENDPOINTS.ANIME_ENTRIES.OpenAnimeEntryInExplorer.endpoint,
        method: API_ENDPOINTS.ANIME_ENTRIES.OpenAnimeEntryInExplorer.methods[0],
        mutationKey: [API_ENDPOINTS.ANIME_ENTRIES.OpenAnimeEntryInExplorer.key],
        onSuccess: async () => {

        },
    })
}

export function useFetchAnimeEntrySuggestions() {
    return useServerMutation<Array<AL_BaseAnime>, FetchAnimeEntrySuggestions_Variables>({
        endpoint: API_ENDPOINTS.ANIME_ENTRIES.FetchAnimeEntrySuggestions.endpoint,
        method: API_ENDPOINTS.ANIME_ENTRIES.FetchAnimeEntrySuggestions.methods[0],
        mutationKey: [API_ENDPOINTS.ANIME_ENTRIES.FetchAnimeEntrySuggestions.key],
        onSuccess: async () => {

        },
    })
}

export function useAnimeEntryManualMatch() {
    const queryClient = useQueryClient()

    return useServerMutation<Array<Anime_LocalFile>, AnimeEntryManualMatch_Variables>({
        endpoint: API_ENDPOINTS.ANIME_ENTRIES.AnimeEntryManualMatch.endpoint,
        method: API_ENDPOINTS.ANIME_ENTRIES.AnimeEntryManualMatch.methods[0],
        mutationKey: [API_ENDPOINTS.ANIME_ENTRIES.AnimeEntryManualMatch.key],
        onSuccess: async () => {
            await queryClient.invalidateQueries({ queryKey: [API_ENDPOINTS.ANIME_COLLECTION.GetLibraryCollection.key] })
            await queryClient.invalidateQueries({ queryKey: [API_ENDPOINTS.ANIME_ENTRIES.GetAnimeEntry.key] })
            queryClient.invalidateQueries({ queryKey: [API_ENDPOINTS.LIBRARY_EXPLORER.GetLibraryExplorerFileTree.key] })
            toast.success("Files matched")
        },
    })
}

export function useGetMissingEpisodes(enabled?: boolean) {
    return useServerQuery<Anime_MissingEpisodes>({
        endpoint: API_ENDPOINTS.ANIME_ENTRIES.GetMissingEpisodes.endpoint,
        method: API_ENDPOINTS.ANIME_ENTRIES.GetMissingEpisodes.methods[0],
        queryKey: [API_ENDPOINTS.ANIME_ENTRIES.GetMissingEpisodes.key],
        enabled: enabled ?? true, // Default to true if not provided
    })
}

export function useGetAnimeEntrySilenceStatus(id: Nullish<string | number>) {
    const { data, ...rest } = useServerQuery({
        endpoint: API_ENDPOINTS.ANIME_ENTRIES.GetAnimeEntrySilenceStatus.endpoint.replace("{id}", String(id)),
        method: API_ENDPOINTS.ANIME_ENTRIES.GetAnimeEntrySilenceStatus.methods[0],
        queryKey: [API_ENDPOINTS.ANIME_ENTRIES.GetAnimeEntrySilenceStatus.key],
        enabled: !!id,
    })

    return { isSilenced: !!data, ...rest }
}

export function useToggleAnimeEntrySilenceStatus() {
    const queryClient = useQueryClient()

    return useServerMutation<boolean, ToggleAnimeEntrySilenceStatus_Variables>({
        endpoint: API_ENDPOINTS.ANIME_ENTRIES.ToggleAnimeEntrySilenceStatus.endpoint,
        method: API_ENDPOINTS.ANIME_ENTRIES.ToggleAnimeEntrySilenceStatus.methods[0],
        mutationKey: [API_ENDPOINTS.ANIME_ENTRIES.ToggleAnimeEntrySilenceStatus.key],
        onSuccess: async () => {
            await queryClient.invalidateQueries({ queryKey: [API_ENDPOINTS.ANIME_ENTRIES.GetAnimeEntrySilenceStatus.key] })
            await queryClient.invalidateQueries({ queryKey: [API_ENDPOINTS.ANIME_ENTRIES.GetMissingEpisodes.key] })
        },
    })
}

export function useUpdateAnimeEntryProgress(id: Nullish<string | number>, episodeNumber: number, showToast: boolean = true) {
    const queryClient = useQueryClient()

    return useServerMutation<boolean, UpdateAnimeEntryProgress_Variables>({
        endpoint: API_ENDPOINTS.ANIME_ENTRIES.UpdateAnimeEntryProgress.endpoint,
        method: API_ENDPOINTS.ANIME_ENTRIES.UpdateAnimeEntryProgress.methods[0],
        mutationKey: [API_ENDPOINTS.ANIME_ENTRIES.UpdateAnimeEntryProgress.key, id, episodeNumber],
        onSuccess: async () => {
            await queryClient.invalidateQueries({ queryKey: [API_ENDPOINTS.ANILIST.GetAnimeCollection.key] })
            await queryClient.invalidateQueries({ queryKey: [API_ENDPOINTS.ANIME_COLLECTION.GetLibraryCollection.key] })
            if (id) {
                await queryClient.invalidateQueries({ queryKey: [API_ENDPOINTS.ANIME_ENTRIES.GetAnimeEntry.key, String(id)] })
            }
            if (showToast) {
                toast.success("Progress updated successfully")
            }
        },
    })
}

export function useUpdateAnimeEntryRepeat(id: Nullish<string | number>) {
    const queryClient = useQueryClient()

    return useServerMutation<boolean, UpdateAnimeEntryRepeat_Variables>({
        endpoint: API_ENDPOINTS.ANIME_ENTRIES.UpdateAnimeEntryRepeat.endpoint,
        method: API_ENDPOINTS.ANIME_ENTRIES.UpdateAnimeEntryRepeat.methods[0],
        mutationKey: [API_ENDPOINTS.ANIME_ENTRIES.UpdateAnimeEntryRepeat.key, id],
        onSuccess: async () => {
            // if (id) {
            //     await queryClient.invalidateQueries({ queryKey: [API_ENDPOINTS.ANIME_ENTRIES.GetAnimeEntry.key, String(id)] })
            // }
            // toast.success("Updated successfully")
        },
    })
}
