import { useGetTorrentstreamBatchHistory } from "@/api/hooks/torrentstream.hooks"
import { __torrentSearch_drawerIsOpenAtom, TorrentSelectionType } from "@/app/(main)/entry/_containers/torrent-search/torrent-search-drawer"
import { Modal } from "@/components/ui/modal"
import { atom, useAtomValue, useSetAtom } from "jotai"
import React from "react"
import { useUpdateEffect } from "react-use"

type TorrentStreamLoaderProps = {
    mediaId: number
}

const __torrentstream_loaderAtom = atom<string | null>(null)

export function TorrentStreamLoader(props: TorrentStreamLoaderProps) {

    const {
        mediaId,
    } = props

    const loaderType = useAtomValue(__torrentstream_loaderAtom)
    const loaderTypeRef = React.useRef(loaderType)

    const { data: batchHistory } = useGetTorrentstreamBatchHistory(mediaId, !!loaderType)

    const [historyOpen, setHistoryOpen] = React.useState(false)

    const setTorrentDrawerIsOpen = useSetAtom(__torrentSearch_drawerIsOpenAtom)

    useUpdateEffect(() => {
        loaderTypeRef.current = loaderType

        if (!loaderType || !batchHistory) return

        if (batchHistory.torrent) {
            setHistoryOpen(true)
        } else {
            cancel()
        }

    }, [loaderType, batchHistory])

    function cancel() {
        setHistoryOpen(false)
        React.startTransition(() => {
            if (loaderTypeRef.current) {
                setTorrentDrawerIsOpen(loaderTypeRef.current as TorrentSelectionType)
            }
        })
    }

    return (
        <Modal
            open={historyOpen}
            onOpenChange={v => {
                if (!v) cancel()
            }}
        >
            <pre>
                {JSON.stringify(batchHistory, null, 2)}
            </pre>
        </Modal>
    )
}

export function useSetTorrentStreamLoader() {
    return useSetAtom(__torrentstream_loaderAtom)
}
