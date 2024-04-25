import { Anime_Playlist } from "@/api/generated/types"
import { useCreatePlaylist, useDeletePlaylist, useUpdatePlaylist } from "@/api/hooks/playlist.hooks"
import { PlaylistManager } from "@/app/(main)/(library)/_containers/playlists/_components/playlist-manager"
import { Button } from "@/components/ui/button"
import { DangerZone } from "@/components/ui/form"
import { Modal } from "@/components/ui/modal"
import { Separator } from "@/components/ui/separator"
import { TextInput } from "@/components/ui/text-input"
import React from "react"
import { toast } from "sonner"

type PlaylistModalProps = {
    playlist?: Anime_Playlist
    trigger: React.ReactElement
}

export function PlaylistModal(props: PlaylistModalProps) {

    const {
        playlist,
        trigger,
    } = props

    const [isOpen, setIsOpen] = React.useState(false)
    const [name, setName] = React.useState(playlist?.name ?? "")
    const [paths, setPaths] = React.useState<string[]>(playlist?.localFiles?.map(l => l.path) ?? [])

    const isUpdate = !!playlist

    const { mutate: createPlaylist, isPending: isCreating } = useCreatePlaylist()

    const { mutate: deletePlaylist, isPending: isDeleting } = useDeletePlaylist()

    const { mutate: updatePlaylist, isPending: isUpdating } = useUpdatePlaylist()

    function reset() {
        setName("")
        setPaths([])
    }

    React.useEffect(() => {
        if (isUpdate && !!playlist && !!playlist.localFiles) {
            setName(playlist.name)
            setPaths(playlist.localFiles.map(l => l.path))
        }
    }, [playlist, isOpen])

    function handleSubmit() {
        if (name.length === 0) {
            toast.error("Please enter a name for the playlist")
            return
        }
        if (isUpdate && !!playlist) {
            updatePlaylist({ dbId: playlist.dbId, name, paths })
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
                    if (isUpdate && !!playlist) {
                        deletePlaylist({ dbId: playlist.dbId }, {
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
