import { HibikeTorrent_BatchEpisodeFiles } from "@/api/generated/types"

export function getBatchSelectionParams(batchEpFiles: HibikeTorrent_BatchEpisodeFiles | undefined, ep: number, aniDBEpisode: string) {
    if (!batchEpFiles) {
        return {
            fileIndex: undefined,
            batchEpisodeFiles: undefined,
        }
    }

    const curr = batchEpFiles.files?.find(f => f.index === batchEpFiles.current)
    let file = batchEpFiles.currentAniDBEpisode === aniDBEpisode ? curr : undefined

    if (curr?.episodeNumber && !file) {
        const episodeOffset = ep - batchEpFiles.currentEpisodeNumber
        const targetEpNum = curr.episodeNumber + episodeOffset
        file = batchEpFiles.files?.find(f => f.episodeNumber === targetEpNum)
    }

    if (!file) {
        return {
            fileIndex: undefined,
            batchEpisodeFiles: batchEpFiles,
        }
    }

    return {
        fileIndex: file.index,
        batchEpisodeFiles: {
            ...batchEpFiles,
            current: file.index,
            currentEpisodeNumber: ep,
            currentAniDBEpisode: aniDBEpisode,
        },
    }
}
