import { useManuallyMatchLocalFiles } from "@/app/(main)/(library)/_containers/unmatched-files/_lib/manually-match-local-files"
import { useFetchMediaEntrySuggestions } from "@/app/(main)/entry/_lib/media-entry"
import { AppLayoutStack } from "@/components/ui/app-layout"
import { Button } from "@/components/ui/button"
import { cn } from "@/components/ui/core/styling"
import { Separator } from "@/components/ui/separator"
import { LoadingSpinner } from "@/components/ui/loading-spinner"
import { Drawer } from "@/components/ui/drawer"
import { NumberInput } from "@/components/ui/number-input"
import { RadioGroup } from "@/components/ui/radio-group"
import { useOpenInExplorer } from "@/lib/server/hooks"
import { UnmatchedGroup } from "@/lib/server/types"
import { BiLeftArrow } from "react-icons/bi"
import { BiRightArrow } from "react-icons/bi"
import { FcFolder } from "react-icons/fc"
import { FiSearch } from "react-icons/fi"
import { atom } from "jotai"
import { useAtom } from "jotai/react"
import Image from "next/image"
import React, { useCallback, useEffect, useState } from "react"
import { toast } from "sonner"

export const _unmatchedFileManagerIsOpen = atom(false)

type UnmatchedFileManagerProps = {
    unmatchedGroups: UnmatchedGroup[]
}

export function UnmatchedFileManager(props: UnmatchedFileManagerProps) {

    const { unmatchedGroups } = props

    const [isOpen, setIsOpen] = useAtom(_unmatchedFileManagerIsOpen)
    const [page, setPage] = useState(0)
    const maxPage = unmatchedGroups.length - 1
    const [currentGroup, setCurrentGroup] = useState(unmatchedGroups?.[0])

    const [anilistId, setAnilistId] = useState(0)

    const { openInExplorer } = useOpenInExplorer()
    const {
        suggestions,
        fetchSuggestions,
        isPending: suggestionsLoading,
        resetSuggestions,
    } = useFetchMediaEntrySuggestions()
    const { manuallyMatchEntry, isPending: matchingLoading } = useManuallyMatchLocalFiles()

    const [_r, setR] = useState(0)

    const handleFetchSuggestions = useCallback(() => {
        fetchSuggestions(currentGroup.dir)
    }, [currentGroup?.dir, fetchSuggestions])

    const handleSelectAnime = useCallback((value: string | null) => {
        if (value && !isNaN(Number(value))) {
            setAnilistId(Number(value))
            setR(r => r + 1)
        }
    }, [])

    useEffect(() => {
        setPage(0)
        setCurrentGroup(unmatchedGroups[0])
    }, [isOpen, unmatchedGroups])

    useEffect(() => {
        setCurrentGroup(unmatchedGroups[page])
        setAnilistId(0)
        resetSuggestions()
    }, [page, unmatchedGroups])

    const AnilistIdInput = useCallback(() => {
        return <NumberInput
            value={anilistId}
            onValueChange={v => setAnilistId(v)}
        />
    }, [currentGroup?.dir, _r])

    function handleManuallyMatchEntry() {
        if (!!currentGroup && anilistId > 0) {
            manuallyMatchEntry({
                dir: currentGroup?.dir,
                mediaId: anilistId,
            }, () => {
                if (page === 0 && unmatchedGroups.length === 1) {
                    setIsOpen(false)
                }
                setAnilistId(0)
                resetSuggestions()
                setPage(0)
                setCurrentGroup(unmatchedGroups[0])
            })
        } else {
            toast.error("Invalid Anilist ID")
        }
    }

    useEffect(() => {
        if (!currentGroup) {
            setIsOpen(false)
        }
    }, [currentGroup])

    if (!currentGroup) return null

    return (
        <Drawer
            open={isOpen}
            onOpenChange={() => setIsOpen(false)}
            size="xl"
            title="Resolve unmatched"

        >
            <AppLayoutStack>

                <div className={cn("flex w-full justify-between", { "hidden": unmatchedGroups.length <= 1 })}>
                    <Button
                        intent="gray-subtle"
                        leftIcon={<BiLeftArrow/>}
                        disabled={page === 0}
                        onClick={() => {
                            setPage(p => p - 1)
                        }}
                        className={cn("transition-opacity", { "opacity-0": page === 0 })}
                    >Previous</Button>
                    <Button
                        intent="gray-subtle"
                        rightIcon={<BiRightArrow/>}
                        disabled={page >= maxPage}
                        onClick={() => {
                            setPage(p => p + 1)
                        }}
                        className={cn("transition-opacity", { "opacity-0": page >= maxPage })}
                    >Next</Button>
                </div>

                <div
                    className="bg-gray-800 border border-[--border] p-2 px-4 rounded-md line-clamp-1 flex gap-2 items-center cursor-pointer transition hover:bg-opacity-80"
                    onClick={() => openInExplorer(currentGroup.dir)}
                >
                    <FcFolder className="text-2xl"/>
                    {currentGroup.dir}
                </div>

                <ul className="list-disc pl-8 bg-[--background] p-2 px-4 rounded-md space-y-1 max-h-60 overflow-y-auto">
                    {currentGroup.localFiles.sort((a, b) => ((Number(a.parsedInfo?.episode ?? 0)) - (Number(b.parsedInfo?.episode ?? 0)))).map(lf => {
                        return <li key={lf.path} className="text-sm">
                            {lf.path}
                        </li>
                    })}
                </ul>

                {/*<Separator />*/}

                <div className="flex gap-2 items-center">
                    <p className="flex-none text-lg mr-2 font-semibold">Anilist ID</p>
                    <AnilistIdInput/>
                    <Button
                        intent="primary-outline"
                        onClick={handleManuallyMatchEntry}
                        loading={matchingLoading}
                    >Match</Button>
                </div>

                <Separator />

                <Button
                    leftIcon={<FiSearch/>}
                    intent="success-subtle"
                    onClick={handleFetchSuggestions}
                >
                    Fetch suggestions
                </Button>

                {suggestionsLoading && <LoadingSpinner/>}

                {(!suggestionsLoading && suggestions.length > 0) && <RadioGroup
                    defaultValue="1"
                    fieldClass="w-full"
                    fieldLabelClass="text-md"
                    label="Select Anime"
                    value={String(anilistId)}
                    onValueChange={handleSelectAnime}
                    options={suggestions.map((media) => (
                        {
                            label: media.title?.userPreferred || media.title?.english || media.title?.romaji || "",
                            value: String(media.id) || "",
                            help: <div className="mt-2 flex w-full gap-4">
                                {media.coverImage?.medium && <div
                                    className="h-28 w-28 flex-none rounded-md object-cover object-center relative overflow-hidden">
                                    <Image
                                        src={media.coverImage.medium}
                                        alt={""}
                                        fill
                                        quality={100}
                                        priority
                                        sizes="10rem"
                                        className="object-cover object-center"
                                    />
                                </div>}
                                <div className="text-[--muted]">
                                    {/*<p className="line-clamp-1">{media.title?.userPreferred || media.title?.english || media.title?.romaji}</p>*/}
                                    <p>Type: <span
                                        className="text-gray-200 font-semibold"
                                    >{media.format}</span>
                                    </p>
                                    <p>Aired: {media.startDate?.year ? new Intl.DateTimeFormat("en-US", {
                                        year: "numeric",
                                    }).format(new Date(media.startDate?.year || 0, media.startDate?.month || 0)) : "-"}</p>
                                    <p>Status: {media.status}</p>
                                    <Button
                                        intent="primary-link"
                                        size="sm"
                                        className="px-0"
                                        onClick={() => window.open(`https://anilist.co/anime/${media.id}`, "_target")}
                                    >Open on AniList</Button>
                                </div>
                            </div>,
                        }
                    ))}
                    // TODO
                    // radioContainerClass="block w-full p-4 cursor-pointer dark:bg-gray-900 transition border border-[--border] rounded-[--radius]
                    // data-[checked=true]:ring-2 ring-[--ring]" radioControlClass="absolute right-2 top-2 h-5 w-5 text-xs" radioHelpClass="text-sm"
                    // radioLabelClass="font-semibold flex-none w-[90%] line-clamp-1" stackClass="grid grid-cols-2 gap-2 space-y-0"
                />}

            </AppLayoutStack>
        </Drawer>
    )

}
