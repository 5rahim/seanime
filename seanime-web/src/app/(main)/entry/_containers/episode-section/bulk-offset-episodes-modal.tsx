import { Button } from "@/components/ui/button"
import { Checkbox } from "@/components/ui/checkbox"
import { Modal } from "@/components/ui/modal"
import { NumberInput } from "@/components/ui/number-input"
import { Select } from "@/components/ui/select"
import { LocalFile, MediaEntry } from "@/lib/server/types"
import { getAniDBEpisodeInteger } from "@/lib/server/utils"
import { Nullish } from "@/types/common"
import { atomWithImmer } from "jotai-immer"
import { useAtom } from "jotai/react"
import React, { useCallback, useEffect, useState } from "react"
import { toast } from "sonner"
import * as upath from "upath"

export type BulkOffsetEpisodesModalProps = {
    entry: MediaEntry
    isOpen: boolean
    onClose: () => void
}

export const _bulkOffsetEpisodesModalIsOpenAtom = atomWithImmer(false)
const _episodeOffsetActionFilesAtom = atomWithImmer<{ file: LocalFile, selected: boolean }[]>([])


export function BulkOffsetEpisodesModal({ entry, isOpen, onClose }: BulkOffsetEpisodesModalProps) {

    return (
        <Modal
            open={isOpen}
            onOpenChange={onClose}
            contentClass="max-w-2xl"
            title={<span>Offset episode numbers</span>}
            titleClass="text-center"

        >
            <Content entry={entry}/>
        </Modal>
    )

}

function Content({ entry }: { entry: MediaEntry }) {

    const [state, setState] = useAtom(_bulkOffsetEpisodesModalIsOpenAtom)
    const [files, setFiles] = useAtom(_episodeOffsetActionFilesAtom)
    const [offset, setOffset] = useState(0)
    const [area, setArea] = useState<"episode" | "aniDBEpisode">("episode")

    const media = entry.media
    const localFiles = entry.localFiles

    useEffect(() => {
        setFiles(localFiles.map(file => ({ file, selected: true })).toSorted((a, b) => a.file.metadata.episode - b.file.metadata.episode))
    }, [state, localFiles])

    function applyOffset() {
        // setLocalFiles(draft => {
        //     for (const { file, selected } of files) {
        //         const index = draft.findIndex(f => f.path === file.path)
        //         if (selected && index !== -1) {
        //             if (area === "episode") {
        //                 draft[index].metadata.episode = calculateOffset(file.metadata.episode)
        //             } else if (area === "aniDBEpisode") {
        //                 draft[index].metadata.aniDBEpisode = String(calculateOffset(localFile_getAniDBEpisodeInteger(file)))
        //             }
        //         }
        //     }
        //     return
        // })
        // rerenderLocalFiles()
        toast.success("Offset applied")
        setState(false)
    }

    const getEpisode = useCallback((file: LocalFile) => {
        if (area === "episode") return file.metadata.episode!
        else if (area === "aniDBEpisode") return (getAniDBEpisodeInteger({ metadata: file.metadata }) || 0)
        else return 0
    }, [area])

    useEffect(() => {
        setOffset(0)
    }, [files, media])

    function calculateOffset(currentEpisode: Nullish<number>) {
        if (currentEpisode === undefined || currentEpisode === null) return 0
        // Make sure it is not less than 0
        return Math.max(0, currentEpisode + offset)
    }

    // Re-render the offset input when a selection changes
    // This is a workaround to make sure that the `files` referenced by the input is always up-to-date
    const OffsetInput = useCallback(() => {
        return !!media && <NumberInput
            label="Offset"
            value={offset}
            onValueChange={value => {
                const episodesArr = files.filter(n => n.selected).map(({ file }) => Math.max(0, getEpisode(file)! + value))
                // Make sure than we can't go any further below if one episode calculated offset is 0
                if (value <= 0) {
                    const minOffset = episodesArr.filter(n => n === 0).length
                    if (minOffset > 1) return
                }
                // Make sure than we can't go any further above if one episode calculated offset is greater than the total number of episodes
                const maxOffset = Math.max(...episodesArr)
                if (!!media.episodes && maxOffset > media.episodes) {
                    return
                }
                setOffset(value)
            }}
            min={-Infinity}
            step={1}
            formatOptions={{
                minimumFractionDigits: 0,
                maximumFractionDigits: 0,

            }}
        />
    }, [files, media, area])

    if (!media) return null

    return (
        <div className="space-y-2 mt-2">
            <div className="max-h-96 overflow-y-auto px-2 space-y-2">
                <Select
                    label="Target"
                    value={area}
                    options={[
                        { label: "Episode number", value: "episode" },
                        { label: "AniDB episode", value: "aniDBEpisode" },
                    ]}
                    onValueChange={v => {
                        setArea(v as any)
                    }}
                />
                {<OffsetInput/>}
                {files.map(({ file, selected }, index) => (
                    <div
                        key={`${file.path}-${index}`}
                        className="p-2 border-b "
                    >
                        <div className="flex items-center">
                            <Checkbox
                                label={`${area === "episode" ? "Episode" : "AniDB Episode"} ${getEpisode(file)}`}
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
                            {selected && <p
                                className="text-[--muted] line-clamp-1 ml-2 flex-none -mt-1"
                            >
                                {`->`} <span
                                className="font-medium text-brand-300"
                            >{calculateOffset(getEpisode(file))}</span>
                            </p>}
                        </div>
                        <div>
                            <p className="text-[--muted] text-sm line-clamp-1">
                                {upath.basename(file.path)}
                            </p>
                        </div>
                    </div>
                ))}
            </div>
            <div className="flex justify-end gap-2 mt-2">
                <Button
                    intent="primary"
                    onClick={() => applyOffset()}
                >
                    Apply
                </Button>
                <Button
                    intent="white"
                    onClick={() => setState(false)}
                >
                    Cancel
                </Button>
            </div>
        </div>
    )
}
