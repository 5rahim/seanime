import { Anime_MediaEntry } from "@/api/generated/types"
import { useAnimeEntryBulkAction } from "@/api/hooks/anime_entries.hooks"
import { IconButton } from "@/components/ui/button"
import { Tooltip } from "@/components/ui/tooltip"
import React from "react"
import { BiLockOpenAlt } from "react-icons/bi"
import { VscVerified } from "react-icons/vsc"

export function BulkToggleLockButton({ entry }: { entry: Anime_MediaEntry }) {

    const allLocked = entry.libraryData?.allFilesLocked

    const { mutate: performBulkAction, isPending } = useAnimeEntryBulkAction(entry.mediaId)

    function toggleLock(mediaId: number) {
        performBulkAction({
            mediaId,
            action: allLocked ? "unlock" : "lock",
        })
    }

    return (
        <Tooltip
            trigger={
                <IconButton
                    icon={entry.libraryData?.allFilesLocked ? <VscVerified /> : <BiLockOpenAlt />}
                    intent={allLocked ? "success-subtle" : "warning-subtle"}
                    size="lg"
                    className="hover:opacity-60"
                    onClick={() => toggleLock(entry.mediaId)}
                    loading={isPending}
                />
            }
        >
            {allLocked ? "Unlock all files" : "Lock all files"}
        </Tooltip>
    )

}
