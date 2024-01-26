"use client"
import { _bulkDeleteFilesModalIsOpenAtom, BulkDeleteFilesModal } from "@/app/(main)/entry/_containers/episode-section/bulk-delete-files-modal"
import { ConfirmationDialog, useConfirmationDialog } from "@/components/application/confirmation-dialog"
import { IconButton } from "@/components/ui/button"
import { DropdownMenu } from "@/components/ui/dropdown-menu"
import { useMediaEntryBulkAction } from "@/lib/server/hooks/library"
import { useOpenDefaultMediaPlayer, useOpenMediaEntryInExplorer } from "@/lib/server/hooks/settings"
import { MediaEntry } from "@/lib/server/types"
import { BiDotsVerticalRounded } from "@react-icons/all-files/bi/BiDotsVerticalRounded"
import { BiRightArrowAlt } from "@react-icons/all-files/bi/BiRightArrowAlt"
import { useSetAtom } from "jotai"
import React from "react"

export function EpisodeSectionMenu({ entry }: { entry: MediaEntry }) {

    const { startDefaultMediaPlayer } = useOpenDefaultMediaPlayer()
    const { openEntryInExplorer } = useOpenMediaEntryInExplorer()

    // const bulkOffsetEpisodeModal = useBoolean(false)

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
            <DropdownMenu trigger={<IconButton icon={<BiDotsVerticalRounded/>} intent={"gray-basic"} size={"xl"}/>}>
                <DropdownMenu.Item
                    onClick={() => openEntryInExplorer(entry.mediaId)}
                >
                    Open folder
                </DropdownMenu.Item>
                <DropdownMenu.Item
                    onClick={startDefaultMediaPlayer}
                >
                    Start video player
                </DropdownMenu.Item>
                <DropdownMenu.Divider/>
                <DropdownMenu.Group title="Bulk actions">
                    {/*<DropdownMenu.Item*/}
                    {/*    onClick={bulkOffsetEpisodeModal.toggle}*/}
                    {/*>*/}
                    {/*    Offset episode numbers*/}
                    {/*</DropdownMenu.Item>*/}
                    <DropdownMenu.Item
                        className="text-red-500 dark:text-red-200 flex justify-between"
                        onClick={confirmDeleteFiles.open}
                        disabled={isPending}
                    >
                        <span>Unmatch all files</span> <BiRightArrowAlt />
                    </DropdownMenu.Item>
                    <DropdownMenu.Item
                        className="text-red-500 dark:text-red-200 flex justify-between"
                        onClick={() => setBulkDeleteFilesModalOpen(true)}
                        disabled={isPending}
                    >
                        <span>Delete some files</span> <BiRightArrowAlt />
                    </DropdownMenu.Item>
                </DropdownMenu.Group>
            </DropdownMenu>

            {/*<BulkOffsetEpisodesModal entry={entry} isOpen={bulkOffsetEpisodeModal.active} onClose={bulkOffsetEpisodeModal.off} />*/}
            <ConfirmationDialog {...confirmDeleteFiles} />
            <BulkDeleteFilesModal entry={entry} />
        </>
    )
}
