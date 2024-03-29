import { MangaChapterContainer, MangaChapterDetails, MangaEntry } from "@/app/(main)/manga/_lib/types"
import { Drawer } from "@/components/ui/drawer"
import { atom } from "jotai"
import { useAtom } from "jotai/react"
import React from "react"

type ChapterDrawerProps = {
    entry: MangaEntry
    chapterContainer: MangaChapterContainer
}

export const __manga_selectedChapterAtom = atom<MangaChapterDetails | undefined>(undefined)

export function ChapterDrawer(props: ChapterDrawerProps) {

    const {
        entry,
        chapterContainer,
        ...rest
    } = props

    const [selectedChapter, setSelectedChapter] = useAtom(__manga_selectedChapterAtom)

    return (
        <Drawer
            open={!!selectedChapter}
            onOpenChange={() => setSelectedChapter(undefined)}
            size="full"
            side="bottom"
            title={`${entry?.media?.title?.userPreferred} - ${selectedChapter?.title || ""}`}
        >


        </Drawer>
    )
}
