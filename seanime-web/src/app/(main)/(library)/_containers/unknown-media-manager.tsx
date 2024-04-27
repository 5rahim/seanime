import { Anime_UnknownGroup } from "@/api/generated/types"
import { useAddUnknownMedia } from "@/api/hooks/anime_collection.hooks"
import { useAnimeEntryBulkAction } from "@/api/hooks/anime_entries.hooks"
import { AppLayoutStack } from "@/components/ui/app-layout"
import { Button } from "@/components/ui/button"
import { Drawer } from "@/components/ui/drawer"
import { atom } from "jotai"
import { useAtom } from "jotai/react"
import Link from "next/link"
import React, { useCallback } from "react"
import { BiLinkExternal } from "react-icons/bi"
import { TbDatabasePlus } from "react-icons/tb"
import { toast } from "sonner"

export const __unknownMedia_drawerIsOpen = atom(false)

type UnknownMediaManagerProps = {
    unknownGroups: Anime_UnknownGroup[]
}

export function UnknownMediaManager(props: UnknownMediaManagerProps) {

    const { unknownGroups } = props

    const [isOpen, setIsOpen] = useAtom(__unknownMedia_drawerIsOpen)

    const { mutate: addUnknownMedia, isPending: isAdding } = useAddUnknownMedia()
    const { mutate: performBulkAction, isPending: isUnmatching } = useAnimeEntryBulkAction()

    /**
     * Add all unknown media to AniList
     */
    const handleAddUnknownMedia = useCallback(() => {
        addUnknownMedia({ mediaIds: unknownGroups.map(n => n.mediaId) })
    }, [unknownGroups])

    /**
     * Close the drawer if there are no unknown groups
     */
    React.useEffect(() => {
        if (unknownGroups.length === 0) {
            setIsOpen(false)
        }
    }, [unknownGroups])

    /**
     * Unmatch all files for a media
     */
    const handleUnmatchMedia = useCallback((mediaId: number) => {
        performBulkAction({
            mediaId,
            action: "unmatch",
        }, {
            onSuccess: () => {
                toast.success("Media unmatched")
                setIsOpen(false)
            },
        })
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
                                    {group.localFiles?.sort((a, b) => ((Number(a.parsedInfo?.episode ?? 0)) - (Number(b.parsedInfo?.episode ?? 0))))
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
