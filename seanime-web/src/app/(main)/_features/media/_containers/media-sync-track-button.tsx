import { useLocalAddTrackedMedia, useLocalGetIsMediaTracked, useLocalRemoveTrackedMedia } from "@/api/hooks/local.hooks"
import { ConfirmationDialog, useConfirmationDialog } from "@/components/shared/confirmation-dialog"
import { IconButton } from "@/components/ui/button"
import { Tooltip } from "@/components/ui/tooltip"
import React from "react"
import { MdOutlineDownloadForOffline, MdOutlineOfflinePin } from "react-icons/md"

type MediaSyncTrackButtonProps = {
    mediaId: number
    type: "anime" | "manga"
    size?: "sm" | "md" | "lg"
}

export function MediaSyncTrackButton(props: MediaSyncTrackButtonProps) {

    const {
        mediaId,
        type,
        size,
        ...rest
    } = props

    const { data: isTracked, isLoading } = useLocalGetIsMediaTracked(mediaId, type)
    const { mutate: addMedia, isPending: isAdding } = useLocalAddTrackedMedia()
    const { mutate: removeMedia, isPending: isRemoving } = useLocalRemoveTrackedMedia()

    function handleToggle() {
        if (isTracked) {
            removeMedia({ mediaId, type })
        } else {
            addMedia({
                media: [{
                    mediaId: mediaId,
                    type: type,
                }],
            })
        }
    }

    const confirmUntrack = useConfirmationDialog({
        title: "Remove offline data",
        description: "This action will remove the offline data for this media entry. Are you sure you want to proceed?",
        onConfirm: () => {
            handleToggle()
        },
    })

    return (
        <>
            <Tooltip
                trigger={<IconButton
                    icon={isTracked ? <MdOutlineOfflinePin /> : <MdOutlineDownloadForOffline />}
                    onClick={() => isTracked ? confirmUntrack.open() : handleToggle()}
                    loading={isLoading || isAdding || isRemoving}
                    intent={isTracked ? "primary-subtle" : "gray-subtle"}
                    size={size}
                    {...rest}
                />}
            >
                {isTracked ? `Remove offline data` : `Save locally`}
            </Tooltip>

            <ConfirmationDialog {...confirmUntrack} />
        </>
    )
}
