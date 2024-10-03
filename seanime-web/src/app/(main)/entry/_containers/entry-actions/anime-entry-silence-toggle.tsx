import { useGetAnimeEntrySilenceStatus, useToggleAnimeEntrySilenceStatus } from "@/api/hooks/anime_entries.hooks"
import { IconButton } from "@/components/ui/button"
import { Tooltip } from "@/components/ui/tooltip"
import React from "react"
import { LuBellOff, LuBellRing } from "react-icons/lu"

type AnimeEntrySilenceToggleProps = {
    mediaId: number
    size?: "sm" | "md" | "lg"
}

export function AnimeEntrySilenceToggle(props: AnimeEntrySilenceToggleProps) {

    const {
        mediaId,
        size = "lg",
        ...rest
    } = props

    const { isSilenced, isLoading } = useGetAnimeEntrySilenceStatus(mediaId)

    const {
        mutate,
        isPending,
    } = useToggleAnimeEntrySilenceStatus()

    function handleToggleSilenceStatus() {
        mutate({ mediaId })
    }

    return (
        <>
            <Tooltip
                trigger={<IconButton
                    icon={isSilenced ? <LuBellOff /> : <LuBellRing />}
                    onClick={handleToggleSilenceStatus}
                    loading={isLoading || isPending}
                    intent={isSilenced ? "warning-subtle" : "gray-subtle"}
                    size={size}
                    {...rest}
                />}
            >
                {isSilenced ? "Un-silence notifications" : "Silence notifications"}
            </Tooltip>
        </>
    )
}
