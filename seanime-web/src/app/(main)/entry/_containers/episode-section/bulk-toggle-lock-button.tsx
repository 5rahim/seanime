import { Tooltip } from "@/components/ui/tooltip"
import { MediaEntry } from "@/lib/server/types"
import { IconButton } from "@/components/ui/button"
import React from "react"
import { BiLockOpenAlt } from "@react-icons/all-files/bi/BiLockOpenAlt"
import { VscVerified } from "@react-icons/all-files/vsc/VscVerified"
import { useMediaEntryBulkAction } from "@/lib/server/hooks/library"

export function BulkToggleLockButton({ entry }: { entry: MediaEntry }) {

    const allLocked = entry.libraryData?.allFilesLocked

    const { toggleLock, isPending } = useMediaEntryBulkAction(entry.mediaId)

    return (
        <Tooltip trigger={
            <IconButton
                icon={entry.libraryData?.allFilesLocked ? <VscVerified/> : <BiLockOpenAlt/>}
                intent={allLocked ? "success-subtle" : "warning-subtle"}
                size={"xl"}
                className={"hover:opacity-60"}
                onClick={() => toggleLock(entry.mediaId)}
                isLoading={isPending}
            />
        }>
            {allLocked ? "Unlock all files" : "Lock all files"}
        </Tooltip>
    )

}