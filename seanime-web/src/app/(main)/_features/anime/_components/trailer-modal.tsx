import { ElectronYoutubeEmbed } from "@/app/(main)/_electron/electron-embed"
import { LuffyError } from "@/components/shared/luffy-error"
import { cn } from "@/components/ui/core/styling"
import { LoadingSpinner } from "@/components/ui/loading-spinner"
import { Modal } from "@/components/ui/modal"
import { __isElectronDesktop__ } from "@/types/constants"
import React from "react"

type PlaylistsModalProps = {
    trigger?: React.ReactElement
    trailerId?: string | null
    isOpen?: boolean
    setIsOpen?: (v: boolean) => void
}

export function TrailerModal(props: PlaylistsModalProps) {

    const {
        trigger,
        trailerId,
        isOpen,
        setIsOpen,
        ...rest
    } = props

    return (
        <>
            <Modal
                open={isOpen}
                onOpenChange={v => setIsOpen?.(v)}
                trigger={trigger}
                contentClass="flex max-w-5xl items-center justify-center"
            >
                <div
                    className="!mt-0 bg-[url(/pattern-2.svg)] z-[-1] w-full h-[5rem] absolute opacity-30 top-0 left-0 bg-no-repeat bg-right bg-cover"
                >
                    <div
                        className="w-full absolute top-0 h-full bg-gradient-to-t from-[--background] to-transparent z-[-2]"
                    />
                </div>

                <Content trailerId={trailerId} />
            </Modal>
        </>
    )
}

type ContentProps = {
    trailerId?: string | null
}

export function Content(props: ContentProps) {

    const {
        trailerId,
        ...rest
    } = props

    const [loaded, setLoaded] = React.useState(true)

    if (!trailerId) return <LuffyError title="No trailer found" />

    return (
        <>
            {!loaded && <LoadingSpinner className="" />}
            <div
                className={cn(
                    "relative aspect-video w-full flex items-center overflow-hidden rounded-xl",
                    !loaded && "hidden",
                )}
            >
                {__isElectronDesktop__ && <ElectronYoutubeEmbed trailerId={trailerId} />}
                {!__isElectronDesktop__ && <iframe
                    src={`https://www.youtube.com/embed/${trailerId}`}
                    title="YouTube Video"
                    className="w-full aspect-video rounded-xl"
                    allowFullScreen
                    loading="lazy" // Lazy load the iframe
                    referrerPolicy="strict-origin-when-cross-origin"
                />}
                {/*<video*/}
                {/*    src={`https://yewtu.be/latest_version?id=${animeDetails?.trailer?.id}&itag=18`}*/}
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

