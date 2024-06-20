"use client"
import React from "react"

export default function Page() {

    const mediaContainer = {
        "filePath": "E:\\ANIME\\WIND BREAKER\\[SubsPlease] Wind Breaker - 01 (1080p) [5D5071F6].mkv",
        "hash": "c4862afed30d91ddaafe678d6a68f5b5a6427cf7",
        "streamType": "transcode",
        "streamUrl": "/api/v1/mediastream/transcode/master.m3u8",
        "mediaInfo": {
            "sha": "c4862afed30d91ddaafe678d6a68f5b5a6427cf7",
            "path": "E:\\ANIME\\WIND BREAKER\\[SubsPlease] Wind Breaker - 01 (1080p) [5D5071F6].mkv",
            "extension": "mkv",
            "mimeCodec": "video/x-matroska; codecs=\"avc1.640028, mp4a.40.2\"",
            "size": 1458683517,
            "duration": 1435.086,
            "container": "matroska,webm",
            "video": {
                "codec": "h264",
                "mimeCodec": "avc1.640028",
                "language": "und",
                "quality": "1080p",
                "width": 1920,
                "height": 1080,
                "bitrate": 8131546,
            },
            "videos": [
                {
                    "codec": "h264",
                    "mimeCodec": "avc1.640028",
                    "language": "und",
                    "quality": "1080p",
                    "width": 1920,
                    "height": 1080,
                    "bitrate": 8131546,
                },
            ],
            "audios": [
                {
                    "index": 0,
                    "title": null,
                    "language": "ja",
                    "codec": "aac",
                    "mimeCodec": "mp4a.40.2",
                    "isDefault": true,
                    "isForced": false,
                    "channels": 0,
                },
            ],
            "subtitles": [
                {
                    "index": 0,
                    "title": "English subs",
                    "language": "en",
                    "codec": "ass",
                    "extension": "ass",
                    "isDefault": true,
                    "isForced": false,
                    "link": "/0.ass",
                },
            ],
            "fonts": [
                "Roboto-Medium.ttf",
                "Roboto-MediumItalic.ttf",
                "arial.ttf",
                "arialbd.ttf",
                "comic.ttf",
                "comicbd.ttf",
                "times.ttf",
                "timesbd.ttf",
                "trebuc.ttf",
                "trebucbd.ttf",
                "verdana.ttf",
                "verdanab.ttf",
                "CONSOLA.TTF",
                "CONSOLAB.TTF",
            ],
            "chapters": [],
        },
    }

    const supported = (() => {
        try {
            if (typeof WebAssembly === "object"
                && typeof WebAssembly.instantiate === "function") {
                const module = new WebAssembly.Module(Uint8Array.of(0x0, 0x61, 0x73, 0x6d, 0x01, 0x00, 0x00, 0x00))
                if (module instanceof WebAssembly.Module)
                    return new WebAssembly.Instance(module) instanceof WebAssembly.Instance
            }
        }
        catch (e) {
        }
        return false
    })()

    console.log(supported ? "WebAssembly is supported" : "WebAssembly is not supported")


    const canPlay = (codec: string) => {
        // most chrome based browser (and safari I think) supports matroska but reports they do not.
        // for those browsers, only check the codecs and not the container.
        if (navigator.userAgent.search("Firefox") === -1)
            codec = codec.replace("video/x-matroska", "video/mp4")
        const videos = document.getElementsByTagName("video")
        const video = videos.item(0) ?? document.createElement("video")
        return !!video.canPlayType(codec)
    }

    React.useEffect(() => {
        console.log(canPlay(mediaContainer.mediaInfo.mimeCodec))
    }, [])

    return (
        <>
            Go away.
        </>
    )
}
