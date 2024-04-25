import { useLocalFileBulkAction, useRemoveEmptyDirectories } from "@/api/hooks/localfiles.hooks"
import { ConfirmationDialog, useConfirmationDialog } from "@/components/shared/confirmation-dialog"
import { AppLayoutStack } from "@/components/ui/app-layout"
import { Button } from "@/components/ui/button"
import { Modal } from "@/components/ui/modal"
import { atom, useAtom } from "jotai"
import React from "react"
import { BiLockAlt, BiLockOpenAlt } from "react-icons/bi"
import { toast } from "sonner"

export const __bulkAction_modalAtomIsOpen = atom<boolean>(false)

export function BulkActionModal() {

    const [isOpen, setIsOpen] = useAtom(__bulkAction_modalAtomIsOpen)

    const { mutate: performBulkAction, isPending } = useLocalFileBulkAction()

    function handleLockFiles() {
        performBulkAction({
            action: "lock",
        }, {
            onSuccess: () => {
                setIsOpen(false)
                toast.success("Files locked")
            },
        })
    }

    function handleUnlockFiles() {
        performBulkAction({
            action: "unlock",
        }, {
            onSuccess: () => {
                setIsOpen(false)
                toast.success("Files unlocked")
            },
        })
    }

    const { mutate: removeEmptyDirectories, isPending: isRemoving } = useRemoveEmptyDirectories()

    function handleRemoveEmptyDirectories() {
        removeEmptyDirectories(undefined, {
            onSuccess: () => {
                setIsOpen(false)
            },
        })
    }

    const confirmRemoveEmptyDirs = useConfirmationDialog({
        title: "Remove empty directories",
        description: "This action will remove all empty directories in the library. Are you sure you want to continue?",
        onConfirm: () => {
            handleRemoveEmptyDirectories()
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
                    onClick={handleLockFiles}
                >
                    Lock all files
                </Button>
                <Button
                    leftIcon={<BiLockOpenAlt />}
                    intent="gray-outline"
                    className="w-full"
                    disabled={isPending || isRemoving}
                    onClick={handleUnlockFiles}
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
