"use client"

import { cn } from "@/components/ui/core/styling"
import * as React from "react"
import { Drawer as VaulPrimitive } from "vaul"

const Vaul = ({
    shouldScaleBackground = false,
    ...props
}: React.ComponentProps<typeof VaulPrimitive.Root>) => (
    <VaulPrimitive.Root
        shouldScaleBackground={shouldScaleBackground}
        {...props}
    />
)
Vaul.displayName = "Vaul"

const VaulTrigger = VaulPrimitive.Trigger

const VaulPortal = VaulPrimitive.Portal

const VaulClose = VaulPrimitive.Close

const VaulOverlay = React.forwardRef<
    React.ElementRef<typeof VaulPrimitive.Overlay>,
    React.ComponentPropsWithoutRef<typeof VaulPrimitive.Overlay>
>(({ className, ...props }, ref) => {
    return (
        <VaulPrimitive.Overlay
            ref={ref}
            className={cn("fixed inset-0 z-50 bg-black/80", className)}
            {...props}
        />
    )
})
VaulOverlay.displayName = VaulPrimitive.Overlay.displayName

const VaulContent = React.forwardRef<
    React.ElementRef<typeof VaulPrimitive.Content>,
    React.ComponentPropsWithoutRef<typeof VaulPrimitive.Content>
>(({ className, children, ...props }, ref) => {
    return (
        <VaulPortal>
            <VaulOverlay />
            <VaulPrimitive.Content
                ref={ref}
                className={cn(
                    "fixed inset-x-0 bottom-0 z-50 mt-24 flex h-auto flex-col rounded-t-[10px] border bg-[var(--background)]",
                    className,
                )}
                {...props}
            >
                <div className="mx-auto mt-4 h-2 w-[100px] rounded-full bg-[--subtle]" />
                {children}
            </VaulPrimitive.Content>
        </VaulPortal>
    )
})
VaulContent.displayName = "VaulContent"

const VaulHeader = ({
    className,
    ...props
}: React.HTMLAttributes<HTMLDivElement>) => {
    return (
        <div
            className={cn("grid gap-1.5 text-center sm:text-left", className)}
            {...props}
        />
    )
}
VaulHeader.displayName = "VaulHeader"

const VaulFooter = ({
    className,
    ...props
}: React.HTMLAttributes<HTMLDivElement>) => {
    return (
        <div
            className={cn("mt-auto flex flex-col gap-2 p-4", className)}
            {...props}
        />
    )
}
VaulFooter.displayName = "VaulFooter"

const VaulTitle = React.forwardRef<
    React.ElementRef<typeof VaulPrimitive.Title>,
    React.ComponentPropsWithoutRef<typeof VaulPrimitive.Title>
>(({ className, ...props }, ref) => {
    return (
        <VaulPrimitive.Title
            ref={ref}
            className={cn(
                "text-2xl font-semibold leading-none tracking-tight",
                className,
            )}
            {...props}
        />
    )
})
VaulTitle.displayName = VaulPrimitive.Title.displayName

const VaulDescription = React.forwardRef<
    React.ElementRef<typeof VaulPrimitive.Description>,
    React.ComponentPropsWithoutRef<typeof VaulPrimitive.Description>
>(({ className, ...props }, ref) => {
    return (
        <VaulPrimitive.Description
            ref={ref}
            className={cn("text-sm text-muted-foreground", className)}
            {...props}
        />
    )
})
VaulDescription.displayName = VaulPrimitive.Description.displayName

export {
    Vaul,
    VaulPortal,
    VaulOverlay,
    VaulTrigger,
    VaulClose,
    VaulContent,
    VaulHeader,
    VaulFooter,
    VaulTitle,
    VaulDescription,
}
