import { useMediaEntryBulkAction } from "@/app/(main)/(library)/_containers/bulk-actions/_lib/media-entry-bulk-actions"
import { IconButton } from "@/components/ui/button"
import { Tooltip } from "@/components/ui/tooltip"
import { MediaEntry } from "@/lib/server/types"
import React from "react"
import { BiLockOpenAlt } from "react-icons/bi"
import { VscVerified } from "react-icons/vsc"

export function BulkToggleLockButton({ entry }: { entry: MediaEntry }) {

    const allLocked = entry.libraryData?.allFilesLocked

    const { toggleLock, isPending } = useMediaEntryBulkAction(entry.mediaId)

    return (
        <Tooltip trigger={
            <IconButton
                icon={entry.libraryData?.allFilesLocked ? <VscVerified/> : <BiLockOpenAlt/>}
                intent={allLocked ? "success-subtle" : "warning-subtle"}
                size="xl"
                className="hover:opacity-60"
                onClick={() => toggleLock(entry.mediaId)}
                loading={isPending}
            />
        }>
            {allLocked ? "Unlock all files" : "Lock all files"}
        </Tooltip>
    )

}
