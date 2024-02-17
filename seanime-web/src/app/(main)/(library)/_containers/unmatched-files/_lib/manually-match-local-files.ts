import { SeaEndpoints } from "@/lib/server/endpoints"
import { useSeaMutation } from "@/lib/server/query"
import { LocalFile } from "@/lib/server/types"
import { useQueryClient } from "@tanstack/react-query"
import toast from "react-hot-toast"

type Props = { dir: string, mediaId: number }

export function useManuallyMatchLocalFiles() {

    const qc = useQueryClient()


    // Return data is ignored
    const { mutate, isPending } = useSeaMutation<LocalFile[], Props>({
        endpoint: SeaEndpoints.MEDIA_ENTRY_MANUAL_MATCH,
        mutationKey: ["media-entry-manual-match"],
        onSuccess: async () => {
            toast.success("Files matched")
            await qc.refetchQueries({ queryKey: ["get-library-collection"] })
        },
    })

    return {
        manuallyMatchEntry: (props: Props, callback: () => void) => {
            mutate(props, {
                onSuccess: async () => {
                    if (props.mediaId) {
                        await qc.refetchQueries({ queryKey: ["get-media-entry", props.mediaId] })
                    }
                    callback()
                },
            })
        },
        isPending,
    }

}
