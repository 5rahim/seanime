import { API_ENDPOINTS } from "@/api/generated/endpoints.ts"
import { isLoginModalOpenAtom } from "@/app/(main)/_atoms/server-status.atoms.ts"
import { useWebsocketMessageListener } from "@/app/(main)/_hooks/handle-websockets.ts"
import { WSEvents } from "@/lib/server/ws-events.ts"
import { useQueryClient } from "@tanstack/react-query"
import { useAtom } from "jotai/react"
import { toast } from "sonner"

export function useAuthEventListeners() {
    const queryClient = useQueryClient()

    const [, setLoginModalOpen] = useAtom(isLoginModalOpenAtom)

    useWebsocketMessageListener<string>({
        type: WSEvents.SERVER_LOGGED_OUT_ANILIST, async onMessage(msg: string) {
            // refetch the status, user should be logged out
            await queryClient.invalidateQueries({ queryKey: [API_ENDPOINTS.STATUS.GetStatus.key] })
            toast.warning(msg)
            setTimeout(() => {
                // open the login modal
                setLoginModalOpen(true)
            }, 1000)
        },
    })

}
