import { Anime_Entry } from "@/api/generated/types"
import { useDebridGetTorrentInfo } from "@/api/hooks/debrid.hooks"
import { useHandleStartDebridStream } from "@/app/(main)/entry/_containers/debrid-stream/_lib/handle-debrid-stream"
import { __torrentSearch_drawerIsOpenAtom } from "@/app/(main)/entry/_containers/torrent-search/torrent-search-drawer"
import { __torrentSearch_torrentstreamSelectedTorrentAtom } from "@/app/(main)/entry/_containers/torrent-stream/torrent-stream-file-selection-modal"
import { useTorrentStreamingSelectedEpisode } from "@/app/(main)/entry/_lib/torrent-streaming.atoms"
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

type DebridStreamFileSelectionModalProps = {
    entry: Anime_Entry
}

export function DebridStreamFileSelectionModal(props: DebridStreamFileSelectionModalProps) {

    const {
        entry,
    } = props

    const [, setter] = useAtom(__torrentSearch_drawerIsOpenAtom)

    const [selectedTorrent, setSelectedTorrent] = useAtom(__torrentSearch_torrentstreamSelectedTorrentAtom)

    const [selectedFileId, setSelectedFileIdx] = React.useState("")

    const { torrentStreamingSelectedEpisode } = useTorrentStreamingSelectedEpisode()

    const { data: torrentInfo, isLoading } = useDebridGetTorrentInfo({
        torrent: selectedTorrent!,
    }, !!selectedTorrent)

    const { handleStreamSelection } = useHandleStartDebridStream()

    function onStream() {
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

    const FileSelection = React.useCallback(() => {
        return <RadioGroup
            value={selectedFileId}
            onValueChange={v => setSelectedFileIdx(v)}
            options={(torrentInfo?.files?.toSorted((a, b) => a.path.localeCompare(b.path))?.map((f, i) => {
                return {
                    label: <div className="w-full">
                        <p className="mb-1 line-clamp-1">
                            {f.name}
                        </p>
                        <Tooltip trigger={<p className="font-normal line-clamp-1 text-sm text-[--muted]">{f.path}</p>}>
                            {f.path}
                        </Tooltip>
                    </div>,
                    value: String(f.id),
                }
            }) || [])}
            itemContainerClass={cn(
                "items-start cursor-pointer transition border-transparent rounded-[--radius] p-4 w-full",
                "bg-gray-50 hover:bg-[--subtle] dark:bg-gray-900",
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
    }, [torrentInfo, selectedFileId])

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
            title="Choose a file to stream"
        >
            {isLoading ? <LoadingSpinner title="Fetching torrent info..." /> : (
                <AppLayoutStack className="mt-4">

                    <div className="flex">
                        <div className="flex flex-1"></div>
                        <Button
                            intent="primary"
                            className=""
                            rightIcon={<IoPlayCircle className="text-xl" />}
                            disabled={selectedFileId === "" || isLoading}
                            onClick={onStream}
                        >
                            Stream
                        </Button>
                    </div>

                    <ScrollArea className="h-[75dvh] overflow-y-auto p-4 border rounded-md">
                        <FileSelection />
                    </ScrollArea>

                </AppLayoutStack>
            )}
        </Modal>
    )
}
