import { useWebsocketMessageListener } from "@/atoms/websocket"
import { SeaEndpoints, WSEvents } from "@/lib/server/endpoints"
import { useSeaMutation, useSeaQuery } from "@/lib/server/query"
import { useQueryClient } from "@tanstack/react-query"
import { toast } from "sonner"

type CreateOfflineSnapshot_QueryVariables = {
    animeMediaIds: number[]
}

export type OfflineSnapshot = {
    id: number
    createdAt: string
}

export function useOfflineSnapshot() {
    const qc = useQueryClient()

    const { mutate, isPending: isCreating } = useSeaMutation<void, CreateOfflineSnapshot_QueryVariables>({
        endpoint: SeaEndpoints.OFFLINE_SNAPSHOT,
        mutationKey: ["create-offline-snapshot"],
        onSuccess: () => {
            toast.info("Creating snapshot...")
        },
    })

    const { data, isLoading } = useSeaQuery<OfflineSnapshot>({
        endpoint: SeaEndpoints.OFFLINE_SNAPSHOT,
        queryKey: ["offline-snapshot"],
    })

    useWebsocketMessageListener({
        type: WSEvents.OFFLINE_SNAPSHOT_CREATED,
        onMessage: _ => {
            qc.refetchQueries({ queryKey: ["offline-snapshot"] })
        },
    })

    return {
        createOfflineSnapshot: mutate,
        snapshot: data,
        isLoading,
        isCreating,
    }
}
