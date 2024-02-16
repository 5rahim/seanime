import { AnimeCollectionQuery } from "@/lib/anilist/gql/graphql"
import { SeaEndpoints } from "@/lib/server/endpoints"
import { useSeaMutation } from "@/lib/server/query"
import { useQueryClient } from "@tanstack/react-query"
import toast from "react-hot-toast"

export function useAddUnknownMedia() {

    const qc = useQueryClient()

    // Return data is ignored
    const { mutate, isPending } = useSeaMutation<AnimeCollectionQuery, { mediaIds: number[] }>({
        endpoint: SeaEndpoints.MEDIA_ENTRY_UNKNOWN_MEDIA,
        mutationKey: ["add-unknown-media"],
        onSuccess: async () => {
            // Refetch library collection
            toast.success("AniList is up-to-date")
            await qc.refetchQueries({ queryKey: ["get-library-collection"] })
            await qc.refetchQueries({ queryKey: ["get-anilist-collection"] })
            await qc.refetchQueries({ queryKey: ["get-missing-episodes"] })
        },
    })

    return {
        addUnknownMedia: mutate,
        isPending,
    }

}
