import { PlaylistManager } from "@/app/(main)/(library)/_containers/playlists/_components/playlist-manager"
import { useCreatePlaylist, useDeletePlaylist, useUpdatePlaylist } from "@/app/(main)/(library)/_containers/playlists/_lib/playlist-actions"
import { Button } from "@/components/ui/button"
import { DangerZone } from "@/components/ui/form"
import { Modal } from "@/components/ui/modal"
import { Separator } from "@/components/ui/separator"
import { TextInput } from "@/components/ui/text-input"
import { Playlist } from "@/lib/server/types"
import React from "react"
import { toast } from "sonner"

type PlaylistModalProps = {
    playlist?: Playlist
    trigger: React.ReactElement
}

export function PlaylistModal(props: PlaylistModalProps) {

    const {
        trigger,
        ...rest
    } = props

    const [isOpen, setIsOpen] = React.useState(false)
    const [name, setName] = React.useState(props.playlist?.name ?? "")
    const [paths, setPaths] = React.useState<string[]>(props.playlist?.localFiles?.map(l => l.path) ?? [])

    const isUpdate = !!props.playlist

    const { createPlaylist, isCreating } = useCreatePlaylist()

    const { deletePlaylist, isDeleting } = useDeletePlaylist()

    const { updatePlaylist, isUpdating } = useUpdatePlaylist()

    function reset() {
        setName("")
        setPaths([])
    }

    React.useEffect(() => {
        if (isUpdate && !!props.playlist) {
            setName(props.playlist.name)
            setPaths(props.playlist.localFiles.map(l => l.path))
        }
    }, [props.playlist, isOpen])

    function handleSubmit() {
        if (name.length === 0) {
            toast.error("Please enter a name for the playlist")
            return
        }
        if (isUpdate && !!props.playlist) {
            updatePlaylist({ dbId: props.playlist.dbId, name, paths }, {
                onSuccess: () => {
                },
            })
        } else {
            setIsOpen(false)
            createPlaylist({ name, paths }, {
                onSuccess: () => {
                    reset()
                },
            })
        }
    }

    return (
        <Modal
            title={isUpdate ? "Edit playlist" : "Create a playlist"}
            trigger={trigger}
            open={isOpen}
            onOpenChange={v => setIsOpen(v)}
            contentClass="max-w-4xl"
        >
            <div className="space-y-4">

                <div className="space-y-4">
                    <TextInput
                        label="Name"
                        value={name}
                        onChange={e => setName(e.target.value)}
                    />

                    <Separator />

                    <PlaylistManager
                        paths={paths}
                        setPaths={setPaths}
                    />
                    <div className="">
                        <Button disabled={paths.length === 0} onClick={handleSubmit} loading={isCreating || isDeleting || isUpdating}>
                            {isUpdate ? "Update" : "Create"}
                        </Button>
                    </div>
                </div>

                {isUpdate && <DangerZone
                    actionText="Delete playlist" onDelete={() => {
                    if (isUpdate && !!props.playlist) {
                        deletePlaylist({ dbId: props.playlist.dbId }, {
                            onSuccess: () => {
                                setIsOpen(false)
                            },
                        })
                    }
                }}
                />}
            </div>
        </Modal>
    )
}
