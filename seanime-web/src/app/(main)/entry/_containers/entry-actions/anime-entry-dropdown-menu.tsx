"use client"
import { Anime_MediaEntry } from "@/api/generated/types"
import { useAnimeEntryBulkAction, useOpenAnimeEntryInExplorer } from "@/api/hooks/anime_entries.hooks"
import { useStartDefaultMediaPlayer } from "@/api/hooks/mediaplayer.hooks"
import { useServerStatus } from "@/app/(main)/_hooks/use-server-status"
import {
    __bulkDeleteFilesModalIsOpenAtom,
    AnimeEntryBulkDeleteFilesModal,
} from "@/app/(main)/entry/_containers/entry-actions/anime-entry-bulk-delete-files-modal"
import { __metadataManager_isOpenAtom, AnimeEntryMetadataManager } from "@/app/(main)/entry/_containers/entry-actions/anime-entry-metadata-manager"
import { useMediastreamMediaToTranscode } from "@/app/(main)/mediastream/_lib/mediastream.atoms"
import { ConfirmationDialog, useConfirmationDialog } from "@/components/shared/confirmation-dialog"
import { IconButton } from "@/components/ui/button"
import { Checkbox } from "@/components/ui/checkbox"
import { DropdownMenu, DropdownMenuItem, DropdownMenuLabel, DropdownMenuSeparator } from "@/components/ui/dropdown-menu"
import { useSetAtom } from "jotai"
import React from "react"
import { BiDotsVerticalRounded, BiRightArrowAlt } from "react-icons/bi"
import { toast } from "sonner"

export function AnimeEntryDropdownMenu({ entry }: { entry: Anime_MediaEntry }) {

    const serverStatus = useServerStatus()
    const setIsMetadataManagerOpen = useSetAtom(__metadataManager_isOpenAtom)

    const inLibrary = !!entry.libraryData

    // Start default media player
    const { mutate: startDefaultMediaPlayer } = useStartDefaultMediaPlayer()
    // Open entry in explorer
    const { mutate: openEntryInExplorer } = useOpenAnimeEntryInExplorer()
    // File bulk actions
    const { mutate: performBulkAction, isPending } = useAnimeEntryBulkAction(entry.mediaId)

    // Media streaming
    const { addMediaToTranscode, mediaToTranscode, removeMediaToTranscode } = useMediastreamMediaToTranscode()

    const confirmDeleteFiles = useConfirmationDialog({
        title: "Unmatch all files",
        description: "Are you sure you want to unmatch all files?",
        onConfirm: () => {
            performBulkAction({
                mediaId: entry.mediaId,
                action: "unmatch",
            }, {
                onSuccess: () => {
                    setIsMetadataManagerOpen(false)
                    toast.success("Files unmatched")
                },
            })
        },
    })

    const setBulkDeleteFilesModalOpen = useSetAtom(__bulkDeleteFilesModalIsOpenAtom)


    return (
        <>
            <DropdownMenu trigger={<IconButton icon={<BiDotsVerticalRounded />} intent="gray-basic" size="lg" />}>

                {inLibrary && <>
                    <DropdownMenuItem
                        onClick={() => openEntryInExplorer({ mediaId: entry.mediaId })}
                    >
                        Open folder
                    </DropdownMenuItem>

                    {serverStatus?.settings?.mediaPlayer?.defaultPlayer != "mpv" && <DropdownMenuItem
                        onClick={() => startDefaultMediaPlayer}
                    >
                        Start video player
                    </DropdownMenuItem>}
                    <DropdownMenuSeparator />
                </>}

                <DropdownMenuItem
                    onClick={() => setIsMetadataManagerOpen(p => !p)}
                >
                    Metadata
                </DropdownMenuItem>

                {/*MEDIASTREAM*/}
                {(inLibrary && serverStatus?.mediastreamSettings?.transcodeEnabled) && <>
                    <DropdownMenuSeparator />

                    <div
                        className="px-2 py-1"
                    >
                        <Checkbox
                            label="Transcode on this device"
                            labelClass="text-sm"
                            containerClass="flex-row-reverse justify-between w-full"
                            value={mediaToTranscode.includes(String(entry.mediaId))}
                            onValueChange={(e) => {
                                if (e) {
                                    addMediaToTranscode(entry.mediaId)
                                } else {
                                    removeMediaToTranscode(entry.mediaId)
                                }
                            }}
                        />
                    </div>
                </>}

                {inLibrary && <>
                    <DropdownMenuSeparator />
                    <DropdownMenuLabel>Bulk actions</DropdownMenuLabel>
                    <DropdownMenuItem
                        className="text-red-500 dark:text-red-200 flex justify-between"
                        onClick={confirmDeleteFiles.open}
                        disabled={isPending}
                    >
                        <span>Unmatch all files</span> <BiRightArrowAlt />
                    </DropdownMenuItem>
                    <DropdownMenuItem
                        className="text-red-500 dark:text-red-200 flex justify-between"
                        onClick={() => setBulkDeleteFilesModalOpen(true)}
                        disabled={isPending}
                    >
                        <span>Delete some files</span> <BiRightArrowAlt />
                    </DropdownMenuItem>
                </>}
            </DropdownMenu>

            <AnimeEntryMetadataManager entry={entry} />
            <ConfirmationDialog {...confirmDeleteFiles} />
            <AnimeEntryBulkDeleteFilesModal entry={entry} />
        </>
    )
}
