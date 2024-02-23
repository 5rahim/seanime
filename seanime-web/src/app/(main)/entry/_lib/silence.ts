import { SeaEndpoints } from "@/lib/server/endpoints"
import { useSeaMutation, useSeaQuery } from "@/lib/server/query"
import { useQueryClient } from "@tanstack/react-query"

export function useMediaEntrySilenceStatus(id: number) {

    const qc = useQueryClient()

    const { data, isLoading } = useSeaQuery({
        queryKey: ["media-entry-silence-status", id],
        endpoint: SeaEndpoints.MEDIA_ENTRY_SILENCE_STATUS.replace("{id}", String(id)),
        enabled: !!id,
        refetchOnWindowFocus: false,
    })

    const { mutate, isPending } = useSeaMutation<boolean, { mediaId: number }>({
        mutationKey: ["media-entry-silence", id],
        endpoint: SeaEndpoints.MEDIA_ENTRY_SILENCE,
        method: "post",
        onSuccess: async () => {
            await qc.refetchQueries({ queryKey: ["media-entry-silence-status", id] })
            await qc.refetchQueries({ queryKey: ["get-missing-episodes"] })
        },
    })

    return {
        isSilenced: !!data,
        silenceStatusIsLoading: isLoading,
        toggleSilenceStatus: () => {
            mutate({ mediaId: id })
        },
        silenceStatusIsUpdating: isPending,
    }

}

