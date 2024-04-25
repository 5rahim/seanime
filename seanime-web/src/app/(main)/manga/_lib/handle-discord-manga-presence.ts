import { AL_BaseManga } from "@/api/generated/types"
import { useCancelDiscordActivity, useSetDiscordMangaActivity } from "@/api/hooks/discord.hooks"
import { useServerStatus } from "@/app/(main)/_hooks/use-server-status"
import { __manga_selectedChapterAtom } from "@/app/(main)/manga/_containers/chapter-reader/chapter-reader-drawer"
import { useAtomValue } from "jotai/react"
import React from "react"

export function useDiscordMangaPresence(entry: { media?: AL_BaseManga } | undefined) {
    const serverStatus = useServerStatus()
    const currentChapter = useAtomValue(__manga_selectedChapterAtom)

    const { mutate } = useSetDiscordMangaActivity()
    const { mutate: cancelActivity } = useCancelDiscordActivity()

    const prevChapter = React.useRef<any>()

    React.useEffect(() => {
        if (serverStatus?.isOffline) return
        if (
            serverStatus?.settings?.discord?.enableRichPresence &&
            serverStatus?.settings?.discord?.enableMangaRichPresence
        ) {

            if (currentChapter && entry) {
                mutate({
                    title: entry.media?.title?.userPreferred || entry.media?.title?.romaji || entry.media?.title?.english || "Reading",
                    image: entry.media?.coverImage?.large || entry.media?.coverImage?.medium || "",
                    chapter: currentChapter.chapterNumber,
                })
            }

            if (!currentChapter) {
                cancelActivity()
            }
        }

        prevChapter.current = currentChapter
    }, [currentChapter, entry])
}
