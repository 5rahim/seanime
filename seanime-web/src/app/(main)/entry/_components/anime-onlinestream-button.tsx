import { Anime_Entry } from "@/api/generated/types"
import { useServerStatus } from "@/app/(main)/_hooks/use-server-status"
import { useAnimeEntryPageView } from "@/app/(main)/entry/_containers/anime-entry-page"
import { Button } from "@/components/ui/button"
import { cn } from "@/components/ui/core/styling"
import React from "react"
import { AiOutlineArrowLeft } from "react-icons/ai"
import { FiPlayCircle } from "react-icons/fi"

type AnimeOnlinestreamButtonProps = {
    children?: React.ReactNode
    entry: Anime_Entry | undefined
}

export function AnimeOnlinestreamButton(props: AnimeOnlinestreamButtonProps) {

    const {
        children,
        entry,
        ...rest
    } = props

    const status = useServerStatus()

    const { isLibraryView, isOnlineStreamingView, toggleOnlineStreamingView } = useAnimeEntryPageView()


    if (
        !entry ||
        entry.media?.status === "NOT_YET_RELEASED" ||
        !status?.settings?.library?.enableOnlinestream ||
        entry.media?.isAdult
    ) return null

    if (!isLibraryView && !isOnlineStreamingView) return null

    // if (!status?.settings?.library?.includeOnlineStreamingInLibrary) return (
    //     <>
    //         <SeaLink href={`/onlinestream?id=${entry?.mediaId}`}>
    //             <Button
    //                 intent="primary-subtle"
    //                 leftIcon={<FiPlayCircle className="text-xl" />}
    //             >
    //                 Stream online
    //             </Button>
    //         </SeaLink>
    //     </>
    // )

    return (
        <Button
            intent={isOnlineStreamingView ? "alert-subtle" : "white-subtle"}
            className={cn((status?.settings?.library?.includeOnlineStreamingInLibrary || isOnlineStreamingView) && "w-full")}
            size="md"
            leftIcon={isOnlineStreamingView ? <AiOutlineArrowLeft className="text-xl" /> : <FiPlayCircle className="text-2xl" />}
            onClick={() => toggleOnlineStreamingView()}
        >
            {isOnlineStreamingView ? "Close Online streaming" : "Stream online"}
        </Button>
    )
}
