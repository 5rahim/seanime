import { VideoCore_InSightCharacter } from "@/api/generated/types"
import { VideoCore_InSightData } from "@/api/generated/types"
import { useVideoCoreInSightGetCharacterDetails } from "@/api/hooks/videocore.hooks"

import { vc_videoElement } from "@/app/(main)/_features/video-core/video-core-atoms"
import { SeaImage } from "@/components/shared/sea-image"
import { IconButton } from "@/components/ui/button"
import { cn } from "@/components/ui/core/styling"
import { Popover } from "@/components/ui/popover"
import { TextInput } from "@/components/ui/text-input"
import { useAtomValue } from "jotai"
import { atom } from "jotai"
import { useAtom } from "jotai/react"
import { useRef } from "react"
import React, { useMemo, useState } from "react"
import { BiX } from "react-icons/bi"

export const vc_inSight_open = atom(false)
export const vc_inSight_data = atom<VideoCore_InSightData | null>(null)

export function useVideoCoreInSight() {
    const [open, setOpen] = useAtom(vc_inSight_open)
    const videoElement = useAtomValue(vc_videoElement)
    const [data, setData] = useAtom(vc_inSight_data)
    const wasPlayingRef = React.useRef(false)

    return {
        toggleOpen: (_open?: boolean) => {
            setOpen(prev => {
                const open = _open ?? !prev
                if (open) {
                    wasPlayingRef.current = !(videoElement?.paused ?? false)
                    if (wasPlayingRef.current) videoElement?.pause()
                } else {
                    if (wasPlayingRef.current) videoElement?.play()
                }
                return open
            })
        },
        open: open,
        data: data,
        setData: setData,
    }
}

export function VideoCoreInSight() {
    const { open, toggleOpen, data } = useVideoCoreInSight()
    const scrollContainerRef = useRef<HTMLDivElement>(null)
    const scrollContentRef = useRef<HTMLDivElement>(null)
    const searchInputRef = useRef<HTMLInputElement>(null)

    const [searchQuery, setSearchQuery] = useState("")

    const characters = useMemo(() => {
        const list = data?.characters || []
        if (!searchQuery) return list
        const res = list.filter(c => c.name?.toLowerCase().includes(searchQuery.toLowerCase()))
        if (res.length === 0) return [{
            mal_id: 0,
            name: "No results",
            images: {
                webp: {
                    image_url: "/no-cover.png",
                },
            },
        }] as VideoCore_InSightCharacter[]
        return res
    }, [data, searchQuery])

    React.useEffect(() => {
        if (!open) return

        const container = scrollContainerRef.current
        if (!container) return

        let scrollTarget = container.scrollLeft
        let animationFrameId: number | null = null
        let isDragging = false
        let startX = 0
        let scrollLeft = 0

        const smoothScroll = () => {
            const current = container.scrollLeft
            const diff = scrollTarget - current

            if (Math.abs(diff) > 0.5) {
                container.scrollLeft = current + diff * 0.15
                animationFrameId = requestAnimationFrame(smoothScroll)
            } else {
                container.scrollLeft = scrollTarget
                animationFrameId = null
            }
        }

        const handleWheel = (e: WheelEvent) => {
            const isTouchpad = Math.abs(e.deltaY) < 50 && e.deltaMode === 0

            if (isTouchpad) {
                e.preventDefault()
                if (Math.abs(e.deltaX) > Math.abs(e.deltaY)) {
                    container.scrollLeft += e.deltaX
                } else {
                    container.scrollLeft += e.deltaY
                }
            } else if (e.deltaY !== 0) {
                e.preventDefault()
                scrollTarget += e.deltaY * 1.2
                scrollTarget = Math.max(0, Math.min(scrollTarget, container.scrollWidth - container.clientWidth))

                if (animationFrameId === null) {
                    animationFrameId = requestAnimationFrame(smoothScroll)
                }
            }
        }

        const handleMouseDown = (e: MouseEvent) => {
            isDragging = true
            startX = e.pageX - container.offsetLeft
            scrollLeft = container.scrollLeft
            container.style.cursor = "grabbing"
            container.style.userSelect = "none"
        }

        const handleMouseLeave = () => {
            isDragging = false
            container.style.cursor = "grab"
        }

        const handleMouseUp = () => {
            isDragging = false
            container.style.cursor = "grab"
        }

        const handleMouseMove = (e: MouseEvent) => {
            if (!isDragging) return
            e.preventDefault()
            const x = e.pageX - container.offsetLeft
            const walk = (x - startX) * 1.5
            container.scrollLeft = scrollLeft - walk
        }

        container.style.cursor = "grab"
        container.addEventListener("wheel", handleWheel, { passive: false })
        container.addEventListener("mousedown", handleMouseDown)
        container.addEventListener("mouseleave", handleMouseLeave)
        container.addEventListener("mouseup", handleMouseUp)
        container.addEventListener("mousemove", handleMouseMove)

        return () => {
            container.style.cursor = ""
            container.style.userSelect = ""
            container.removeEventListener("wheel", handleWheel)
            container.removeEventListener("mousedown", handleMouseDown)
            container.removeEventListener("mouseleave", handleMouseLeave)
            container.removeEventListener("mouseup", handleMouseUp)
            container.removeEventListener("mousemove", handleMouseMove)
            if (animationFrameId !== null) {
                cancelAnimationFrame(animationFrameId)
            }
        }
    }, [open])

    // Auto-focus search input when InSight opens
    React.useEffect(() => {
        if (open && searchInputRef.current) {
            setTimeout(() => {
                searchInputRef.current?.focus()
            }, 100)
        }
    }, [open])

    React.useEffect(() => {
        if (!open) return

        const handleKeyDown = (e: KeyboardEvent) => {
            if (e.key === "Escape" && open) {
                e.preventDefault()
                e.stopPropagation()
                toggleOpen(false)
            }
        }

        document.addEventListener("keydown", handleKeyDown, true)

        return () => {
            document.removeEventListener("keydown", handleKeyDown, true)
        }
    }, [open, toggleOpen])

    if (!open) return null

    return (
        <div
            id="in-sight-container" data-vc-element="in-sight" className={cn(
            "absolute z-[50] bottom-32 left-0 w-full flex flex-col gap-2 items-end pointer-events-none",
        )}
        >
            <div className="absolute z-[1] -bottom-32 w-full h-full opacity-90 bg-gradient-to-t from-black to-transparent"></div>
            <div className="px-12 relative z-10 pointer-events-auto w-full flex">
                <div className="w-fit flex items-center gap-3">
                    <div className="w-fit">
                        <p className="text-2xl font-semibold text-white text-shadow-md">
                            Characters
                        </p>
                        <p className="text-white/60">May contain spoilers.</p>
                    </div>
                    <IconButton
                        icon={<BiX />}
                        intent="white-subtle"
                        size="md"
                        className="rounded-full"
                        onClick={() => toggleOpen()}
                    />
                </div>
                <div className="flex-1"></div>
                <TextInput
                    ref={searchInputRef}
                    type="text"
                    placeholder="Search characters..."
                    value={searchQuery}
                    onValueChange={(v) => setSearchQuery(v)}
                    fieldClass="w-[300px] !rounded-full"
                    className="bg-gray-950/70"
                />
            </div>
            <div
                ref={scrollContainerRef}
                data-vc-element="in-sight-scroll-container"
                className={cn(
                    "overflow-x-scroll z-[2] scrollbar-hide max-w-full w-full h-auto mask-image-fade pointer-events-auto",
                )}
                style={{
                    maskImage: "linear-gradient(to right, transparent, black 40px, black calc(100% - 40px), transparent)",
                    WebkitMaskImage: "linear-gradient(to right, transparent, black 40px, black calc(100% - 40px), transparent)",
                }}
            >
                <div
                    ref={scrollContentRef}
                    data-vc-element="in-sight-scroll-content"
                    className={cn(
                        "flex gap-4 flex-nowrap relative px-12 pb-4 py-6 !pr-20 h-[20rem] items-start",
                    )}
                >
                    {characters?.map(character => (
                        <Popover
                            key={character.mal_id}
                            className="z-[100] bg-gray-950/95 max-h-[14rem] w-[25rem] overflow-y-auto"
                            side="top"
                            sideOffset={8}
                            trigger={<div
                                data-vc-element="in-sight-character" className={cn(
                                "group/in-sight-character flex-none cursor-pointer",
                            )}
                            >
                                <div
                                    data-vc-element="in-sight-character-image"
                                    className={cn(
                                        "w-32 pointer-events-none aspect-[2/3] overflow-hidden rounded-3xl relative shadow-lg bg-gray-900 border border-gray-900/20 transition-all duration-300",
                                        "scale-90 opacity-90 group-hover/in-sight-character:scale-110 group-hover/in-sight-character:opacity-100 ease-in-out group-hover/in-sight-character:rounded-3xl origin-bottom z-0 group-hover/in-sight-character:z-10",
                                    )}
                                >
                                    <SeaImage
                                        src={character.images?.webp?.image_url ?? "/no-cover.png"}
                                        fill
                                        className="object-cover object-center transition-transform duration-500 ease-in-out group-hover/in-sight-character:scale-105"
                                    />
                                </div>
                                <div className="w-32 p-1.5 rounded-lg tracking-wide bg-gray-950 group-hover/in-sight-character:bg-opacity-70 bg-opacity-50 backdrop-blur-sm line-clamp-3 text-base font-semibold mt-2 text-center text-white transition-opacity duration-300 text-shadow-md">
                                    {character.name}
                                </div>
                            </div>}
                        >
                            <CharacterPopoverContent character={character} />
                        </Popover>
                    ))}
                    <div className="w-[40px] flex-none h-full">

                    </div>
                </div>
            </div>
        </div>
    )
}

function CharacterPopoverContent({ character }: { character: VideoCore_InSightCharacter }) {
    const { data: details, isLoading } = useVideoCoreInSightGetCharacterDetails(character.mal_id)

    if (isLoading) return <p>Loading...</p>

    return (
        <p className="text-justify text-white">
            {details?.about || "No details available."}
        </p>
    )
}
