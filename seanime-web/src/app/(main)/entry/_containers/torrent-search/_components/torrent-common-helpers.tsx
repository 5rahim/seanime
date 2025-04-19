import { Torrent_TorrentMetadata } from "@/api/generated/types"
import { Button } from "@/components/ui/button"
import { Checkbox } from "@/components/ui/checkbox"
import { Popover } from "@/components/ui/popover"
import React, { useState } from "react"
import { LiaMicrophoneSolid } from "react-icons/lia"
import { PiChatCircleTextDuotone, PiChatsTeardropDuotone } from "react-icons/pi"
import { TbArrowsSort, TbFilter, TbSortAscending, TbSortDescending } from "react-icons/tb"

// Define sort types
export type SortField = "seeders" | "size" | "date" | "resolution" | null
export type SortDirection = "asc" | "desc" | null

// Define filter types
export type TorrentFilters = {
    multiSubs: boolean,
    dualAudio: boolean,
    dubbed: boolean
}

// Helper to get sort icon for a field
export const getSortIcon = (sortField: SortField, field: SortField, sortDirection: SortDirection) => {
    if (sortField !== field) return <TbArrowsSort className="opacity-50 text-lg" />
    return sortDirection === "asc" ?
        <TbSortAscending className="text-brand-200 text-lg" /> :
        <TbSortDescending className="text-brand-200 text-lg" />
}

export const getFilterIcon = (active: boolean) => {
    return active ? <TbFilter className="text-brand-200 animate-bounce text-lg" /> : <TbFilter className="opacity-50 text-lg" />
}

// Sort handler function
export const handleSort = (
    field: SortField,
    sortField: SortField,
    sortDirection: SortDirection,
    setSortField: (field: SortField) => void,
    setSortDirection: (direction: SortDirection) => void,
) => {
    if (sortField === field) {
        if (sortDirection === "desc") {
            setSortDirection("asc")
        } else if (sortDirection === "asc") {
            setSortField(null)
            setSortDirection(null)
        } else {
            setSortDirection("desc")
        }
    } else {
        setSortField(field)
        setSortDirection("desc")
    }
}

// Helper functions for checking torrent properties
export const hasTorrentMultiSubs = (metadata: Torrent_TorrentMetadata | undefined): boolean => {
    if (!metadata) return false
    return !!metadata.metadata?.subtitles?.some(n => n.toLocaleLowerCase().includes("multi"))
}

export const hasTorrentDualAudio = (metadata: Torrent_TorrentMetadata | undefined): boolean => {
    if (!metadata) return false
    return !!metadata.metadata?.audio_term?.some(term =>
        term.toLowerCase().includes("dual") || term.toLowerCase().includes("multi"))
}

export const hasTorrentDubbed = (metadata: Torrent_TorrentMetadata | undefined): boolean => {
    if (!metadata) return false
    return !!metadata.metadata?.subtitles?.some(n => n.toLocaleLowerCase().includes("dub"))
}

// Generic interface for torrent-like objects
interface TorrentLike {
    seeders?: number
    size?: number
    date: string
    resolution?: string
    infoHash?: string
}

// Generic interface for preview-like objects
interface PreviewLike {
    torrent?: {
        seeders?: number
        size?: number
        date: string
        resolution?: string
        infoHash?: string
    }
}

// Generic sort function that works with both torrent types
export function sortItems<T extends TorrentLike | PreviewLike>(
    items: T[],
    sortField: SortField,
    sortDirection: SortDirection,
): T[] {
    if (!sortField || !sortDirection) return items

    return [...items].sort((a, b) => {
        let valueA: number, valueB: number

        // Handle both direct torrents and preview torrents
        const torrentA = "torrent" in a ? a.torrent : a as TorrentLike
        const torrentB = "torrent" in b ? b.torrent : b as TorrentLike

        if (!torrentA || !torrentB) return 0

        switch (sortField) {
            case "seeders":
                valueA = torrentA.seeders || 0
                valueB = torrentB.seeders || 0
                break
            case "size":
                valueA = torrentA.size || 0
                valueB = torrentB.size || 0
                break
            case "date":
                valueA = new Date(torrentA.date).getTime()
                valueB = new Date(torrentB.date).getTime()
                break
            case "resolution":
                // Convert resolution to numeric value for sorting
                valueA = torrentA.resolution ? parseInt(torrentA.resolution.replace(/[^\d]/g, "") || "0") : 0
                valueB = torrentB.resolution ? parseInt(torrentB.resolution.replace(/[^\d]/g, "") || "0") : 0
                break
            default:
                return 0
        }

        return sortDirection === "asc"
            ? valueA - valueB
            : valueB - valueA
    })
}

// Generic filter function that works with both torrent types
export function filterItems<T extends TorrentLike | PreviewLike>(
    items: T[],
    torrentMetadata: Record<string, Torrent_TorrentMetadata> | undefined,
    filters: TorrentFilters,
): T[] {
    if (!torrentMetadata || (!filters.multiSubs && !filters.dualAudio && !filters.dubbed)) {
        return items
    }

    return items.filter(item => {
        // Handle both direct torrents and preview torrents
        const torrent = "torrent" in item ? item.torrent : item as TorrentLike
        if (!torrent?.infoHash || !torrentMetadata[torrent.infoHash]) return true

        const metadata = torrentMetadata[torrent.infoHash]

        // Apply filters
        if (filters.multiSubs && !hasTorrentMultiSubs(metadata)) return false
        if (filters.dualAudio && !hasTorrentDualAudio(metadata)) return false
        if (filters.dubbed && !hasTorrentDubbed(metadata)) return false

        return true
    })
}

// Hook for managing sorting state
export function useTorrentSorting() {
    const [sortField, setSortField] = useState<SortField>("seeders")
    const [sortDirection, setSortDirection] = useState<SortDirection>("desc")

    const handleSortChange = (field: SortField) => {
        handleSort(field, sortField, sortDirection, setSortField, setSortDirection)
    }

    return {
        sortField,
        sortDirection,
        handleSortChange,
    }
}

// Hook for managing filtering state
export function useTorrentFiltering() {
    const [filters, setFilters] = useState<TorrentFilters>({
        multiSubs: false,
        dualAudio: false,
        dubbed: false,
    })

    const handleFilterChange = (filterName: keyof TorrentFilters, value: boolean | "indeterminate") => {
        if (typeof value === "boolean") {
            setFilters(prev => ({
                ...prev,
                [filterName]: value,
            }))
        }
    }

    const isAnyFilterActive = filters.multiSubs || filters.dualAudio || filters.dubbed

    return {
        filters,
        handleFilterChange,
        isAnyFilterActive,
    }
}

// UI Component for filter and sort controls
export const TorrentFilterSortControls: React.FC<{
    resultCount: number,
    sortField: SortField,
    sortDirection: SortDirection,
    filters: TorrentFilters,
    onSortChange: (field: SortField) => void,
    onFilterChange: (filterName: keyof TorrentFilters, value: boolean | "indeterminate") => void
}> = ({
    resultCount,
    sortField,
    sortDirection,
    filters,
    onSortChange,
    onFilterChange,
}) => {
    const isAnyFilterActive = filters.multiSubs || filters.dualAudio || filters.dubbed

    return (
        <div className="flex items-center justify-between gap-4">
            <p className="text-sm text-[--muted] flex-none">{resultCount} results</p>
            <div className="flex items-center gap-1 flex-wrap">
                <Popover
                    trigger={<Button
                        size="xs"
                        intent="gray-basic"
                        leftIcon={<>
                            {getFilterIcon(isAnyFilterActive)}
                        </>}
                    >
                        Filters
                    </Button>}
                >
                    <p className="text-sm text-[--muted] flex-none pb-2">
                        Filters may miss some results
                    </p>
                    <div className="space-y-1">
                        <Checkbox
                            label={<div className="flex items-center gap-1">
                                <PiChatCircleTextDuotone className="text-lg text-[--orange]" /> Multi Subs
                            </div>}
                            value={filters.multiSubs}
                            onValueChange={(value) => onFilterChange("multiSubs", value)}
                        />
                        <Checkbox
                            label={<div className="flex items-center gap-1">
                                <PiChatsTeardropDuotone className="text-lg text-[--rose]" /> Dual Audio
                            </div>}
                            value={filters.dualAudio}
                            onValueChange={(value) => onFilterChange("dualAudio", value)}
                        />
                        <Checkbox
                            label={<div className="flex items-center gap-1">
                                <LiaMicrophoneSolid className="text-lg text-[--red]" /> Dubbed
                            </div>}
                            value={filters.dubbed}
                            onValueChange={(value) => onFilterChange("dubbed", value)}
                        />
                    </div>
                </Popover>
                <Button
                    size="xs"
                    intent="gray-basic"
                    leftIcon={<>
                        {getSortIcon(sortField, "seeders", sortDirection)}
                    </>}
                    onClick={() => onSortChange("seeders")}
                >
                    Seeders
                </Button>
                <Button
                    size="xs"
                    intent="gray-basic"
                    leftIcon={<>
                        {getSortIcon(sortField, "size", sortDirection)}
                    </>}
                    onClick={() => onSortChange("size")}
                >
                    Size
                </Button>
                <Button
                    size="xs"
                    intent="gray-basic"
                    leftIcon={<>
                        {getSortIcon(sortField, "date", sortDirection)}
                    </>}
                    onClick={() => onSortChange("date")}
                >
                    Date
                </Button>
                <Button
                    size="xs"
                    intent="gray-basic"
                    leftIcon={<>
                        {getSortIcon(sortField, "resolution", sortDirection)}
                    </>}
                    onClick={() => onSortChange("resolution")}
                >
                    Resolution
                </Button>
            </div>
        </div>
    )
}
