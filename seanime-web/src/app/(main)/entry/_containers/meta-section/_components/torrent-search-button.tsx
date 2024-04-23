import { Anime_MediaEntry } from "@/api/generated/types"
import { torrentSearchDrawerIsOpenAtom } from "@/app/(main)/entry/_containers/torrent-search/torrent-search-drawer"
import { Button } from "@/components/ui/button"
import { useSetAtom } from "jotai/react"
import React, { useMemo } from "react"
import { BiDownload } from "react-icons/bi"
import { FiSearch } from "react-icons/fi"

export function TorrentSearchButton({ entry }: { entry: Anime_MediaEntry }) {

    const setter = useSetAtom(torrentSearchDrawerIsOpenAtom)
    const count = entry.downloadInfo?.episodesToDownload?.length
    const isMovie = useMemo(() => entry.media?.format === "MOVIE", [entry.media?.format])

    return (
        <div>
            {entry.downloadInfo?.hasInaccurateSchedule && <p className="text-orange-200 text-center mb-3">
                <span className="block">Could not retrieve accurate scheduling information for this show.</span>
                <span className="block text-[--muted]">Please check the schedule online for more information.</span>
            </p>}
            <Button
                className="w-full"
                intent={!entry.downloadInfo?.hasInaccurateSchedule ? (!!count ? "white" : "gray-subtle") : "warning-subtle"}
                size="lg"
                leftIcon={(!!count) ? <BiDownload /> : <FiSearch />}
                iconClass="text-2xl"
                onClick={() => setter(true)}
            >
                {(!entry.downloadInfo?.hasInaccurateSchedule && !!count) ? <>
                    {(!isMovie) && `Download ${entry.downloadInfo?.batchAll ? "batch /" : "next"} ${count > 1 ? `${count} episodes` : "episode"}`}
                    {(isMovie) && `Download movie`}
                </> : <>
                    Search torrents
                </>}
            </Button>
        </div>
    )
}
