import { Anime_Entry, HibikeTorrent_BatchEpisodeFiles } from "@/api/generated/types"
import { useDebridGetTorrentFilePreviews } from "@/api/hooks/debrid.hooks"
import { useAutoPlaySelectedTorrent } from "@/app/(main)/_features/autoplay/autoplay"
import { useSelectedDebridService } from "@/app/(main)/_hooks/use-server-status"
import { useHandleStartDebridStream } from "@/app/(main)/entry/_containers/debrid-stream/_lib/handle-debrid-stream"
import { useTorrentSearchSelectedStreamEpisode } from "@/app/(main)/entry/_containers/torrent-search/_lib/handle-torrent-selection"
import { __torrentSearch_selectionAtom } from "@/app/(main)/entry/_containers/torrent-search/torrent-search-drawer"
import { __torrentSearch_fileSelectionTorrentAtom } from "@/app/(main)/entry/_containers/torrent-stream/torrent-stream-file-selection-modal"
import { FileTreeSelector } from "@/components/shared/file-tree-selector"
import { AppLayoutStack } from "@/components/ui/app-layout"
import { Button } from "@/components/ui/button"
import { LoadingSpinner } from "@/components/ui/loading-spinner"
import { ScrollArea } from "@/components/ui/scroll-area"
import { Vaul, VaulContent } from "@/components/vaul"
import { logger } from "@/lib/helpers/debug"
import { DEBRID_SERVICE } from "@/lib/server/settings"
import { useAtom } from "jotai/react"
import React from "react"
import { IoPlayCircle } from "react-icons/io5"

const log = logger("DEBRID STREAM FILE SELECTION")

type DebridStreamFileSelectionModalProps = {
    entry: Anime_Entry
}

export function DebridStreamFileSelectionModal(props: DebridStreamFileSelectionModalProps) {

    const {
        entry,
    } = props

    const [, setter] = useAtom(__torrentSearch_selectionAtom)

    const [selectedTorrent, setSelectedTorrent] = useAtom(__torrentSearch_fileSelectionTorrentAtom)

    const [selectedFileId, setSelectedFileIdx] = React.useState("")

    const { torrentSearchStreamEpisode } = useTorrentSearchSelectedStreamEpisode()

    const { selectedDebridService } = useSelectedDebridService()

    const { data: previews, isLoading } = useDebridGetTorrentFilePreviews({
        torrent: selectedTorrent!,
        media: entry.media,
        episodeNumber: torrentSearchStreamEpisode?.episodeNumber,
    }, !!selectedTorrent)

    const { setAutoPlayTorrent } = useAutoPlaySelectedTorrent()

    const { handleStreamSelection } = useHandleStartDebridStream()

    function onStream(selectedFileId: string) {
        if (selectedFileId == "" || !selectedTorrent || !torrentSearchStreamEpisode || !torrentSearchStreamEpisode.aniDBEpisode) return

        // save to autoplay
        // autoplay will increment selectedFileIdx by 1 to play the next file
        // Devnote: Torbox isn't supported because we can't use indexes to identify files
        let batchFiles: HibikeTorrent_BatchEpisodeFiles | undefined = undefined
        if (selectedDebridService !== DEBRID_SERVICE.TORBOX) {
            batchFiles = {
                current: parseInt(selectedFileId),
                files: previews?.map(n => { return { index: n.index, name: n.displayPath, path: n.path } }) || [],
                currentEpisodeNumber: torrentSearchStreamEpisode.episodeNumber,
                currentAniDBEpisode: torrentSearchStreamEpisode.aniDBEpisode,
            }
            log.info("Saving torrent for auto play", { batchFiles })
            setAutoPlayTorrent(selectedTorrent, entry, batchFiles)
        }

        handleStreamSelection({
            torrent: selectedTorrent,
            mediaId: entry.mediaId,
            aniDBEpisode: torrentSearchStreamEpisode.aniDBEpisode,
            episodeNumber: torrentSearchStreamEpisode.episodeNumber,
            chosenFileId: selectedFileId,
            batchEpisodeFiles: batchFiles,
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

    const handleFileSelect = React.useCallback((value: string | number) => {
        setSelectedFileIdx(String(value))
    }, [])

    const getFileValue = React.useCallback((filePreview: any) => {
        return String(filePreview.fileId)
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
            }, 1000) // Increased timeout to ensure DOM is ready
            return () => clearTimeout(t)
        }
    }, [hasOneLikelyMatch, likelyMatchRef.current])

    return (
        <Vaul
            open={!!selectedTorrent}
            onOpenChange={open => {
                if (!open) {
                    setSelectedTorrent(undefined)
                    setSelectedFileIdx("")
                }
            }}
        >
            <VaulContent className="max-w-5xl mx-auto">
                <AppLayoutStack className="mt-4 p-3 lg:p-6">
                    {(isLoading || previews?.length === 1) ? <LoadingSpinner
                        title={previews?.length === 1 ? "Launching stream..." : "Fetching torrent info..."}
                    /> : (
                        <AppLayoutStack className="mt-4">

                            <ScrollArea viewportRef={scrollRef} className="h-[75dvh] overflow-y-auto p-4 border rounded-[--radius-md]">
                                <FileTreeSelector
                                    filePreviews={previews || []}
                                    selectedValue={selectedFileId}
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
                                disabled={selectedFileId === "" || isLoading}
                                onClick={() => onStream(selectedFileId)}
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
