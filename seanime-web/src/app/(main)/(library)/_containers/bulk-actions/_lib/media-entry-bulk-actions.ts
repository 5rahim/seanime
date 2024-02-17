import { SeaEndpoints } from "@/lib/server/endpoints"
import { useSeaMutation } from "@/lib/server/query"
import { LocalFile } from "@/lib/server/types"
import { useQueryClient } from "@tanstack/react-query"
import toast from "react-hot-toast"

export type MediaEntryBulkAction = "unmatch" | "toggle-lock"

export function useMediaEntryBulkAction(mId?: number) {

    const qc = useQueryClient()

    // Return data is ignored
    const { mutate, isPending } = useSeaMutation<LocalFile[], { mediaId: number, action: MediaEntryBulkAction }>({
        endpoint: SeaEndpoints.MEDIA_ENTRY_BULK_ACTION,
        mutationKey: ["media-entry-bulk-action"],
        method: "patch",
        onSuccess: async () => {
            await qc.refetchQueries({ queryKey: ["get-library-collection"] })
            if (mId) {
                await qc.refetchQueries({ queryKey: ["get-media-entry", mId] })
            }
        },
    })

    return {
        toggleLock: (mId: number) => mutate({
            mediaId: mId,
            action: "toggle-lock",
        }),
        unmatchAll: (mId: number) => mutate({
            mediaId: mId,
            action: "unmatch",
        }, {
            onSuccess: () => {
                toast.success("Files unmatched")
            },
        }),
        isPending,
    }

}

