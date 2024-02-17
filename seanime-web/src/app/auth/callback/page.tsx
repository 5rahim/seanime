"use client"
import { serverStatusAtom } from "@/atoms/server-status"
import { LoadingOverlay } from "@/components/ui/loading-spinner"
import { SeaEndpoints } from "@/lib/server/endpoints"
import { useSeaMutation } from "@/lib/server/query"
import { ServerStatus } from "@/lib/server/types"
import { useSetAtom } from "jotai/react"
import { useRouter } from "next/navigation"
import React, { useEffect, useState } from "react"
import toast from "react-hot-toast"
import { useUpdateEffect } from "react-use"

export default function Page() {
    const router = useRouter()

    const [token, setToken] = useState<string | null>(null)

    const { mutate: login, data, error } = useSeaMutation<ServerStatus, { token: string }>({
        mutationKey: ["login"],
        endpoint: SeaEndpoints.LOGIN,
    })

    useEffect(() => {
        if (window !== undefined) {
            const token = window?.location?.hash?.replace("#access_token=", "")?.replace(/&.*/, "")
            setToken(token)
            if (!!token) {
                login({ token })
            } else {
                toast.error("Invalid token")
                router.push("/")
            }
        }
    }, [])

    const setServerStatus = useSetAtom(serverStatusAtom)

    useUpdateEffect(() => {
        if (error) {
            toast.error(error.message)
            router.push("/")
        }
    }, [error])

    useEffect(() => {
        if (window !== undefined && !!data && !!token) {
            setServerStatus(data)
            const t = setTimeout(() => {
                toast.success("Successfully authenticated")
                router.push("/")
            }, 1000)

            return () => clearTimeout(t)
        }
    }, [data])

    return (
        <div>
            <LoadingOverlay className={"fixed w-full h-full z-[80]"}>
                <h3 className={"mt-2"}>Authenticating...</h3>
            </LoadingOverlay>
        </div>
    )
}
