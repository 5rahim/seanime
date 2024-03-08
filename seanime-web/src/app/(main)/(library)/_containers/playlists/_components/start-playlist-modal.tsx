import { Button } from "@/components/ui/button"
import { Modal } from "@/components/ui/modal"
import { SeaEndpoints } from "@/lib/server/endpoints"
import { useSeaMutation } from "@/lib/server/query"
import { Playlist } from "@/lib/server/types"
import { useQueryClient } from "@tanstack/react-query"
import React from "react"
import { FaPlay } from "react-icons/fa"
import { toast } from "sonner"

type StartPlaylistModalProps = {
    trigger?: React.ReactElement
    playlist: Playlist
    canStart?: boolean
    onPlaylistLoaded: () => void
}

export function StartPlaylistModal(props: StartPlaylistModalProps) {

    const {
        trigger,
        playlist,
        canStart,
        onPlaylistLoaded,
        ...rest
    } = props

    const qc = useQueryClient()

    const { mutate: startPlaylist, isPending } = useSeaMutation<void, { dbId: number }>({
        endpoint: SeaEndpoints.PLAYBACK_MANAGER_START_PLAYLIST,
        method: "post",
        onSuccess: async () => {
            toast.success("Playlist loaded")
            onPlaylistLoaded()
            await qc.refetchQueries({ queryKey: ["get-playlists"] })
        },
    })

    return (
        <Modal
            title="Start playlist"
            titleClass="text-center"
            trigger={trigger}
        >
            <p className="text-center">
                You are about to start the playlist <strong>"{playlist.name}"</strong>,
                which contains {playlist.localFiles.length} episode{playlist.localFiles.length > 1 ? "s" : ""}.
            </p>
            <p className="text-[--muted] text-center">
                Reminder: The playlist will be deleted once you start it, whether you finish it or not.
            </p>
            {!canStart && (
                <p className="text-orange-300 text-center">
                    Please enable "Automatically update progress" to start
                </p>
            )}
            <Button
                className="w-full flex-none"
                leftIcon={<FaPlay />}
                intent="primary"
                size="lg"
                loading={isPending}
                // disabled={!canStart} // TODO Re-enable this
                onClick={() => startPlaylist({ dbId: playlist.dbId })}
            >Start</Button>
        </Modal>
    )
}
