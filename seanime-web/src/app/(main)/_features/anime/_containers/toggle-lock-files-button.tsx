import { useAnimeEntryBulkAction } from "@/api/hooks/anime_entries.hooks"
import { IconButton, IconButtonProps } from "@/components/ui/button"
import { Tooltip } from "@/components/ui/tooltip"
import React, { memo } from "react"
import { BiLockOpenAlt } from "react-icons/bi"
import { VscVerified } from "react-icons/vsc"

type ToggleLockFilesButtonProps = {
    mediaId: number
    allFilesLocked: boolean
    size?: IconButtonProps["size"]
}

export const ToggleLockFilesButton = memo((props: ToggleLockFilesButtonProps) => {
    const { mediaId, allFilesLocked, size = "sm" } = props
    const [isLocked, setIsLocked] = React.useState(allFilesLocked)
    const { mutate: performBulkAction, isPending } = useAnimeEntryBulkAction(mediaId)

    const handleToggle = React.useCallback(() => {
        performBulkAction({
            mediaId,
            action: "toggle-lock",
        })
        setIsLocked(p => !p)
    }, [mediaId])

    return (
        <Tooltip
            trigger={
                <IconButton
                    icon={isLocked ? <VscVerified /> : <BiLockOpenAlt />}
                    intent={isLocked ? "success-subtle" : "warning-subtle"}
                    size={size}
                    className="hover:opacity-60"
                    loading={isPending}
                    onClick={handleToggle}
                />
            }
        >
            {isLocked ? "Unlock all files" : "Lock all files"}
        </Tooltip>
    )
})
