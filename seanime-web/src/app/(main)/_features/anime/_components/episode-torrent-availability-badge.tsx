import type { Anime_EpisodeTorrentAvailability } from "@/api/generated/types"
import { Badge } from "@/components/ui/badge"
import { LuCircleCheck, LuCircleHelp, LuClock3, LuLoaderCircle } from "react-icons/lu"

export function EpisodeTorrentAvailabilityBadge({ status }: { status?: Anime_EpisodeTorrentAvailability }) {
    if (status === "available") {
        return <Badge size="sm" intent="success-solid" leftIcon={<LuCircleCheck />} title="A matching torrent was found">
            Torrent available
        </Badge>
    }
    if (status === "checking") {
        return <Badge
            size="sm"
            intent="primary-solid"
            leftIcon={<LuLoaderCircle className="animate-spin" />}
            title="Checking the selected torrent provider"
        >
            Checking torrents
        </Badge>
    }
    if (status === "waiting") {
        return <Badge size="sm" intent="warning-solid" leftIcon={<LuClock3 />} title="No matching torrent was found yet">
            Waiting for torrent
        </Badge>
    }
    if (status === "unknown") {
        return <Badge size="sm" intent="gray-solid" leftIcon={<LuCircleHelp />} title="The torrent provider could not be checked">
            Availability unknown
        </Badge>
    }
    return null
}
