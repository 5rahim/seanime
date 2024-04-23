"use client"
import { serverStatusAtom } from "@/app/(main)/_atoms/server-status.atoms"
import { websocketConnectedAtom } from "@/components/application/websocket-provider"
import { LoadingOverlay } from "@/components/ui/loading-spinner"
import { __DEV_SERVER_PORT } from "@/lib/anilist/config"
import { SeaEndpoints } from "@/lib/server/endpoints"

import { ServerStatus } from "@/lib/types/server-status.types"
import { useMutation } from "@tanstack/react-query"
import axios from "axios"
import { useAtom, useAtomValue } from "jotai/react"
import { useRouter } from "next/navigation"
import React from "react"
import { toast } from "sonner"

export const dynamic = "force-static"

export default function CallbackPage() {
    const router = useRouter()
    const [status, setServerStatus] = useAtom(serverStatusAtom)
    const isConnected = useAtomValue(websocketConnectedAtom)

    const [token, setToken] = React.useState<string | null>(null)

    const { mutate: login } = useMutation<ServerStatus, any, { token: string }>({
        mutationKey: ["login", token],
        mutationFn: async (variables) => {
            const res = await axios(typeof window !== "undefined" ? ("http://" + (process.env.NODE_ENV === "development"
                ? `${window?.location?.hostname}:${__DEV_SERVER_PORT}`
                : window?.location?.host) + "/api/v1" + SeaEndpoints.LOGIN) : "", {
                method: "POST",
                data: variables,
            })
            return await res.data?.data
        },
        onSuccess: (data) => {
            console.log(data)
            setServerStatus(data)
            toast.success("Successfully authenticated")
            router.push("/")
        },
        onError: (error) => {
            toast.error(error.message)
            router.push("/")
        },
    })

    const called = React.useRef(false)

    React.useEffect(() => {
        if (typeof window !== "undefined" && isConnected) {
            const _token = window?.location?.hash?.replace("#access_token=", "")?.replace(/&.*/, "")
            setToken(_token)
            if (!!_token && !called.current) {
                login({ token: _token })
                called.current = true
            } else {
                toast.error("Invalid token")
                router.push("/")
            }
        }
    }, [isConnected])

    return (
        <div>
            <LoadingOverlay className="fixed w-full h-full z-[80]">
                <h3 className="mt-2">Authenticating...</h3>
            </LoadingOverlay>
        </div>
    )
}
