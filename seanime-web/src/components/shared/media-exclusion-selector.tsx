"use client"

import { Anime_LibraryCollectionEntry } from "@/api/generated/types"
import { animeLibraryCollectionAtom } from "@/app/(main)/_atoms/anime-library-collection.atoms"
import { imageShimmer } from "@/components/shared/image-helpers"
import { BasicField } from "@/components/ui/basic-field"
import { Button } from "@/components/ui/button"
import { cn } from "@/components/ui/core/styling"
import { LoadingSpinner } from "@/components/ui/loading-spinner"
import { Modal } from "@/components/ui/modal"
import { useAtomValue } from "jotai/react"
import Image from "next/image"
import React from "react"
import { BiEdit } from "react-icons/bi"
import { RiCloseCircleFill } from "react-icons/ri"

export type MediaExclusionSelectorProps = {
    value?: number[]
    onChange?: (value: number[]) => void
    onBlur?: () => void
    disabled?: boolean
    error?: string
    label?: string
    help?: string
    required?: boolean
    name?: string
}

export const MediaExclusionSelector = React.forwardRef<HTMLDivElement, MediaExclusionSelectorProps>(
    (props, ref) => {
        const {
            value = [],
            onChange,
            onBlur,
            disabled = false,
            error,
            label,
            help,
            required,
            name,
            ...rest
        } = props

        const _animeLibraryCollection = useAtomValue(animeLibraryCollectionAtom)
        const animeLibraryCollectionEntries = _animeLibraryCollection?.lists?.flatMap(n => n.entries)?.filter(n => !!n?.libraryData)?.filter(Boolean)
        const [selectedIds, setSelectedIds] = React.useState<number[]>(value)
        const [modalOpen, setModalOpen] = React.useState(false)

        React.useEffect(() => {
            setSelectedIds(value)
        }, [value])

        const handleToggleMedia = React.useCallback((mediaId: number) => {
            const newSelectedIds = selectedIds.includes(mediaId)
                ? selectedIds.filter(id => id !== mediaId)
                : [...selectedIds, mediaId]

            setSelectedIds(newSelectedIds)
            onChange?.(newSelectedIds)
        }, [selectedIds, onChange])

        const handleSelectAll = React.useCallback(() => {
            if (!animeLibraryCollectionEntries) return

            const allMediaIds: number[] = []
            animeLibraryCollectionEntries?.forEach(entry => {
                if (entry.mediaId && !allMediaIds.includes(entry.mediaId)) {
                    allMediaIds.push(entry.mediaId)
                }
            })

            setSelectedIds(allMediaIds)
            onChange?.(allMediaIds)
        }, [animeLibraryCollectionEntries, onChange])

        const handleDeselectAll = React.useCallback(() => {
            setSelectedIds([])
            onChange?.([])
        }, [onChange])

        const handleSelectAdult = React.useCallback(() => {
            if (!animeLibraryCollectionEntries) return

            const adultMediaIds: number[] = []
            animeLibraryCollectionEntries?.forEach(entry => {
                if (entry.media?.isAdult && entry.mediaId && !adultMediaIds.includes(entry.mediaId)) {
                    adultMediaIds.push(entry.mediaId)
                }
            })

            const newSelectedIds = [...new Set([...selectedIds, ...adultMediaIds])]
            setSelectedIds(newSelectedIds)
            onChange?.(newSelectedIds)
        }, [animeLibraryCollectionEntries, selectedIds, onChange])

        const lists = React.useMemo(() => {
            if (!animeLibraryCollectionEntries) return {
                CURRENT: [],
                PLANNING: [],
                COMPLETED: [],
                PAUSED: [],
                DROPPED: [],
            }

            return {
                CURRENT: animeLibraryCollectionEntries
                    ?.filter(Boolean)
                    ?.toSorted((a, b) => a.media!.title!.userPreferred!.localeCompare(b.media!.title!.userPreferred!)) ?? [],
                // PLANNING: animeLibraryCollection.lists
                //     .find(n => n.type === "PLANNING")
                //     ?.entries?.filter(Boolean)
                //     ?.toSorted((a, b) => a.media!.title!.userPreferred!.localeCompare(b.media!.title!.userPreferred!)) ?? [],
                // COMPLETED: animeLibraryCollection.lists
                //     .find(n => n.type === "COMPLETED")
                //     ?.entries?.filter(Boolean)
                //     ?.toSorted((a, b) => a.media!.title!.userPreferred!.localeCompare(b.media!.title!.userPreferred!)) ?? [],
                // PAUSED: animeLibraryCollection.lists
                //     .find(n => n.type === "PAUSED")
                //     ?.entries?.filter(Boolean)
                //     ?.toSorted((a, b) => a.media!.title!.userPreferred!.localeCompare(b.media!.title!.userPreferred!)) ?? [],
                // DROPPED: animeLibraryCollection.lists
                //     .find(n => n.type === "DROPPED")
                //     ?.entries?.filter(Boolean)
                //     ?.toSorted((a, b) => a.media!.title!.userPreferred!.localeCompare(b.media!.title!.userPreferred!)) ?? [],
            }
        }, [animeLibraryCollectionEntries])

        // Get preview items for display
        const selectedEntries = React.useMemo(() => {
            if (!animeLibraryCollectionEntries) return []

            const allEntries: Anime_LibraryCollectionEntry[] = []
            animeLibraryCollectionEntries?.forEach(entry => {
                    if (selectedIds.includes(entry.mediaId)) {
                        allEntries.push(entry)
                    }
                })
            return allEntries.slice(0, 5)
        }, [animeLibraryCollectionEntries, selectedIds])

        if (!animeLibraryCollectionEntries) {
            return (
                <BasicField
                    label={label}
                    help={help}
                    error={error}
                    required={required}
                >
                    <div className="flex items-center justify-center p-8">
                        <LoadingSpinner />
                    </div>
                </BasicField>
            )
        }

        return (
            <BasicField
                label={label}
                help={help}
                error={error}
                required={required}
                ref={ref}
                {...rest}
            >
                <div className="space-y-3">
                    <div className="flex items-center gap-3 p-4 border rounded-[--radius-md] bg-gray-900">
                        <div className="flex-1">
                            <div className="flex items-center gap-2 mb-2">
                                <span className="text-sm font-medium">
                                    {selectedIds.length} anime excluded from sharing
                                </span>
                                {selectedIds.length > 0 && (
                                    <span className="text-xs text-[--muted]">(will not be visible to other clients)</span>
                                )}
                            </div>

                            {selectedEntries.length > 0 && (
                                <div className="flex items-center gap-2">
                                    <div className="flex -space-x-1">
                                        {selectedEntries.map(entry => (
                                            <div
                                                key={entry.mediaId}
                                                className="size-8 rounded-md overflow-hidden border-2 border-white dark:border-gray-900"
                                            >
                                                <Image
                                                    src={entry.media?.coverImage?.medium || entry.media?.coverImage?.large || ""}
                                                    placeholder={imageShimmer(200, 280)}
                                                    width={32}
                                                    height={32}
                                                    alt=""
                                                    className="object-cover size-full"
                                                />
                                            </div>
                                        ))}
                                    </div>
                                    {selectedIds.length > 5 && (
                                        <span className="text-xs text-[--muted]">
                                            +{selectedIds.length - 5} more
                                        </span>
                                    )}
                                </div>
                            )}
                        </div>

                        <Modal
                            title="Select anime to exclude from sharing"
                            contentClass="max-w-6xl"
                            open={modalOpen}
                            onOpenChange={setModalOpen}
                            trigger={
                                <Button
                                    type="button"
                                    intent="gray-subtle"
                                    size="sm"
                                    leftIcon={<BiEdit />}
                                    disabled={disabled}
                                >
                                    {selectedIds.length > 0 ? "Edit selection" : "Select anime"}
                                </Button>
                            }
                        >
                            <div className="space-y-4">
                                <p className="text-[--muted]">
                                    Select anime that you don't want to share with other clients. Selected anime will not be visible to connected
                                    clients.
                                </p>

                                <div className="flex items-center gap-2 flex-wrap p-4 bg-[--subtle] rounded-[--radius-md]">
                                    <Button
                                        type="button"
                                        intent="gray-subtle"
                                        size="sm"
                                        onClick={handleSelectAll}
                                        disabled={disabled}
                                    >
                                        Select all
                                    </Button>
                                    <Button
                                        type="button"
                                        intent="gray-subtle"
                                        size="sm"
                                        onClick={handleDeselectAll}
                                        disabled={disabled}
                                    >
                                        Deselect all
                                    </Button>
                                    <Button
                                        type="button"
                                        intent="gray-subtle"
                                        size="sm"
                                        onClick={handleSelectAdult}
                                        disabled={disabled}
                                    >
                                        Select adult
                                    </Button>
                                    <div className="flex-1" />
                                    <span className="text-sm text-[--muted]">
                                        {selectedIds.length} selected (will not be shared)
                                    </span>
                                </div>

                                <div className="space-y-6 max-h-[60vh] overflow-y-auto p-1">
                                    {!!lists.CURRENT.length && (
                                        <MediaSection
                                            title="All"
                                            entries={lists.CURRENT}
                                            selectedIds={selectedIds}
                                            onToggle={handleToggleMedia}
                                            disabled={disabled}
                                        />
                                    )}
                                    {/*{!!lists.PAUSED.length && (*/}
                                    {/*    <MediaSection*/}
                                    {/*        title="Paused"*/}
                                    {/*        entries={lists.PAUSED}*/}
                                    {/*        selectedIds={selectedIds}*/}
                                    {/*        onToggle={handleToggleMedia}*/}
                                    {/*        disabled={disabled}*/}
                                    {/*    />*/}
                                    {/*)}*/}
                                    {/*{!!lists.PLANNING.length && (*/}
                                    {/*    <MediaSection*/}
                                    {/*        title="Planning"*/}
                                    {/*        entries={lists.PLANNING}*/}
                                    {/*        selectedIds={selectedIds}*/}
                                    {/*        onToggle={handleToggleMedia}*/}
                                    {/*        disabled={disabled}*/}
                                    {/*    />*/}
                                    {/*)}*/}
                                    {/*{!!lists.COMPLETED.length && (*/}
                                    {/*    <MediaSection*/}
                                    {/*        title="Completed"*/}
                                    {/*        entries={lists.COMPLETED}*/}
                                    {/*        selectedIds={selectedIds}*/}
                                    {/*        onToggle={handleToggleMedia}*/}
                                    {/*        disabled={disabled}*/}
                                    {/*    />*/}
                                    {/*)}*/}
                                    {/*{!!lists.DROPPED.length && (*/}
                                    {/*    <MediaSection*/}
                                    {/*        title="Dropped"*/}
                                    {/*        entries={lists.DROPPED}*/}
                                    {/*        selectedIds={selectedIds}*/}
                                    {/*        onToggle={handleToggleMedia}*/}
                                    {/*        disabled={disabled}*/}
                                    {/*    />*/}
                                    {/*)}*/}
                                </div>

                                <div className="flex justify-end pt-4 border-t">
                                    <Button
                                        type="button"
                                        intent="primary"
                                        onClick={() => setModalOpen(false)}
                                    >
                                        Done ({selectedIds.length} selected)
                                    </Button>
                                </div>
                            </div>
                        </Modal>
                    </div>
                </div>
            </BasicField>
        )
    },
)

MediaExclusionSelector.displayName = "MediaExclusionSelector"

function MediaSection(props: {
    title: string
    entries: Anime_LibraryCollectionEntry[]
    selectedIds: number[]
    onToggle: (mediaId: number) => void
    disabled?: boolean
}) {
    const { title, entries, selectedIds, onToggle, disabled } = props

    return (
        <div className="space-y-2">
            <h4 className="border-b pb-1 mb-1">{title}</h4>
            <div className="grid grid-cols-3 md:grid-cols-6 2xl:grid-cols-7 gap-2">
                {entries.map(entry => (
                    <MediaExclusionItem
                        key={entry.mediaId}
                        entry={entry}
                        isSelected={selectedIds.includes(entry.mediaId)}
                        onToggle={() => onToggle(entry.mediaId)}
                        disabled={disabled}
                    />
                ))}
            </div>
        </div>
    )
}

function MediaExclusionItem(props: {
    entry: Anime_LibraryCollectionEntry
    isSelected: boolean
    onToggle: () => void
    disabled?: boolean
}) {
    const { entry, isSelected, onToggle, disabled } = props

    return (
        <div
            className={cn(
                "col-span-1 aspect-[6/7] rounded-[--radius-md] overflow-hidden relative bg-[var(--background)] cursor-pointer transition-all select-none group",
                disabled && "pointer-events-none opacity-50",
                isSelected && "ring-2 ring-red-500",
            )}
            onClick={onToggle}
        >
            <Image
                src={entry.media?.coverImage?.large || entry.media?.bannerImage || ""}
                placeholder={imageShimmer(700, 475)}
                sizes="10rem"
                fill
                alt=""
                className={cn(
                    "object-center object-cover rounded-[--radius-md] transition-opacity",
                    isSelected ? "opacity-50" : "opacity-90 group-hover:opacity-100",
                )}
            />

            <p className="line-clamp-2 text-sm absolute m-2 bottom-0 font-semibold z-[10] text-white drop-shadow-lg">
                {entry.media?.title?.userPreferred || entry.media?.title?.romaji}
            </p>

            {entry.media?.isAdult && (
                <div className="absolute top-2 left-2 bg-red-600 text-white text-xs px-1.5 py-0.5 rounded font-semibold z-[10]">
                    18+
                </div>
            )}

            <div
                className={cn(
                    "absolute top-2 right-2 size-6 rounded-full flex items-center justify-center z-[10] transition-all",
                    isSelected
                        ? "bg-red-500 text-white"
                        : "bg-black/50 text-white/70 group-hover:bg-black/70",
                )}
            >
                {isSelected ? (
                    <RiCloseCircleFill className="size-4" />
                ) : (
                    <div className="size-3 border border-current rounded-full" />
                )}
            </div>

            <div className="z-[5] absolute bottom-0 w-full h-[80%] bg-gradient-to-t from-black/80 to-transparent" />
            {!isSelected && (
                <div className="z-[5] absolute top-0 w-full h-[80%] bg-gradient-to-b from-black/50 to-transparent opacity-100 group-hover:opacity-60 transition-opacity" />
            )}
        </div>
    )
}
