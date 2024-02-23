import { useMediaEntrySilenceStatus } from "@/app/(main)/entry/_lib/silence"
import { IconButton } from "@/components/ui/button"
import { Tooltip } from "@/components/ui/tooltip"
import React from "react"
import { LuBellOff, LuBellRing } from "react-icons/lu"

type MediaEntrySilenceToggleProps = {
    mediaId: number
}

export function MediaEntrySilenceToggle(props: MediaEntrySilenceToggleProps) {

    const {
        mediaId,
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
                    isLoading={silenceStatusIsLoading || silenceStatusIsUpdating}
                    intent={isSilenced ? "warning-subtle" : "primary-subtle"}
                    {...rest}
                />}
            >
                {isSilenced ? "Un-silence notifications" : "Silence notifications"}
            </Tooltip>
        </>
    )
}
