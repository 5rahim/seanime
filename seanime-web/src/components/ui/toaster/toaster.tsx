"use client"

import { cva } from "class-variance-authority"
import * as React from "react"
import { Toaster as Sonner } from "sonner"
import { cn, defineStyleAnatomy } from "../core/styling"

/* -------------------------------------------------------------------------------------------------
 * Anatomy
 * -----------------------------------------------------------------------------------------------*/

export const ToasterAnatomy = defineStyleAnatomy({
    toaster: cva(["group toaster z-[150]"]),
    toast: cva([
        "group/toast",
        "select-none cursor-default",
        "group-[.toaster]:py-4 group-[.toaster]:px-6 group-[.toaster]:gap-3",
        "group-[.toaster]:text-sm group-[.toaster]:font-medium",
        "group-[.toaster]:rounded-xl group-[.toaster]:border group-[.toaster]:backdrop-blur-sm",
        // "group-[.toaster]:ring-1 group-[.toaster]:ring-inset",
        "group-[.toaster]:transition-all group-[.toaster]:duration-200",
        // Default/Base style
        // "group-[.toaster]:bg-gradient-to-br group-[.toaster]:from-[--paper] group-[.toaster]:to-[--paper]/80",
        "group-[.toaster]:text-[--foreground] group-[.toaster]:border-[--border]",
        "group-[.toaster]:ring-[--border]",
        // Success
        "group-[.toaster]:data-[type=success]:bg-gradient-to-br",
        "group-[.toaster]:data-[type=success]:from-emerald-950/95 group-[.toaster]:data-[type=success]:to-emerald-900/60",
        "group-[.toaster]:data-[type=success]:text-emerald-100",
        "group-[.toaster]:data-[type=success]:border-emerald-800/50",
        "group-[.toaster]:data-[type=success]:ring-emerald-700/40",
        // Warning
        "group-[.toaster]:data-[type=warning]:bg-gradient-to-br",
        "group-[.toaster]:data-[type=warning]:from-amber-950/95 group-[.toaster]:data-[type=warning]:to-amber-900/60",
        "group-[.toaster]:data-[type=warning]:text-amber-100",
        "group-[.toaster]:data-[type=warning]:border-amber-800/50",
        "group-[.toaster]:data-[type=warning]:ring-amber-700/40",
        // Error
        "group-[.toaster]:data-[type=error]:bg-gradient-to-br",
        "group-[.toaster]:data-[type=error]:from-red-950/95 group-[.toaster]:data-[type=error]:to-red-900/60",
        "group-[.toaster]:data-[type=error]:text-red-100",
        "group-[.toaster]:data-[type=error]:border-red-800/50",
        "group-[.toaster]:data-[type=error]:ring-red-700/40",
        // Info
        "group-[.toaster]:data-[type=info]:bg-gradient-to-br",
        "group-[.toaster]:data-[type=info]:from-blue-950/95 group-[.toaster]:data-[type=info]:to-blue-900/60",
        "group-[.toaster]:data-[type=info]:text-blue-100",
        "group-[.toaster]:data-[type=info]:border-blue-800/50",
        "group-[.toaster]:data-[type=info]:ring-blue-700/40",
    ]),
    description: cva([
        "group/toast:text-xs group/toast:font-normal group/toast:mt-1",
        "group/toast:opacity-80",
        "group-data-[type=success]/toast:text-emerald-300",
        "group-data-[type=warning]/toast:text-amber-300",
        "group-data-[type=error]/toast:text-red-300",
        "group-data-[type=info]/toast:text-blue-300",
        "cursor-default",
    ]),
    actionButton: cva([
        "group/toast:bg-[--subtle] group/toast:text-[--foreground]",
        "group/toast:rounded-lg group/toast:px-3 group/toast:py-1.5",
        "group/toast:text-xs group/toast:font-medium",
        "group/toast:transition-colors group/toast:hover:bg-[--subtle-hover]",
        "group/toast:ring-1 group/toast:ring-[--border]/20",
    ]),
    cancelButton: cva([
        "group/toast:bg-transparent group/toast:text-[--muted]",
        "group/toast:rounded-lg group/toast:px-3 group/toast:py-1.5",
        "group/toast:text-xs group/toast:font-medium",
        "group/toast:transition-colors group/toast:hover:bg-[--subtle]",
        "group/toast:ring-1 group/toast:ring-transparent group/toast:hover:ring-[--border]/20",
    ]),
})

/* -------------------------------------------------------------------------------------------------
 * Toaster
 * -----------------------------------------------------------------------------------------------*/

export type ToasterProps = React.ComponentProps<typeof Sonner>

export const Toaster = ({ position = "top-center", ...props }: ToasterProps) => {

    const allProps = React.useMemo(() => ({
        position,
        visibleToasts: 4,
        className: cn(ToasterAnatomy.toaster()),
        toastOptions: {
            classNames: {
                toast: cn(ToasterAnatomy.toast()),
                description: cn(ToasterAnatomy.description()),
                actionButton: cn(ToasterAnatomy.actionButton()),
                cancelButton: cn(ToasterAnatomy.cancelButton()),
            },
        },
        ...props,
    } as ToasterProps), [])

    return (
        <>
            <Sonner theme="dark" {...allProps} />
        </>
    )
}

Toaster.displayName = "Toaster"
