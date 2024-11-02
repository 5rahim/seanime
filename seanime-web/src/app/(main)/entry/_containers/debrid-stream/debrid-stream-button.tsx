import { Anime_Entry } from "@/api/generated/types"
import { useServerStatus } from "@/app/(main)/_hooks/use-server-status"
import { useAnimeEntryPageView } from "@/app/(main)/entry/_containers/anime-entry-page"
import { Button } from "@/components/ui/button"
import React from "react"
import { AiOutlineArrowLeft } from "react-icons/ai"
import { HiOutlineServerStack } from "react-icons/hi2"

type DebridStreamButtonProps = {
    children?: React.ReactNode
    entry: Anime_Entry
}

export function DebridStreamButton(props: DebridStreamButtonProps) {

    const {
        children,
        entry,
        ...rest
    } = props

    const serverStatus = useServerStatus()

    const { isLibraryView, isDebridStreamingView, toggleDebridStreamingView } = useAnimeEntryPageView()

    if (
        !entry ||
        entry.media?.status === "NOT_YET_RELEASED" ||
        !serverStatus?.debridSettings?.enabled
    ) return null

    if (!isLibraryView && !isDebridStreamingView) return null

    return (
        <>
            <Button
                intent={isDebridStreamingView ? "alert-subtle" : "white-subtle"}
                className="w-full"
                size="md"
                leftIcon={isDebridStreamingView ? <AiOutlineArrowLeft className="text-xl" /> : <HiOutlineServerStack className="text-2xl" />}
                onClick={() => toggleDebridStreamingView()}
            >
                {isDebridStreamingView ? "Close Debrid streaming" : "Stream with Debrid"}
            </Button>
        </>
    )
}
