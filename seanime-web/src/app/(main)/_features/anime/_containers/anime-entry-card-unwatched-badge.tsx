import { AL_BaseAnime, Anime_EntryLibraryData, Nullish } from "@/api/generated/types"
import { Badge } from "@/components/ui/badge"
import { cn } from "@/components/ui/core/styling"
import { anilist_getCurrentEpisodes } from "@/lib/helpers/media"
import { useThemeSettings } from "@/lib/theme/hooks"
import React from "react"
import { MdOutlinePlayCircleOutline } from "react-icons/md"

type AnimeEntryCardUnwatchedBadgeProps = {
    progress: number
    media: Nullish<AL_BaseAnime>
    libraryData: Nullish<Anime_EntryLibraryData>
}

export function AnimeEntryCardUnwatchedBadge(props: AnimeEntryCardUnwatchedBadgeProps) {

    const {
        media,
        libraryData,
        progress,
        ...rest
    } = props

    const { showAnimeUnwatchedCount } = useThemeSettings()

    if (!showAnimeUnwatchedCount) return null

    const progressTotal = anilist_getCurrentEpisodes(media)
    const unwatched = progressTotal - (progress ?? 0)

    const unwatchedFromLibrary = libraryData?.unwatchedCount ?? 0
    const isInLibrary = !!libraryData?.mainFileCount

    const unwatchedCount = isInLibrary ? unwatchedFromLibrary : unwatched

    if (unwatchedCount <= 0) return null

    return (
        <div
            data-anime-entry-card-unwatched-badge-container
            className={cn(
                "flex w-full z-[5]",
            )}
        >
            <Badge
                intent="unstyled"
                size="lg"
                className="text-sm tracking-wide flex gap-1 items-center rounded-[--radius-md] border-0 bg-transparent px-1.5"
                data-anime-entry-card-unwatched-badge
            >
                <MdOutlinePlayCircleOutline className="text-lg" /><span className="text-[--foreground] font-normal">{unwatchedCount}</span>
            </Badge>
        </div>
    )

    // return (
    //     <MediaEntryProgressBadge progress={progress} progressTotal={progressTotal} {...rest} />
    // )
}
