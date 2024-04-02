import { LocalFile } from "@/app/(main)/(library)/_lib/anime-library.types"
import { SeaEndpoints } from "@/lib/server/endpoints"
import { useSeaMutation } from "@/lib/server/query"
import { useQueryClient } from "@tanstack/react-query"
import { toast } from "sonner"

export type LocalFileBulkAction = "lock" | "unlock"

export function useLocalFileBulkAction({ onSuccess }: { onSuccess: () => void }) {

    const qc = useQueryClient()

    // Return data is ignored
    const { mutate, isPending } = useSeaMutation<LocalFile[], { action: LocalFileBulkAction }>({
        endpoint: SeaEndpoints.LOCAL_FILES,
        mutationKey: ["local-file-bulk-action"],
        method: "post",
        onSuccess: async () => {
            await qc.refetchQueries({ queryKey: ["get-library-collection"] })
            onSuccess()
        },
    })

    return {
        lockFiles: () => mutate({
            action: "lock",
        }, {
            onSuccess: () => {
                toast.success("Files locked")
            },
        }),
        unlockFiles: () => mutate({
            action: "unlock",
        }, {
            onSuccess: () => {
                toast.success("Files unlocked")
            },
        }),
        isPending,
    }

}

export function useRemoveEmptyDirectories({ onSuccess }: { onSuccess: () => void }) {

    // Return data is ignored
    const { mutate, isPending } = useSeaMutation<null>({
        endpoint: SeaEndpoints.EMPTY_DIRECTORIES,
        mutationKey: ["remove-empty-directories"],
        method: "delete",
        onSuccess: async () => {
            toast.success("Empty directories removed")
            onSuccess()
        },
    })

    return {
        removeEmptyDirectories: () => mutate(),
        isPending,
    }

}
