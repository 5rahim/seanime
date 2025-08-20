import { Anime_Entry } from "@/api/generated/types"
import { useDebridGetTorrentFilePreviews } from "@/api/hooks/debrid.hooks"
import { useHandleStartDebridStream } from "@/app/(main)/entry/_containers/debrid-stream/_lib/handle-debrid-stream"
import { useTorrentSearchSelectedStreamEpisode } from "@/app/(main)/entry/_containers/torrent-search/_lib/handle-torrent-selection"
import { __torrentSearch_selectionAtom } from "@/app/(main)/entry/_containers/torrent-search/torrent-search-drawer"
import { __torrentSearch_torrentstreamSelectedTorrentAtom } from "@/app/(main)/entry/_containers/torrent-stream/torrent-stream-file-selection-modal"
import { AppLayoutStack } from "@/components/ui/app-layout"
import { Button } from "@/components/ui/button"
import { cn } from "@/components/ui/core/styling"
import { LoadingSpinner } from "@/components/ui/loading-spinner"
import { Modal } from "@/components/ui/modal"
import { RadioGroup } from "@/components/ui/radio-group"
import { ScrollArea } from "@/components/ui/scroll-area"
import { Tooltip } from "@/components/ui/tooltip"
import { useAtom } from "jotai/react"
import React from "react"
import { IoPlayCircle } from "react-icons/io5"
import { MdVerified } from "react-icons/md"

type DebridStreamFileSelectionModalProps = {
    entry: Anime_Entry
}

export function DebridStreamFileSelectionModal(props: DebridStreamFileSelectionModalProps) {

    const {
        entry,
    } = props

    const [, setter] = useAtom(__torrentSearch_selectionAtom)

    const [selectedTorrent, setSelectedTorrent] = useAtom(__torrentSearch_torrentstreamSelectedTorrentAtom)

    const [selectedFileId, setSelectedFileIdx] = React.useState("")

    const { torrentStreamingSelectedEpisode } = useTorrentSearchSelectedStreamEpisode()

    const { data: previews, isLoading } = useDebridGetTorrentFilePreviews({
        torrent: selectedTorrent!,
        media: entry.media,
        episodeNumber: torrentStreamingSelectedEpisode?.episodeNumber,
    }, !!selectedTorrent)

    const { handleStreamSelection } = useHandleStartDebridStream()

    function onStream(selectedFileId: string) {
        if (selectedFileId == "" || !selectedTorrent || !torrentStreamingSelectedEpisode || !torrentStreamingSelectedEpisode.aniDBEpisode) return

        handleStreamSelection({
            torrent: selectedTorrent,
            entry,
            aniDBEpisode: torrentStreamingSelectedEpisode.aniDBEpisode,
            episodeNumber: torrentStreamingSelectedEpisode.episodeNumber,
            chosenFileId: selectedFileId,
        })

        setSelectedTorrent(undefined)
        setSelectedFileIdx("")
        setter(undefined)
    }

    React.useEffect(() => {
        if (previews && previews.length === 1) {
            setSelectedFileIdx(String(previews[0].fileId))
            React.startTransition(() => {
                onStream(String(previews[0].fileId))
            })
        }
    }, [previews])

    const hasLikelyMatch = previews?.some(f => f.isLikely)
    const hasOneLikelyMatch = hasLikelyMatch && previews?.filter(f => f.isLikely).length === 1
    const likelyMatchRef = React.useRef<HTMLDivElement>(null)

    const FileSelection = React.useCallback(() => {
        return <RadioGroup
            value={selectedFileId}
            onValueChange={v => setSelectedFileIdx(v)}
            options={(previews?.toSorted((a, b) => a.path.localeCompare(b.path))?.map((f, i) => {
                return {
                    label: <div
                        className={cn(
                            "w-full",
                            (hasLikelyMatch && !f.isLikely) && "opacity-60",
                        )}
                        ref={hasOneLikelyMatch && f.isLikely ? likelyMatchRef : undefined}
                    >
                        <p className="mb-1 line-clamp-1">
                            {f.displayTitle}
                        </p>
                        {f.isLikely && <p className="flex items-center">
                            <MdVerified className="text-[--green] mr-1" />
                            <span className="text-white">Likely match</span>
                        </p>}
                        <Tooltip trigger={<p className="font-normal line-clamp-1 text-sm text-[--muted]">{f.displayPath}</p>}>
                            {f.path}
                        </Tooltip>
                    </div>,
                    value: String(f.fileId),
                }
            }) || [])}
            itemContainerClass={cn(
                "items-start cursor-pointer transition border-transparent rounded-[--radius] p-2 w-full",
                "hover:bg-[--subtle] bg-gray-900 hover:bg-gray-950",
                "data-[state=checked]:bg-white dark:data-[state=checked]:bg-gray-950",
                "focus:ring-2 ring-brand-100 dark:ring-brand-900 ring-offset-1 ring-offset-[--background] focus-within:ring-2 transition",
                "border border-transparent data-[state=checked]:border-[--brand] data-[state=checked]:ring-offset-0",
            )}
            itemClass={cn(
                "border-transparent absolute top-2 right-2 bg-transparent dark:bg-transparent dark:data-[state=unchecked]:bg-transparent",
                "data-[state=unchecked]:bg-transparent data-[state=unchecked]:hover:bg-transparent dark:data-[state=unchecked]:hover:bg-transparent",
                "focus-visible:ring-0 focus-visible:ring-offset-0 focus-visible:ring-offset-transparent",
            )}
            itemIndicatorClass="hidden"
            itemLabelClass="font-medium flex flex-col items-center data-[state=checked]:text-[--brand] cursor-pointer"
            stackClass="flex flex-col gap-2 space-y-0"
        />
    }, [previews, selectedFileId, hasLikelyMatch])

    const scrollRef = React.useRef<HTMLDivElement>(null)

    // Scroll to the likely match on mount
    React.useEffect(() => {
        if (hasOneLikelyMatch && likelyMatchRef.current && scrollRef.current) {
            const t = setTimeout(() => {
                const element = likelyMatchRef.current
                const container = scrollRef.current

                if (element && container) {
                    const elementRect = element.getBoundingClientRect()
                    const containerRect = container.getBoundingClientRect()

                    const scrollTop = elementRect.top - containerRect.top + container.scrollTop - 16 // 16px offset for padding

                    container.scrollTo({
                        top: scrollTop,
                        behavior: "smooth",
                    })
                }
            }, 1000) // Increased timeout to ensure DOM is ready
            return () => clearTimeout(t)
        }
    }, [hasOneLikelyMatch, likelyMatchRef.current])

    return (
        <Modal
            open={!!selectedTorrent}
            onOpenChange={open => {
                if (!open) {
                    setSelectedTorrent(undefined)
                    setSelectedFileIdx("")
                }
            }}
            // size="xl"
            contentClass="max-w-5xl"
            title={previews?.length !== 1 ? "Choose a file to stream" : "Launching stream..."}
        >
            {(isLoading || previews?.length === 1) ? <LoadingSpinner
                title={previews?.length === 1 ? "Launching stream..." : "Fetching torrent info..."}
            /> : (
                <AppLayoutStack className="mt-4">

                    <div className="flex">
                        <div className="flex flex-1"></div>
                        <Button
                            intent="primary"
                            className=""
                            rightIcon={<IoPlayCircle className="text-xl" />}
                            disabled={selectedFileId === "" || isLoading}
                            onClick={() => onStream(selectedFileId)}
                        >
                            Stream
                        </Button>
                    </div>

                    <ScrollArea viewportRef={scrollRef} className="h-[75dvh] overflow-y-auto p-4 border rounded-[--radius-md]">
                        <FileSelection />
                    </ScrollArea>

                </AppLayoutStack>
            )}
        </Modal>
    )
}
