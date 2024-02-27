import { useLocalFileBulkAction, useRemoveEmptyDirectories } from "@/app/(main)/(library)/_containers/bulk-actions/_lib/local-file-bulk-actions"
import { ConfirmationDialog, useConfirmationDialog } from "@/components/application/confirmation-dialog"
import { AppLayoutStack } from "@/components/ui/app-layout"
import { Button } from "@/components/ui/button"
import { Modal } from "@/components/ui/modal"
import { atom, useAtom } from "jotai"
import React from "react"
import { BiLockAlt, BiLockOpenAlt } from "react-icons/bi"

export const bulkActionModalAtomIsOpen = atom<boolean>(false)

export function BulkActionModal() {

    const [isOpen, setIsOpen] = useAtom(bulkActionModalAtomIsOpen)

    const { lockFiles, unlockFiles, isPending } = useLocalFileBulkAction({
        onSuccess: () => {
            setIsOpen(false)
        },
    })

    const { removeEmptyDirectories, isPending: isRemoving } = useRemoveEmptyDirectories({
        onSuccess: () => {
            setIsOpen(false)
        },
    })

    const confirmRemoveEmptyDirs = useConfirmationDialog({
        title: "Remove empty directories",
        description: "This action will remove all empty directories in the library. Are you sure you want to continue?",
        onConfirm: () => {
            removeEmptyDirectories()
        },
    })

    return (
        <Modal
            open={isOpen} onOpenChange={() => setIsOpen(false)} title="Bulk actions"
            contentClass="space-y-4"
        >
            <AppLayoutStack spacing="sm">
                {/*<p>These actions do not affect ignored files.</p>*/}
                <Button
                    leftIcon={<BiLockAlt />}
                    intent="gray-outline"
                    className="w-full"
                    disabled={isPending || isRemoving}
                    onClick={lockFiles}
                >
                    Lock all files
                </Button>
                <Button
                    leftIcon={<BiLockOpenAlt />}
                    intent="gray-outline"
                    className="w-full"
                    disabled={isPending || isRemoving}
                    onClick={unlockFiles}
                >
                    Unlock all files
                </Button>
                <Button
                    intent="gray-outline"
                    className="w-full"
                    disabled={isPending}
                    loading={isRemoving}
                    onClick={() => confirmRemoveEmptyDirs.open()}
                >
                    Remove empty directories
                </Button>
            </AppLayoutStack>
            <ConfirmationDialog {...confirmRemoveEmptyDirs} />
        </Modal>
    )

}
