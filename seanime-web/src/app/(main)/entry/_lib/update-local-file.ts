import { SeaEndpoints } from "@/lib/server/endpoints"
import { useSeaMutation } from "@/lib/server/query"
import { LocalFile, LocalFileMetadata } from "@/lib/server/types"
import { useQueryClient } from "@tanstack/react-query"

type UpdateLocalFileVariables = {
    path: string
    metadata?: LocalFileMetadata
    locked: boolean
    ignored: boolean
    mediaId: number
}

function getDefaultLocalFileVariables(lf: LocalFile): UpdateLocalFileVariables {
    return {
        path: lf.path,
        metadata: lf.metadata,
        locked: lf.locked,
        ignored: lf.ignored,
        mediaId: lf.mediaId,
    }
}

export function useUpdateLocalFile(mId?: number) {

    const qc = useQueryClient()

    // Return data is ignored
    const { mutate, isPending } = useSeaMutation<LocalFile[], UpdateLocalFileVariables>({
        endpoint: SeaEndpoints.LOCAL_FILE,
        mutationKey: ["patch-local-file"],
        method: "patch",
        onSuccess: async () => {
            await qc.refetchQueries({ queryKey: ["get-library-collection"] })
            if (mId) {
                await qc.refetchQueries({ queryKey: ["get-media-entry", mId] })
            }
        },
    })

    return {
        updateLocalFile: (lf: LocalFile, variables: Partial<UpdateLocalFileVariables>, onSuccess?: () => void) => {
            mutate({
                ...getDefaultLocalFileVariables(lf),
                ...variables,
            }, {
                onSuccess: () => {
                    onSuccess && onSuccess()
                },
            })
        },
        isPending,
    }

}
