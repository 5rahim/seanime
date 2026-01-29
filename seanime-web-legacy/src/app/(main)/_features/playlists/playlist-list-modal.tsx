import { Anime_LibraryCollection, Anime_LibraryCollectionList } from "@/api/generated/types"
import { useGetLibraryCollection } from "@/api/hooks/anime_collection.hooks"
import { useGetPlaylists } from "@/api/hooks/playlist.hooks"
import { MediaCardBodyBottomGradient } from "@/app/(main)/_features/custom-ui/item-bottom-gradients"
import { PlaylistEditorModal } from "@/app/(main)/_features/playlists/_components/playlist-editor-modal"
import { usePlaylistManager } from "@/app/(main)/_features/playlists/_containers/global-playlist-manager"
import { usePlaylistEditorManager } from "@/app/(main)/_features/playlists/lib/playlist-editor-manager"
import { useServerStatus } from "@/app/(main)/_hooks/use-server-status"
import { imageShimmer } from "@/components/shared/image-helpers"
import { SeaImage } from "@/components/shared/sea-image"
import { Button } from "@/components/ui/button"
import { Carousel, CarouselContent, CarouselDotButtons, CarouselItem } from "@/components/ui/carousel"
import { cn } from "@/components/ui/core/styling"
import { Drawer } from "@/components/ui/drawer"
import { LoadingSpinner } from "@/components/ui/loading-spinner"
import React from "react"
import { BiEditAlt } from "react-icons/bi"
import { FaCirclePlay } from "react-icons/fa6"
import { LuPlus } from "react-icons/lu"
import { MdOutlineVideoLibrary } from "react-icons/md"
import { toast } from "sonner"

// todo: select checkbox, group actions

export function PlaylistListModal() {
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

    // const isNakama = _data.


    const allEntries = libraryCollection?.lists?.flatMap(n => n.entries) ?? []

    React.useEffect(() => {
        if (selectedMedia) {
            if (!allEntries.find(n => n?.mediaId === selectedMedia)) {
                toast.warning("This anime is not in your library or currently watching collection.")
                setSelectedMedia(null)
                React.startTransition(() => {
                    setModalOpen(false)
                })
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
                            <h4 className="flex items-center">Playlists</h4>
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

                    <div className="">
                        <PlaylistLists libraryCollection={libraryCollection} />
                    </div>
                </div>
            </Drawer>
        </>
    )
}


function PlaylistLists({ libraryCollection }: { libraryCollection: Anime_LibraryCollection | undefined }) {

    const { data: playlists, isLoading } = useGetPlaylists()
    const { selectedMedia, setModalOpen } = usePlaylistEditorManager()

    const { startPlaylist } = usePlaylistManager()

    if (isLoading) return <LoadingSpinner />

    if (!playlists?.length) {
        return (
            <div className="text-center text-[--muted] space-y-1 py-6">
                <MdOutlineVideoLibrary className="mx-auto text-5xl text-[--muted]" />
                <div>
                    No playlists
                </div>
            </div>
        )
    }

    return (
        <Carousel
            className="w-full max-w-full"
            gap="none"
            opts={{
                align: "start",
            }}
        >
            <CarouselDotButtons />
            <CarouselContent>
                {playlists.map(p => {

                    const mainMedia = p.episodes?.[0]?.episode?.baseAnime

                    return (
                        <CarouselItem
                            key={p.dbId}
                            className={cn(
                                "md:basis-1/3 lg:basis-1/4 2xl:basis-1/6 min-[2000px]:basis-1/6",
                                "aspect-[7/6] p-2",
                            )}
                            // onClick={() => handleSelect(lf.path)}
                        >
                            <div className="group/playlist-item flex gap-3 h-full justify-between items-center bg-gray-950 rounded-md transition relative overflow-hidden">
                                {(mainMedia?.coverImage?.large || mainMedia?.bannerImage) && <SeaImage
                                    src={mainMedia?.coverImage?.extraLarge || mainMedia?.bannerImage || ""}
                                    placeholder={imageShimmer(700, 475)}
                                    sizes="10rem"
                                    fill
                                    alt=""
                                    className="object-center object-cover z-[1]"
                                />}

                                <div className="absolute inset-0 z-[2] bg-gray-900 opacity-50 hover:opacity-70 transition-opacity flex items-center justify-center" />
                                <div className="absolute inset-0 z-[6] flex items-center justify-center">
                                    {/*<StartPlaylistModal*/}
                                    {/*    canStart={serverStatus?.settings?.library?.autoUpdateProgress}*/}
                                    {/*    playlist={p}*/}
                                    {/*    onPlaylistLoaded={handlePlaylistLoaded}*/}
                                    {/*    trigger={*/}
                                    {/*        <FaCirclePlay className="block text-5xl cursor-pointer opacity-50 hover:opacity-100 transition-opacity" />}*/}
                                    {/*/>*/}
                                    {!selectedMedia && <div
                                        onClick={() => {
                                            startPlaylist(p)
                                            setModalOpen(false)
                                        }}
                                    >
                                        <FaCirclePlay className="block text-5xl cursor-pointer opacity-95 hover:opacity-70 hover:scale-[1.1] transition-all" />
                                    </div>}
                                </div>
                                <div className="absolute top-2 right-2 z-[6] flex items-center justify-center">
                                    <PlaylistEditorModal
                                        libraryCollection={libraryCollection}
                                        trigger={<Button
                                            className={cn(
                                                "w-full flex-none rounded-full",
                                                selectedMedia && "animate-pulse",
                                            )}
                                            leftIcon={selectedMedia ? <LuPlus /> : <BiEditAlt />}
                                            intent={selectedMedia ? "white" : "white-subtle"}
                                            size="sm"

                                        >{selectedMedia ? "Add to Playlist" : "Edit"}</Button>} playlist={p}
                                    />
                                </div>
                                <div className="absolute w-full bottom-0 h-fit z-[6]">
                                    <div className="space-y-0 pb-3 items-center">
                                        <p className="text-md font-bold text-white max-w-lg truncate text-center">{p.name}</p>
                                        {p.episodes &&
                                            <p className="text-sm text-[--muted] font-normal line-clamp-1 text-center">{p.episodes.length} episode{p.episodes.length > 1
                                                ? `s`
                                                : ""}</p>}
                                    </div>
                                </div>

                                <MediaCardBodyBottomGradient />
                            </div>
                        </CarouselItem>
                    )
                })}
            </CarouselContent>
        </Carousel>
    )
}
