import { useLogin } from "@/api/hooks/auth.hooks"
import { websocketConnectedAtom } from "@/app/websocket-provider"
import { LoadingOverlay } from "@/components/ui/loading-spinner"
import { useAtomValue } from "jotai/react"
import { useRouter } from "next/navigation"
import React from "react"
import { toast } from "sonner"

type CallbackPageProps = {}

/**
 * @description
 * - Logs the user in using the AniList token present in the URL hash
 */
export function CallbackPage(props: CallbackPageProps) {
    const router = useRouter()
    const {} = props

    const websocketConnected = useAtomValue(websocketConnectedAtom)

    const { mutate: login } = useLogin()

    const called = React.useRef(false)

    React.useEffect(() => {
        if (typeof window !== "undefined" && websocketConnected) {
            /**
             * Get the AniList token from the URL hash
             */
            const _token = window?.location?.hash?.replace("#access_token=", "")?.replace(/&.*/, "")
            if (!!_token && !called.current) {
                login({ token: _token })
                called.current = true
            } else {
                toast.error("Invalid token")
                router.push("/")
            }
        }
    }, [websocketConnected])

    return (
        <div>
            <LoadingOverlay className="fixed w-full h-full z-[80]">
                <h3 className="mt-2">Authenticating...</h3>
            </LoadingOverlay>
        </div>
    )
}
