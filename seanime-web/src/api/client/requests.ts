"use client"
import { getServerBaseUrl } from "@/api/client/server-url"
import { serverAuthTokenAtom } from "@/app/(main)/_atoms/server-status.atoms"
import { useMutation, UseMutationOptions, useQuery, UseQueryOptions } from "@tanstack/react-query"
import axios, { AxiosError, InternalAxiosRequestConfig } from "axios"
import { useAtomValue } from "jotai"
import { usePathname } from "next/navigation"
import { useEffect } from "react"
import { toast } from "sonner"

type SeaError = AxiosError<{ error: string }>

type SeaQuery<D> = {
    endpoint: string
    method: "POST" | "GET" | "PATCH" | "DELETE" | "PUT"
    data?: D
    params?: D
    password?: string
}

/**
 * Create axios query to the server
 * - First generic: Return type
 * - Second generic: Params/Data type
 */
export async function buildSeaQuery<T, D extends any = any>(
    {
        endpoint,
        method,
        data,
        params,
        password,
    }: SeaQuery<D>): Promise<T | undefined> {

    axios.interceptors.request.use((request: InternalAxiosRequestConfig) => {
            if (password) {
                request.headers.set("X-Seanime-Token", password)
            }
            return request
        },
    )

    const res = await axios<T>({
        url: getServerBaseUrl() + endpoint,
        method,
        data,
        params,
    })
    const response = _handleSeaResponse<T>(res.data)
    return response.data
}

type ServerMutationProps<R, V = void> = UseMutationOptions<R | undefined, SeaError, V, unknown> & {
    endpoint: string
    method: "POST" | "GET" | "PATCH" | "DELETE" | "PUT"
}

/**
 * Create mutation hook to the server
 * - First generic: Return type
 * - Second generic: Params/Data type
 */
export function useServerMutation<R = void, V = void>(
    {
        endpoint,
        method,
        ...options
    }: ServerMutationProps<R, V>) {

    const password = useAtomValue(serverAuthTokenAtom)

    return useMutation<R | undefined, SeaError, V>({
        onError: error => {
            console.log("Mutation error", error)
            toast.error(_handleSeaError(error.response?.data))
        },
        mutationFn: async (variables) => {
            return buildSeaQuery<R, V>({
                endpoint: endpoint,
                method: method,
                data: variables,
                password: password,
            })
        },
        ...options,
    })
}


type ServerQueryProps<R, V> = UseQueryOptions<R | undefined, SeaError, R | undefined> & {
    endpoint: string
    method: "POST" | "GET" | "PATCH" | "DELETE" | "PUT"
    params?: V
    data?: V
    muteError?: boolean
}

/**
 * Create query hook to the server
 * - First generic: Return type
 * - Second generic: Params/Data type
 */
export function useServerQuery<R, V = any>(
    {
        endpoint,
        method,
        params,
        data,
        muteError,
        ...options
    }: ServerQueryProps<R | undefined, V>) {

    const pathname = usePathname()
    const password = useAtomValue(serverAuthTokenAtom)

    const props = useQuery<R | undefined, SeaError>({
        queryFn: async () => {
            return buildSeaQuery<R, V>({
                endpoint: endpoint,
                method: method,
                params: params,
                data: data,
                password: password,
            })
        },
        ...options,
    })

    useEffect(() => {
        if (!muteError && props.isError) {
            if (props.error?.response?.data?.error === "UNAUTHENTICATED" && pathname !== "/public/auth") {
                window.location.href = "/public/auth"
                return
            }
            console.log("Server error", props.error)
            toast.error(_handleSeaError(props.error?.response?.data))
        }
    }, [props.error, props.isError, muteError])

    return props
}

//----------------------------------------------------------------------------------------------------------------------

function _handleSeaError(data: any): string {
    if (typeof data === "string") return "Server Error: " + data

    const err = data?.error as string

    if (!err) return "Unknown error"

    if (err.includes("Too many requests"))
        return "AniList: Too many requests, please wait a moment and try again."

    try {
        const graphqlErr = JSON.parse(err) as any
        console.log("AniList error", graphqlErr)
        if (graphqlErr.graphqlErrors && graphqlErr.graphqlErrors.length > 0 && !!graphqlErr.graphqlErrors[0]?.message) {
            return "AniList error: " + graphqlErr.graphqlErrors[0]?.message
        }
        return "AniList error"
    }
    catch (e) {
        return "Error: " + err
    }
}

function _handleSeaResponse<T>(res: unknown): { data: T | undefined, error: string | undefined } {

    if (typeof res === "object" && !!res && "error" in res && typeof res.error === "string") {
        return { data: undefined, error: res.error }
    }
    if (typeof res === "object" && !!res && "data" in res) {
        return { data: res.data as T, error: undefined }
    }

    return { data: undefined, error: "No response from the server" }

}
