import { Badge } from "@/components/ui/badge"
import { Tooltip } from "@/components/ui/tooltip"
import React from "react"
import { LuGauge } from "react-icons/lu"

export function TorrentResolutionBadge({ resolution }: { resolution?: string }) {

    if (!resolution) return null

    return (
        <Badge
            className="rounded-[--radius-md] border-transparent bg-transparent px-0"
            intent={resolution?.includes("1080")
            ? "warning"
            : (resolution?.includes("2160") || resolution?.toLowerCase().includes("4k"))
                ? "success"
                    : "blue"}
        >
            {resolution}
        </Badge>
    )
}

export function TorrentSeedersBadge({ seeders }: { seeders: number }) {

    if (seeders === 0) return null

    return (
        <Badge
            className="rounded-[--radius-md] border-transparent bg-transparent px-0"
            intent={(seeders) > 4 ? (seeders) > 19 ? "primary" : "success" : "gray"}
        >
            <span className="text-sm">{seeders}</span> seeder{seeders > 1 ? "s" : ""}
        </Badge>
    )

}


export function TorrentDebridInstantAvailabilityBadge() {

    return (
        <Tooltip
            trigger={<Badge
                className="rounded-[--radius-md] bg-transparent dark:text-[--green]"
                intent="white"
                leftIcon={<LuGauge className="text-lg" />}
        >
            Cached
            </Badge>}
        >
            Instantly available on Debrid service
        </Tooltip>
    )

}
