import { SeaEndpoints } from "@/lib/server/endpoints"
import { useMutation, UseMutationOptions, UseMutationResult, useQuery, UseQueryOptions, UseQueryResult } from "@tanstack/react-query"
import axios, { AxiosError } from "axios"
import { useEffect } from "react"
import { toast } from "sonner"

type SeaError = AxiosError<{ error: string }>

type SeaQuery<D> = {
    endpoint: SeaEndpoints | string
    method: "post" | "get" | "patch" | "delete"
    data?: D
    params?: D
}

/**
 * Create axios query to the server
 * - First generic: Return type
 * - Second generic: Params/Data type
 * @param endpoint
 * @param method
 * @param data
 * @param params
 */
export async function buildSeaQuery<T, D extends any = any>(
    {
        endpoint,
        method,
        data,
        params,
    }: SeaQuery<D>): Promise<T | undefined> {
    const res = await axios<T>({
        url: "http://" + (process.env.NODE_ENV === "development" ? `${window.location.hostname}:43211` : window.location.host) + "/api/v1" + endpoint,
        method,
        data,
        params,
    })
    const response = _handleSeaResponse<T>(res.data)
    return response.data
}

type SeaMutationProps<R, V = unknown> = UseMutationOptions<R, SeaError, V, unknown> & {
    endpoint: SeaEndpoints | string
    method?: "post" | "get" | "patch" | "delete"
}

export function useSeaMutation<R, V = void>(
    {
        endpoint,
        method = "post",
        ...options
    }: SeaMutationProps<R | undefined, V>): UseMutationResult<R | undefined, SeaError, V, unknown> {
    return useMutation({
        onError: error => {
            toast.error(_handleSeaError(error.response?.data?.error))
        },
        mutationFn: async (variables) => {
            return buildSeaQuery<R, V>({
                endpoint: endpoint,
                method: method,
                data: variables,
            })
        },
        ...options,
    })
}


type SeaQueryProps<TData, TParams> = UseQueryOptions<TData | undefined, SeaError, TData | undefined> & {
    endpoint: SeaEndpoints | string
    method?: "post" | "get"
    params?: TParams
    data?: TParams
}

export function useSeaQuery<TData, TParams = any>(
    {
        endpoint,
        method = "get",
        params,
        data,
        ...options
    }: SeaQueryProps<TData | undefined, TParams>): UseQueryResult<TData | undefined, SeaError> {
    const props = useQuery({
        queryFn: async () => {
            return buildSeaQuery<TData, TParams>({
                endpoint: endpoint,
                method: method,
                params: params,
                data: data,
            })
        },
        ...options,
    })

    useEffect(() => {
        if (props.isError) {
            toast.error(_handleSeaError(props.error?.response?.data?.error))
        }
    }, [props.error, props.isError])

    return props
}

//----------------------------------------------------------------------------------------------------------------------

function _handleSeaError(err: string | null | undefined): string {
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
