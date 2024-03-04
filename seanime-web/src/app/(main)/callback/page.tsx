"use client"
import { serverStatusAtom } from "@/atoms/server-status"
import { LoadingOverlay } from "@/components/ui/loading-spinner"
import { SeaEndpoints } from "@/lib/server/endpoints"
import { useSeaMutation } from "@/lib/server/query"
import { ServerStatus } from "@/lib/server/types"
import { useSetAtom } from "jotai/react"
import { useRouter } from "next/navigation"
import React from "react"
import { toast } from "sonner"

export default function Page() {
    const router = useRouter()

    const [token, setToken] = React.useState<string | null>(null)

    const { mutate: login, data, error } = useSeaMutation<ServerStatus, { token: string }>({
        mutationKey: ["login"],
        endpoint: SeaEndpoints.LOGIN,
    })

    React.useEffect(() => {
        const _token = window?.location?.hash?.replace("#access_token=", "")?.replace(/&.*/, "")
        setToken(_token)
        if (!!_token) {
            login({ token: _token })
        } else {
            toast.error("Invalid token")
            router.push("/")
        }
    }, [])

    const setServerStatus = useSetAtom(serverStatusAtom)

    React.useEffect(() => {
        if (error) {
            toast.error(error.message)
            router.push("/")
        }
    }, [error])

    React.useEffect(() => {
        if (!!data && !!token) {
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
            <LoadingOverlay className="fixed w-full h-full z-[80]">
                <h3 className="mt-2">Authenticating...</h3>
            </LoadingOverlay>
        </div>
    )
}
