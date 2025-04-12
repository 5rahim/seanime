import { usePlaybackPlayRandomVideo } from "@/api/hooks/playback_manager.hooks"
import { IconButton } from "@/components/ui/button"
import { Tooltip } from "@/components/ui/tooltip"
import React from "react"
import { LiaRandomSolid } from "react-icons/lia"

type PlayRandomEpisodeButtonProps = {
    children?: React.ReactNode
}

export function PlayRandomEpisodeButton(props: PlayRandomEpisodeButtonProps) {

    const {
        children,
        ...rest
    } = props

    const { mutate: playRandom, isPending } = usePlaybackPlayRandomVideo()

    return (
        <>
            <Tooltip
                trigger={<IconButton
                    data-play-random-episode-button
                    intent={"white-subtle"}
                    icon={<LiaRandomSolid className="text-2xl" />}
                    loading={isPending}
                    onClick={() => playRandom()}
                />}
            >Play random anime</Tooltip>
        </>
    )
}
