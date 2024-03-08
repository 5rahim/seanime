import { SeaEndpoints } from "@/lib/server/endpoints"
import { useSeaMutation, useSeaQuery } from "@/lib/server/query"
import { Playlist } from "@/lib/server/types"
import { useQueryClient } from "@tanstack/react-query"
import { toast } from "sonner"

export const useGetPlaylists = () => {

    const { data, isLoading, isError, error } = useSeaQuery<Playlist[]>({
        endpoint: SeaEndpoints.PLAYLISTS,
        queryKey: ["get-playlists"],
    })

    return {
        playlists: data,
        isLoading,
        isError,
        error,
    }
}

type CreatePlaylistProps = {
    name: string
    paths: string[]
}

export const useCreatePlaylist = () => {
    const qc = useQueryClient()
    const { mutate, isPending } = useSeaMutation<void, CreatePlaylistProps>({
        endpoint: SeaEndpoints.PLAYLIST,
        method: "post",
        onSuccess: () => {
            qc.refetchQueries({ queryKey: ["get-playlists"] })
            toast.success("Playlist created")
        },
    })
    return {
        createPlaylist: mutate,
        isCreating: isPending,
    }
}

type UpdatePlaylistProps = {
    dbId: number
    name: string
    paths: string[]
}

export const useUpdatePlaylist = () => {
    const qc = useQueryClient()
    const { mutate, isPending } = useSeaMutation<void, UpdatePlaylistProps>({
        endpoint: SeaEndpoints.PLAYLIST,
        method: "patch",
        onSuccess: () => {
            qc.refetchQueries({ queryKey: ["get-playlists"] })
            toast.success("Playlist updated")
        },
    })
    return {
        updatePlaylist: mutate,
        isUpdating: isPending,
    }
}

export const useDeletePlaylist = () => {
    const qc = useQueryClient()
    const { mutate, isPending } = useSeaMutation<void, { dbId: number }>({
        endpoint: SeaEndpoints.PLAYLIST,
        method: "delete",
        onSuccess: () => {
            qc.refetchQueries({ queryKey: ["get-playlists"] })
            toast.success("Playlist deleted")
        },
    })
    return {
        deletePlaylist: mutate,
        isDeleting: isPending,
    }
}
