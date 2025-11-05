import { Torrent_TorrentMetadata } from "@/api/generated/types"
import { Button } from "@/components/ui/button"
import { Checkbox } from "@/components/ui/checkbox"
import { Popover } from "@/components/ui/popover"
import { Separator } from "@/components/ui/separator"
import React, { useState } from "react"
import { LiaMicrophoneSolid } from "react-icons/lia"
import { PiChatCircleDotsDuotone } from "react-icons/pi"
import { TbArrowsSort, TbFilter, TbSortAscending, TbSortDescending } from "react-icons/tb"

// Define sort types
export type SortField = "seeders" | "size" | "date" | "resolution" | null
export type SortDirection = "asc" | "desc" | null

// Define filter types
export type TorrentFilters = {
    multiSubs: boolean,
    dubbed: boolean,
    videoAvc: boolean,   // H.264/AVC
    videoHevc: boolean,  // H.265/HEVC
    videoAv1: boolean,   // AV1
    audioAac: boolean,
    audioAc3: boolean,
    audioDts: boolean,
    audioOpus: boolean,
    audioEac3: boolean,
    audioFlac: boolean,
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

export const hasVideoTerm = (term: string, metadata: Torrent_TorrentMetadata | undefined): boolean => {
    if (!metadata) return false
    const terms = metadata.metadata?.video_term?.map(n => n.toLocaleLowerCase())
    switch (term.toLocaleLowerCase()) {
        case "avc":
            return terms?.some(n => /264|avc|h264/mi.test(n)) || false
        case "hevc":
            return terms?.some(n => /265|hevc|h265/mi.test(n)) || false
        case "av1":
            return terms?.some(n => /av1|vp8/mi.test(n)) || false
        default:
            return false
    }
}

export const hasAudioTerm = (term: string, metadata: Torrent_TorrentMetadata | undefined): boolean => {
    if (!metadata) return false
    const terms = metadata.metadata?.audio_term?.map(n => n.toLocaleLowerCase())
    switch (term.toLocaleLowerCase()) {
        case "aac":
            return terms?.some(n => /aac|aac_latm/mi.test(n)) || false
        case "ac3":
            return terms?.some(n => /ac3|ac-3/mi.test(n)) || false
        case "dts":
            return terms?.some(n => /dts|dca/mi.test(n)) || false
        case "opus":
            return terms?.some(n => /opus|vorbis/mi.test(n)) || false
        case "eac3":
            return terms?.some(n => /(eac3|e-ac3|e-ac-3)/mi.test(n)) || false
        case "flac":
            return terms?.some(n => /flac|alac/mi.test(n)) || false
        default:
            return false
    }
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

function anyFilterActive(filters: TorrentFilters) {
    return Object.values(filters).some(value => value === true)
}

// Generic filter function that works with both torrent types
export function filterItems<T extends TorrentLike | PreviewLike>(
    items: T[],
    torrentMetadata: Record<string, Torrent_TorrentMetadata> | undefined,
    filters: TorrentFilters,
): T[] {
    if (!torrentMetadata || !anyFilterActive(filters)) {
        return items
    }

    return items.filter(item => {
        // Handle both direct torrents and preview torrents
        const torrent = "torrent" in item ? item.torrent : item as TorrentLike
        if (!torrent?.infoHash || !torrentMetadata[torrent.infoHash]) return true

        const metadata = torrentMetadata[torrent.infoHash]

        // Apply filters
        if (filters.multiSubs && !hasTorrentMultiSubs(metadata)) return false
        if (filters.videoHevc && !hasVideoTerm("hevc", metadata)) return false
        if (filters.videoAvc && !hasVideoTerm("avc", metadata)) return false
        if (filters.videoAv1 && !hasVideoTerm("av1", metadata)) return false
        if (filters.audioAc3 && !hasAudioTerm("ac3", metadata)) return false
        if (filters.audioEac3 && !hasAudioTerm("eac3", metadata)) return false
        if (filters.audioAac && !hasAudioTerm("aac", metadata)) return false
        if (filters.audioDts && !hasAudioTerm("dts", metadata)) return false
        if (filters.audioFlac && !hasAudioTerm("flac", metadata)) return false
        if (filters.audioOpus && !hasAudioTerm("opus", metadata)) return false
        if (filters.dubbed && !hasTorrentDubbed(metadata) && !hasTorrentDualAudio(metadata)) return false

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
        dubbed: false,
        videoAvc: false,
        videoHevc: false,
        videoAv1: false,
        audioOpus: false,
        audioFlac: false,
        audioDts: false,
        audioAac: false,
        audioEac3: false,
        audioAc3: false,
    })

    const handleFilterChange = (filterName: keyof TorrentFilters, value: boolean | "indeterminate") => {
        if (typeof value === "boolean") {
            setFilters(prev => ({
                ...prev,
                [filterName]: value,
            }))
        }
    }

    const isAnyFilterActive = anyFilterActive(filters)

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
    const isAnyFilterActive = anyFilterActive(filters)

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
                    <p className="text-xs text-[--muted] flex-none pb-2">
                        Filters are based on torrent names and can miss some results.
                    </p>
                    <div className="space-y-1">
                        <Checkbox
                            label={<div className="flex items-center gap-1">
                                <PiChatCircleDotsDuotone className="text-lg text-[--blue]" /> Multi Subs
                            </div>}
                            value={filters.multiSubs}
                            onValueChange={(value) => onFilterChange("multiSubs", value)}
                        />
                        {/*<Checkbox*/}
                        {/*    label={<div className="flex items-center gap-1">*/}
                        {/*        <LiaMicrophoneSolid className="text-lg text-[--rose]" /> Dual Audio*/}
                        {/*    </div>}*/}
                        {/*    value={filters.dualAudio}*/}
                        {/*    onValueChange={(value) => onFilterChange("dualAudio", value)}*/}
                        {/*/>*/}

                        <Checkbox
                            label={<div className="flex items-center gap-1">
                                <LiaMicrophoneSolid className="text-lg text-[--red]" /> Dubbed
                            </div>}
                            value={filters.dubbed}
                            onValueChange={(value) => onFilterChange("dubbed", value)}
                        />
                        <Separator className="!my-2" />
                        <div className="grid grid-cols-2 gap-2">
                            <Checkbox
                                label={<div className="flex items-center gap-1">
                                    HEVC/H.265
                                </div>}
                                value={filters.videoHevc}
                                onValueChange={(value) => onFilterChange("videoHevc", value)}
                                size="sm"
                            />
                            <Checkbox
                                label={<div className="flex items-center gap-1">
                                    AV1
                                </div>}
                                value={filters.videoAv1}
                                onValueChange={(value) => onFilterChange("videoAv1", value)}
                                size="sm"
                            />
                            <Checkbox
                                label={<div className="flex items-center gap-1">
                                    AAC
                                </div>}
                                value={filters.audioAac}
                                onValueChange={(value) => onFilterChange("audioAac", value)}
                                size="sm"
                            />
                            <Checkbox
                                label={<div className="flex items-center gap-1">
                                    AC3/AC-3
                                </div>}
                                value={filters.audioAc3}
                                onValueChange={(value) => onFilterChange("audioAc3", value)}
                                size="sm"
                            />
                            <Checkbox
                                label={<div className="flex items-center gap-1">
                                    DTS/DCA
                                </div>}
                                value={filters.audioDts}
                                onValueChange={(value) => onFilterChange("audioDts", value)}
                                size="sm"
                            />
                            <Checkbox
                                label={<div className="flex items-center gap-1">
                                    EAC3
                                </div>}
                                value={filters.audioEac3}
                                onValueChange={(value) => onFilterChange("audioEac3", value)}
                                size="sm"
                            />
                            <Checkbox
                                label={<div className="flex items-center gap-1">
                                    Opus/Vorbis
                                </div>}
                                value={filters.audioOpus}
                                onValueChange={(value) => onFilterChange("audioOpus", value)}
                                size="sm"
                            />
                            <Checkbox
                                label={<div className="flex items-center gap-1">
                                    FLAC/ALAC
                                </div>}
                                value={filters.audioFlac}
                                onValueChange={(value) => onFilterChange("audioFlac", value)}
                                size="sm"
                            />
                        </div>
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
