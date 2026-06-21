import React from "react"
import { IconButton } from "@/components/ui/button"
import { FiMinimize2 } from "react-icons/fi"
import { BiExpand, BiX } from "react-icons/bi"
import { cn } from "@/components/ui/core/styling"

export interface MpvCoreFloatingButtonsProps {
    part: "video" | "loading"
    fullscreen: boolean
    miniPlayer: boolean
    onEnterMiniPlayer: () => void
    onExpand: () => void
    onTerminate: () => void
    onExitFullscreen: () => void
}

export function MpvCoreFloatingButtons(props: MpvCoreFloatingButtonsProps) {
    const {
        part,
        fullscreen,
        miniPlayer,
        onEnterMiniPlayer,
        onExpand,
        onTerminate,
        onExitFullscreen,
    } = props

    if (fullscreen && part === "video") return null

    const content = fullscreen ? (
        <IconButton
            data-vc-element="floating-button-exit-fullscreen"
            data-vc-for={part}
            icon={<FiMinimize2 className="text-2xl" />}
            intent="gray-basic"
            className="rounded-full absolute top-0 flex-none right-4 z-[999]"
            onClick={onExitFullscreen}
        />
    ) : (
        <>
            {!miniPlayer && (
                <IconButton
                    data-vc-element="floating-button-miniplayer"
                    data-vc-for={part}
                    icon={<FiMinimize2 className="text-2xl" />}
                    intent="gray-basic"
                    className="rounded-full absolute top-0 flex-none right-4 z-[999]"
                    onClick={onEnterMiniPlayer}
                />
            )}
            {miniPlayer && (
                <>
                    <IconButton
                        data-vc-element="floating-button-expand"
                        data-vc-for={part}
                        type="button"
                        intent="gray"
                        size="sm"
                        className="rounded-full text-xl flex-none absolute z-[999] right-4 top-4 pointer-events-auto bg-black/30 hover:bg-black/40"
                        icon={<BiExpand />}
                        onClick={onExpand}
                    />
                    <IconButton
                        data-vc-element="floating-button-terminate"
                        data-vc-for={part}
                        type="button"
                        intent="alert-subtle"
                        size="sm"
                        className="rounded-full text-xl flex-none absolute z-[999] left-4 top-4 pointer-events-auto"
                        icon={<BiX />}
                        onClick={onTerminate}
                    />
                </>
            )}
        </>
    )

    if (part === "loading") {
        return (
            <div
                data-vc-element="loading-floating-buttons-container"
                data-vc-for={part}
                className={cn("absolute top-8 w-full z-[100]", miniPlayer && "top-0")}
            >
                {content}
            </div>
        )
    }

    return content
}
