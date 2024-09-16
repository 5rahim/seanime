import { Anime_Entry } from "@/api/generated/types"
import { useEmptyTVDBEpisodes, usePopulateFillerData, usePopulateTVDBEpisodes, useRemoveFillerData } from "@/api/hooks/metadata.hooks"
import { Button } from "@/components/ui/button"
import { Modal } from "@/components/ui/modal"
import { Separator } from "@/components/ui/separator"
import { atom } from "jotai"
import { useAtom } from "jotai/react"
import React from "react"
import { IoReloadCircle, IoSearchCircle } from "react-icons/io5"

type AnimeEntryMetadataManagerProps = {
    entry: Anime_Entry
}

export const __metadataManager_isOpenAtom = atom(false)

export function AnimeEntryMetadataManager(props: AnimeEntryMetadataManagerProps) {

    const { entry } = props

    const [isOpen, setOpen] = useAtom(__metadataManager_isOpenAtom)
    const { mutate: tvdb_populate, isPending: tvdb_isPopulating } = usePopulateTVDBEpisodes()
    const { mutate: tvdb_empty, isPending: tvdb_isEmptying } = useEmptyTVDBEpisodes()

    const { mutate: filler_populate, isPending: filler_isPopulating } = usePopulateFillerData()
    const { mutate: filler_remove, isPending: filler_isRemoving } = useRemoveFillerData()

    const cannotAddMetadata = entry.media?.format !== "TV" && entry.media?.format !== "TV_SHORT" && entry.media?.format !== "ONA"

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

            <h3 className="text-center">TVDB</h3>

            <div className="flex sm:flex-row flex-col gap-2">
                <Button
                    className="w-full"
                    intent="success-subtle"
                    onClick={() => tvdb_populate({ mediaId: entry.mediaId })}
                    loading={tvdb_isPopulating}
                    disabled={tvdb_isPopulating || tvdb_isEmptying || cannotAddMetadata}
                    leftIcon={<IoReloadCircle className="text-xl" />}
                >
                    {tvdb_isPopulating ? "Populating..." : "Fetch / Reload TVDB metadata"}
                </Button>
                <Button
                    className="w-full"
                    intent="gray-subtle"
                    onClick={() => tvdb_empty({ mediaId: entry.mediaId })}
                    loading={tvdb_isEmptying}
                    disabled={tvdb_isPopulating || tvdb_isEmptying || cannotAddMetadata}
                >
                    {tvdb_isEmptying ? "Emptying..." : "Remove TVDB metadata"}
                </Button>
            </div>

            <Separator />

            <h3 className="text-center">AnimeFillerList</h3>

            <div className="flex lg:flex-row flex-col gap-2">
                <Button
                    className="w-full"
                    intent="primary-subtle"
                    leftIcon={<IoSearchCircle className="text-xl" />}
                    onClick={() => filler_populate({ mediaId: entry.mediaId })}
                    loading={filler_isPopulating}
                    disabled={filler_isPopulating || filler_isRemoving || cannotAddMetadata}
                >
                    {filler_isPopulating ? "Fetching..." : "Fetch filler info"}
                </Button>
                <Button
                    className="w-full"
                    intent="gray-subtle"
                    onClick={() => filler_remove({ mediaId: entry.mediaId })}
                    loading={filler_isRemoving}
                    disabled={filler_isPopulating || filler_isRemoving || cannotAddMetadata}
                >
                    {filler_isRemoving ? "Removing..." : "Remove filler info"}
                </Button>
            </div>
        </Modal>
    )
}
