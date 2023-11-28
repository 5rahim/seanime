"use client"

import React from "react"
import { cn, ComponentWithAnatomy, defineStyleAnatomy } from "../core"
import { cva } from "class-variance-authority"

/* -------------------------------------------------------------------------------------------------
 * Anatomy
 * -----------------------------------------------------------------------------------------------*/

export const TimelineAnatomy = defineStyleAnatomy({
    root: cva([
        "UI-Timeline__root",
    ]),
    item: cva([
        "UI-Timeline__item",
        "flex text-md"
    ]),
    leftSection: cva([
        "UI-Timeline__leftSection",
        "flex flex-col items-center mr-4"
    ]),
    icon: cva([
        "UI-Timeline__icon",
        "flex items-center justify-center w-8 h-8 border border-[--border] rounded-full"
    ]),
    line: cva([
        "UI-Timeline__line",
        "w-px h-full bg-[--border]"
    ]),
    detailsSection: cva([
        "UI-Timeline__detailsSection",
        "pb-8"
    ]),
    title: cva([
        "UI-Timeline__title",
        "text-md font-semibold"
    ]),
    description: cva([
        "UI-Timeline__description",
        "text-[--muted] text-sm"
    ]),
})

/* -------------------------------------------------------------------------------------------------
 * Timeline
 * -----------------------------------------------------------------------------------------------*/

export interface TimelineProps extends React.ComponentPropsWithRef<"div">, ComponentWithAnatomy<typeof TimelineAnatomy> {
    children?: React.ReactNode
    items?: {
        title: React.ReactNode
        description?: React.ReactNode
        content?: React.ReactNode
        icon: React.ReactNode
        unstyledTitle?: boolean
        unstyledDescription?: boolean
        unstyledIcon?: boolean
        titleClassName?: string
        descriptionClassName?: string
        iconClassName?: string
        lineClassName?: string
    }[]
}

export const Timeline = React.forwardRef<HTMLDivElement, TimelineProps>((props, ref) => {

    const {
        children,
        rootClassName,
        itemClassName,
        leftSectionClassName,
        descriptionClassName,
        detailsSectionClassName,
        titleClassName,
        lineClassName,
        iconClassName,
        className,
        items,
        ...rest
    } = props

    return (
        <div
            className={cn(TimelineAnatomy.root(), rootClassName, className)}
            {...rest}
            ref={ref}
        >
            {items?.map((item, idx) => (
                <div
                    key={`${item.title}-${idx}`}
                    className={cn(
                        TimelineAnatomy.item(), itemClassName
                    )}
                >
                    {/*Left section*/}
                    <div className={cn(
                        TimelineAnatomy.leftSection(), leftSectionClassName
                    )}>
                        <div>
                            <div className={cn(
                                item.unstyledIcon ? null : TimelineAnatomy.icon(), iconClassName, item.iconClassName
                            )}>
                                {item.icon}
                            </div>
                        </div>
                        {(idx < items.length - 1) &&
                            <div className={cn(TimelineAnatomy.line(), lineClassName, item.lineClassName)}/>}
                    </div>

                    {/*Details section*/}
                    <div className={cn(
                        TimelineAnatomy.detailsSection(), detailsSectionClassName
                    )}>

                        <p className={cn(
                            item.unstyledTitle ? null : TimelineAnatomy.title(), titleClassName, item.titleClassName
                        )}>{item.title}</p>

                        {item.description && <p className={cn(
                            item.unstyledDescription ? null : TimelineAnatomy.description(), descriptionClassName, item.descriptionClassName
                        )}>
                            {item.description}
                        </p>}

                        {item.content}

                    </div>
                </div>
            ))}
        </div>
    )

})

Timeline.displayName = "Timeline"
