import { HibikeTorrent_AnimeTorrent, Torrent_Preview } from "@/api/generated/types"
import React from "react"
import { TbArrowsSort, TbSortAscending, TbSortDescending } from "react-icons/tb"

// Define sort types
export type SortField = "seeders" | "size" | "date" | "resolution" | null
export type SortDirection = "asc" | "desc" | null

// Helper to get sort icon for a field
export const getSortIcon = (sortField: SortField, field: SortField, sortDirection: SortDirection) => {
    if (sortField !== field) return <TbArrowsSort className="opacity-50 text-lg" />
    return sortDirection === "asc" ?
        <TbSortAscending className="text-brand-200 text-lg" /> :
        <TbSortDescending className="text-brand-200 text-lg" />
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

// Sort torrents function
export const sortTorrents = <T extends HibikeTorrent_AnimeTorrent>(
    torrents: T[],
    sortField: SortField,
    sortDirection: SortDirection,
): T[] => {
    if (!sortField || !sortDirection) return torrents

    return [...torrents].sort((a, b) => {
        if (!a || !b) return 0

        let valueA, valueB

        switch (sortField) {
            case "seeders":
                valueA = a.seeders || 0
                valueB = b.seeders || 0
                break
            case "size":
                valueA = a.size || 0
                valueB = b.size || 0
                break
            case "date":
                valueA = new Date(a.date).getTime()
                valueB = new Date(b.date).getTime()
                break
            case "resolution":
                // Convert resolution to numeric value for sorting
                valueA = a.resolution ? parseInt(a.resolution.replace(/[^\d]/g, "") || "0") : 0
                valueB = b.resolution ? parseInt(b.resolution.replace(/[^\d]/g, "") || "0") : 0
                break
            default:
                return 0
        }

        return sortDirection === "asc"
            ? valueA - valueB
            : valueB - valueA
    })
}

// Sort preview torrents function
export const sortPreviewTorrents = (
    previews: Torrent_Preview[],
    sortField: SortField,
    sortDirection: SortDirection,
): Torrent_Preview[] => {
    if (!sortField || !sortDirection) return previews

    return [...previews].sort((a, b) => {
        if (!a.torrent || !b.torrent) return 0

        let valueA, valueB

        switch (sortField) {
            case "seeders":
                valueA = a.torrent.seeders || 0
                valueB = b.torrent.seeders || 0
                break
            case "size":
                valueA = a.torrent.size || 0
                valueB = b.torrent.size || 0
                break
            case "date":
                valueA = new Date(a.torrent.date).getTime()
                valueB = new Date(b.torrent.date).getTime()
                break
            case "resolution":
                // Convert resolution to numeric value for sorting
                valueA = a.torrent.resolution ? parseInt(a.torrent.resolution.replace(/[^\d]/g, "") || "0") : 0
                valueB = b.torrent.resolution ? parseInt(b.torrent.resolution.replace(/[^\d]/g, "") || "0") : 0
                break
            default:
                return 0
        }

        return sortDirection === "asc"
            ? valueA - valueB
            : valueB - valueA
    })
}
