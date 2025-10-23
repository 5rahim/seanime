import { Anime_LocalFile, LibraryExplorer_FileTreeNodeJSON } from "@/api/generated/types"
import { atom } from "jotai"
import { useAtom } from "jotai/react"

export const libraryExplorer_drawerOpenAtom = atom(false)

export const libraryExplorer_selectedNodeAtom = atom<LibraryExplorer_FileTreeNodeJSON | null>(null)

export const libraryExplorer_isSelectingPathsAtom = atom(false)
export const libraryExplorer_selectedPathsAtom = atom<Set<string>>(new Set<string>())

export type LibraryExplorer_Filter = "UNLOCKED" | "UNMATCHED" | "IGNORED" | "UNKNOWN_MEDIA" | undefined

export const libraryExplorer_selectedFilterAtom = atom<LibraryExplorer_Filter>(undefined)

export const libraryExplorer_matchLocalFilesAtom = atom<Anime_LocalFile[]>([])
export const libraryExplorer_resolveUnknownLocalFilesAtom = atom<Anime_LocalFile[]>([])

export const libraryExplorer_openDirectoryAtom = atom<string | null>(null)

export const libraryExplorer_superUpdateDrawerOpenAtom = atom(false)

export function useLibraryExplorer() {
    const [open, setOpen] = useAtom(libraryExplorer_drawerOpenAtom)
    const [openDirInLibraryExplorer, setOpenDirInLibraryExplorer] = useAtom(libraryExplorer_openDirectoryAtom)
    return {
        openDirInLibraryExplorer: (path: string) => {
            setOpenDirInLibraryExplorer(path)
            setTimeout(() => {
                setOpen(true)
            }, 200)
        },
    }
}
