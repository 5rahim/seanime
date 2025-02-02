"use client"

import * as ScrollAreaPrimitive from "@radix-ui/react-scroll-area"
import { cva } from "class-variance-authority"
import * as React from "react"
import { cn, ComponentAnatomy, defineStyleAnatomy } from "../core/styling"

/* -------------------------------------------------------------------------------------------------
 * Anatomy
 * -----------------------------------------------------------------------------------------------*/

export const ScrollAreaAnatomy = defineStyleAnatomy({
    root: cva([
        "UI-ScrollArea__root",
        "relative overflow-hidden",
    ]),
    viewport: cva([
        "UI-ScrollArea__viewport",
        "h-full w-full rounded-[inherit]",
        "[&>div]:!block",
    ]),
    scrollbar:
        cva([
            "UI-ScrollArea__scrollbar",
            "flex touch-none select-none transition-colors",
        ], {
            variants: {
                orientation: {
                    vertical: "h-full w-2.5 border-l border-l-transparent p-[1px]",
                    horizontal: "h-2.5 flex-col border-t border-t-transparent p-[1px]",
                },
            },
            defaultVariants: {
                orientation: "vertical",
            },
        }),
    thumb: cva([
        "UI-ScrollArea__thumb",
        "relative flex-1 rounded-full bg-[--border]",
    ]),
})

/* -------------------------------------------------------------------------------------------------
 * ScrollArea
 * -----------------------------------------------------------------------------------------------*/

export type ScrollAreaProps =
    React.ComponentPropsWithoutRef<typeof ScrollAreaPrimitive.Root>
    & ComponentAnatomy<typeof ScrollAreaAnatomy> &
    {
        orientation?: "vertical" | "horizontal",
        viewportRef?: React.RefObject<HTMLDivElement>
    }

export const ScrollArea = React.forwardRef<HTMLDivElement, ScrollAreaProps>((props, ref) => {
    const {
        className,
        scrollbarClass,
        thumbClass,
        viewportClass,
        children,
        orientation = "vertical",
        viewportRef,
        ...rest
    } = props
    return (
        <ScrollAreaPrimitive.Root
            ref={ref}
            className={cn(ScrollAreaAnatomy.root(), className)}
            {...rest}
        >
            <ScrollAreaPrimitive.Viewport
                ref={viewportRef}
                className={cn(ScrollAreaAnatomy.viewport(), viewportClass)}
            >
                {children}
            </ScrollAreaPrimitive.Viewport>
            <ScrollBar
                className={scrollbarClass}
                thumbClass={thumbClass}
                orientation={orientation}
            />
            <ScrollAreaPrimitive.Corner />
        </ScrollAreaPrimitive.Root>
    )

})
ScrollArea.displayName = "ScrollArea"

/* -------------------------------------------------------------------------------------------------
 * ScrollBar
 * -----------------------------------------------------------------------------------------------*/

type ScrollBarProps =
    React.ComponentPropsWithoutRef<typeof ScrollAreaPrimitive.ScrollAreaScrollbar> &
    Pick<ComponentAnatomy<typeof ScrollAreaAnatomy>, "thumbClass">

const ScrollBar = React.forwardRef<HTMLDivElement, ScrollBarProps>((props, ref) => {
    const {
        className,
        thumbClass,
        orientation = "vertical",
        ...rest
    } = props

    return (
        <ScrollAreaPrimitive.ScrollAreaScrollbar
            ref={ref}
            orientation={orientation}
            className={cn(ScrollAreaAnatomy.scrollbar({ orientation }), className)}
            {...rest}
        >
            <ScrollAreaPrimitive.ScrollAreaThumb className={cn(ScrollAreaAnatomy.thumb(), thumbClass)} />
        </ScrollAreaPrimitive.ScrollAreaScrollbar>
    )
})
ScrollBar.displayName = "ScrollBar"
