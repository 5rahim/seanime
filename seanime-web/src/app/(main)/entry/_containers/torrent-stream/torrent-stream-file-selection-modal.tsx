import { Anime_Entry, HibikeTorrent_AnimeTorrent } from "@/api/generated/types"
import { useGetTorrentstreamTorrentFilePreviews } from "@/api/hooks/torrentstream.hooks"
import { useTorrentSearchSelectedStreamEpisode } from "@/app/(main)/entry/_containers/torrent-search/_lib/handle-torrent-selection"
import { __torrentSearch_selectionAtom } from "@/app/(main)/entry/_containers/torrent-search/torrent-search-drawer"
import { useHandleStartTorrentStream } from "@/app/(main)/entry/_containers/torrent-stream/_lib/handle-torrent-stream"
import { AppLayoutStack } from "@/components/ui/app-layout"
import { Button } from "@/components/ui/button"
import { cn } from "@/components/ui/core/styling"
import { LoadingSpinner } from "@/components/ui/loading-spinner"
import { Modal } from "@/components/ui/modal"
import { RadioGroup } from "@/components/ui/radio-group"
import { ScrollArea } from "@/components/ui/scroll-area"
import { Tooltip } from "@/components/ui/tooltip"
import { atom } from "jotai"
import { useAtom } from "jotai/react"
import React from "react"
import { IoPlayCircle } from "react-icons/io5"
import { MdVerified } from "react-icons/md"

export const __torrentSearch_torrentstreamSelectedTorrentAtom = atom<HibikeTorrent_AnimeTorrent | undefined>(undefined)

export function TorrentstreamFileSelectionModal({ entry }: { entry: Anime_Entry }) {
    const [, setter] = useAtom(__torrentSearch_selectionAtom)

    const [selectedTorrent, setSelectedTorrent] = useAtom(__torrentSearch_torrentstreamSelectedTorrentAtom)

    const [selectedFileIdx, setSelectedFileIdx] = React.useState(-1)

    const { torrentStreamingSelectedEpisode } = useTorrentSearchSelectedStreamEpisode()

    const { data: filePreviews, isLoading } = useGetTorrentstreamTorrentFilePreviews({
        torrent: selectedTorrent,
        episodeNumber: torrentStreamingSelectedEpisode?.episodeNumber,
        media: entry.media,
    }, !!selectedTorrent)

    const { handleManualTorrentStreamSelection } = useHandleStartTorrentStream()

    function onStream() {
        if (selectedFileIdx == -1 || !selectedTorrent || !torrentStreamingSelectedEpisode || !torrentStreamingSelectedEpisode.aniDBEpisode) return

        handleManualTorrentStreamSelection({
            torrent: selectedTorrent,
            entry,
            aniDBEpisode: torrentStreamingSelectedEpisode.aniDBEpisode,
            episodeNumber: torrentStreamingSelectedEpisode.episodeNumber,
            chosenFileIndex: selectedFileIdx,
        })

        setSelectedTorrent(undefined)
        setSelectedFileIdx(-1)
        setter(undefined)
    }

    const hasLikelyMatch = filePreviews?.some(f => f.isLikely)
    const hasOneLikelyMatch = filePreviews?.filter(f => f.isLikely).length === 1

    const likelyMatchRef = React.useRef<HTMLDivElement>(null)

    const FileSelection = React.useCallback(() => {
        return <RadioGroup
            value={String(selectedFileIdx)}
            onValueChange={v => setSelectedFileIdx(Number(v))}
            options={(filePreviews?.toSorted((a, b) => a.path.localeCompare(b.path))?.map((f, i) => {
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
                    value: String(f.index),
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
    }, [filePreviews, selectedFileIdx])

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
                    setSelectedFileIdx(-1)
                }
            }}
            // size="xl"
            contentClass="max-w-5xl"
            title="Choose a file to stream"
        >
            <AppLayoutStack className="mt-4">
                {isLoading ? <LoadingSpinner /> : (
                    <AppLayoutStack className="pb-0">

                        <div className="flex">
                            <div className="flex flex-1"></div>
                            <Button
                                intent="primary"
                                className=""
                                rightIcon={<IoPlayCircle className="text-xl" />}
                                disabled={selectedFileIdx === -1 || isLoading}
                                onClick={onStream}
                            >
                                Stream
                            </Button>
                        </div>

                        <ScrollArea
                            viewportRef={scrollRef}
                            className="h-[75dvh] overflow-y-auto p-4 border rounded-[--radius-md]"
                        >
                            <FileSelection />
                        </ScrollArea>

                    </AppLayoutStack>
                )}
            </AppLayoutStack>
        </Modal>
    )

}
