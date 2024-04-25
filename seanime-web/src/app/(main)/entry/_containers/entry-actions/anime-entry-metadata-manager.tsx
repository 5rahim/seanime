import { Anime_MediaEntry } from "@/api/generated/types"
import { useEmptyTVDBEpisodes, usePopulateTVDBEpisodes } from "@/api/hooks/metadata.hooks"
import { Button } from "@/components/ui/button"
import { Modal } from "@/components/ui/modal"
import { Separator } from "@/components/ui/separator"
import { atom } from "jotai"
import { useAtom } from "jotai/react"
import React from "react"
import { IoReloadCircle } from "react-icons/io5"

type AnimeEntryMetadataManagerProps = {
    entry: Anime_MediaEntry
}

export const __metadataManager_isOpenAtom = atom(false)

export function AnimeEntryMetadataManager(props: AnimeEntryMetadataManagerProps) {

    const { entry } = props

    const [isOpen, setOpen] = useAtom(__metadataManager_isOpenAtom)
    const { mutate: populate, isPending: isPopulating } = usePopulateTVDBEpisodes()
    const { mutate: empty, isPending: isEmptying } = useEmptyTVDBEpisodes()

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
                onClick={() => populate({ mediaId: entry.mediaId })}
                loading={isPopulating}
                disabled={isPopulating || isEmptying || cannotAddMetadata}
                leftIcon={<IoReloadCircle className="text-xl" />}
            >
                {isPopulating ? "Populating..." : "Fetch / Reload TVDB metadata"}
            </Button>
            <Button
                intent="gray-subtle"
                onClick={() => empty({ mediaId: entry.mediaId })}
                loading={isEmptying}
                disabled={isPopulating || isEmptying || cannotAddMetadata}
            >
                {isEmptying ? "Emptying..." : "Remove TVDB metadata"}
            </Button>
        </Modal>
    )
}
