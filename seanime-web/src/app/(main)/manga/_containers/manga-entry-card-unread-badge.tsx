import { __mangaLibrary_chapterCountsAtom } from "@/app/(main)/manga/_lib/handle-manga-collection"
import { Badge } from "@/components/ui/badge"
import { useThemeSettings } from "@/lib/theme/hooks"
import { useAtom } from "jotai"
import React from "react"
import { IoBookOutline } from "react-icons/io5"

type MangaEntryCardUnreadBadgeProps = {
    mediaId: number
    progress?: number
    progressTotal?: number
}

export function MangaEntryCardUnreadBadge(props: MangaEntryCardUnreadBadgeProps) {

    const {
        mediaId,
        progress,
        progressTotal: _progressTotal,
        ...rest
    } = props

    const { showMangaUnreadCount } = useThemeSettings()
    const [chapterCounts] = useAtom(__mangaLibrary_chapterCountsAtom)

    const [progressTotal, setProgressTotal] = React.useState(_progressTotal || 0)

    React.useEffect(() => {
        if (chapterCounts[mediaId]) {
            setProgressTotal(chapterCounts[mediaId])
        }
    }, [chapterCounts])

    if (!showMangaUnreadCount) return null

    const unread = progressTotal - (progress || 0)

    if (unread <= 0) return null

    return (
        <div className="flex w-full z-[5]">
            <Badge
                intent="unstyled"
                size="lg"
                className="text-sm tracking-wide rounded-[--radius-md] flex gap-1 items-center rounded-tr-none rounded-bl-none border-0 px-1.5"
            >
                <IoBookOutline className="text-sm" /><span className="text-[--foreground] font-normal">{unread}</span>
            </Badge>
        </div>
    )

    // return (
    //     <MediaEntryProgressBadge progress={progress} progressTotal={progressTotal} {...rest} />
    // )
}
