import { Badge } from "@/components/ui/badge"
import React from "react"

export function TorrentResolutionBadge({ resolution }: { resolution?: string }) {

    if (!resolution) {
        return (
            <Badge intent="gray">
                Unknown
            </Badge>
        )
    }

    return (
        <Badge intent={resolution?.includes("1080")
            ? "warning"
            : (resolution?.includes("2160") || resolution?.toLowerCase().includes("4k"))
                ? "success"
                : "gray"}
        >
            {resolution}
        </Badge>
    )
}

export function TorrentSeedersBadge({ seeders }: { seeders: string }) {

    return (
        <Badge
            intent={parseInt(seeders) > 20 ? parseInt(seeders) > 200 ? "primary" : "success" : "gray"}
            // leftIcon={<FcLineChart/>}
        >
            <span className="text-sm">{seeders}</span> seeders
        </Badge>
    )

}