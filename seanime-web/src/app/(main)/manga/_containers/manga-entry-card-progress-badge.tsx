import { __mangaLibrary_chapterCountsAtom } from "@/app/(main)/manga/_lib/handle-manga-collection"
import { Badge } from "@/components/ui/badge"
import { useAtom } from "jotai"
import React from "react"
import { IoBookOutline } from "react-icons/io5"

type MangaEntryCardProgressBadgeProps = {
    mediaId: number
    progress?: number
    progressTotal?: number
}

export function MangaEntryCardProgressBadge(props: MangaEntryCardProgressBadgeProps) {

    const {
        mediaId,
        progress: _progress,
        progressTotal: _progressTotal,
        ...rest
    } = props

    const [chapterCounts] = useAtom(__mangaLibrary_chapterCountsAtom)

    const [progress, setProgress] = React.useState(_progress)
    const [progressTotal, setProgressTotal] = React.useState(_progressTotal || 0)

    React.useEffect(() => {
        if (chapterCounts[mediaId]) {
            setProgressTotal(chapterCounts[mediaId])
        }
    }, [chapterCounts])

    if (!progress) return null

    const unread = progressTotal - progress

    if (unread <= 0) return null

    return (
        <div className="flex absolute text-sm left-0 top-0 w-full z-[5]">
            <Badge
                intent="unstyled"
                size="lg"
                className="text-base tracking-wide rounded-md rounded-tr-none rounded-bl-none border-0 bg-zinc-950/80 px-1.5 gap-0"
            >
                <span className="text-blue-100 font-normal">{unread}</span>&nbsp;&nbsp;<IoBookOutline className="" />
            </Badge>
        </div>
    )

    // return (
    //     <MediaEntryProgressBadge progress={progress} progressTotal={progressTotal} {...rest} />
    // )
}
