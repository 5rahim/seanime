import { AL_AnimeDetailsById_Media, Anime_AnimeEntry } from "@/api/generated/types"
import { useHandlePlayMedia } from "@/app/(main)/entry/_lib/handle-play-media"
import { usePlayNextVideoOnMount } from "@/app/(main)/entry/_lib/handle-play-on-mount"
import { useMemo } from "react"

export function useHandleEpisodeSection(props: { entry: Anime_AnimeEntry, details: AL_AnimeDetailsById_Media | undefined }) {
    const { entry, details } = props
    const media = entry.media


    const { playMediaFile } = useHandlePlayMedia()

    usePlayNextVideoOnMount({
        onPlay: () => {
            if (entry.nextEpisode) {
                playMediaFile({ path: entry.nextEpisode.localFile?.path ?? "", mediaId: entry.mediaId })
                // playVideo({ path: entry.nextEpisode.localFile?.path ?? "" })
            }
        },
    })

    const mainEpisodes = useMemo(() => {
        return entry.episodes?.filter(ep => ep.type === "main") ?? []
    }, [entry.episodes])

    const specialEpisodes = useMemo(() => {
        return entry.episodes?.filter(ep => ep.type === "special") ?? []
    }, [entry.episodes])

    const ncEpisodes = useMemo(() => {
        return entry.episodes?.filter(ep => ep.type === "nc") ?? []
    }, [entry.episodes])

    const hasInvalidEpisodes = useMemo(() => {
        return entry.episodes?.some(ep => ep.isInvalid) ?? false
    }, [entry.episodes])

    const episodesToWatch = useMemo(() => {
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
