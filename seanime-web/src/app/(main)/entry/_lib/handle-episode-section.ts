import { Anime_Entry } from "@/api/generated/types"
import { useNakamaPlayVideo } from "@/api/hooks/nakama.hooks"
import { useHandlePlayMedia } from "@/app/(main)/entry/_lib/handle-play-media"
import { usePlayNextVideoOnMount } from "@/app/(main)/entry/_lib/handle-play-on-mount"
import React from "react"

export function useHandleEpisodeSection(props: { entry: Anime_Entry }) {
    const { entry } = props
    const media = entry.media

    const { playMediaFile } = useHandlePlayMedia()

    usePlayNextVideoOnMount({
        onPlay: () => {
            if (entry.nextEpisode) {
                playMediaFile({ path: entry.nextEpisode.localFile?.path ?? "", mediaId: entry.mediaId, episode: entry.nextEpisode })
            }
        },
    }, !!entry.nextEpisode)

    const mainEpisodes = React.useMemo(() => {
        return entry.episodes?.filter(ep => ep.type === "main") ?? []
    }, [entry.episodes])

    const specialEpisodes = React.useMemo(() => {
        return (entry.episodes?.filter(ep => ep.type === "special") ?? [])
            .sort((a, b) => a.displayTitle.localeCompare(b.displayTitle, undefined, { numeric: true }))
    }, [entry.episodes])

    const ncEpisodes = React.useMemo(() => {
        return (entry.episodes?.filter(ep => ep.type === "nc" && !!ep.localFile?.path) ?? []).sort((a,
            b,
        ) => a.localFile!.path!.localeCompare(b.localFile!.path!, undefined, { numeric: true }))
    }, [entry.episodes])

    const hasInvalidEpisodes = React.useMemo(() => {
        return entry.episodes?.some(ep => ep.isInvalid) ?? false
    }, [entry.episodes])

    const episodesToWatch = React.useMemo(() => {

        const ret = mainEpisodes.filter(ep => {
            if (!entry.nextEpisode) {
                return true
            } else {
                return ep.progressNumber > (entry.listData?.progress ?? 0)
            }
        })
        return (!!entry.listData?.progress && !entry.nextEpisode) ? ret.reverse() : ret
    }, [mainEpisodes, entry.nextEpisode, entry.listData?.progress])

    return {
        media,
        playMediaFile,
        mainEpisodes,
        specialEpisodes,
        ncEpisodes,
        hasInvalidEpisodes,
        episodesToWatch,
    }
}
