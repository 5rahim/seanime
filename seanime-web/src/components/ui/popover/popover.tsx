"use client"

import React from "react"
import { cn, ComponentWithAnatomy, defineStyleAnatomy } from "../core"
import { cva } from "class-variance-authority"
import type { PopoverContentProps as PopoverPrimitiveContentProps } from "@radix-ui/react-popover"
import * as PopoverPrimitive from "@radix-ui/react-popover"


/* -------------------------------------------------------------------------------------------------
 * Anatomy
 * -----------------------------------------------------------------------------------------------*/

export const PopoverAnatomy = defineStyleAnatomy({
    popover: cva([
        "UI-Popover__popover relative",
        "w-72",
        "z-50 rounded-[--radius] border border-[--border] bg-[--paper] p-4 shadow-md outline-none animate-in",
        "data-[side=bottom]:slide-in-from-bottom-2 data-[side=left]:slide-in-from-right-2 data-[side=right]:slide-in-from-left-2 data-[side=top]:slide-in-from-top-2",
        "data-[side=bottom]:mt-2 data-[side=top]:mb-2 data-[side=left]:mr-2 data-[side=right]:ml-2"
    ])
})

/* -------------------------------------------------------------------------------------------------
 * Popover
 * -----------------------------------------------------------------------------------------------*/

export interface PopoverProps extends React.ComponentPropsWithRef<"div">,
    ComponentWithAnatomy<typeof PopoverAnatomy>, PopoverPrimitiveContentProps {
    trigger: React.ReactNode
}

export const Popover: React.FC<PopoverProps> = React.forwardRef<HTMLDivElement, PopoverProps>((props, ref) => {

    const {
        children,
        trigger,
        popoverClassName,
        className,
        ...rest
    } = props

    return (
        <PopoverPrimitive.Root>
            <PopoverPrimitive.Trigger asChild>
                {trigger}
            </PopoverPrimitive.Trigger>

            <PopoverPrimitive.Portal>
                <PopoverPrimitive.Content
                    className={cn([
                        PopoverAnatomy.popover(),
                        popoverClassName,
                        className,
                    ])}
                    ref={ref}
                    {...rest}
                >
                    {children}
                </PopoverPrimitive.Content>
            </PopoverPrimitive.Portal>
        </PopoverPrimitive.Root>
    )

})

Popover.displayName = "Popover"
