"use client"

import * as TooltipPrimitive from "@radix-ui/react-tooltip"
import { cva } from "class-variance-authority"
import * as React from "react"
import { cn, defineStyleAnatomy } from "../core/styling"

/* -------------------------------------------------------------------------------------------------
 * Anatomy
 * -----------------------------------------------------------------------------------------------*/

export const TooltipAnatomy = defineStyleAnatomy({
    root: cva([
        "UI-Tooltip__root",
        "z-50 overflow-hidden rounded-[--radius] px-3 py-1.5 text-sm shadow-md animate-in fade-in-50",
        "bg-gray-800 text-white",
        "data-[side=bottom]:slide-in-from-top-1 data-[side=left]:slide-in-from-right-1",
        "data-[side=right]:slide-in-from-left-1 data-[side=top]:slide-in-from-bottom-1",
    ]),
})

/* -------------------------------------------------------------------------------------------------
 * Tooltip
 * -----------------------------------------------------------------------------------------------*/

export type TooltipProps = React.ComponentPropsWithoutRef<typeof TooltipPrimitive.Content> &
    React.ComponentPropsWithoutRef<typeof TooltipPrimitive.Root> & {
    /**
     * The trigger that toggles the tooltip.
     * - Passed props: `data-state`	("closed" | "delayed-open" | "instant-open")
     */
    trigger: React.ReactElement
    /**
     * Portal container for custom mounting (useful for fullscreen mode)
     */
    portalContainer?: HTMLElement
}

export const Tooltip = React.forwardRef<HTMLDivElement, TooltipProps>((props, ref) => {

    const {
        children,
        className,
        trigger,
        // Root
        delayDuration = 50,
        disableHoverableContent,
        defaultOpen,
        open,
        onOpenChange,
        // Portal
        portalContainer,
        ...rest
    } = props

    return (
        <TooltipProvider>
            <TooltipPrimitive.Root
                delayDuration={delayDuration}
                disableHoverableContent={disableHoverableContent}
                defaultOpen={defaultOpen}
                open={open}
                onOpenChange={onOpenChange}
            >
                <TooltipPrimitive.Trigger asChild>
                    {trigger}
                </TooltipPrimitive.Trigger>
                <TooltipPrimitive.Portal container={portalContainer}>
                    <TooltipPrimitive.Content
                        ref={ref}
                        className={cn(TooltipAnatomy.root(), className)}
                        {...rest}
                    >
                        {children}
                    </TooltipPrimitive.Content>
                </TooltipPrimitive.Portal>
            </TooltipPrimitive.Root>
        </TooltipProvider>
    )

})

Tooltip.displayName = "Tooltip"

/* -------------------------------------------------------------------------------------------------
 * TooltipProvider
 * -----------------------------------------------------------------------------------------------*/

/**
 * Wraps your app to provide global functionality to your tooltips.
 */
export const TooltipProvider = TooltipPrimitive.Provider
