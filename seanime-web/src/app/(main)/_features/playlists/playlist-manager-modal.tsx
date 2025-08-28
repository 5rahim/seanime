import { Anime_LibraryCollectionList } from "@/api/generated/types"
import { useGetLibraryCollection } from "@/api/hooks/anime_collection.hooks"
import { PlaylistEditorModal } from "@/app/(main)/_features/playlists/_components/playlist-editor-modal"
import { PlaylistManagerLists } from "@/app/(main)/_features/playlists/_components/playlist-manager-lists"
import { usePlaylistEditorManager } from "@/app/(main)/_features/playlists/lib/playlist-manager"
import { useServerStatus } from "@/app/(main)/_hooks/use-server-status"
import { Alert } from "@/components/ui/alert"
import { Button } from "@/components/ui/button"
import { cn } from "@/components/ui/core/styling"
import { Drawer } from "@/components/ui/drawer"
import React from "react"
import { toast } from "sonner"

export function PlaylistManagerModal() {
    const serverStatus = useServerStatus()
    const { isModalOpen, setModalOpen, setSelectedMedia, selectedMedia } = usePlaylistEditorManager()

    const { data: _data, isLoading: isLibraryLoading } = useGetLibraryCollection()

    const libraryCollection = React.useMemo(() => {
        if (!_data?.lists) return undefined
        if (!!_data?.stream) {
            // Add to current list
            let currentList = _data.lists?.find(n => n.type === "CURRENT")
            let entries = [...(currentList?.entries ?? [])]
            for (let anime of (_data.stream.anime ?? [])) {
                if (!entries.some(e => e.mediaId === anime.id)) {
                    entries.push({
                        media: anime,
                        mediaId: anime.id,
                        listData: _data.stream.listData?.[anime.id],
                        libraryData: undefined,
                    })
                }
            }
            return {
                ..._data,
                lists: [
                    {
                        type: "CURRENT",
                        status: "CURRENT",
                        entries: entries,
                    } as Anime_LibraryCollectionList,
                    ...(_data.lists ?? [])?.filter(n => n.type !== "CURRENT") ?? [],
                ].filter(Boolean),
            }
        } else {
            return _data
        }
    }, [_data])


    const allEntries = libraryCollection?.lists?.flatMap(n => n.entries) ?? []

    React.useEffect(() => {
        if (selectedMedia) {
            if (!allEntries.find(n => n?.mediaId === selectedMedia)) {
                toast.warning("This anime is not in your collection")
                setSelectedMedia(null)
            }
        }
    }, [selectedMedia, allEntries])

    return (
        <>
            <Drawer
                open={isModalOpen}
                onOpenChange={v => {
                    setModalOpen(v)
                    if (!v) {
                        setSelectedMedia(null)
                    }
                }}
                size="lg"
                side="bottom"
                contentClass=""
            >

                <div className="space-y-6">
                    <div className="flex flex-col md:flex-row justify-between items-center gap-4">
                        <div>
                            <h4>Playlists</h4>
                        </div>
                        <div className="flex gap-2 items-center md:pr-8">
                            <PlaylistEditorModal
                                libraryCollection={libraryCollection}
                                trigger={
                                    <Button
                                        intent="white"
                                        className={cn("rounded-full", selectedMedia && "animate-pulse")}
                                    >
                                        {selectedMedia ? "Add to new Playlist" : "Create a Playlist"}
                                    </Button>
                                }
                            />
                        </div>
                    </div>

                    {!serverStatus?.settings?.library?.autoUpdateProgress && <Alert
                        className="max-w-2xl mx-auto"
                        intent="warning"
                        description={<>
                            <p>
                                You need to enable "Automatically update progress" to use playlists.
                            </p>
                        </>}
                    />}

                    <div className="">
                        <PlaylistManagerLists libraryCollection={libraryCollection} />
                    </div>
                </div>
            </Drawer>
        </>
    )
}
