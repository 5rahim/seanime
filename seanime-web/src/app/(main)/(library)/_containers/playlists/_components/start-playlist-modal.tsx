import { Anime_Playlist } from "@/api/generated/types"
import { usePlaybackStartPlaylist } from "@/api/hooks/playback_manager.hooks"
import { Button } from "@/components/ui/button"
import { Modal } from "@/components/ui/modal"
import React from "react"
import { FaPlay } from "react-icons/fa"

type StartPlaylistModalProps = {
    trigger?: React.ReactElement
    playlist: Anime_Playlist
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

    const { mutate: startPlaylist, isPending } = usePlaybackStartPlaylist({
        onSuccess: onPlaylistLoaded,
    })

    if (!playlist?.localFiles?.length) return null

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
                disabled={!canStart}
                onClick={() => startPlaylist({ dbId: playlist.dbId })}
            >Start</Button>
        </Modal>
    )
}
