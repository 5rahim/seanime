import { Anime_Entry, HibikeTorrent_AnimeTorrent } from "@/api/generated/types"
import { useTorrentClientDownload, useTorrentClientGetFiles } from "@/api/hooks/torrent_client.hooks"
import { useLibraryPathSelection } from "@/app/(main)/_hooks/use-library-path-selection"
import { useServerStatus } from "@/app/(main)/_hooks/use-server-status"
import { __torrentSearch_selectedTorrentsAtom } from "@/app/(main)/entry/_containers/torrent-search/torrent-search-container"
import { __torrentSearch_selectionAtom } from "@/app/(main)/entry/_containers/torrent-search/torrent-search-drawer"
import { DirectorySelector } from "@/components/shared/directory-selector"
import { FileTreeMultiSelector } from "@/components/shared/file-tree-selector"
import { AppLayoutStack } from "@/components/ui/app-layout"
import { Button } from "@/components/ui/button"
import { LoadingSpinner } from "@/components/ui/loading-spinner"
import { ScrollArea } from "@/components/ui/scroll-area"
import { Vaul, VaulContent } from "@/components/vaul"
import { logger } from "@/lib/helpers/debug"
import { upath } from "@/lib/helpers/upath"
import { atom } from "jotai"
import { useAtom, useAtomValue, useSetAtom } from "jotai/react"
import { useRouter } from "next/navigation"
import React from "react"
import { BiDownload } from "react-icons/bi"
import { FcFolder } from "react-icons/fc"

const log = logger("TORRENT DOWNLOAD FILE SELECTION")

export type TorrentDownloadFileSelection = {
    torrent: HibikeTorrent_AnimeTorrent
    destination: string
}

export const __torrentDownload_fileSelectionAtom = atom<TorrentDownloadFileSelection | undefined>(undefined)

export function getDefaultDestination(entry: Anime_Entry, libraryPath?: string): string {
    const fPath = entry.localFiles?.findLast(n => n)?.path // file path
    const newPath = libraryPath ? upath.join(libraryPath, sanitizeDirectoryName(entry.media?.title?.romaji || "")) : ""
    return fPath ? upath.normalize(upath.dirname(fPath)) : newPath
}

export function sanitizeDirectoryName(input: string): string {
    const disallowedChars = /[<>:"/\\|?*\x00-\x1F]/g // Pattern for disallowed characters
    // Replace disallowed characters with an underscore
    const sanitized = input.replace(disallowedChars, " ")
    // Remove leading/trailing spaces and dots (periods) which are not allowed
    const trimmed = sanitized.trim().replace(/^\.+|\.+$/g, "").replace(/\s+/g, " ")
    // Ensure the directory name is not empty after sanitization
    return trimmed || "Untitled"
}

export function TorrentDownloadFileSelection({ entry }: { entry: Anime_Entry }) {
    const router = useRouter()
    const serverStatus = useServerStatus()
    const libraryPath = serverStatus?.settings?.library?.libraryPath

    const setTorrentDrawerIsOpen = useSetAtom(__torrentSearch_selectionAtom)

    const [fileSelection, setFileSelection] = useAtom(__torrentDownload_fileSelectionAtom)
    const selectedTorrents = useAtomValue(__torrentSearch_selectedTorrentsAtom)

    const [selectedFileIndices, setSelectedFileIndices] = React.useState<number[]>([])

    const animeFolderName = React.useMemo(() => {
        return sanitizeDirectoryName(entry.media?.title?.romaji || "")
    }, [entry.media?.title?.romaji])

    const selectedTorrent = fileSelection?.torrent
    const destination = fileSelection?.destination ?? getDefaultDestination(entry, libraryPath)

    const handleDestinationChange = React.useCallback((newDestination: string) => {
        if (fileSelection) {
            setFileSelection({
                ...fileSelection,
                destination: newDestination,
            })
        }
    }, [fileSelection, setFileSelection])

    const libraryPathSelectionProps = useLibraryPathSelection({
        destination,
        setDestination: handleDestinationChange,
        animeFolderName,
    })

    const handleLibraryPathSelect = React.useCallback((selectedLibraryPath: string) => {
        if (fileSelection) {
            libraryPathSelectionProps.handleLibraryPathSelect(selectedLibraryPath)
        }
    }, [fileSelection, libraryPathSelectionProps.handleLibraryPathSelect])

    const { data: filepaths, isLoading } = useTorrentClientGetFiles({ torrent: selectedTorrent, provider: selectedTorrent?.provider })

    // download via torrent client
    const { mutate, isPending } = useTorrentClientDownload(() => {
        setFileSelection(undefined)
        setTorrentDrawerIsOpen(undefined)
        router.push("/torrent-list")
    })

    // Convert file paths to file previews format
    const filePreviews = React.useMemo(() => {
        if (!filepaths) return []
        return filepaths.map((path, index) => ({
            index,
            path,
            displayTitle: path.split("/").pop() || path,
            displayPath: path,
            isLikely: false,
        }))
    }, [filepaths])

    // Select all files by default when file previews are loaded
    // React.useEffect(() => {
    //     if (filePreviews.length > 0 && selectedFileIndices.length === 0) {
    //         const allIndices = filePreviews.map(file => file.index)
    //         setSelectedFileIndices(allIndices)
    //     }
    // }, [filePreviews, selectedFileIndices.length])

    const deselectedIndices = React.useMemo(() => {
        if (!filepaths) return []
        return filepaths.map((_, index) => index).filter(index => !selectedFileIndices.includes(index))
    }, [filepaths, selectedFileIndices])

    const getFileValue = React.useCallback((filePreview: any) => {
        return filePreview.index
    }, [])

    const scrollRef = React.useRef<HTMLDivElement>(null)

    const handleDownload = () => {
        if (!selectedTorrent || selectedFileIndices.length === 0) return

        mutate({
            torrents: [selectedTorrent],
            destination,
            smartSelect: {
                enabled: false,
                missingEpisodeNumbers: [],
            },
            deselect: {
                enabled: true,
                indices: deselectedIndices,
            },
            media: entry.media,
        })
    }

    return (
        <Vaul
            open={!!selectedTorrent}
            onOpenChange={open => {
                if (!open) {
                    setFileSelection(undefined)
                    setSelectedFileIndices([])
                }
            }}
        >
            <VaulContent className="max-w-5xl mx-auto">
                <AppLayoutStack className="mt-4 p-3 lg:p-6">
                    <h4 className="text-center mb-4">
                        Select files to download
                    </h4>

                    <DirectorySelector
                        name="destination"
                        label="Destination"
                        leftIcon={<FcFolder />}
                        value={destination}
                        defaultValue={destination}
                        onSelect={handleDestinationChange}
                        shouldExist={false}
                        libraryPathSelectionProps={{ ...libraryPathSelectionProps, handleLibraryPathSelect }}
                    />

                    {isLoading ? <LoadingSpinner /> : (
                        <AppLayoutStack className="pb-0">

                            <ScrollArea
                                viewportRef={scrollRef}
                                className="h-[60dvh] lg:h-[50dvh] overflow-y-auto p-4 border rounded-[--radius-md]"
                            >
                                <FileTreeMultiSelector
                                    filePreviews={filePreviews}
                                    selectedIndices={selectedFileIndices}
                                    onSelectionChange={setSelectedFileIndices}
                                    getFileValue={getFileValue}
                                />
                            </ScrollArea>

                            <div className="text-sm text-[--muted] mb-2">
                                {selectedFileIndices.length} of {filePreviews.length} files selected
                            </div>

                            <Button
                                intent="white"
                                className="w-full"
                                rightIcon={<BiDownload className="text-xl" />}
                                disabled={selectedFileIndices.length === 0 || isLoading || isPending}
                                loading={isPending}
                                onClick={handleDownload}
                            >
                                Download selected files
                            </Button>

                        </AppLayoutStack>
                    )}
                </AppLayoutStack>
            </VaulContent>
        </Vaul>
    )

}
