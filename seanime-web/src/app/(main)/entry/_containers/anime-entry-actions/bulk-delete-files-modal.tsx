import { ConfirmationDialog, useConfirmationDialog } from "@/components/application/confirmation-dialog"
import { Button } from "@/components/ui/button"
import { Checkbox } from "@/components/ui/checkbox"
import { Modal } from "@/components/ui/modal"
import { Separator } from "@/components/ui/separator"
import { SeaEndpoints } from "@/lib/server/endpoints"
import { useSeaMutation } from "@/lib/server/query"
import { LocalFile, MediaEntry } from "@/lib/server/types"
import { useQueryClient } from "@tanstack/react-query"
import { atom } from "jotai"
import { atomWithImmer } from "jotai-immer"
import { useAtom } from "jotai/react"
import React, { useEffect } from "react"
import { toast } from "sonner"
import * as upath from "upath"

export type BulkDeleteFilesModalProps = {
    entry: MediaEntry
}

export const _bulkDeleteFilesModalIsOpenAtom = atom(false)
const _episodeDeleteActionFilesAtom = atomWithImmer<{ file: LocalFile, selected: boolean }[]>([])


export function BulkDeleteFilesModal({ entry }: BulkDeleteFilesModalProps) {

    const [open, setOpen] = useAtom(_bulkDeleteFilesModalIsOpenAtom)

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

function Content({ entry }: { entry: MediaEntry }) {

    const [state, setState] = useAtom(_bulkDeleteFilesModalIsOpenAtom)
    const [files, setFiles] = useAtom(_episodeDeleteActionFilesAtom)

    const media = entry.media
    const localFiles = entry.localFiles

    useEffect(() => {
        setFiles(localFiles.map(file => ({ file, selected: true })).toSorted((a, b) => a.file.metadata.episode - b.file.metadata.episode))
    }, [state, localFiles])

    const qc = useQueryClient()

    const { mutate: deleteFiles, isPending: isDeleting } = useSeaMutation<any, { paths: string[] }>({
        endpoint: SeaEndpoints.LOCAL_FILES,
        mutationKey: ["delete-local-files"],
        method: "delete",
        onSuccess: async () => {
            toast.success("Files removed")
            await qc.refetchQueries({ queryKey: ["get-media-entry", media?.id] })
            await qc.refetchQueries({ queryKey: ["get-library-collection"] })
            setState(false)
        },
    })

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
