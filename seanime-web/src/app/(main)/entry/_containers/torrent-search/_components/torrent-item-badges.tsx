import { Torrent_TorrentMetadata } from "@/api/generated/types"
import { Badge } from "@/components/ui/badge"
import { Tooltip } from "@/components/ui/tooltip"
import { startCase } from "lodash"
import React from "react"
import { LiaMicrophoneSolid } from "react-icons/lia"
import { LuGauge } from "react-icons/lu"
import { PiChatCircleTextDuotone, PiChatsTeardropDuotone, PiChatTeardropDuotone } from "react-icons/pi"

export function TorrentResolutionBadge({ resolution }: { resolution?: string }) {

    if (!resolution) return null

    return (
        <Badge
            data-torrent-item-resolution-badge
            className="rounded-[--radius-md] border-transparent bg-transparent px-0"
            intent={resolution?.includes("1080")
                ? "warning"
                : (resolution?.includes("2160") || resolution?.toLowerCase().includes("4k"))
                    ? "success"
                    : (resolution?.includes("720")
                        ? "blue"
                        : "gray")}
        >
            {resolution}
        </Badge>
    )
}

export function TorrentSeedersBadge({ seeders }: { seeders: number }) {

    if (seeders === 0) return null

    return (
        <Badge
            data-torrent-item-seeders-badge
            className="rounded-[--radius-md] border-transparent bg-transparent px-0"
            intent={(seeders) > 4 ? (seeders) > 19 ? "primary" : "success" : "gray"}
        >
            <span className="text-sm">{seeders}</span> seeder{seeders > 1 ? "s" : ""}
        </Badge>
    )

}


export function TorrentParsedMetadata({ metadata }: { metadata: Torrent_TorrentMetadata | undefined }) {

    if (!metadata) return null

    const hasDubs = metadata.metadata?.subtitles?.some(n => n.toLocaleLowerCase().includes("dub"))
    // const hasSubs = metadata.metadata?.subtitles?.some(n => n.toLocaleLowerCase().includes("sub"))
    const hasMultiSubs = metadata.metadata?.subtitles?.some(n => n.toLocaleLowerCase().includes("multi"))

    const languages = !!metadata.metadata?.language?.length ? [...new Set(metadata.metadata?.language)] : []

    return (
        <div className="flex flex-row gap-1 flex-wrap justify-end w-full lg:absolute bottom-0 right-0">
            {!!languages?.length && languages.length == 2 ? languages.slice(0, 2)?.map(term => (
                <Badge
                    key={term}
                    className="rounded-md bg-transparent border-transparent px-1"
                >
                    <PiChatTeardropDuotone className="text-lg text-[--blue]" /> {term}
                </Badge>
            )) : null}
            {!!languages?.length && languages.length > 2 ? <Tooltip
                trigger={<Badge
                    className="rounded-md bg-transparent border-transparent px-1"
                >
                    <PiChatTeardropDuotone className="text-lg text-[--blue]" /> Multiple languages
                </Badge>}
            >
                <span>
                    {languages.join(", ")}
                </span>
            </Tooltip> : null}
            {metadata.metadata?.video_term?.map(term => (
                <Badge
                    key={term}
                    className="rounded-md border-transparent bg-[--subtle] !text-[--muted] px-1"
                >
                    {term}
                </Badge>
            ))}
            {metadata.metadata?.audio_term?.filter(term => term.toLowerCase().includes("dual") || term.toLowerCase().includes("multi")).map(term => (
                <Badge
                    key={term}
                    className="rounded-md border-transparent bg-[--subtle] px-1"
                >
                    {/* <LuAudioWaveform className="text-lg text-[--blue]" /> {term} */}
                    <PiChatsTeardropDuotone className="text-lg text-[--rose]" /> {startCase(term)}
                </Badge>
            ))}
            {hasDubs && (
                <Badge
                    className="rounded-md border-transparent bg-indigo-300 px-1"
                >
                    <LiaMicrophoneSolid className="text-lg text-[--red]" /> Dubbed
                </Badge>
            )}
            {hasMultiSubs && (
                <Badge
                    className="rounded-md border-transparent bg-indigo-300 px-1"
                >
                    <PiChatCircleTextDuotone className="text-lg text-[--orange]" /> Multi Subs
                </Badge>
            )}
        </div>
    )
}


export function TorrentDebridInstantAvailabilityBadge() {

    return (
        <Tooltip
            trigger={<Badge
                data-torrent-item-debrid-instant-availability-badge
                className="rounded-[--radius-md] bg-transparent border-transparent dark:text-[--white] animate-pulse"
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
