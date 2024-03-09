import { PlaylistModal } from "@/app/(main)/(library)/_containers/playlists/_components/playlist-modal"
import { StartPlaylistModal } from "@/app/(main)/(library)/_containers/playlists/_components/start-playlist-modal"
import { useGetPlaylists } from "@/app/(main)/(library)/_containers/playlists/_lib/playlist-actions"
import { __playlists_modalOpenAtom } from "@/app/(main)/(library)/_containers/playlists/playlists-modal"
import { anilistUserMediaAtom } from "@/app/(main)/_loaders/anilist-user-media"
import { serverStatusAtom } from "@/atoms/server-status"
import { imageShimmer } from "@/components/shared/styling/image-helpers"
import { Button } from "@/components/ui/button"
import { Carousel, CarouselContent, CarouselDotButtons, CarouselItem } from "@/components/ui/carousel"
import { cn } from "@/components/ui/core/styling"
import { LoadingSpinner } from "@/components/ui/loading-spinner"
import { useAtomValue, useSetAtom } from "jotai/react"
import Image from "next/image"
import React from "react"
import { FaCirclePlay } from "react-icons/fa6"
import { PiMonitorPlayFill } from "react-icons/pi"

type PlaylistsListProps = {
    children?: React.ReactNode
}

export function PlaylistsList(props: PlaylistsListProps) {

    const {
        children,
        ...rest
    } = props

    const { playlists, isLoading } = useGetPlaylists()
    const userMedia = useAtomValue(anilistUserMediaAtom)
    const serverStatus = useAtomValue(serverStatusAtom)

    const setOpen = useSetAtom(__playlists_modalOpenAtom)

    const handlePlaylistLoaded = React.useCallback(() => {
        setOpen(false)
    }, [])

    if (isLoading) {
        return (
            <LoadingSpinner />
        )
    }

    if (!playlists?.length) {
        return (
            <div className="text-center text-[--muted] space-y-1">
                <PiMonitorPlayFill className="mx-auto text-6xl text-white" />
                <div>
                    You haven't set up any playlists yet
                </div>
            </div>
        )
    }


    return (
        // <div className="grid grid-cols-2 md:grid-cols-4 lg:grid-cols-5 xl:grid-cols-6 gap-4">
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

                    const mainMedia = userMedia?.find(m => m.id === p.localFiles[0]?.mediaId)

                    return (
                        <CarouselItem
                            key={p.dbId}
                            className={cn(
                                "md:basis-1/3 lg:basis-1/4 2xl:basis-1/6 min-[2000px]:basis-1/6",
                                "aspect-[7/6] p-2",
                            )}
                            // onClick={() => handleSelect(lf.path)}
                        >
                            <div className="group/playlist-item flex gap-3 h-full justify-between items-center bg-gray-950 rounded-md border transition relative overflow-hidden">
                                {/*{(mainMedia?.coverImage?.large || mainMedia?.bannerImage) && <Image*/}
                                {/*    src={mainMedia?.coverImage?.extraLarge || mainMedia?.bannerImage || ""}*/}
                                {/*    placeholder={imageShimmer(700, 475)}*/}
                                {/*    sizes="10rem"*/}
                                {/*    fill*/}
                                {/*    alt=""*/}
                                {/*    className="object-center object-cover z-[1] transition-opacity lg:group-hover/playlist-item:opacity-0"*/}
                                {/*/>}*/}
                                <div
                                    className={cn(
                                        "absolute inset-0 bg-gray-950 transition-opacity",
                                        "opacity-100 z-[2] p-2 space-y-3",
                                        "flex flex-col",
                                    )}
                                >
                                    <StartPlaylistModal
                                        canStart={serverStatus?.settings?.library?.autoUpdateProgress}
                                        trigger={<div className="w-full h-full rounded-md overflow-hidden relative cursor-pointer">
                                            {(mainMedia?.coverImage?.large || mainMedia?.bannerImage) && <Image
                                                src={mainMedia?.coverImage?.extraLarge || mainMedia?.bannerImage || ""}
                                                placeholder={imageShimmer(700, 475)}
                                                sizes="10rem"
                                                fill
                                                alt=""
                                                className="object-center object-cover"
                                            />}
                                            <div className="absolute inset-0 bg-gray-900 opacity-50 hover:opacity-70 transition-opacity flex items-center justify-center">
                                                <FaCirclePlay className="block text-4xl" />
                                            </div>
                                        </div>}
                                        playlist={p}
                                        onPlaylistLoaded={handlePlaylistLoaded}
                                    />
                                    <div className="space-y-1 items-center">
                                        <p className="text-md font-bold text-white max-w-lg truncate text-center">{p.name}</p>
                                        <p className="text-lg font-normal line-clamp-1 text-center">{p.localFiles.length} episode{p.localFiles.length > 1
                                            ? `s`
                                            : ""}</p>
                                    </div>
                                    <PlaylistModal
                                        trigger={<Button
                                            className="w-full flex-none"
                                            // leftIcon={<FaRegEye />}
                                            intent="white-subtle"
                                            size="sm"

                                        >Edit</Button>} playlist={p}
                                    />
                                </div>
                            </div>
                        </CarouselItem>
                    )
                })}
            </CarouselContent>
        </Carousel>
    )
}
