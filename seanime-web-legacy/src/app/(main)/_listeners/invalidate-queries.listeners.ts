import { WSEvents } from "@/lib/server/ws-events"
import { useQueryClient } from "@tanstack/react-query"
import { useWebsocketMessageListener } from "../_hooks/handle-websockets"

export function useInvalidateQueriesListener() {

    const queryClient = useQueryClient()

    useWebsocketMessageListener<string[]>({
        type: WSEvents.INVALIDATE_QUERIES,
        onMessage: async (data) => {
            await Promise.all(data.map(async (queryKey) => {
                await queryClient.invalidateQueries({ queryKey: [queryKey] })
            }))
        },
    })

}