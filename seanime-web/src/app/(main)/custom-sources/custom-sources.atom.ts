import { atomWithImmer } from "jotai-immer"

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
