import { cva } from "class-variance-authority"
import * as React from "react"
import { cn, ComponentAnatomy, defineStyleAnatomy } from "../core/styling"

/* -------------------------------------------------------------------------------------------------
 * Anatomy
 * -----------------------------------------------------------------------------------------------*/

export const TimelineAnatomy = defineStyleAnatomy({
    root: cva([
        "UI-Timeline__root",
    ]),
    item: cva([
        "UI-Timeline__item",
        "flex text-md",
    ]),
    leftSection: cva([
        "UI-Timeline__leftSection",
        "flex flex-col items-center mr-4",
    ]),
    icon: cva([
        "UI-Timeline__icon",
        "flex items-center justify-center w-8 h-8 border rounded-full flex-none",
    ]),
    line: cva([
        "UI-Timeline__line",
        "w-px h-full bg-[--border]",
    ]),
    detailsSection: cva([
        "UI-Timeline__detailsSection",
        "pb-8",
    ]),
    title: cva([
        "UI-Timeline__title",
        "text-md font-semibold",
    ]),
    description: cva([
        "UI-Timeline__description",
        "text-[--muted] text-sm",
    ]),
})

/* -------------------------------------------------------------------------------------------------
 * Timeline
 * -----------------------------------------------------------------------------------------------*/

export type TimelineItem = {
    title: React.ReactNode
    description?: React.ReactNode
    content?: React.ReactNode
    icon: React.ReactNode
    unstyledTitle?: boolean
    unstyledDescription?: boolean
    unstyledIcon?: boolean
    titleClass?: string
    descriptionClass?: string
    iconClass?: string
    lineClass?: string
}

export type TimelineProps = React.ComponentPropsWithoutRef<"div"> & ComponentAnatomy<typeof TimelineAnatomy> & {
    children?: React.ReactNode
    items: TimelineItem[]
}

export const Timeline = React.forwardRef<HTMLDivElement, TimelineProps>((props, ref) => {

    const {
        children,
        itemClass,
        leftSectionClass,
        descriptionClass,
        detailsSectionClass,
        titleClass,
        lineClass,
        iconClass,
        className,
        items,
        ...rest
    } = props

    return (
        <div
            ref={ref}
            className={cn(TimelineAnatomy.root(), className)}
            {...rest}
        >
            {items.map((item, idx) => (
                <div
                    key={`${idx}`}
                    className={cn(TimelineAnatomy.item(), itemClass)}
                >
                    {/*Left section*/}
                    <div className={cn(TimelineAnatomy.leftSection(), leftSectionClass)}>
                        <div
                            className={cn(
                                item.unstyledIcon ?
                                    null :
                                    TimelineAnatomy.icon(),
                                iconClass,
                                item.iconClass,
                            )}
                        >
                            {item.icon}
                        </div>
                        {(idx < items.length - 1) && <div className={cn(TimelineAnatomy.line(), lineClass, item.lineClass)} />}
                    </div>

                    {/*Details section*/}
                    <div className={cn(TimelineAnatomy.detailsSection(), detailsSectionClass)}>

                        <div
                            className={cn(
                                item.unstyledTitle ?
                                    null :
                                    TimelineAnatomy.title(),
                                titleClass,
                                item.titleClass,
                            )}
                        >
                            {item.title}
                        </div>

                        {item.description && <div
                            className={cn(
                                item.unstyledDescription ?
                                    null :
                                    TimelineAnatomy.description(),
                                descriptionClass,
                                item.descriptionClass,
                            )}
                        >
                            {item.description}
                        </div>}

                        {item.content}

                    </div>
                </div>
            ))}
        </div>
    )

})

Timeline.displayName = "Timeline"
