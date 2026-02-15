import { useQuery } from "@tanstack/react-query"

/* -------------------------------------------------------------------------------------------------
 * @link https://github.com/lexesjan/typescript-aniskip-extension/blob/main/src/api/aniskip-http-client/aniskip-http-client.types.ts
 * -----------------------------------------------------------------------------------------------*/

export const SKIP_TYPE_NAMES: Record<AniSkipType, string> = {
    op: "Opening",
    ed: "Ending",
    "mixed-op": "Mixed opening",
    "mixed-ed": "Mixed ending",
    recap: "Recap",
} as const

export const SKIP_TYPES = [
    "op",
    "ed",
    "mixed-op",
    "mixed-ed",
    "recap",
] as const

export type AniSkipType = (typeof SKIP_TYPES)[number]

export type AniSkipTime = {
    interval: {
        startTime: number
        endTime: number
    }
    skipType: AniSkipType
    skipId: string
    episodeLength: number
}

export function useSkipData(mediaMalId: number | null | undefined, episodeNumber: number | null | undefined = -1) {
    const res = useQuery({
        queryKey: ["skip-data", mediaMalId, episodeNumber],
        queryFn: async () => {
            const result = await fetch(
                `https://api.aniskip.com/v2/skip-times/${mediaMalId}/${episodeNumber}?types[]=ed&types[]=mixed-ed&types[]=mixed-op&types[]=op&types[]=recap&episodeLength=`,
            )
            const skip = (await result.json()) as {
                found: boolean,
                results: AniSkipTime[]
            }
            if (!!skip.results && skip.found) return {
                op: skip.results?.find((item) => item.skipType === "op") || null,
                ed: skip.results?.find((item) => item.skipType === "ed") || null,
            }
            return { op: null, ed: null }
        },
        refetchOnWindowFocus: false,
        enabled: !!mediaMalId && episodeNumber != -1,
    })
    return { data: res.data, isLoading: res.isLoading || res.isFetching, isError: res.isError }
}
