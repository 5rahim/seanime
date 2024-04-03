import { MangaChapterDetails, MangaEntry } from "@/app/(main)/manga/_lib/manga.types"
import { __manga_selectedChapterAtom } from "@/app/(main)/manga/entry/_containers/chapter-drawer/chapter-drawer"
import { SeaEndpoints } from "@/lib/server/endpoints"
import { useSeaMutation } from "@/lib/server/query"
import { useAtomValue } from "jotai/react"
import React from "react"

export type DiscordPresenceRoute_Body = {
    title: string
    image: string
    chapter: string
}

export function useDiscordMangaPresence(entry: MangaEntry | undefined) {
    const currentChapter = useAtomValue(__manga_selectedChapterAtom)

    const { mutate } = useSeaMutation<boolean, DiscordPresenceRoute_Body>({
        endpoint: SeaEndpoints.DISCORD_PRESENCE_MANGA,
    })
    const { mutate: cancelActivity } = useSeaMutation({
        endpoint: SeaEndpoints.DISCORD_PRESENCE_CANCEL,
    })

    const prevChapter = React.useRef<MangaChapterDetails | undefined>()

    React.useEffect(() => {
        if (currentChapter && entry) {
            mutate({
                title: entry.media.title?.userPreferred || entry.media.title?.romaji || entry.media.title?.english || "Reading",
                image: entry.media.coverImage?.large || entry.media.coverImage?.medium || "",
                chapter: currentChapter.chapter,
            })
        }

        if (!currentChapter && prevChapter.current) {
            cancelActivity()
        }

        prevChapter.current = currentChapter
    }, [currentChapter, entry])
}
