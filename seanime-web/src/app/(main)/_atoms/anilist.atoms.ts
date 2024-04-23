import { BaseMediaFragment } from "@/lib/anilist/gql/graphql"
import { atom } from "jotai/index"

export const anilistUserMediaAtom = atom<BaseMediaFragment[] | undefined>(undefined)
