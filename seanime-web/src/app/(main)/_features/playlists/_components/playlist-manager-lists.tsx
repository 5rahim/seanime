import { Anime_LibraryCollection } from "@/api/generated/types"
import { useGetPlaylists } from "@/api/hooks/playlist.hooks"
import { MediaCardBodyBottomGradient } from "@/app/(main)/_features/custom-ui/item-bottom-gradients"
import { PlaylistEditorModal } from "@/app/(main)/_features/playlists/_components/playlist-editor-modal"
import { usePlaylistEditorManager } from "@/app/(main)/_features/playlists/lib/playlist-manager"
import { imageShimmer } from "@/components/shared/image-helpers"
import { Button } from "@/components/ui/button"
import { Carousel, CarouselContent, CarouselDotButtons, CarouselItem } from "@/components/ui/carousel"
import { cn } from "@/components/ui/core/styling"
import { LoadingSpinner } from "@/components/ui/loading-spinner"
import Image from "next/image"
import React from "react"
import { BiEditAlt } from "react-icons/bi"
import { LuPlus } from "react-icons/lu"
import { MdOutlineVideoLibrary } from "react-icons/md"

export function PlaylistManagerLists({ libraryCollection }: { libraryCollection: Anime_LibraryCollection | undefined }) {

    const { data: playlists, isLoading } = useGetPlaylists()
    const { selectedMedia } = usePlaylistEditorManager()

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
                                {(mainMedia?.coverImage?.large || mainMedia?.bannerImage) && <Image
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
