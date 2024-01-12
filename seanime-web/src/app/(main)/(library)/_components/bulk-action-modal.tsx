import { ConfirmationDialog, useConfirmationDialog } from "@/components/application/confirmation-dialog"
import { AppLayoutStack } from "@/components/ui/app-layout"
import { Button } from "@/components/ui/button"
import { Modal } from "@/components/ui/modal"
import { useLocalFileBulkAction, useRemoveEmptyDirectories } from "@/lib/server/hooks/library"
import { BiLockAlt } from "@react-icons/all-files/bi/BiLockAlt"
import { BiLockOpenAlt } from "@react-icons/all-files/bi/BiLockOpenAlt"
import { atom, useAtom } from "jotai"
import React from "react"

export const bulkActionModalAtomIsOpen = atom<boolean>(false)

export function BulkActionModal() {

    const [isOpen, setIsOpen] = useAtom(bulkActionModalAtomIsOpen)

    const { lockFiles, unlockFiles, isPending } = useLocalFileBulkAction()

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
        <Modal isOpen={isOpen} onClose={() => setIsOpen(false)} isClosable title={"Bulk actions"}
               bodyClassName={"space-y-4"}>
            <AppLayoutStack spacing={"sm"}>
                {/*<p>These actions do not affect ignored files.</p>*/}
                <Button
                    leftIcon={<BiLockAlt/>}
                    intent={"white-link"}
                    className={"w-full"}
                    isDisabled={isPending || isRemoving}
                    onClick={lockFiles}
                >
                    Lock all files
                </Button>
                <Button
                    leftIcon={<BiLockOpenAlt/>}
                    intent={"white-link"}
                    className={"w-full"}
                    isDisabled={isPending || isRemoving}
                    onClick={unlockFiles}
                >
                    Unlock all files
                </Button>
                <Button
                    intent={"white-link"}
                    className={"w-full"}
                    isDisabled= {isPending}
                    isLoading={isRemoving}
                    onClick={() => confirmRemoveEmptyDirs.open()}
                >
                    Remove empty directories
                </Button>
            </AppLayoutStack>
            <ConfirmationDialog {...confirmRemoveEmptyDirs} />
        </Modal>
    )

}
