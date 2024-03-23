import { useTVDBMetadata } from "@/app/(main)/entry/_lib/media-entry"
import { Button } from "@/components/ui/button"
import { Modal } from "@/components/ui/modal"
import { Separator } from "@/components/ui/separator"
import { atom } from "jotai"
import { useAtom } from "jotai/react"
import React from "react"

type MetadataManagerProps = {
    mediaId: number
}

export const __metadataManager_isOpenAtom = atom(false)

export function MetadataManager(props: MetadataManagerProps) {

    const { mediaId } = props

    const [isOpen, setOpen] = useAtom(__metadataManager_isOpenAtom)
    const { populate, empty, isPopulating, isEmptying } = useTVDBMetadata(mediaId)

    return (
        <Modal
            open={isOpen}
            onOpenChange={setOpen}
            title="Metadata"
            contentClass="max-w-xl"
            titleClass=""
        >
            <p className="text-[--muted]">
                You can add alternative metadata for this media entry. This can be useful if some images are missing.
            </p>

            <Separator />
            <h3 className="text-center">TVDB</h3>

            <Button
                intent="success-subtle"
                onClick={() => populate()}
                loading={isPopulating}
                disabled={isPopulating || isEmptying}
            >
                {isPopulating ? "Populating..." : "Fetch metadata"}
            </Button>
            <Button
                intent="gray-subtle"
                onClick={() => empty()}
                loading={isEmptying}
                disabled={isPopulating || isEmptying}
            >
                {isEmptying ? "Emptying..." : "Delete metadata"}
            </Button>
        </Modal>
    )
}
