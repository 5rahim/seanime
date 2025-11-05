import { Anime_Entry, HibikeTorrent_AnimeTorrent, HibikeTorrent_BatchEpisodeFiles } from "@/api/generated/types"
import { useGetTorrentstreamTorrentFilePreviews } from "@/api/hooks/torrentstream.hooks"
import { useAutoPlaySelectedTorrent } from "@/app/(main)/_features/autoplay/autoplay"
import { useTorrentSearchSelectedStreamEpisode } from "@/app/(main)/entry/_containers/torrent-search/_lib/handle-torrent-selection"
import { __torrentSearch_selectionAtom } from "@/app/(main)/entry/_containers/torrent-search/torrent-search-drawer"
import { useHandleStartTorrentStream } from "@/app/(main)/entry/_containers/torrent-stream/_lib/handle-torrent-stream"
import { FileTreeSelector } from "@/components/shared/file-tree-selector"
import { AppLayoutStack } from "@/components/ui/app-layout"
import { Button } from "@/components/ui/button"
import { LoadingSpinner } from "@/components/ui/loading-spinner"
import { ScrollArea } from "@/components/ui/scroll-area"
import { Vaul, VaulContent } from "@/components/vaul"
import { logger } from "@/lib/helpers/debug"
import { atom } from "jotai"
import { useAtom } from "jotai/react"
import React from "react"
import { IoPlayCircle } from "react-icons/io5"

const log = logger("TORRENT STREAM FILE SELECTION")

export const __torrentSearch_fileSelectionTorrentAtom = atom<HibikeTorrent_AnimeTorrent | undefined>(undefined)

export function TorrentstreamFileSelectionModal({ entry }: { entry: Anime_Entry }) {
    const [, setter] = useAtom(__torrentSearch_selectionAtom)

    const [selectedTorrent, setSelectedTorrent] = useAtom(__torrentSearch_fileSelectionTorrentAtom)

    const [selectedFileIdx, setSelectedFileIdx] = React.useState(-1)

    const { torrentSearchStreamEpisode } = useTorrentSearchSelectedStreamEpisode()

    const { data: filePreviews, isLoading } = useGetTorrentstreamTorrentFilePreviews({
        torrent: selectedTorrent,
        episodeNumber: torrentSearchStreamEpisode?.episodeNumber,
        media: entry.media,
    }, !!selectedTorrent)

    const { handleStreamSelection } = useHandleStartTorrentStream()

    const { setAutoPlayTorrent } = useAutoPlaySelectedTorrent()

    function onStream() {
        if (selectedFileIdx == -1 || !selectedTorrent || !torrentSearchStreamEpisode || !torrentSearchStreamEpisode.aniDBEpisode) return

        // save to autoplay
        // autoplay will increment selectedFileIdx by 1 to play the next file
        const batchFiles: HibikeTorrent_BatchEpisodeFiles = {
            current: selectedFileIdx,
            files: filePreviews?.map(n => { return { index: n.index, name: n.displayPath, path: n.path } }) || [],
            currentEpisodeNumber: torrentSearchStreamEpisode.episodeNumber,
            currentAniDBEpisode: torrentSearchStreamEpisode.aniDBEpisode,
        }
        log.info("Saving torrent for auto play", { batchFiles })
        setAutoPlayTorrent(selectedTorrent, entry, batchFiles)

        // start stream
        handleStreamSelection({
            torrent: selectedTorrent,
            mediaId: entry.mediaId,
            aniDBEpisode: torrentSearchStreamEpisode.aniDBEpisode,
            episodeNumber: torrentSearchStreamEpisode.episodeNumber,
            chosenFileIndex: selectedFileIdx,
            batchEpisodeFiles: batchFiles,
        })

        setSelectedTorrent(undefined)
        setSelectedFileIdx(-1)
        setter(undefined)
    }

    const hasLikelyMatch = filePreviews?.some(f => f.isLikely)
    const hasOneLikelyMatch = filePreviews?.filter(f => f.isLikely).length === 1

    const likelyMatchRef = React.useRef<HTMLDivElement>(null)

    const handleFileSelect = React.useCallback((value: string | number) => {
        setSelectedFileIdx(Number(value))
    }, [])

    const getFileValue = React.useCallback((filePreview: any) => {
        return filePreview.index
    }, [])

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
            }, 500)
            return () => clearTimeout(t)
        }
    }, [hasOneLikelyMatch, likelyMatchRef.current])


    return (
        <Vaul
            open={!!selectedTorrent}
            onOpenChange={open => {
                if (!open) {
                    setSelectedTorrent(undefined)
                    setSelectedFileIdx(-1)
                }
            }}
            // size="xl"
        >
            <VaulContent className="max-w-5xl mx-auto">
                <AppLayoutStack className="mt-4 p-3 lg:p-6">
                    {isLoading ? <LoadingSpinner /> : (
                        <AppLayoutStack className="pb-0">

                            <ScrollArea
                                viewportRef={scrollRef}
                                className="h-[80dvh] lg:h-[60dvh] overflow-y-auto p-4 border rounded-[--radius-md]"
                            >
                                <FileTreeSelector
                                    filePreviews={filePreviews || []}
                                    selectedValue={selectedFileIdx}
                                    onFileSelect={handleFileSelect}
                                    getFileValue={getFileValue}
                                    hasLikelyMatch={hasLikelyMatch || false}
                                    hasOneLikelyMatch={hasOneLikelyMatch || false}
                                    likelyMatchRef={likelyMatchRef}
                                />
                            </ScrollArea>

                            <Button
                                intent="primary"
                                className="w-full"
                                rightIcon={<IoPlayCircle className="text-xl" />}
                                disabled={selectedFileIdx === -1 || isLoading}
                                onClick={onStream}
                            >
                                Stream
                            </Button>

                        </AppLayoutStack>
                    )}
                </AppLayoutStack>
            </VaulContent>
        </Vaul>
    )

}
