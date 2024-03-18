"use client"
import { serverStatusAtom } from "@/atoms/server-status"
import { websocketConnectedAtom } from "@/components/application/websocket-provider"
import { LoadingOverlay } from "@/components/ui/loading-spinner"
import { SeaEndpoints } from "@/lib/server/endpoints"
import { ServerStatus } from "@/lib/server/types"
import { useMutation } from "@tanstack/react-query"
import axios from "axios"
import { useAtom, useAtomValue } from "jotai/react"
import { useRouter } from "next/navigation"
import React from "react"
import { toast } from "sonner"

export default function CallbackPage() {
    const router = useRouter()
    const [status, setServerStatus] = useAtom(serverStatusAtom)
    const isConnected = useAtomValue(websocketConnectedAtom)


    const [token, setToken] = React.useState<string | null>(null)

    const { mutate: login, error } = useMutation<ServerStatus, any, { token: string }>({
        mutationKey: ["login", token],
        mutationFn: async (variables) => {
            const res = await axios(typeof window !== "undefined" ? ("http://" + (process.env.NODE_ENV === "development"
                ? `${window?.location?.hostname}:43211`
                : window?.location?.host) + "/api/v1" + SeaEndpoints.LOGIN) : "", {
                method: "POST",
                data: variables,
            })
            return await res.data?.data
        },
        onSuccess: (data) => {
            console.log(data)
            setServerStatus(data)
            const t = setTimeout(() => {
                toast.success("Successfully authenticated")
                router.push("/")
            }, 200)
        },
        onError: (error) => {
            toast.error(error.message)
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

    React.useEffect(() => {
        if (!!status?.user) {
            router.push("/")
        }
    }, [status])

    return (
        <div>
            <LoadingOverlay className="fixed w-full h-full z-[80]">
                <h3 className="mt-2">Authenticating...</h3>
            </LoadingOverlay>
        </div>
    )
}
