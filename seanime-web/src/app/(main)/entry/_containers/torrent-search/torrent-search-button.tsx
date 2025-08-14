import { Anime_Entry } from "@/api/generated/types"
import { AnimeMetaActionButton } from "@/app/(main)/entry/_components/meta-section"
import { __torrentSearch_selectionAtom } from "@/app/(main)/entry/_containers/torrent-search/torrent-search-drawer"
import { useSetAtom } from "jotai/react"
import React, { useMemo } from "react"
import { BiDownload } from "react-icons/bi"
import { FiSearch } from "react-icons/fi"

export function TorrentSearchButton({ entry }: { entry: Anime_Entry }) {

    const setter = useSetAtom(__torrentSearch_selectionAtom)
    const count = entry.downloadInfo?.episodesToDownload?.length
    const isMovie = useMemo(() => entry.media?.format === "MOVIE", [entry.media?.format])

    return (
        <div className="contents" data-torrent-search-button-container>
            <AnimeMetaActionButton
                intent={!entry.downloadInfo?.hasInaccurateSchedule ? (!!count ? "white" : "gray-subtle") : "white-subtle"}
                size="md"
                leftIcon={(!!count) ? <BiDownload /> : <FiSearch />}
                iconClass="text-2xl"
                onClick={() => setter("download")}
                data-torrent-search-button
            >
                {(!entry.downloadInfo?.hasInaccurateSchedule && !!count) ? <>
                    {(!isMovie) && `Download ${entry.downloadInfo?.batchAll ? "batch /" : "next"} ${count > 1 ? `${count} episodes` : "episode"}`}
                    {(isMovie) && `Download movie`}
                </> : <>
                    Search torrents
                </>}
            </AnimeMetaActionButton>
        </div>
    )
}
