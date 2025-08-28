import { atom } from "jotai"
import { useAtom } from "jotai/react"
import React from "react"

const playlistEditorManager_isOpenAtom = atom(false)
const playlistEditorManager_selectedMediaAtom = atom<number | null>(null)
const playlistEditorManager_episodeToAddAtom = atom<string | null>(null)

export function usePlaylistEditorManager() {
    const [isModalOpen, setModalOpen] = useAtom(playlistEditorManager_isOpenAtom)
    const [selectedMedia, setSelectedMedia] = useAtom(playlistEditorManager_selectedMediaAtom)
    const [episodeToAdd, setEpisodeToAdd] = useAtom(playlistEditorManager_episodeToAddAtom)

    function selectMediaAndOpenEditor(mediaId: number) {
        setSelectedMedia(mediaId)
        React.startTransition(() => {
            setModalOpen(true)
        })
    }

    function selectEpisodeToAddAndOpenEditor(mediaId: number, anidbEpisode: string) {
        setSelectedMedia(mediaId)
        setEpisodeToAdd(anidbEpisode)
        React.startTransition(() => {
            setModalOpen(true)
        })
    }

    return {
        selectedMedia,
        setSelectedMedia,
        episodeToAdd,
        setEpisodeToAdd,
        isModalOpen,
        setModalOpen,
        selectMediaAndOpenEditor,
        selectEpisodeToAddAndOpenEditor,
    }
}
