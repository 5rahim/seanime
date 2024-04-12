import { OfflineSnapshot } from "@/app/(main)/(offline)/offline/_lib/offline-snapshot.types"
import { SeaEndpoints } from "@/lib/server/endpoints"
import { useSeaQuery } from "@/lib/server/query"
import { useQueryClient } from "@tanstack/react-query"

export function useGetOfflineSnapshot() {
    const qc = useQueryClient()

    const { data, isLoading } = useSeaQuery<OfflineSnapshot>({
        endpoint: SeaEndpoints.OFFLINE_SNAPSHOT,
        queryKey: ["get-offline-snapshot"],
    })

    return {
        snapshot: data,
        isLoading,
    }
}
