"use client"

import * as PopoverPrimitive from "@radix-ui/react-popover"
import { cva } from "class-variance-authority"
import * as React from "react"
import { cn, defineStyleAnatomy } from "../core/styling"

/* -------------------------------------------------------------------------------------------------
 * Anatomy
 * -----------------------------------------------------------------------------------------------*/

export const PopoverAnatomy = defineStyleAnatomy({
    root: cva([
        "UI-Popover__root",
        "z-50 w-72 rounded-[--radius] border bg-[--background] p-4 text-base shadow-sm outline-none",
        "data-[state=open]:animate-in data-[state=closed]:animate-out data-[state=closed]:fade-out-0",
        "data-[state=open]:fade-in-50 data-[state=closed]:zoom-out-100 data-[state=open]:zoom-in-95",
        "data-[side=bottom]:slide-in-from-top-2 data-[side=left]:slide-in-from-right-2",
        "data-[side=right]:slide-in-from-left-2 data-[side=top]:slide-in-from-bottom-2",
    ]),
})

/* -------------------------------------------------------------------------------------------------
 * Popover
 * -----------------------------------------------------------------------------------------------*/

export type PopoverProps =
    React.ComponentPropsWithoutRef<typeof PopoverPrimitive.Root> &
    Omit<React.ComponentPropsWithoutRef<typeof PopoverPrimitive.Content>, "asChild"> &
    {
        /**
         * The trigger element that opens the popover
         */
        trigger: React.ReactElement,
        /**
         * Additional props for the trigger element
         */
        triggerProps?: React.ComponentPropsWithoutRef<typeof PopoverPrimitive.Trigger>,
        /**
         * Portal container for custom mounting (useful for fullscreen mode)
         */
        portalContainer?: HTMLElement
    }

export const Popover = React.forwardRef<HTMLDivElement, PopoverProps>((props, ref) => {
    const {
        trigger,
        triggerProps,
        // Root
        defaultOpen,
        open,
        onOpenChange,
        modal = true,
        // Content
        className,
        align = "center",
        sideOffset = 8,
        // Portal
        portalContainer,
        ...contentProps
    } = props

    return (
        <PopoverPrimitive.Root
            defaultOpen={defaultOpen}
            open={open}
            onOpenChange={onOpenChange}
            modal={modal}
        >
            <PopoverPrimitive.Trigger
                asChild
                {...triggerProps}
            >
                {trigger}
            </PopoverPrimitive.Trigger>
            <PopoverPrimitive.Portal container={portalContainer}>
                <PopoverPrimitive.Content
                    ref={ref}
                    align={align}
                    sideOffset={sideOffset}
                    className={cn(PopoverAnatomy.root(), className)}
                    onOpenAutoFocus={(e) => e.preventDefault()}
                    {...contentProps}
                />
            </PopoverPrimitive.Portal>
        </PopoverPrimitive.Root>
    )
})

Popover.displayName = "Popover"

