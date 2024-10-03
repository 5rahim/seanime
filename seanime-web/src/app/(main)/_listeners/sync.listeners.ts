import { API_ENDPOINTS } from "@/api/generated/endpoints"
import { Sync_QueueState } from "@/api/generated/types"
import { useSyncIsActive } from "@/app/(main)/_atoms/sync.atoms"
import { useWebsocketMessageListener } from "@/app/(main)/_hooks/handle-websockets"
import { logger } from "@/lib/helpers/debug"
import { WSEvents } from "@/lib/server/ws-events"
import { useQueryClient } from "@tanstack/react-query"
import React from "react"

export function useSyncListener() {
    const qc = useQueryClient()

    useWebsocketMessageListener({
        type: WSEvents.SYNC_LOCAL_FINISHED,
        onMessage: _ => {
            qc.invalidateQueries({ queryKey: [API_ENDPOINTS.SYNC.SyncGetTrackedMediaItems.key] })
        },
    })

    const [queueState, setQueueState] = React.useState<Sync_QueueState | null>(null)
    useWebsocketMessageListener<Sync_QueueState>({
        type: WSEvents.SYNC_LOCAL_QUEUE_STATE,
        onMessage: data => {
            logger("SYNC").info("Queue state", queueState)
            setQueueState(data)
        },
    })

    const { setSyncIsActive } = useSyncIsActive()

    React.useEffect(() => {
        setSyncIsActive(!!queueState && (Object.keys(queueState.animeTasks!).length > 0 || Object.keys(queueState.mangaTasks!).length > 0))
    }, [queueState])
}
