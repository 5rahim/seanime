import { useServerMutation } from "@/api/client/requests"
import { Login_Variables } from "@/api/generated/endpoint.types"
import { API_ENDPOINTS } from "@/api/generated/endpoints"
import { Status } from "@/api/generated/types"
import { useSetServerStatus } from "@/app/(main)/_hooks/use-server-status"
import { useQueryClient } from "@tanstack/react-query"
import { useRouter } from "next/navigation"
import { toast } from "sonner"

export function useLogin() {
    const queryClient = useQueryClient()
    const router = useRouter()
    const setServerStatus = useSetServerStatus()

    return useServerMutation<Status, Login_Variables>({
        endpoint: API_ENDPOINTS.AUTH.Login.endpoint,
        method: API_ENDPOINTS.AUTH.Login.methods[0],
        mutationKey: [API_ENDPOINTS.AUTH.Login.key],
        onSuccess: async data => {
            if (data) {
                toast.success("Successfully authenticated")
                await queryClient.invalidateQueries({ queryKey: [API_ENDPOINTS.ANIME_COLLECTION.GetLibraryCollection.key] })
                await queryClient.invalidateQueries({ queryKey: [API_ENDPOINTS.ANILIST.GetRawAnimeCollection.key] })
                await queryClient.invalidateQueries({ queryKey: [API_ENDPOINTS.ANILIST.GetAnimeCollection.key] })
                await queryClient.invalidateQueries({ queryKey: [API_ENDPOINTS.MANGA.GetMangaCollection.key] })
                setServerStatus(data)
                router.push("/")
                queryClient.invalidateQueries({ queryKey: [API_ENDPOINTS.ANIME_ENTRIES.GetMissingEpisodes.key] })
                queryClient.invalidateQueries({ queryKey: [API_ENDPOINTS.ANIME_ENTRIES.GetAnimeEntry.key] })
                queryClient.invalidateQueries({ queryKey: [API_ENDPOINTS.MANGA.GetMangaEntry.key] })
            }
        },
        onError: async error => {
            toast.error(error.message)
            router.push("/")
        },
    })
}

export function useLogout() {
    const queryClient = useQueryClient()
    const router = useRouter()
    const setServerStatus = useSetServerStatus()

    return useServerMutation<Status>({
        endpoint: API_ENDPOINTS.AUTH.Logout.endpoint,
        method: API_ENDPOINTS.AUTH.Logout.methods[0],
        mutationKey: [API_ENDPOINTS.AUTH.Logout.key],
        onSuccess: async () => {
            toast.success("Successfully logged out")
            await queryClient.invalidateQueries({ queryKey: [API_ENDPOINTS.ANIME_COLLECTION.GetLibraryCollection.key] })
            await queryClient.invalidateQueries({ queryKey: [API_ENDPOINTS.ANILIST.GetRawAnimeCollection.key] })
            await queryClient.invalidateQueries({ queryKey: [API_ENDPOINTS.ANILIST.GetAnimeCollection.key] })
            await queryClient.invalidateQueries({ queryKey: [API_ENDPOINTS.MANGA.GetMangaCollection.key] })
            router.push("/")
            queryClient.invalidateQueries({ queryKey: [API_ENDPOINTS.ANIME_ENTRIES.GetMissingEpisodes.key] })
            queryClient.invalidateQueries({ queryKey: [API_ENDPOINTS.ANIME_ENTRIES.GetAnimeEntry.key] })
            queryClient.invalidateQueries({ queryKey: [API_ENDPOINTS.MANGA.GetMangaEntry.key] })
        },
    })
}
