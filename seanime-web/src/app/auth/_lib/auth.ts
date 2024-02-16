import { SeaEndpoints } from "@/lib/server/endpoints"
import { buildSeaQuery } from "@/lib/server/query"
import { ServerStatus } from "@/lib/server/types"
import { useQuery } from "@tanstack/react-query"
import { useEffect, useState } from "react"

async function q_login(token: string) {
    return buildSeaQuery<ServerStatus, { token: string }>({
        endpoint: SeaEndpoints.LOGIN,
        method: "post",
        data: {
            token,
        },
    })
}

export function useAuth() {

    const [token, setToken] = useState<string | null>(null)

    const { data, error } = useQuery({
        queryKey: ["login"],
        queryFn: async () => q_login(token ?? ""),
        enabled: !!token,
    })

    useEffect(() => {
        if (window !== undefined) {
            setToken(window?.location?.hash?.replace("#access_token=", "")?.replace(/&.*/, ""))
        }
    }, [])

    return { data, error, token }

}
