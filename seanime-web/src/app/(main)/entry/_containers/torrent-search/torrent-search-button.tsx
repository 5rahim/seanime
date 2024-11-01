import { Anime_Entry } from "@/api/generated/types"
import { __torrentSearch_drawerIsOpenAtom } from "@/app/(main)/entry/_containers/torrent-search/torrent-search-drawer"
import { Button } from "@/components/ui/button"
import { useSetAtom } from "jotai/react"
import React, { useMemo } from "react"
import { BiDownload } from "react-icons/bi"
import { FiSearch } from "react-icons/fi"

export function TorrentSearchButton({ entry }: { entry: Anime_Entry }) {

    const setter = useSetAtom(__torrentSearch_drawerIsOpenAtom)
    const count = entry.downloadInfo?.episodesToDownload?.length
    const isMovie = useMemo(() => entry.media?.format === "MOVIE", [entry.media?.format])

    return (
        <div className="w-full">
            <Button
                className="w-full"
                intent={!entry.downloadInfo?.hasInaccurateSchedule ? (!!count ? "white" : "gray-subtle") : "white-subtle"}
                size="md"
                leftIcon={(!!count) ? <BiDownload /> : <FiSearch />}
                iconClass="text-2xl"
                onClick={() => setter("download")}
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
