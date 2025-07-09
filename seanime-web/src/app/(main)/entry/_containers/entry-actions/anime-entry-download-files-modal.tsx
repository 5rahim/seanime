import { getServerBaseUrl } from "@/api/client/server-url"
import { Anime_Entry } from "@/api/generated/types"
import { FilepathSelector } from "@/app/(main)/_features/media/_components/filepath-selector"
import { useServerHMACAuth } from "@/app/(main)/_hooks/use-server-status"
import { Button } from "@/components/ui/button"
import { Modal } from "@/components/ui/modal"
import { Separator } from "@/components/ui/separator"
import { openTab } from "@/lib/helpers/browser"
import { atom } from "jotai/index"
import { useAtom } from "jotai/react"
import React from "react"

export type AnimeEntryDownloadFilesModalProps = {
    entry: Anime_Entry
}

export const __animeEntryDownloadFilesModalIsOpenAtom = atom(false)


export function AnimeEntryDownloadFilesModal({ entry }: AnimeEntryDownloadFilesModalProps) {

    const [open, setOpen] = useAtom(__animeEntryDownloadFilesModalIsOpenAtom)


    return (
        <Modal
            open={open}
            onOpenChange={() => setOpen(false)}
            contentClass="max-w-2xl"
            title={<span>Select files to download</span>}
            titleClass="text-center"

        >
            <Content entry={entry} />
        </Modal>
    )

}

function Content({ entry }: { entry: Anime_Entry }) {

    const [open, setOpen] = useAtom(__animeEntryDownloadFilesModalIsOpenAtom)
    const [filepaths, setFilepaths] = React.useState<string[]>([])
    const { getHMACTokenQueryParam } = useServerHMACAuth()

    async function handleDownload() {
        for (const filepath of filepaths) {
            const endpoint = "/api/v1/mediastream/file?path=" + encodeURIComponent(filepath)
            const tokenQueryParam = await getHMACTokenQueryParam("/api/v1/mediastream/file", "&")
            const url = getServerBaseUrl() + endpoint + tokenQueryParam
            openTab(url)
        }
        setOpen(false)
    }

    if (!entry.media) return null

    return (
        <div className="space-y-2 mt-2">

            <p className="text-[--muted]">
                Seanime will open a new tab for each file you download. Make sure your browser allows popups.
            </p>

            <Separator />

            <FilepathSelector
                className="max-h-96"
                filepaths={filepaths}
                allFilepaths={entry.localFiles?.map(n => n.path) ?? []}
                onFilepathSelected={setFilepaths}
                showFullPath
            />
            <div className="flex justify-end gap-2 mt-2">
                <Button
                    intent="white"
                    onClick={() => handleDownload()}
                >
                    Download
                </Button>
            </div>
        </div>
    )
}
