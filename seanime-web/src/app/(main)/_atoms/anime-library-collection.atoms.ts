import { Anime_LibraryCollection } from "@/api/generated/types"
import { atom } from "jotai/index"

export const animeLibraryCollectionAtom = atom<Anime_LibraryCollection | undefined>(undefined)

export const getAtomicLibraryEntryAtom = atom(get => get(animeLibraryCollectionAtom)?.lists?.length,
    (get, set, payload: number) => {
        const lists = get(animeLibraryCollectionAtom)?.lists
        if (!lists) {
            return undefined
        }
        return lists.flatMap(n => n.entries)?.filter(Boolean).find(n => n.mediaId === payload)
    },
)
