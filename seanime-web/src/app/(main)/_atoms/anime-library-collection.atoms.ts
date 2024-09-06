import { Anime_LibraryCollection } from "@/api/generated/types"
import { atom } from "jotai/index"

export const libraryCollectionAtom = atom<Anime_LibraryCollection | undefined>(undefined)

export const getAtomicLibraryEntryAtom = atom(get => get(libraryCollectionAtom)?.lists?.length,
    (get, set, payload: number) => {
        const lists = get(libraryCollectionAtom)?.lists
        if (!lists) {
            return undefined
        }
        return lists.flatMap(n => n.entries)?.filter(Boolean).find(n => n.mediaId === payload)
    },
)
