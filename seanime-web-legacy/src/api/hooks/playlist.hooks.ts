import { useServerMutation, useServerQuery } from "@/api/client/requests"
import { CreatePlaylist_Variables, DeletePlaylist_Variables, UpdatePlaylist_Variables } from "@/api/generated/endpoint.types"
import { API_ENDPOINTS } from "@/api/generated/endpoints"
import { Anime_Playlist, Anime_PlaylistEpisode } from "@/api/generated/types"
import { Nullish } from "@/types/common"
import { useQueryClient } from "@tanstack/react-query"
import { toast } from "sonner"

export function useCreatePlaylist() {
    const queryClient = useQueryClient()

    return useServerMutation<Anime_Playlist, CreatePlaylist_Variables>({
        endpoint: API_ENDPOINTS.PLAYLIST.CreatePlaylist.endpoint,
        method: API_ENDPOINTS.PLAYLIST.CreatePlaylist.methods[0],
        mutationKey: [API_ENDPOINTS.PLAYLIST.CreatePlaylist.key],
        onSuccess: async () => {
            await queryClient.invalidateQueries({ queryKey: [API_ENDPOINTS.PLAYLIST.GetPlaylists.key] })
            toast.success("Playlist created")
        },
    })
}

export function useGetPlaylists() {
    return useServerQuery<Array<Anime_Playlist>>({
        endpoint: API_ENDPOINTS.PLAYLIST.GetPlaylists.endpoint,
        method: API_ENDPOINTS.PLAYLIST.GetPlaylists.methods[0],
        queryKey: [API_ENDPOINTS.PLAYLIST.GetPlaylists.key],
        enabled: true,
    })
}

export function useUpdatePlaylist() {
    const queryClient = useQueryClient()

    return useServerMutation<Anime_Playlist, UpdatePlaylist_Variables>({
        endpoint: API_ENDPOINTS.PLAYLIST.UpdatePlaylist.endpoint,
        method: API_ENDPOINTS.PLAYLIST.UpdatePlaylist.methods[0],
        mutationKey: [API_ENDPOINTS.PLAYLIST.UpdatePlaylist.key],
        onSuccess: async () => {
            await queryClient.refetchQueries({ queryKey: [API_ENDPOINTS.PLAYLIST.GetPlaylists.key] })
            toast.success("Playlist updated")
        },
    })
}

export function useDeletePlaylist() {
    const queryClient = useQueryClient()

    return useServerMutation<boolean, DeletePlaylist_Variables>({
        endpoint: API_ENDPOINTS.PLAYLIST.DeletePlaylist.endpoint,
        method: API_ENDPOINTS.PLAYLIST.DeletePlaylist.methods[0],
        mutationKey: [API_ENDPOINTS.PLAYLIST.DeletePlaylist.key],
        onSuccess: async () => {
            await queryClient.invalidateQueries({ queryKey: [API_ENDPOINTS.PLAYLIST.GetPlaylists.key] })
            toast.success("Playlist deleted")
        },
    })
}

export function useGetPlaylistEpisodes(id: Nullish<number>) {
    return useServerQuery<Array<Anime_PlaylistEpisode>>({
        endpoint: API_ENDPOINTS.PLAYLIST.GetPlaylistEpisodes.endpoint.replace("{id}", String(id)),
        method: API_ENDPOINTS.PLAYLIST.GetPlaylistEpisodes.methods[0],
        queryKey: [API_ENDPOINTS.PLAYLIST.GetPlaylistEpisodes.key, String(id)],
        enabled: !!id,
    })
}

