"use client"
import { Anime_Entry } from "@/api/generated/types"
import { useOpenAnimeEntryInExplorer } from "@/api/hooks/anime_entries.hooks"
import { useStartDefaultMediaPlayer } from "@/api/hooks/mediaplayer.hooks"
import { useServerStatus } from "@/app/(main)/_hooks/use-server-status"
import {
    __bulkDeleteFilesModalIsOpenAtom,
    AnimeEntryBulkDeleteFilesModal,
} from "@/app/(main)/entry/_containers/entry-actions/anime-entry-bulk-delete-files-modal"
import {
    __animeEntryDownloadFilesModalIsOpenAtom,
    AnimeEntryDownloadFilesModal,
} from "@/app/(main)/entry/_containers/entry-actions/anime-entry-download-files-modal"
import { __metadataManager_isOpenAtom, AnimeEntryMetadataManager } from "@/app/(main)/entry/_containers/entry-actions/anime-entry-metadata-manager"
import {
    __animeEntryUnmatchFilesModalIsOpenAtom,
    AnimeEntryUnmatchFilesModal,
} from "@/app/(main)/entry/_containers/entry-actions/anime-entry-unmatch-files-modal"
import { IconButton } from "@/components/ui/button"
import { DropdownMenu, DropdownMenuItem, DropdownMenuLabel, DropdownMenuSeparator } from "@/components/ui/dropdown-menu"
import { useSetAtom } from "jotai"
import React from "react"
import { BiDotsVerticalRounded, BiRightArrowAlt } from "react-icons/bi"

export function AnimeEntryDropdownMenu({ entry }: { entry: Anime_Entry }) {

    const serverStatus = useServerStatus()
    const setIsMetadataManagerOpen = useSetAtom(__metadataManager_isOpenAtom)

    const inLibrary = !!entry.libraryData

    // Start default media player
    const { mutate: startDefaultMediaPlayer } = useStartDefaultMediaPlayer()
    // Open entry in explorer
    const { mutate: openEntryInExplorer } = useOpenAnimeEntryInExplorer()

    const setBulkDeleteFilesModalOpen = useSetAtom(__bulkDeleteFilesModalIsOpenAtom)
    const setAnimeEntryUnmatchFilesModalOpen = useSetAtom(__animeEntryUnmatchFilesModalIsOpenAtom)
    const setDownloadFilesModalOpen = useSetAtom(__animeEntryDownloadFilesModalIsOpenAtom)


    return (
        <>
            <DropdownMenu trigger={<IconButton icon={<BiDotsVerticalRounded />} intent="gray-basic" size="lg" />}>

                {inLibrary && <>
                    <DropdownMenuItem
                        onClick={() => openEntryInExplorer({ mediaId: entry.mediaId })}
                    >
                        Open directory
                    </DropdownMenuItem>

                    {serverStatus?.settings?.mediaPlayer?.defaultPlayer != "mpv" && <DropdownMenuItem
                        onClick={() => startDefaultMediaPlayer()}
                    >
                        Start video player
                    </DropdownMenuItem>}
                    <DropdownMenuSeparator />
                </>}

                <DropdownMenuItem
                    onClick={() => setIsMetadataManagerOpen(p => !p)}
                >
                    Metadata
                </DropdownMenuItem>


                {inLibrary && <>
                    <DropdownMenuSeparator />
                    <DropdownMenuLabel>Bulk actions</DropdownMenuLabel>
                    <DropdownMenuItem
                        className="flex justify-between"
                        onClick={() => setDownloadFilesModalOpen(p => !p)}
                    >
                        <span>Download some files</span> <BiRightArrowAlt />
                    </DropdownMenuItem>
                    <DropdownMenuItem
                        className="text-orange-500 dark:text-orange-200 flex justify-between"
                        onClick={() => setAnimeEntryUnmatchFilesModalOpen(true)}
                    >
                        <span>Unmatch some files</span> <BiRightArrowAlt />
                    </DropdownMenuItem>
                    <DropdownMenuItem
                        className="text-red-500 dark:text-red-200 flex justify-between"
                        onClick={() => setBulkDeleteFilesModalOpen(true)}
                    >
                        <span>Delete some files</span> <BiRightArrowAlt />
                    </DropdownMenuItem>
                </>}
            </DropdownMenu>

            <AnimeEntryDownloadFilesModal entry={entry} />
            <AnimeEntryMetadataManager entry={entry} />
            <AnimeEntryBulkDeleteFilesModal entry={entry} />
            <AnimeEntryUnmatchFilesModal entry={entry} />
        </>
    )
}
