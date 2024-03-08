import { Button } from "@/components/ui/button"
import { Modal } from "@/components/ui/modal"
import { Playlist } from "@/lib/server/types"
import React from "react"
import { FaPlay } from "react-icons/fa"

type StartPlaylistModalProps = {
    trigger?: React.ReactElement
    playlist: Playlist
    canStart?: boolean
}

export function StartPlaylistModal(props: StartPlaylistModalProps) {

    const {
        trigger,
        playlist,
        canStart,
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
                disabled={!canStart}
            >Start</Button>
        </Modal>
    )
}
