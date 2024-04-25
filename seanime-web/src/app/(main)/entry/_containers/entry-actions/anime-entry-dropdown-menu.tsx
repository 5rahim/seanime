"use client"
import { Anime_MediaEntry } from "@/api/generated/types"
import { useAnimeEntryBulkAction, useOpenAnimeEntryInExplorer } from "@/api/hooks/anime_entries.hooks"
import { useStartDefaultMediaPlayer } from "@/api/hooks/mediaplayer.hooks"
import { serverStatusAtom } from "@/app/(main)/_atoms/server-status.atoms"
import {
    __bulkDeleteFilesModalIsOpenAtom,
    AnimeEntryBulkDeleteFilesModal,
} from "@/app/(main)/entry/_containers/entry-actions/anime-entry-bulk-delete-files-modal"
import { __metadataManager_isOpenAtom, AnimeEntryMetadataManager } from "@/app/(main)/entry/_containers/entry-actions/anime-entry-metadata-manager"
import { ConfirmationDialog, useConfirmationDialog } from "@/components/shared/confirmation-dialog"
import { IconButton } from "@/components/ui/button"
import { DropdownMenu, DropdownMenuItem, DropdownMenuLabel, DropdownMenuSeparator } from "@/components/ui/dropdown-menu"
import { useSetAtom } from "jotai"
import { useAtomValue } from "jotai/react"
import React from "react"
import { BiDotsVerticalRounded, BiRightArrowAlt } from "react-icons/bi"

export function AnimeEntryDropdownMenu({ entry }: { entry: Anime_MediaEntry }) {

    const serverStatus = useAtomValue(serverStatusAtom)
    const setIsMetadataManagerOpen = useSetAtom(__metadataManager_isOpenAtom)

    const { mutate: startDefaultMediaPlayer } = useStartDefaultMediaPlayer()
    const { mutate: openEntryInExplorer } = useOpenAnimeEntryInExplorer()

    const { mutate: performBulkAction, isPending } = useAnimeEntryBulkAction(entry.mediaId)

    const confirmDeleteFiles = useConfirmationDialog({
        title: "Unmatch all files",
        description: "Are you sure you want to unmatch all files?",
        onConfirm: () => {
            performBulkAction({
                mediaId: entry.mediaId,
                action: "unmatch",
            })
        },
    })

    const setBulkDeleteFilesModalOpen = useSetAtom(__bulkDeleteFilesModalIsOpenAtom)


    return (
        <>
            <DropdownMenu trigger={<IconButton icon={<BiDotsVerticalRounded />} intent="gray-basic" size="lg" />}>

                <DropdownMenuItem
                    onClick={() => openEntryInExplorer({ mediaId: entry.mediaId })}
                >
                    Open folder
                </DropdownMenuItem>

                {serverStatus?.settings?.mediaPlayer?.defaultPlayer != "mpv" && <DropdownMenuItem
                    onClick={() => startDefaultMediaPlayer}
                >
                    Start video player
                </DropdownMenuItem>}

                <DropdownMenuSeparator />
                <DropdownMenuItem
                    onClick={() => setIsMetadataManagerOpen(p => !p)}
                >
                    Metadata
                </DropdownMenuItem>

                <DropdownMenuSeparator />
                <DropdownMenuLabel>Bulk actions</DropdownMenuLabel>
                <DropdownMenuItem
                    className="text-red-500 dark:text-red-200 flex justify-between"
                    onClick={confirmDeleteFiles.open}
                    disabled={isPending}
                >
                    <span>Unmatch all files</span> <BiRightArrowAlt />
                </DropdownMenuItem>
                <DropdownMenuItem
                    className="text-red-500 dark:text-red-200 flex justify-between"
                    onClick={() => setBulkDeleteFilesModalOpen(true)}
                    disabled={isPending}
                >
                    <span>Delete some files</span> <BiRightArrowAlt />
                </DropdownMenuItem>
            </DropdownMenu>

            <AnimeEntryMetadataManager entry={entry} />
            <ConfirmationDialog {...confirmDeleteFiles} />
            <AnimeEntryBulkDeleteFilesModal entry={entry} />
        </>
    )
}
