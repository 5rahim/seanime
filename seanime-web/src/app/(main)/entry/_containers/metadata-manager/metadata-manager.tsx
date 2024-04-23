import { Anime_MediaEntry } from "@/api/generated/types"
import { useTVDBMetadata } from "@/app/(main)/entry/_lib/media-entry"
import { Button } from "@/components/ui/button"
import { Modal } from "@/components/ui/modal"
import { Separator } from "@/components/ui/separator"
import { atom } from "jotai"
import { useAtom } from "jotai/react"
import React from "react"
import { IoReloadCircle } from "react-icons/io5"

type MetadataManagerProps = {
    entry: Anime_MediaEntry
}

export const __metadataManager_isOpenAtom = atom(false)

export function MetadataManager(props: MetadataManagerProps) {

    const { entry } = props

    const [isOpen, setOpen] = useAtom(__metadataManager_isOpenAtom)
    const { populate, empty, isPopulating, isEmptying } = useTVDBMetadata(entry.mediaId)

    const cannotAddMetadata = entry.media?.format !== "TV" && entry.media?.format !== "TV_SHORT"

    return (
        <Modal
            open={isOpen}
            onOpenChange={setOpen}
            title="Metadata"
            contentClass="max-w-xl"
            titleClass=""
        >
            <p className="text-[--muted]">
                Having issues with missing images? Try fetching metadata from other sources.
            </p>

            <Separator />
            <h3 className="text-center">TVDB</h3>

            <Button
                intent="success-subtle"
                onClick={() => populate()}
                loading={isPopulating}
                disabled={isPopulating || isEmptying || cannotAddMetadata}
                leftIcon={<IoReloadCircle className="text-xl" />}
            >
                {isPopulating ? "Populating..." : "Fetch / Reload TVDB metadata"}
            </Button>
            <Button
                intent="gray-subtle"
                onClick={() => empty()}
                loading={isEmptying}
                disabled={isPopulating || isEmptying || cannotAddMetadata}
            >
                {isEmptying ? "Emptying..." : "Remove TVDB metadata"}
            </Button>
        </Modal>
    )
}
