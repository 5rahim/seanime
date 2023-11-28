import { MediaEntry } from "@/lib/server/types"
import { ConfirmationDialog, useConfirmationDialog } from "@/components/application/confirmation-dialog"
import { IconButton } from "@/components/ui/button"
import React from "react"
import { DropdownMenu } from "@/components/ui/dropdown-menu"
import { BiDotsVerticalRounded } from "@react-icons/all-files/bi/BiDotsVerticalRounded"
import { useOpenDefaultMediaPlayer, useOpenMediaEntryInExplorer } from "@/lib/server/hooks/settings"
import { useMediaEntryBulkAction } from "@/lib/server/hooks/library"

export function EpisodeSectionMenu({ entry }: { entry: MediaEntry }) {

    const { startDefaultMediaPlayer } = useOpenDefaultMediaPlayer()
    const { openEntryInExplorer } = useOpenMediaEntryInExplorer()

    // const bulkOffsetEpisodeModal = useBoolean(false)

    const { unmatchAll, isPending } = useMediaEntryBulkAction(entry.mediaId)

    const confirmUnmatch = useConfirmationDialog({
        title: "Unmatch all files",
        description: "Are you sure you want to unmatch all files?",
        onConfirm: () => {
            unmatchAll(entry.mediaId)
        },
    })


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
                        className="text-red-500 dark:text-red-200"
                        onClick={confirmUnmatch.open}
                        disabled={isPending}
                    >
                        Unmatch all files
                    </DropdownMenu.Item>
                </DropdownMenu.Group>
            </DropdownMenu>

            {/*<BulkOffsetEpisodesModal entry={entry} isOpen={bulkOffsetEpisodeModal.active} onClose={bulkOffsetEpisodeModal.off} />*/}
            <ConfirmationDialog {...confirmUnmatch} />
        </>
    )
}