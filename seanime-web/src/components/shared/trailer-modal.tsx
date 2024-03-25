import { useMediaDetails } from "@/app/(main)/entry/_lib/media-entry"
import { LuffyError } from "@/components/shared/luffy-error"
import { cn } from "@/components/ui/core/styling"
import { Drawer } from "@/components/ui/drawer"
import { LoadingSpinner } from "@/components/ui/loading-spinner"
import React from "react"

type PlaylistsModalProps = {
    trigger?: React.ReactElement
    mediaId: number
    isOpen?: boolean
    setIsOpen?: (v: boolean) => void
}

export function TrailerModal(props: PlaylistsModalProps) {

    const {
        trigger,
        mediaId,
        isOpen,
        setIsOpen,
        ...rest
    } = props

    return (
        <>
            <Drawer
                open={isOpen}
                onOpenChange={v => setIsOpen?.(v)}
                trigger={trigger}
                size="full"
                side="bottom"
                contentClass="flex items-center justify-center w-full h-full"
            >
                <div
                    className="!mt-0 bg-[url(/pattern-2.svg)] z-[-1] w-full h-[5rem] absolute opacity-30 top-0 left-0 bg-no-repeat bg-right bg-cover"
                >
                    <div
                        className="w-full absolute top-0 h-full bg-gradient-to-t from-[--background] to-transparent z-[-2]"
                    />
                </div>

                <Content mediaId={mediaId} />
            </Drawer>
        </>
    )
}

type ContentProps = {
    mediaId: number
}

export function Content(props: ContentProps) {

    const {
        mediaId,
        ...rest
    } = props

    const { mediaDetails, mediaDetailsLoading } = useMediaDetails(mediaId)
    const [loaded, setLoaded] = React.useState(true)
    const [muted, setMuted] = React.useState(true)

    if (mediaDetailsLoading) return <LoadingSpinner className="" />

    if (!mediaDetails?.trailer) return <LuffyError title="No trailer found" />

    return (
        <>
            {!loaded && <LoadingSpinner className="" />}
            <div
                className={cn(
                    "relative aspect-video h-[85dvh] overflow-hidden rounded-xl",
                    !loaded && "hidden",
                )}
            >
                <iframe
                    src={`https://www.youtube.com/embed/${mediaDetails?.trailer?.id}`}
                    title="YouTube Video"
                    className="w-full h-full"
                    allowFullScreen
                    loading="lazy" // Lazy load the iframe
                />
                {/*<video*/}
                {/*    src={`https://yewtu.be/latest_version?id=${mediaDetails?.trailer?.id}&itag=18`}*/}
                {/*    className={cn(*/}
                {/*        "w-full h-full absolute left-0",*/}
                {/*    )}*/}
                {/*    playsInline*/}
                {/*    preload="none"*/}
                {/*    loop*/}
                {/*    autoPlay*/}
                {/*    muted={muted}*/}
                {/*    onLoadedData={() => setLoaded(true)}*/}
                {/*/>*/}
                {/*{<IconButton*/}
                {/*    intent="white-basic"*/}
                {/*    className="absolute bottom-4 left-4"*/}
                {/*    icon={muted ? <FaVolumeMute /> : <FaVolumeHigh />}*/}
                {/*    onClick={() => setMuted(p => !p)}*/}
                {/*/>}*/}
            </div>
        </>
    )
}

