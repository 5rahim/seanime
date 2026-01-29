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
        "z-50 overflow-hidden rounded-xl px-3 py-1.5 text-sm shadow-md animate-in fade-in-50",
        "bg-gray-900 border text-white",
        "data-[side=bottom]:slide-in-from-top-1 data-[side=left]:slide-in-from-right-1",
        "data-[side=right]:slide-in-from-left-1 data-[side=top]:slide-in-from-bottom-1",
    ]),
})

const TooltipContent = React.memo(
    React.forwardRef<
        React.ElementRef<typeof TooltipPrimitive.Content>,
        React.ComponentPropsWithoutRef<typeof TooltipPrimitive.Content> & {
        container?: HTMLElement
    }
    >(({ className, children, container, ...props }, ref) => {
        return (
            <TooltipPrimitive.Portal container={container}>
                <TooltipPrimitive.Content
                    ref={ref}
                    className={cn(TooltipAnatomy.root(), className)}
                    {...props}
                >
                    {children}
                </TooltipPrimitive.Content>
            </TooltipPrimitive.Portal>
        )
    }),
)

TooltipContent.displayName = "TooltipContent"

/* -------------------------------------------------------------------------------------------------
 * Tooltip
 * -----------------------------------------------------------------------------------------------*/

export type TooltipProps = React.ComponentPropsWithoutRef<typeof TooltipPrimitive.Content> &
    React.ComponentPropsWithoutRef<typeof TooltipPrimitive.Root> & {
    trigger: React.ReactElement
    portalContainer?: HTMLElement
}

export const Tooltip = React.memo(
    React.forwardRef<React.ElementRef<typeof TooltipPrimitive.Content>, TooltipProps>((props, ref) => {
        const {
            children,
            className,
            trigger,
            // Root props
            delayDuration = 50,
            disableHoverableContent,
            defaultOpen,
            open,
            onOpenChange,
            // Portal prop
            portalContainer,
            ...contentProps
        } = props

        return (
            <TooltipProvider delayDuration={delayDuration}>
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

                    <TooltipContent
                        ref={ref}
                        container={portalContainer}
                        className={className}
                        {...contentProps}
                    >
                        {children}
                    </TooltipContent>
                </TooltipPrimitive.Root>
            </TooltipProvider>
        )
    })
)

Tooltip.displayName = "Tooltip"

/* -------------------------------------------------------------------------------------------------
 * TooltipProvider
 * -----------------------------------------------------------------------------------------------*/

export const TooltipProvider = TooltipPrimitive.Provider
