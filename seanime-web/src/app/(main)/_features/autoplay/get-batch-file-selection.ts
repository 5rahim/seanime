import { HibikeTorrent_BatchEpisodeFiles } from "@/api/generated/types"

export function getBatchFileSelection(
    batchFiles: HibikeTorrent_BatchEpisodeFiles | undefined,
    episodeNumber: number,
    aniDBEpisode: string,
) {
    if (!batchFiles) {
        return { fileIndex: undefined, batchEpisodeFiles: undefined }
    }

    const currentFile = batchFiles.files?.find(file => file.index === batchFiles.current)
    let file = batchFiles.currentAniDBEpisode === aniDBEpisode
        ? currentFile
        : undefined

    if (!file && currentFile?.episodeNumber) {
        const episodeOffset = episodeNumber - batchFiles.currentEpisodeNumber
        const targetFileEpisodeNumber = currentFile.episodeNumber + episodeOffset
        file = batchFiles.files?.find(file => file.episodeNumber === targetFileEpisodeNumber)
    }

    if (!file) {
        return { fileIndex: undefined, batchEpisodeFiles: batchFiles }
    }

    return {
        fileIndex: file.index,
        batchEpisodeFiles: {
            ...batchFiles,
            current: file.index,
            currentEpisodeNumber: episodeNumber,
            currentAniDBEpisode: aniDBEpisode,
        },
    }
}
