import { useMediaEntryBulkAction } from "@/app/(main)/(library)/_containers/bulk-actions/_lib/media-entry-bulk-actions"
import { useAddUnknownMedia } from "@/app/(main)/(library)/_containers/unknown-media/_lib/add-unknown-media"
import { AppLayoutStack } from "@/components/ui/app-layout"
import { Button } from "@/components/ui/button"
import { Drawer } from "@/components/ui/drawer"

import { UnknownGroup } from "@/app/(main)/(library)/_lib/anime-library.types"
import { atom } from "jotai"
import { useAtom } from "jotai/react"
import Link from "next/link"
import React, { useCallback } from "react"
import { BiLinkExternal } from "react-icons/bi"
import { TbDatabasePlus } from "react-icons/tb"

export const _unknownMediaManagerIsOpen = atom(false)

type UnknownMediaManagerProps = {
    unknownGroups: UnknownGroup[]
}

export function UnknownMediaManager(props: UnknownMediaManagerProps) {

    const { unknownGroups } = props

    const [isOpen, setIsOpen] = useAtom(_unknownMediaManagerIsOpen)

    const { addUnknownMedia, isPending: isAdding } = useAddUnknownMedia()
    const { unmatchAll, isPending: isUnmatching } = useMediaEntryBulkAction()

    const handleAddUnknownMedia = useCallback(() => {
        addUnknownMedia({ mediaIds: unknownGroups.map(n => n.mediaId) })
    }, [unknownGroups])

    React.useEffect(() => {
        if (unknownGroups.length === 0) {
            setIsOpen(false)
        }
    }, [unknownGroups])

    const handleUnmatchMedia = useCallback((mediaId: number) => {
        unmatchAll(mediaId)
    }, [])

    if (unknownGroups.length === 0) return null

    return (
        <Drawer
            open={isOpen}
            onOpenChange={o => {
                if (!isAdding) {
                    setIsOpen(o)
                }
            }}
            size="xl"
            title="Resolve hidden media"

        >
            <AppLayoutStack className="mt-4">

                <p className="">
                    Seanime matched {unknownGroups.length} group{unknownGroups.length === 1 ? "" : "s"} to media that {unknownGroups.length === 1
                    ? "is"
                    : "are"} absent from your
                    AniList collection.<br />
                    Add the media to be able to see the entry in your library or unmatch them if incorrect.
                </p>

                <Button
                    leftIcon={<TbDatabasePlus />}
                    onClick={handleAddUnknownMedia}
                    loading={isAdding}
                    disabled={isUnmatching}
                >
                    Add all to AniList
                </Button>

                <div className="divide divide-y divide-[--border] space-y-4">

                    {unknownGroups.map(group => {
                        return (
                            <div key={group.mediaId} className="pt-4 space-y-2">
                                <div className="flex items-center w-full justify-between">
                                    <h4 className="font-semibold flex gap-2 items-center">
                                        <span>Anilist ID:{" "}</span>
                                        <Link
                                            href={`https://anilist.co/anime/${group.mediaId}`}
                                            target="_blank"
                                            className="underline text-brand-200 flex gap-1.5 items-center"
                                        >
                                            {group.mediaId} <BiLinkExternal />
                                        </Link>
                                    </h4>
                                    <div>
                                        <Button
                                            size="sm"
                                            intent="alert-subtle"
                                            disabled={isUnmatching}
                                            onClick={() => handleUnmatchMedia(group.mediaId)}
                                        >
                                            Unmatch media
                                        </Button>
                                    </div>
                                </div>
                                <div className="bg-gray-900 border p-2 px-2 rounded-md space-y-1 max-h-28 overflow-y-auto text-sm">
                                    {group.localFiles.sort((a, b) => ((Number(a.parsedInfo?.episode ?? 0)) - (Number(b.parsedInfo?.episode ?? 0))))
                                        .map(lf => {
                                            return <p key={lf.path} className="text-[--muted] line-clamp-1 tracking-wide">
                                                {lf.path}
                                            </p>
                                        })}
                                </div>
                            </div>
                        )

                    })}
                </div>

            </AppLayoutStack>
        </Drawer>
    )

}
