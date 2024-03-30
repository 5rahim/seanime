import { useMangaChapterContainer } from "@/app/(main)/manga/_lib/queries"
import { MangaChapterDetails, MangaEntry } from "@/app/(main)/manga/_lib/types"
import { __manga_selectedChapterAtom, ChapterDrawer } from "@/app/(main)/manga/entry/_containers/chapter-drawer"
import { Accordion, AccordionContent, AccordionItem, AccordionTrigger } from "@/components/ui/accordion"
import { IconButton } from "@/components/ui/button"
import { Card } from "@/components/ui/card"
import { cn } from "@/components/ui/core/styling"
import { LoadingSpinner } from "@/components/ui/loading-spinner"
import { MangaDetailsByIdQuery } from "@/lib/anilist/gql/graphql"
import { useSetAtom } from "jotai/react"
import React from "react"
import { BiBookAlt } from "react-icons/bi"
import { FaBookOpenReader } from "react-icons/fa6"

type ChaptersListProps = {
    mediaId: string | null
    entry: MangaEntry
    details?: MangaDetailsByIdQuery["Media"]
}

export function ChaptersList(props: ChaptersListProps) {

    const {
        mediaId,
        entry,
        details,
        ...rest
    } = props

    const { chapterContainer, chapterIdToNumbersMap, chapterContainerLoading } = useMangaChapterContainer(mediaId)

    const retainUnreadChapters = React.useCallback((chapter: MangaChapterDetails) => {
        if (!entry.listData || !chapterIdToNumbersMap.has(chapter.id) || !entry.listData?.progress) return true

        const chapterNumber = chapterIdToNumbersMap.get(chapter.id)
        return chapterNumber && chapterNumber > entry.listData?.progress
    }, [chapterIdToNumbersMap, chapterContainer, entry])

    if (!chapterContainer || chapterContainerLoading) return <LoadingSpinner />

    return (
        <div
            className="space-y-2"
        >

            <Accordion
                type="single"
                className=""
                triggerClass="text-[--muted] dark:data-[state=open]:text-white px-0 dark:hover:bg-transparent hover:bg-transparent dark:hover:text-white hover:text-black"
                itemClass="border-b"
                contentClass="pb-8"
                collapsible
            >
                <AccordionItem value="all">
                    <AccordionTrigger>
                        <h3 className="flex gap-2 items-center"><BiBookAlt className="text-gray-300" /> All chapters</h3>
                    </AccordionTrigger>
                    <AccordionContent className="p-0 py-4 space-y-2">
                        {chapterContainer?.chapters?.toReversed()?.map((chapter, index) => (
                            <ChapterItem chapter={chapter} key={chapter.id} />
                        ))}
                    </AccordionContent>
                </AccordionItem>
            </Accordion>


            <h3>Unread chapters</h3>
            {chapterContainer?.chapters?.filter(ch => retainUnreadChapters(ch)).map((chapter, index) => (
                <ChapterItem chapter={chapter} key={chapter.id} />
            ))}

            <ChapterDrawer entry={entry} chapterContainer={chapterContainer} />
        </div>
    )
}


type ChapterItemProps = {
    chapter: MangaChapterDetails
}

export function ChapterItem(props: ChapterItemProps) {

    const {
        chapter,
        ...rest
    } = props

    const setSelectedChapter = useSetAtom(__manga_selectedChapterAtom)

    return (
        <>
            <Card
                key={chapter.id}
                className={cn(
                    "p-3 flex w-full gap-2 items-center",
                    "hover:bg-[--subtle]",
                )}
            >
                <p>{chapter.title}</p>
                <div className="flex flex-1"></div>
                <IconButton
                    intent="gray-basic"
                    size="sm"
                    onClick={() => setSelectedChapter(chapter)}
                    icon={<FaBookOpenReader />}
                />
            </Card>
        </>
    )
}
