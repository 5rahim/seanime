import { LibraryCollection } from "@/app/(main)/(library)/_lib/anime-library.types"
import { atom } from "jotai/index"

export const libraryCollectionAtom = atom<LibraryCollection | undefined>(undefined)

export const getAtomicLibraryEntryAtom = atom(get => get(libraryCollectionAtom),
    (get, set, payload: number) => {
        const lists = get(libraryCollectionAtom)?.lists
        if (!lists) {
            return undefined
        }
        return lists.flatMap(n => n.entries).find(n => n.mediaId === payload)
    },
)
