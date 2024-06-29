import { cn } from "@/components/ui/core/styling"
import { HorizontalDraggableScroll } from "@/components/ui/horizontal-draggable-scroll"
import { StaticTabs, StaticTabsItem } from "@/components/ui/tabs"
import React from "react"

type MediaGenreSelectorProps = {
    items: StaticTabsItem[]
    className?: string
    staticTabsClass?: string,
    staticTabsTriggerClass?: string
}

export function MediaGenreSelector(props: MediaGenreSelectorProps) {

    const {
        items,
        className,
        staticTabsClass,
        staticTabsTriggerClass,
        ...rest
    } = props

    return (
        <>
            <HorizontalDraggableScroll
                className={cn(
                    "scroll-pb-1 flex",
                    className,
                )}
            >
                <div className="flex flex-1"></div>
                <StaticTabs
                    className={cn(
                        "px-2 overflow-visible gap-2 py-4 w-fit",
                        staticTabsClass,
                    )}
                    triggerClass={cn(
                        "text-base rounded-md ring-2 ring-transparent data-[current=true]:ring-brand-500 data-[current=true]:text-brand-300",
                        staticTabsTriggerClass,
                    )}
                    items={items}
                />
                <div className="flex flex-1"></div>
            </HorizontalDraggableScroll>
        </>
    )
}
