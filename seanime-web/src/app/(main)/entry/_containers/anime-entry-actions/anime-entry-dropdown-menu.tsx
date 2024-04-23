"use client"
import { useMediaEntryBulkAction } from "@/app/(main)/(library)/_containers/bulk-actions/_lib/media-entry-bulk-actions"

import { MediaEntry } from "@/app/(main)/(library)/_lib/anime-library.types"
import { serverStatusAtom } from "@/app/(main)/_atoms/server-status"
import { _bulkDeleteFilesModalIsOpenAtom, BulkDeleteFilesModal } from "@/app/(main)/entry/_containers/anime-entry-actions/bulk-delete-files-modal"
import { __metadataManager_isOpenAtom, MetadataManager } from "@/app/(main)/entry/_containers/metadata-manager/metadata-manager"
import { useOpenDefaultMediaPlayer } from "@/app/(main)/entry/_lib/media-player"
import { ConfirmationDialog, useConfirmationDialog } from "@/components/application/confirmation-dialog"
import { IconButton } from "@/components/ui/button"
import { DropdownMenu, DropdownMenuItem, DropdownMenuLabel, DropdownMenuSeparator } from "@/components/ui/dropdown-menu"
import { useOpenMediaEntryInExplorer } from "@/lib/server/hooks"
import { useSetAtom } from "jotai"
import { useAtomValue } from "jotai/react"
import React from "react"
import { BiDotsVerticalRounded, BiRightArrowAlt } from "react-icons/bi"

export function AnimeEntryDropdownMenu({ entry }: { entry: MediaEntry }) {

    const serverStatus = useAtomValue(serverStatusAtom)
    const setIsMetadataManagerOpen = useSetAtom(__metadataManager_isOpenAtom)

    const { startDefaultMediaPlayer } = useOpenDefaultMediaPlayer()
    const { openEntryInExplorer } = useOpenMediaEntryInExplorer()

    const { unmatchAll, isPending } = useMediaEntryBulkAction(entry.mediaId)

    const confirmDeleteFiles = useConfirmationDialog({
        title: "Unmatch all files",
        description: "Are you sure you want to unmatch all files?",
        onConfirm: () => {
            unmatchAll(entry.mediaId)
        },
    })

    const setBulkDeleteFilesModalOpen = useSetAtom(_bulkDeleteFilesModalIsOpenAtom)


    return (
        <>
            <DropdownMenu trigger={<IconButton icon={<BiDotsVerticalRounded />} intent="gray-basic" size="lg" />}>

                <DropdownMenuItem
                    onClick={() => openEntryInExplorer(entry.mediaId)}
                >
                    Open folder
                </DropdownMenuItem>

                {serverStatus?.settings?.mediaPlayer?.defaultPlayer != "mpv" && <DropdownMenuItem
                    onClick={startDefaultMediaPlayer}
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

            <MetadataManager entry={entry} />
            <ConfirmationDialog {...confirmDeleteFiles} />
            <BulkDeleteFilesModal entry={entry} />
        </>
    )
}
