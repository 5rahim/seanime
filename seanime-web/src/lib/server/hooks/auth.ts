import { useQuery } from "@tanstack/react-query"
import { q_login } from "@/lib/server/queries/general"
import { useEffect, useState } from "react"

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