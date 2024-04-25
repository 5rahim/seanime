import { Anime_LocalFile, Anime_MediaEntry } from "@/api/generated/types"
import { useDeleteLocalFiles } from "@/api/hooks/localfiles.hooks"
import { ConfirmationDialog, useConfirmationDialog } from "@/components/shared/confirmation-dialog"
import { Button } from "@/components/ui/button"
import { Checkbox } from "@/components/ui/checkbox"
import { Modal } from "@/components/ui/modal"
import { Separator } from "@/components/ui/separator"
import { atom } from "jotai"
import { atomWithImmer } from "jotai-immer"
import { useAtom } from "jotai/react"
import React from "react"
import * as upath from "upath"

export type AnimeEntryBulkDeleteFilesModalProps = {
    entry: Anime_MediaEntry
}

export const __bulkDeleteFilesModalIsOpenAtom = atom(false)
const __episodeDeleteActionFilesAtom = atomWithImmer<{ file: Anime_LocalFile, selected: boolean }[]>([])


export function AnimeEntryBulkDeleteFilesModal({ entry }: AnimeEntryBulkDeleteFilesModalProps) {

    const [open, setOpen] = useAtom(__bulkDeleteFilesModalIsOpenAtom)

    return (
        <Modal
            open={open}
            onOpenChange={() => setOpen(false)}
            contentClass="max-w-2xl"
            title={<span>Select files to delete</span>}
            titleClass="text-center"

        >
            <Content entry={entry} />
        </Modal>
    )

}

function Content({ entry }: { entry: Anime_MediaEntry }) {

    const [state, setState] = useAtom(__bulkDeleteFilesModalIsOpenAtom)
    const [files, setFiles] = useAtom(__episodeDeleteActionFilesAtom)

    const media = entry.media
    const localFiles = entry.localFiles

    React.useEffect(() => {
        if (localFiles) {
            setFiles(localFiles
                .filter(f => !!f.metadata)
                .map(file => ({ file, selected: true }))
                .toSorted((a, b) => a.file.metadata!.episode - b.file.metadata!.episode))
        }
    }, [state, localFiles])


    const { mutate: deleteFiles, isPending: isDeleting } = useDeleteLocalFiles(entry.mediaId)

    const confirmUnmatch = useConfirmationDialog({
        title: "Delete files",
        description: "This action cannot be undone.",
        onConfirm: () => {
            deleteFiles({ paths: files.filter(({ selected }) => selected).map(({ file }) => file.path) })
        },
    })


    const allFilesChecked = files.every(({ selected }) => selected)
    const noneFilesChecked = files.every(({ selected }) => !selected)

    if (!media) return null

    return (
        <div className="space-y-2 mt-2">
            <div className="max-h-96 overflow-y-auto px-2 space-y-2">

                <div className="p-2">
                    <Checkbox
                        label={`Select all files`}
                        value={allFilesChecked ? true : noneFilesChecked ? false : "indeterminate"}
                        onValueChange={checked => {
                            if (typeof checked === "boolean") {
                                setFiles(prev => !prev.every(({ selected }) => selected)
                                    ? prev.map(({ file }) => ({ file, selected: true }))
                                    : prev.map(({ file }) => ({ file, selected: false })))
                            }
                        }}
                        fieldClass="w-[fit-content]"
                    />
                </div>

                <Separator />

                {files.map(({ file, selected }, index) => (
                    <div
                        key={`${file.path}-${index}`}
                        className="p-2 border-b "
                    >
                        <div className="flex items-center">
                            <Checkbox
                                label={`${upath.basename(file.path)}`}
                                value={selected}
                                onValueChange={checked => {
                                    if (typeof checked === "boolean") {
                                        setFiles(draft => {
                                            draft[index].selected = checked
                                            return
                                        })
                                    }
                                }}
                                fieldClass="w-[fit-content]"
                            />
                        </div>
                    </div>
                ))}
            </div>
            <div className="flex justify-end gap-2 mt-2">
                <Button
                    intent="alert"
                    onClick={() => confirmUnmatch.open()}
                    loading={isDeleting}
                >
                    Delete
                </Button>
                <Button
                    intent="white"
                    onClick={() => setState(false)}
                    disabled={isDeleting}
                >
                    Cancel
                </Button>
            </div>
            <ConfirmationDialog {...confirmUnmatch} />
        </div>
    )
}
