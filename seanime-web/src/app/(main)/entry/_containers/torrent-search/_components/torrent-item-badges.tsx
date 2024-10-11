import { Badge } from "@/components/ui/badge"
import React from "react"
import { HiOutlineServerStack } from "react-icons/hi2"

export function TorrentResolutionBadge({ resolution }: { resolution?: string }) {

    if (!resolution) return null

    return (
        <Badge
            className="rounded-md"
            intent={resolution?.includes("1080")
            ? "warning"
            : (resolution?.includes("2160") || resolution?.toLowerCase().includes("4k"))
                ? "success"
                : "gray"}
        >
            {resolution}
        </Badge>
    )
}

export function TorrentSeedersBadge({ seeders }: { seeders: number }) {

    if (seeders === 0) return null

    return (
        <Badge
            className="rounded-md"
            intent={(seeders) > 4 ? (seeders) > 19 ? "primary" : "success" : "gray"}
        >
            <span className="text-sm">{seeders}</span> seeders
        </Badge>
    )

}


export function TorrentDebridInstantAvailabilityBadge() {

    return (
        <Badge
            className="rounded-md"
            intent="white-solid"
            leftIcon={<HiOutlineServerStack className="text-xl" />}
        >
            Cached
        </Badge>
    )

}
