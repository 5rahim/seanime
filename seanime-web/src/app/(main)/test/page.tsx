"use client"

import { useCustomSourceListAnime } from "@/api/hooks/custom_source.hooks"
import { useListCustomSourceExtensions } from "@/api/hooks/extensions.hooks"
import { MediaCardLazyGrid } from "@/app/(main)/_features/media/_components/media-card-grid"
import { MediaEntryCard } from "@/app/(main)/_features/media/_components/media-entry-card"
import { AppLayoutStack } from "@/components/ui/app-layout"
import { upath } from "@/lib/helpers/upath"
import React from "react"

export default function TestPage() {

    const { data: customSources } = useListCustomSourceExtensions()

    const { data, mutate: listAnime, isPending } = useCustomSourceListAnime()

    React.useEffect(() => {
        listAnime({
            provider: "usermedia-test",
            page: 1,
            perPage: 100,
        })
    }, [])

    return <AppLayoutStack className="h-full w-full relative">

        <pre>
            {JSON.stringify(customSources, null, 2)}
        </pre>
        <MediaCardLazyGrid itemCount={1}>
            {data?.media?.map(media => {
                return (
                    <MediaEntryCard type="anime" media={media} />
                )
            })}
        </MediaCardLazyGrid>

        {/*<VideoCoreProvider>*/}
        {/*    <VideoCore*/}
        {/*        active={true}*/}
        {/*        src="https://stream.mux.com/fXNzVtmtWuyz00xnSrJg4OJH6PyNo6D02UzmgeKGkP5YQ/high.mp4"*/}
        {/*    />*/}
        {/*</VideoCoreProvider>*/}

    </AppLayoutStack>
}

