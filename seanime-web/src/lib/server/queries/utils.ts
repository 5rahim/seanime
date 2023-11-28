import { SeaEndpoints } from "@/lib/server/endpoints"
import axios, { AxiosError } from "axios"
import { SEANIME_SERVER_URI } from "@/lib/server/constants"
import {
    useMutation,
    UseMutationOptions,
    UseMutationResult,
    useQuery,
    UseQueryOptions,
    UseQueryResult,
} from "@tanstack/react-query"
import toast from "react-hot-toast"
import { useEffect } from "react"

type SeaError = AxiosError<{ error: string }>

type SeaQuery<D> = {
    endpoint: SeaEndpoints | string
    method: "post" | "get" | "patch"
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
        url: SEANIME_SERVER_URI + endpoint,
        method,
        data,
        params,
    })
    const response = _handleSeaResponse<T>(res.data)
    return response.data
}

type SeaMutationProps<R, V = unknown> = UseMutationOptions<R, SeaError, V, unknown> & {
    endpoint: SeaEndpoints
    method?: "post" | "get" | "patch"
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
                data: params,
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
    if (!err) return ""
    try {
        const graphqlErr = JSON.parse(err)
        return "AniList error"
    } catch (e) {
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