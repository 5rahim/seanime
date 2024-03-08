import { Button } from "@/components/ui/button"
import { Modal } from "@/components/ui/modal"
import { Playlist } from "@/lib/server/types"
import React from "react"
import { FaPlay } from "react-icons/fa"

type StartPlaylistModalProps = {
    trigger?: React.ReactElement
    playlist: Playlist
}

export function StartPlaylistModal(props: StartPlaylistModalProps) {

    const {
        trigger,
        playlist,
        ...rest
    } = props

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
            <Button
                className="w-full flex-none"
                leftIcon={<FaPlay />}
                intent="white-subtle"
                size="lg"

            >Start</Button>
        </Modal>
    )
}
