"use client"

import * as HoverCardPrimitive from "@radix-ui/react-hover-card"
import { cva } from "class-variance-authority"
import * as React from "react"
import { cn, defineStyleAnatomy } from "../core/styling"

/* -------------------------------------------------------------------------------------------------
 * Anatomy
 * -----------------------------------------------------------------------------------------------*/

export const HoverCardAnatomy = defineStyleAnatomy({
    root: cva([
        "UI-HoverCard__root",
        "z-50 w-64 rounded-[--radius-md] border bg-[--paper] p-4 shadow-sm outline-none",
        "data-[state=open]:animate-in data-[state=closed]:animate-out data-[state=closed]:fade-out-0",
        "data-[state=open]:fade-in-0 data-[state=closed]:zoom-out-100 data-[state=open]:zoom-in-95",
        "data-[side=bottom]:slide-in-from-top-2 data-[side=left]:slide-in-from-right-2",
        "data-[side=right]:slide-in-from-left-2 data-[side=top]:slide-in-from-bottom-2",
    ]),
})

/* -------------------------------------------------------------------------------------------------
 * HoverCard
 * -----------------------------------------------------------------------------------------------*/

export type HoverCardProps = React.ComponentPropsWithoutRef<typeof HoverCardPrimitive.Content> & {
    trigger: React.ReactElement
    openDelay?: number
    closeDelay?: number
}

export const HoverCard = React.forwardRef<HTMLDivElement, HoverCardProps>((props, ref) => {
    const {
        className,
        align = "center",
        sideOffset = 8,
        openDelay = 1,
        closeDelay = 0,
        ...rest
    } = props

    return (
        <HoverCardPrimitive.Root openDelay={openDelay} closeDelay={closeDelay}>
            <HoverCardPrimitive.Trigger asChild>
                {props.trigger}
            </HoverCardPrimitive.Trigger>

            <HoverCardPrimitive.Content
                ref={ref}
                align={align}
                sideOffset={sideOffset}
                className={cn(HoverCardAnatomy.root(), className)}
                {...rest}
            />
        </HoverCardPrimitive.Root>
    )
})

HoverCard.displayName = "HoverCard"

