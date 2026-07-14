import { __mangaLibrary_latestChapterNumbersAtom as __mangaLibrary_currentMangaDataAtom } from "@/app/(main)/manga/_lib/handle-manga-collection"
import { getMangaEntryUnreadState } from "@/app/(main)/manga/_lib/manga-unread"
import { Badge } from "@/components/ui/badge"
import { useThemeSettings } from "@/lib/theme/theme-hooks"
import { useAtom } from "jotai"
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
    } = props

    const { showMangaUnreadCount } = useThemeSettings()
    const [mangaData] = useAtom(__mangaLibrary_currentMangaDataAtom)

    if (!showMangaUnreadCount) return null

    const { known, unread } = getMangaEntryUnreadState(mediaId, progress || 0,
        mangaData.latestChapterNumbers, mangaData.storedProviders, mangaData.storedFilters)

    if (!known || unread <= 0) return null

    return (
        <div className="flex w-full z-[5]" data-manga-entry-card-unread-badge-container>
            <Badge
                intent="unstyled"
                size="lg"
                className="text-sm tracking-wide rounded-[--radius-md] flex gap-1 items-center rounded-tr-none rounded-bl-none border-0 px-1.5"
                data-manga-entry-card-unread-badge
            >
                <IoBookOutline className="text-sm" /><span className="text-[--foreground] font-normal">{unread}</span>
            </Badge>
        </div>
    )

}
