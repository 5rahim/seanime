import { useSyncAddMedia, useSyncGetIsMediaTracked, useSyncRemoveMedia } from "@/api/hooks/sync.hooks"
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

    const { data: isTracked, isLoading } = useSyncGetIsMediaTracked(mediaId, type)
    const { mutate: addMedia, isPending: isAdding } = useSyncAddMedia()
    const { mutate: removeMedia, isPending: isRemoving } = useSyncRemoveMedia()

    function handleToggle() {
        if (isTracked) {
            removeMedia({ mediaId, type })
        } else {
            addMedia({ mediaId, type })
        }
    }

    return (
        <>
            <Tooltip
                trigger={<IconButton
                    icon={isTracked ? <MdOutlineOfflinePin /> : <MdOutlineDownloadForOffline />}
                    onClick={handleToggle}
                    loading={isLoading || isAdding || isRemoving}
                    intent={isTracked ? "primary-basic" : "gray-subtle"}
                    size={size}
                    {...rest}
                />}
            >
                {isTracked ? `Un-track ${type}` : `Sync offline`}
            </Tooltip>
        </>
    )
}
