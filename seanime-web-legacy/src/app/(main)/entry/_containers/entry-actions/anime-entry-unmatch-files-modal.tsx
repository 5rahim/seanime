import { Anime_Entry } from "@/api/generated/types"
import { useUpdateLocalFiles } from "@/api/hooks/localfiles.hooks"
import { FilepathSelector } from "@/app/(main)/_features/media/_components/filepath-selector"
import { ConfirmationDialog, useConfirmationDialog } from "@/components/shared/confirmation-dialog"
import { Button } from "@/components/ui/button"
import { Modal } from "@/components/ui/modal"
import { atom } from "jotai/index"
import { useAtom } from "jotai/react"
import React from "react"

export type AnimeEntryUnmatchFilesModalProps = {
    entry: Anime_Entry
}

export const __animeEntryUnmatchFilesModalIsOpenAtom = atom(false)


export function AnimeEntryUnmatchFilesModal({ entry }: AnimeEntryUnmatchFilesModalProps) {

    const [open, setOpen] = useAtom(__animeEntryUnmatchFilesModalIsOpenAtom)

    return (
        <Modal
            open={open}
            onOpenChange={() => setOpen(false)}
            contentClass="max-w-2xl"
            title={<span>Select files to unmatch</span>}
            titleClass="text-center"

        >
            <Content entry={entry} />
        </Modal>
    )

}

function Content({ entry }: { entry: Anime_Entry }) {

    const [open, setOpen] = useAtom(__animeEntryUnmatchFilesModalIsOpenAtom)

    const [filepaths, setFilepaths] = React.useState<string[]>([])

    const media = entry.media

    React.useEffect(() => {
        if (entry.localFiles) {
            setFilepaths(entry.localFiles.map(f => f.path))
        }
    }, [entry.localFiles])


    const { mutate: updateFiles, isPending: isDeleting } = useUpdateLocalFiles()

    const confirmUnmatch = useConfirmationDialog({
        title: "Unmatch files",
        onConfirm: () => {
            if (filepaths.length === 0) return

            updateFiles({
                paths: filepaths,
                action: "unmatch",
            }, {
                onSuccess: () => {
                    setOpen(false)
                },
            })
        },
    })

    if (!media) return null

    return (
        <div className="space-y-2 mt-2">

            <FilepathSelector
                className="max-h-96"
                filepaths={filepaths}
                allFilepaths={entry.localFiles?.map(n => n.path) ?? []}
                onFilepathSelected={setFilepaths}
                showFullPath
            />
            <div className="flex justify-end gap-2 mt-2">
                <Button
                    intent="warning"
                    onClick={() => confirmUnmatch.open()}
                    loading={isDeleting}
                >
                    Unmatch
                </Button>
                <Button
                    intent="white"
                    onClick={() => setOpen(false)}
                    disabled={isDeleting}
                >
                    Cancel
                </Button>
            </div>
            <ConfirmationDialog {...confirmUnmatch} />
        </div>
    )
}
