import { useMediaEntrySilenceStatus } from "@/app/(main)/entry/_lib/silence"
import { IconButton } from "@/components/ui/button"
import { Tooltip } from "@/components/ui/tooltip"
import React from "react"
import { LuBellOff, LuBellRing } from "react-icons/lu"

type MediaEntrySilenceToggleProps = {
    mediaId: number
    size?: "sm" | "md" | "lg"
}

export function MediaEntrySilenceToggle(props: MediaEntrySilenceToggleProps) {

    const {
        mediaId,
        size = "lg",
        ...rest
    } = props

    const {
        isSilenced,
        silenceStatusIsLoading,
        toggleSilenceStatus,
        silenceStatusIsUpdating,
    } = useMediaEntrySilenceStatus(mediaId)

    return (
        <>
            <Tooltip
                trigger={<IconButton
                    icon={isSilenced ? <LuBellOff /> : <LuBellRing />}
                    onClick={toggleSilenceStatus}
                    loading={silenceStatusIsLoading || silenceStatusIsUpdating}
                    intent={isSilenced ? "warning-subtle" : "primary-subtle"}
                    size={size}
                    {...rest}
                />}
            >
                {isSilenced ? "Un-silence notifications" : "Silence notifications"}
            </Tooltip>
        </>
    )
}
