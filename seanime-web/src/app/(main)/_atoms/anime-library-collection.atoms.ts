import { Anime_LibraryCollection } from "@/api/generated/types"
import { derive } from "jotai-derive"
import { atom } from "jotai/index"

export const animeLibraryCollectionAtom = atom<Anime_LibraryCollection | undefined>(undefined)
export const animeLibraryCollectionWithoutStreamsAtom = derive([animeLibraryCollectionAtom], (animeLibraryCollection) => {
    if (!animeLibraryCollection) {
        return undefined
    }
    return {
        ...animeLibraryCollection,
        lists: animeLibraryCollection.lists?.map(list => ({
            ...list,
            entries: list.entries?.filter(n => !!n.libraryData),
        })),
    } as Anime_LibraryCollection
})

export const getAtomicLibraryEntryAtom = atom(get => get(animeLibraryCollectionAtom)?.lists?.length,
    (get, set, payload: number) => {
        const lists = get(animeLibraryCollectionAtom)?.lists
        if (!lists) {
            return undefined
        }
        return lists.flatMap(n => n.entries)?.filter(Boolean).find(n => n.mediaId === payload)
    },
)
