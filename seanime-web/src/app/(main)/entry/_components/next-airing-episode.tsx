import { AL_BaseAnime } from "@/api/generated/types"
import { cn } from "@/components/ui/core/styling"
import { ThemeMediaPageInfoBoxSize, useThemeSettings } from "@/lib/theme/hooks"
import { addSeconds, format, formatDistanceToNow } from "date-fns"
import React from "react"
import { BiCalendarAlt } from "react-icons/bi"

export function NextAiringEpisode(props: { media: AL_BaseAnime }) {
    const distance = formatDistanceToNow(addSeconds(new Date(), props.media.nextAiringEpisode?.timeUntilAiring || 0), { addSuffix: true })
    const day = format(addSeconds(new Date(), props.media.nextAiringEpisode?.timeUntilAiring || 0), "EEEE")
    const ts = useThemeSettings()
    return <>
        {!!props.media.nextAiringEpisode && (
            <div
                className={cn(
                    "flex gap-2 items-center justify-center text-lg",
                    ts.mediaPageBannerInfoBoxSize === ThemeMediaPageInfoBoxSize.Fluid && "justify-start",
                )}
            >
                <span className="font-semibold">Episode {props.media.nextAiringEpisode?.episode}</span> {distance}
                <BiCalendarAlt className="text-lg text-[--muted]" />
                <span className="text-[--muted]">{day}</span>
            </div>
        )}
    </>
}
