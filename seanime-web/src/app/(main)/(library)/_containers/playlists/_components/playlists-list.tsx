import { PlaylistModal } from "@/app/(main)/(library)/_containers/playlists/_components/playlist-modal"
import { useGetPlaylists } from "@/app/(main)/(library)/_containers/playlists/_lib/playlist-actions"
import { Button, IconButton } from "@/components/ui/button"
import { cn } from "@/components/ui/core/styling"
import { LoadingSpinner } from "@/components/ui/loading-spinner"
import React from "react"
import { FaRegEye } from "react-icons/fa"
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
        <div className="flex flex-col gap-2" data-vaul-no-drag>
            {playlists.map(p => (
                <div
                    key={p.dbId}
                    className={cn(
                        "flex gap-3 justify-between items-center px-4 py-2 bg-[#0c0c0c] rounded-md border cursor-pointer",
                        "transition",
                    )}
                    data-vaul-no-drag
                    // onClick={() => handleSelect(lf.path)}
                >
                    <div className="flex gap-3 items-center" data-vaul-no-drag>
                        <div data-vaul-no-drag>
                            <FaCirclePlay className="text-2xl text-white" />
                        </div>
                        <div data-vaul-no-drag>
                            <p className="text-xl font-bold text-white" data-vaul-no-drag>{p.name}</p>
                            <p className="text-base text-[--muted] font-normal line-clamp-1" data-vaul-no-drag>{p.localFiles.length} episodes</p>
                        </div>
                    </div>
                    <div data-vaul-no-drag className="flex items-center gap-2">
                        <Button
                            className=""
                            leftIcon={<FaCirclePlay />}
                            intent="white"
                            size="sm"
                            data-vaul-no-drag
                        >Play</Button>
                        <PlaylistModal
                            trigger={<IconButton
                                className=""
                                icon={<FaRegEye />}
                                intent="white-subtle"
                                size="sm"
                                data-vaul-no-drag
                            />} playlist={p}
                        />
                    </div>
                </div>
            ))}
        </div>
    )
}
