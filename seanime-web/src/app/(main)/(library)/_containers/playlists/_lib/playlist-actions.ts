import { SeaEndpoints } from "@/lib/server/endpoints"
import { useSeaMutation, useSeaQuery } from "@/lib/server/query"
import { Playlist } from "@/lib/server/types"

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
    const { mutate, isPending } = useSeaMutation<CreatePlaylistProps>({
        endpoint: SeaEndpoints.PLAYLIST,
        method: "post",
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
    const { mutate, isPending } = useSeaMutation<UpdatePlaylistProps>({
        endpoint: SeaEndpoints.PLAYLIST,
        method: "patch",
    })
    return {
        updatePlaylist: mutate,
        isUpdating: isPending,
    }
}

export const useDeletePlaylist = () => {
    const { mutate, isPending } = useSeaMutation<{ dbId: number }>({
        endpoint: SeaEndpoints.PLAYLIST,
        method: "delete",
    })
    return {
        deletePlaylist: mutate,
        isDeleting: isPending,
    }
}
