"use client"
import { serverStatusAtom } from "@/atoms/server-status"
import { LoadingOverlay } from "@/components/ui/loading-spinner"
import { SeaEndpoints } from "@/lib/server/endpoints"
import { ServerStatus } from "@/lib/server/types"
import { useMutation } from "@tanstack/react-query"
import axios from "axios"
import { useAtom } from "jotai/react"
import { useRouter } from "next/navigation"
import React from "react"
import { useUpdateEffect } from "react-use"
import { toast } from "sonner"

export default function CallbackPage() {
    const router = useRouter()
    const [status, setServerStatus] = useAtom(serverStatusAtom)

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
            }, 1000)
        },
    })

    React.useEffect(() => {
        if (typeof window !== "undefined") {
            console.log(window?.location?.hash)
            const _token = window?.location?.hash?.replace("#access_token=", "")?.replace(/&.*/, "")
            setToken(_token)
            console.log(_token)
            if (!!_token) {
                login({ token: _token })
            } else {
                toast.error("Invalid token")
                router.push("/")
            }
        }
    }, [])


    useUpdateEffect(() => {
        if (!!error) {
            toast.error(error.message)
            router.push("/")
        }
    }, [error])

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
