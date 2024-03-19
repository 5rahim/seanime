"use client"
import "@vidstack/react/player/styles/default/theme.css"
import "@vidstack/react/player/styles/default/layouts/video.css"
import { useOnlinestreamEpisodes, useOnlinestreamEpisodeSources } from "@/app/(main)/onlinestream/_lib/episodes"
import { MediaPlayer, MediaProvider, Track } from "@vidstack/react"
import { defaultLayoutIcons, DefaultVideoLayout } from "@vidstack/react/player/layouts/default"
import { useSearchParams } from "next/navigation"
import React from "react"

export default function Page() {

    const searchParams = useSearchParams()
    const mediaId = searchParams.get("id")

    const [episodeNumber, setEpisodeNumber] = React.useState(searchParams.get("episode") ? Number(searchParams.get("episode")) : 1)

    const { episodes } = useOnlinestreamEpisodes(mediaId, false)

    const { sources } = useOnlinestreamEpisodeSources(mediaId, episodeNumber, false)

    console.log(sources)

    return (
        <div>
            <MediaPlayer
                crossOrigin="anonymous"
                src={{
                    src: sources?.[0]?.sources?.[1].url || "",
                    type: "application/x-mpegurl",
                }}
            >
                <MediaProvider>
                    {sources?.[0]?.subtitles?.map((sub) => {
                        return <Track
                            {...{
                                id: sub.language,
                                label: sub.language,
                                kind: "subtitles",
                                src: sub.url,
                                language: sub.language,
                                default: sub.language
                                    ? sub.language?.toLowerCase() === "english" || sub.language?.toLowerCase() === "en-us"
                                    : sub.language?.toLowerCase() === "english" || sub.language?.toLowerCase() === "en-us",
                            }}
                        />
                    })}
                </MediaProvider>
                <DefaultVideoLayout icons={defaultLayoutIcons} />
            </MediaPlayer>
        </div>
    )

}
