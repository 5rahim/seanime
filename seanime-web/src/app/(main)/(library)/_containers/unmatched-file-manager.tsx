import { Anime_UnmatchedGroup } from "@/api/generated/types"
import { useAnimeEntryManualMatch, useFetchAnimeEntrySuggestions } from "@/api/hooks/anime_entries.hooks"
import { useOpenInExplorer } from "@/api/hooks/explorer.hooks"
import { useUpdateLocalFiles } from "@/api/hooks/localfiles.hooks"
import { SeaLink } from "@/components/shared/sea-link"
import { AppLayoutStack } from "@/components/ui/app-layout"
import { Button } from "@/components/ui/button"
import { Checkbox } from "@/components/ui/checkbox"
import { cn } from "@/components/ui/core/styling"
import { Drawer } from "@/components/ui/drawer"
import { LoadingSpinner } from "@/components/ui/loading-spinner"
import { NumberInput } from "@/components/ui/number-input"
import { RadioGroup } from "@/components/ui/radio-group"
import { upath } from "@/lib/helpers/upath"
import { atom } from "jotai"
import { useAtom } from "jotai/react"
import Image from "next/image"
import React from "react"
import { FaArrowLeft, FaArrowRight } from "react-icons/fa"
import { FcFolder } from "react-icons/fc"
import { FiSearch } from "react-icons/fi"
import { TbFileSad } from "react-icons/tb"
import { toast } from "sonner"

export const __unmatchedFileManagerIsOpen = atom(false)

type UnmatchedFileManagerProps = {
    unmatchedGroups: Anime_UnmatchedGroup[]
}

export function UnmatchedFileManager(props: UnmatchedFileManagerProps) {

    const { unmatchedGroups } = props

    const [isOpen, setIsOpen] = useAtom(__unmatchedFileManagerIsOpen)
    const [page, setPage] = React.useState(0)
    const maxPage = unmatchedGroups.length - 1
    const [currentGroup, setCurrentGroup] = React.useState(unmatchedGroups?.[0])

    const [selectedPaths, setSelectedPaths] = React.useState<string[]>([])

    const [anilistId, setAnilistId] = React.useState(0)

    const { mutate: openInExplorer } = useOpenInExplorer()

    const {
        data: suggestions,
        mutate: fetchSuggestions,
        isPending: suggestionsLoading,
        reset: resetSuggestions,
    } = useFetchAnimeEntrySuggestions()

    const { mutate: updateLocalFiles, isPending: isUpdatingFile } = useUpdateLocalFiles()

    const { mutate: manualMatch, isPending: isMatching } = useAnimeEntryManualMatch()

    const isUpdating = isUpdatingFile || isMatching

    const [_r, setR] = React.useState(0)

    const handleSelectAnime = React.useCallback((value: string | null) => {
        if (value && !isNaN(Number(value))) {
            setAnilistId(Number(value))
            setR(r => r + 1)
        }
    }, [])

    // Reset the selected paths when the current group changes
    React.useLayoutEffect(() => {
        setSelectedPaths(currentGroup?.localFiles?.map(lf => lf.path) ?? [])
    }, [currentGroup])

    // Reset the current group and page when the drawer is opened
    React.useEffect(() => {
        setPage(0)
        setCurrentGroup(unmatchedGroups[0])
    }, [isOpen, unmatchedGroups])

    // Set the current group when the page changes
    React.useEffect(() => {
        setCurrentGroup(unmatchedGroups[page])
        setAnilistId(0)
        resetSuggestions()
    }, [page, unmatchedGroups])

    const AnilistIdInput = React.useCallback(() => {
        return <NumberInput
            value={anilistId}
            onValueChange={v => setAnilistId(v)}
            formatOptions={{
                useGrouping: false,
            }}
        />
    }, [currentGroup?.dir, _r])

    function onActionSuccess() {
        if (page === 0 && unmatchedGroups.length === 1) {
            setIsOpen(false)
        }
        setAnilistId(0)
        resetSuggestions()
        setPage(0)
        setCurrentGroup(unmatchedGroups[0])
    }

    /**
     * Manually match the current group with the specified Anilist ID.
     * If the current group is the last group and there are no more unmatched groups, close the drawer.
     */
    function handleMatchSelected() {
        if (!!currentGroup && anilistId > 0 && selectedPaths.length > 0) {
            manualMatch({
                paths: selectedPaths,
                mediaId: anilistId,
            }, {
                onSuccess: () => {
                    onActionSuccess()
                },
            })
        }
    }

    const handleFetchSuggestions = React.useCallback(() => {
        fetchSuggestions({
            dir: currentGroup.dir,
        })
    }, [currentGroup?.dir, fetchSuggestions])

    function handleIgnoreSelected() {
        if (selectedPaths.length > 0) {
            updateLocalFiles({
                paths: selectedPaths,
                action: "ignore",
            }, {
                onSuccess: () => {
                    onActionSuccess()
                    toast.success("Files ignored")
                },
            })
        }
    }

    React.useEffect(() => {
        if (!currentGroup) {
            setIsOpen(false)
        }
    }, [currentGroup])

    if (!currentGroup) return null

    return (
        <Drawer
            data-unmatched-file-manager-drawer
            open={isOpen}
            onOpenChange={() => setIsOpen(false)}
            // contentClass="max-w-5xl"
            size="xl"
            title="Unmatched files"
        >
            <AppLayoutStack className="mt-4">

                <div className={cn("flex w-full justify-between", { "hidden": unmatchedGroups.length <= 1 })}>
                    <Button
                        intent="gray-subtle"
                        leftIcon={<FaArrowLeft />}
                        disabled={page === 0}
                        onClick={() => {
                            setPage(p => p - 1)
                        }}
                        className={cn("transition-opacity", { "opacity-0": page === 0 })}
                    >Previous</Button>

                    <p>
                        {page + 1} / {maxPage + 1}
                    </p>

                    <Button
                        intent="gray-subtle"
                        rightIcon={<FaArrowRight />}
                        disabled={page >= maxPage}
                        onClick={() => {
                            setPage(p => p + 1)
                        }}
                        className={cn("transition-opacity", { "opacity-0": page >= maxPage })}
                    >Next</Button>
                </div>

                <div
                    className="bg-gray-900 border  p-2 px-4 rounded-[--radius-md] line-clamp-1 flex gap-2 items-center cursor-pointer transition hover:bg-opacity-80"
                    onClick={() => openInExplorer({
                        path: currentGroup.dir,
                    })}
                >
                    <FcFolder className="text-2xl" />
                    {currentGroup.dir}
                </div>

                <div className="flex items-center flex-wrap gap-2">
                    <div className="flex gap-2 items-center w-full">
                        <p className="flex-none text-lg mr-2 font-semibold">Anilist ID</p>
                        <AnilistIdInput />
                        <Button
                            intent="white"
                            onClick={handleMatchSelected}
                            disabled={isUpdating}
                        >Match selection</Button>
                    </div>

                    {/*<div className="flex flex-1">*/}
                    {/*</div>*/}
                </div>

                <div className="bg-gray-950 border p-2 px-2 divide-y divide-[--border] rounded-[--radius-md] max-h-[50vh] max-w-full overflow-x-auto overflow-y-auto text-sm">

                    <div className="p-2">
                        <Checkbox
                            label={`Select all files`}
                            value={(selectedPaths.length === currentGroup?.localFiles?.length) ? true : (selectedPaths.length === 0
                                ? false
                                : "indeterminate")}
                            onValueChange={checked => {
                                if (typeof checked === "boolean") {
                                    setSelectedPaths(draft => {
                                        if (draft.length === currentGroup?.localFiles?.length) {
                                            return []
                                        } else {
                                            return currentGroup?.localFiles?.map(lf => lf.path) ?? []
                                        }
                                    })
                                }
                            }}
                            fieldClass="w-[fit-content]"
                        />
                    </div>

                    {currentGroup.localFiles?.sort((a, b) => ((Number(a.parsedInfo?.episode ?? 0)) - (Number(b.parsedInfo?.episode ?? 0))))
                        .map((lf, index) => (
                            <div
                                key={`${lf.path}-${index}`}
                                className="p-2 "
                            >
                                <div className="flex items-center">
                                    <Checkbox
                                        label={`${upath.basename(lf.path)}`}
                                        value={selectedPaths.includes(lf.path)}
                                        onValueChange={checked => {
                                            if (typeof checked === "boolean") {
                                                setSelectedPaths(draft => {
                                                    if (checked) {
                                                        return [...draft, lf.path]
                                                    } else {
                                                        return draft.filter(p => p !== lf.path)
                                                    }
                                                })
                                            }
                                        }}
                                        labelClass="text-sm tracking-wide data-[checked=false]:opacity-50"
                                        fieldClass="w-[fit-content]"
                                    />
                                </div>
                            </div>
                        ))}
                </div>

                {/*<Separator />*/}

                {/*<Separator />*/}


                <div className="flex flex-wrap items-center gap-1">
                    <Button
                        leftIcon={<FiSearch />}
                        intent="primary-subtle"
                        onClick={handleFetchSuggestions}
                    >
                        Fetch suggestions
                    </Button>

                    <SeaLink
                        target="_blank"
                        href={`https://anilist.co/search/anime?search=${encodeURIComponent(currentGroup?.localFiles?.[0]?.parsedInfo?.title || currentGroup?.localFiles?.[0]?.parsedFolderInfo?.[0]?.title || "")}`}
                    >
                        <Button
                            intent="white-link"
                        >
                            Search on AniList
                        </Button>
                    </SeaLink>

                    <div className="flex flex-1"></div>

                    <Button
                        leftIcon={<TbFileSad className="text-lg" />}
                        intent="warning-subtle"
                        size="sm"
                        rounded
                        disabled={isUpdating}
                        onClick={handleIgnoreSelected}
                    >
                        Ignore selection
                    </Button>
                </div>

                {suggestionsLoading && <LoadingSpinner />}

                {(!suggestionsLoading && !!suggestions?.length) && <RadioGroup
                    defaultValue="1"
                    fieldClass="w-full"
                    fieldLabelClass="text-md"
                    label="Select Anime"
                    value={String(anilistId)}
                    onValueChange={handleSelectAnime}
                    options={suggestions?.map((media) => (
                        {
                            label: <div>
                                <p className="text-base md:text-md font-medium !-mt-1.5 line-clamp-1">{media.title?.userPreferred || media.title?.english || media.title?.romaji || "N/A"}</p>
                                <div className="mt-2 flex w-full gap-4">
                                    {media.coverImage?.medium && <div
                                        className="h-28 w-28 flex-none rounded-[--radius-md] object-cover object-center relative overflow-hidden"
                                    >
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
                                        <p>Type: <span
                                            className="text-gray-200 font-semibold"
                                        >{media.format}</span>
                                        </p>
                                        <p>Aired: {media.startDate?.year ? new Intl.DateTimeFormat("en-US", {
                                            year: "numeric",
                                        }).format(new Date(media.startDate?.year || 0, media.startDate?.month || 0)) : "-"}</p>
                                        <p>Status: {media.status}</p>
                                        <SeaLink href={`https://anilist.co/anime/${media.id}`} target="_blank">
                                            <Button
                                                intent="primary-link"
                                                size="sm"
                                                className="px-0"
                                            >Open on AniList</Button>
                                        </SeaLink>
                                    </div>
                                </div>
                            </div>,
                            value: String(media.id) || "",
                        }
                    ))}
                    stackClass="grid grid-cols-1 md:grid-cols-2 gap-2 space-y-0"
                    itemContainerClass={cn(
                        "items-start cursor-pointer transition border-transparent rounded-[--radius] p-4 w-full",
                        "bg-gray-50 hover:bg-[--subtle] dark:bg-gray-900",
                        "data-[state=checked]:bg-white dark:data-[state=checked]:bg-gray-950",
                        "focus:ring-2 ring-brand-100 dark:ring-brand-900 ring-offset-1 ring-offset-[--background] focus-within:ring-2 transition",
                        "border border-transparent data-[state=checked]:border-[--brand] data-[state=checked]:ring-offset-0",
                    )}
                    itemClass={cn(
                        "border-transparent absolute top-2 right-2 bg-transparent dark:bg-transparent dark:data-[state=unchecked]:bg-transparent",
                        "data-[state=unchecked]:bg-transparent data-[state=unchecked]:hover:bg-transparent dark:data-[state=unchecked]:hover:bg-transparent",
                        "focus-visible:ring-0 focus-visible:ring-offset-0 focus-visible:ring-offset-transparent",
                    )}
                    itemIndicatorClass="hidden"
                    itemLabelClass="font-medium flex flex-col items-center data-[state=checked]:text-[--brand] cursor-pointer"
                />}

            </AppLayoutStack>
        </Drawer>
    )

}
