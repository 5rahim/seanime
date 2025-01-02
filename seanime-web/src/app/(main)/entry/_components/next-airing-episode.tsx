import { AL_BaseAnime } from "@/api/generated/types"
import { Badge } from "@/components/ui/badge"
import { cn } from "@/components/ui/core/styling"
import { ThemeMediaPageInfoBoxSize, useThemeSettings } from "@/lib/theme/hooks"
import { addSeconds, format, formatDistanceToNow } from "date-fns"
import React from "react"

export function NextAiringEpisode(props: { media: AL_BaseAnime }) {
    const distance = formatDistanceToNow(addSeconds(new Date(), props.media.nextAiringEpisode?.timeUntilAiring || 0), { addSuffix: true })
    const day = format(addSeconds(new Date(), props.media.nextAiringEpisode?.timeUntilAiring || 0), "EEEE")
    const ts = useThemeSettings()
    return <>
        {!!props.media.nextAiringEpisode && (
            <div
                className={cn(
                    "flex gap-2 items-center justify-center",
                    ts.mediaPageBannerInfoBoxSize === ThemeMediaPageInfoBoxSize.FullWidth && "justify-start",
                )}
            >
                <p className="text-xl min-[2000px]:text-xl">Episode <Badge
                    size="lg"
                    className="rounded-md px-1.5"
                >{props.media.nextAiringEpisode?.episode}</Badge> {distance} <span
                    className="text-[--muted]"
                >({day})</span></p>
            </div>
        )}
    </>
}
