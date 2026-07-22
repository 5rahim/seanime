import { atomWithImmer } from "jotai-immer"
import { atomWithStorage } from "jotai/utils"

type Params = {
    search: string
    page: number
    perPage: number
    type: "anime" | "manga"
}

export const __customSources_paramsAtom = atomWithImmer<Params>({
    type: "anime",
    search: "",
    page: 1,
    perPage: 100,
})

export const __customSources_providerAtom = atomWithStorage<string | null>("sea-custom-sources-provider", null, undefined, { getOnInit: true })
