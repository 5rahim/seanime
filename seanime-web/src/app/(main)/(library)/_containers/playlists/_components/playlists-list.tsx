import { useGetPlaylists } from "@/app/(main)/(library)/_containers/playlists/_lib/playlist-actions"
import { LoadingSpinner } from "@/components/ui/loading-spinner"
import React from "react"
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
        <>

        </>
    )
}
