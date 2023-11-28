import { atom, useAtom } from "jotai"
import { Modal } from "@/components/ui/modal"
import React from "react"
import { Button } from "@/components/ui/button"
import { AppLayoutStack } from "@/components/ui/app-layout"
import { BiLockAlt } from "@react-icons/all-files/bi/BiLockAlt"
import { BiLockOpenAlt } from "@react-icons/all-files/bi/BiLockOpenAlt"
import { useLocalFileBulkAction } from "@/lib/server/hooks/library"

export const bulkActionModalAtomIsOpen = atom<boolean>(false)

export function BulkActionModal() {

    const [isOpen, setIsOpen] = useAtom(bulkActionModalAtomIsOpen)

    const { lockFiles, unlockFiles, isPending } = useLocalFileBulkAction()

    return (
        <Modal isOpen={isOpen} onClose={() => setIsOpen(false)} isClosable title={"Bulk actions"}
               bodyClassName={"space-y-4"}>
            <AppLayoutStack spacing={"sm"}>
                {/*<p>These actions do not affect ignored files.</p>*/}
                <Button
                    leftIcon={<BiLockAlt/>}
                    intent={"white-subtle"}
                    className={"w-full"}
                    isLoading={isPending}
                    onClick={lockFiles}
                >
                    Lock all files
                </Button>
                <Button
                    leftIcon={<BiLockOpenAlt/>}
                    intent={"white-subtle"}
                    className={"w-full"}
                    isLoading={isPending}
                    onClick={unlockFiles}
                >
                    Unlock all files
                </Button>
            </AppLayoutStack>
        </Modal>
    )

}