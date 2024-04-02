import { LocalFile } from "@/app/(main)/(library)/_lib/anime-library.types"

export type Playlist = {
    dbId: number
    name: string
    localFiles: LocalFile[]
}
