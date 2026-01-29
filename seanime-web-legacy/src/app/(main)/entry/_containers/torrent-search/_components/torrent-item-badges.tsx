import { Habari_Metadata } from "@/api/generated/types"
import { Badge } from "@/components/ui/badge"
import { cn } from "@/components/ui/core/styling"
import { Tooltip } from "@/components/ui/tooltip"
import { startCase } from "lodash"
import React from "react"
import { LiaMicrophoneSolid } from "react-icons/lia"
import { LuGauge } from "react-icons/lu"
import {
    PiBatteryFullDuotone,
    PiBatteryHighDuotone,
    PiBatteryLowDuotone,
    PiBatteryMediumDuotone,
    PiChatCircleDotsDuotone,
    PiChatTeardropDuotone,
    PiChatTextDuotone,
    PiSmileyNervousDuotone,
} from "react-icons/pi"

export function TorrentResolutionBadge({ resolution }: { resolution?: string }) {

    if (!resolution) return null

    return (
        <Badge
            data-torrent-item-resolution-badge
            className="rounded-[--radius-md] border-transparent bg-gray-900/50 px-1 text-md"
            intent={resolution?.includes("1080")
                ? "indigo"
                : (resolution?.includes("2160") || resolution?.toLowerCase().includes("4k"))
                    ? "blue"
                    : (resolution?.includes("720")
                        ? "success"
                        : "gray")}
        >
            {resolution}
        </Badge>
    )
}

export function TorrentSeedersBadge({ seeders }: { seeders: number }) {

    let Icon = seeders >= 50 ? PiBatteryFullDuotone : seeders >= 20 ? PiBatteryHighDuotone :
        seeders >= 10 ? PiBatteryMediumDuotone :
            seeders >= 5 ? PiBatteryMediumDuotone :
                seeders > 0 ? PiBatteryLowDuotone :
                    PiSmileyNervousDuotone

    if (seeders === -1) return null

    return (
        <Badge
            data-torrent-item-seeders-badge
            className="rounded-[--radius-md] border-transparent bg-transparent px-0 gap-1 font-normal opacity-80"
            // intent={(seeders) > 4 ? (seeders) > 19 ? "primary" : "success" : "gray"}
            intent={"gray"}
            leftIcon={<Icon
                className={cn(
                    "text-xl mr-0.5",
                    seeders >= 50 ? "text-[--indigo]" : seeders >= 10 ? "text-[--green]" : seeders >= 5 ? "text-orange-300" : "text-[--red]",
                )}
            />}
        >
            <span
                className={cn("text-[.9rem] font-normal",
                    seeders >= 50 ? "text-[--indigo]" : seeders >= 10 ? "text-[--green]" : seeders >= 5 ? "text-orange-300" : "text-[--red]",
                )}
            >{seeders || "No"}</span><span className="text-[--muted] text-[.9rem]">seeder{seeders != 1
            ? "s"
            : ""}</span>
        </Badge>
    )

}


export function TorrentParsedMetadata({ metadata }: { metadata: Habari_Metadata | undefined }) {

    if (!metadata) return null

    const hasDubs = metadata?.subtitles?.some(n => n.toLocaleLowerCase().includes("dub"))
    // const hasSubs = metadata?.subtitles?.some(n => n.toLocaleLowerCase().includes("sub"))
    const hasMultiSubs = metadata?.subtitles?.some(n => n.toLocaleLowerCase().includes("multi"))

    const languages = !!metadata?.language?.length ? [...new Set(metadata?.language)] : []

    const filterHEVC = (n: string) => {
        return !(n.toLocaleLowerCase().includes("265") && metadata?.video_term?.map(n => n.toLocaleLowerCase()).includes("hevc"))
    }

    return (
        <div className="flex flex-row gap-1 flex-wrap justify-end w-full lg:absolute top-0 right-0">
            {!!languages?.length && languages.length == 2 ? languages.slice(0, 2)?.map(term => (
                <Badge
                    key={term}
                    className="rounded-md bg-transparent border-transparent px-1"
                >
                    <PiChatTeardropDuotone className="text-lg text-[--blue]" /> {term}
                </Badge>
            )) : null}
            {metadata?.video_term?.filter(filterHEVC).map(term => (
                <Badge
                    key={term}
                    className="rounded-md border-transparent bg-transparent text-[.8rem] text-[--foreground] px-1"
                >
                    {term}
                </Badge>
            ))}
            {metadata?.audio_term?.filter(term => !term.toLowerCase().includes("dual") && !term.toLowerCase().includes("multi"))
                .map(term => (
                    <Badge
                        key={term}
                        className="rounded-md border-transparent bg-transparent text-[.8rem] text-[--foreground] px-1 opacity-60"
                    >
                        {term}
                    </Badge>
                ))}
            {!!languages?.length && languages.length > 2 ? <Tooltip
                trigger={<Badge
                    className="rounded-md bg-transparent border-transparent px-1"
                >
                    <PiChatTextDuotone className="text-lg text-[--blue]" /> Languages
                </Badge>}
            >
                <span>
                    {languages.join(", ")}
                </span>
            </Tooltip> : null}
            {metadata?.audio_term?.filter(term => term.toLowerCase().includes("dual") || term.toLowerCase().includes("multi")).map(term => (
                <Badge
                    key={term}
                    className="rounded-md border-transparent bg-[--subtle] text-[.8rem] px-1"
                >
                    {/* <LuAudioWaveform className="text-lg text-[--blue]" /> {term} */}
                    <LiaMicrophoneSolid className="text-lg text-[--rose]" /> {term.toLowerCase().includes("dual")
                    ? "Original + Dub"
                    : startCase(term)}
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
                    <PiChatCircleDotsDuotone className="text-lg text-[--blue]" /> Multi Subs
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
                className="rounded-[--radius-md] bg-transparent border-transparent dark:text-[--indigo] animate-pulse p-0"
                intent="white"
            >
                <LuGauge className="text-xl" />
            </Badge>}
        >
            Instantly available on Debrid service
        </Tooltip>
    )

}
