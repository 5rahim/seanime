import { Anime_LibraryCollection, Anime_Playlist, Anime_PlaylistEpisode } from "@/api/generated/types"
import { useCreatePlaylist, useDeletePlaylist, useUpdatePlaylist } from "@/api/hooks/playlist.hooks"
import { PlaylistEditor, PlaylistMediaEntry } from "@/app/(main)/_features/playlists/_components/playlist-editor"
import { Button } from "@/components/ui/button"
import { DangerZone } from "@/components/ui/form"
import { Modal } from "@/components/ui/modal"
import { TextInput } from "@/components/ui/text-input"
import React from "react"
import { toast } from "sonner"

type PlaylistEditorModalProps = {
    libraryCollection: Anime_LibraryCollection | undefined
    playlist?: Anime_Playlist
    trigger: React.ReactElement
}

export function PlaylistEditorModal(props: PlaylistEditorModalProps) {

    const {
        libraryCollection,
        playlist,
        trigger,
    } = props

    const [isOpen, setIsOpen] = React.useState(false)
    const [name, setName] = React.useState(playlist?.name ?? "")
    const [episodes, setEpisodes] = React.useState<Anime_PlaylistEpisode[]>([])

    const isUpdate = !!playlist

    const { mutate: createPlaylist, isPending: isCreating } = useCreatePlaylist()
    const { mutate: deletePlaylist, isPending: isDeleting } = useDeletePlaylist()
    const { mutate: updatePlaylist, isPending: isUpdating } = useUpdatePlaylist()


    function reset() {
        setName("")
        setEpisodes([])
    }

    React.useEffect(() => {
        if (isUpdate && !!playlist && !!playlist.episodes) {
            setName(playlist.name)
            setEpisodes(playlist.episodes)
        }
    }, [playlist, isOpen])

    function handleSubmit() {
        if (name.length === 0) {
            toast.error("Please enter a name for the playlist")
            return
        }
        if (isUpdate && !!playlist) {
            updatePlaylist({ dbId: playlist.dbId, name, episodes })
        } else {
            setIsOpen(false)
            createPlaylist({ name, episodes }, {
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
            onInteractOutside={e => e.preventDefault()}
            contentClass="max-w-4xl"
        >
            <div className="space-y-4">

                <div className="space-y-4">
                    <TextInput
                        label="Name"
                        value={name}
                        onChange={e => setName(e.target.value)}
                    />

                    <PlaylistEditor
                        episodes={episodes}
                        setEpisodes={setEpisodes}
                        libraryCollection={libraryCollection}
                    />

                    {libraryCollection?.lists?.flatMap(n => n.entries)?.filter(Boolean)?.map(entry => {
                        return (
                            <PlaylistMediaEntry key={entry.mediaId} entry={entry} episodes={episodes} setEpisodes={setEpisodes} />
                        )
                    })}

                    <div className="w-full">
                        <Button
                            disabled={episodes.length === 0}
                            onClick={handleSubmit}
                            loading={isCreating || isDeleting || isUpdating}
                            className="w-full"
                        >
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
