import { Badge } from "@/components/ui/badge"
import React from "react"

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
            // leftIcon={<FcLineChart/>}
        >
            <span className="text-sm">{seeders}</span> seeders
        </Badge>
    )

}
