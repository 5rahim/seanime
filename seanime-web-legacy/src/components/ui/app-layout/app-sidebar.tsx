"use client"

import { cva, VariantProps } from "class-variance-authority"
import * as React from "react"
import { AppLayoutAnatomy } from "."
import { cn, ComponentAnatomy, defineStyleAnatomy } from "../core/styling"
import { Drawer, DrawerProps } from "../drawer"

/* -------------------------------------------------------------------------------------------------
 * Context
 * -----------------------------------------------------------------------------------------------*/

export const __AppSidebarContext = React.createContext<{
    open: boolean,
    setOpen: (open: boolean) => void,
    size: VariantProps<typeof AppLayoutAnatomy.root>["sidebarSize"]
    setSize: (size: VariantProps<typeof AppLayoutAnatomy.root>["sidebarSize"]) => void,
    isBelowBreakpoint: boolean,
}>({
    open: false,
    setOpen: () => {},
    setSize: () => {},
    size: "md",
    isBelowBreakpoint: false,
})

export function useAppSidebarContext() {
    const ctx = React.useContext(__AppSidebarContext)
    if (!ctx) throw new Error("useAppSidebarContext must be used within a AppSidebarProvider")
    return ctx
}

/* -------------------------------------------------------------------------------------------------
 * Anatomy
 * -----------------------------------------------------------------------------------------------*/

export const AppSidebarAnatomy = defineStyleAnatomy({
    sidebar: cva([
        "UI-AppSidebar__sidebar",
        "flex flex-grow flex-col overflow-y-auto border-r border-transparent bg-[--background]",
    ]),
})

export const AppSidebarTriggerAnatomy = defineStyleAnatomy({
    trigger: cva([
        "UI-AppSidebarTrigger__trigger",
        "block lg:hidden",
        "items-center justify-center rounded-[--radius] p-2 text-[--muted] hover:bg-[--subtle] hover:text-[--foreground] transition-colors",
        "focus:outline-none focus:ring-2 focus:ring-inset focus:ring-[--ring]",
    ]),
})

/* -------------------------------------------------------------------------------------------------
 * AppSidebar
 * -----------------------------------------------------------------------------------------------*/

export type AppSidebarProps = React.ComponentPropsWithoutRef<"div"> & ComponentAnatomy<typeof AppSidebarAnatomy> & {
    mobileDrawerProps?: Partial<DrawerProps>
}

export const AppSidebar = React.forwardRef<HTMLDivElement, AppSidebarProps>((props, ref) => {

    const {
        children,
        className,
        ...rest
    } = props

    const ctx = React.useContext(__AppSidebarContext)

    return (
        <>
            <div
                ref={ref}
                className={cn(
                    AppSidebarAnatomy.sidebar(),
                    // __isDesktop__ && "pt-4",
                    className,
                )}
                {...rest}
            >
                {children}
            </div>
            <Drawer
                open={ctx.open}
                onOpenChange={v => ctx.setOpen(v)}
                side="left"
            >
                {children}
            </Drawer>
        </>
    )

})

AppSidebar.displayName = "AppSidebar"

/* -------------------------------------------------------------------------------------------------
 * AppSidebarTrigger
 * -----------------------------------------------------------------------------------------------*/

export type AppSidebarTriggerProps = React.ComponentPropsWithoutRef<"button"> & ComponentAnatomy<typeof AppSidebarTriggerAnatomy>

export const AppSidebarTrigger = React.forwardRef<HTMLButtonElement, AppSidebarTriggerProps>((props, ref) => {

    const {
        children,
        className,
        ...rest
    } = props

    const ctx = React.useContext(__AppSidebarContext)

    return (
        <button
            ref={ref}
            className={cn(AppSidebarTriggerAnatomy.trigger(), className)}
            onClick={() => ctx.setOpen(!ctx.open)}
            {...rest}
        >
            <span className="sr-only">Open main menu</span>
            {ctx.open ? (
                <svg
                    xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none"
                    stroke="currentColor"
                    strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" className="block h-6 w-6"
                >
                    <line x1="18" x2="6" y1="6" y2="18"></line>
                    <line x1="6" x2="18" y1="6" y2="18"></line>
                </svg>
            ) : (
                <svg
                    xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none"
                    stroke="currentColor"
                    strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" className="block h-6 w-6"
                >
                    <line x1="4" x2="20" y1="12" y2="12"></line>
                    <line x1="4" x2="20" y1="6" y2="6"></line>
                    <line x1="4" x2="20" y1="18" y2="18"></line>
                </svg>
            )}
        </button>
    )

})

AppSidebarTrigger.displayName = "AppSidebarTrigger"


/* -------------------------------------------------------------------------------------------------
 * AppSidebarProvider
 * -----------------------------------------------------------------------------------------------*/

export type AppSidebarProviderProps = {
    children?: React.ReactNode,
    open?: boolean,
    onOpenChange?: (open: boolean) => void,
    onSizeChange?: (size: VariantProps<typeof AppLayoutAnatomy.root>["sidebarSize"]) => void,
}

export const AppSidebarProvider: React.FC<AppSidebarProviderProps> = ({
    children,
    onOpenChange,
    onSizeChange,
}) => {

    const [open, setOpen] = React.useState(false)
    const [size, setSize] = React.useState<VariantProps<typeof AppLayoutAnatomy.root>["sidebarSize"]>(undefined)

    const [isBelowBreakpoint, setIsBelowBreakpoint] = React.useState<boolean>(false)

    React.useEffect(() => {
        const handleResize = () => setIsBelowBreakpoint(window.innerWidth < 1024) // lg breakpoint
        handleResize()
        window.addEventListener("resize", handleResize)
        return () => window.removeEventListener("resize", handleResize)
    }, [isBelowBreakpoint])

    return (
        <__AppSidebarContext.Provider
            value={{
                open,
                setOpen: (open: boolean) => {
                    onOpenChange?.(open)
                    setOpen(open)
                },
                setSize: (size: VariantProps<typeof AppLayoutAnatomy.root>["sidebarSize"]) => {
                    onSizeChange?.(size)
                    setSize(size)
                },
                size: size,
                isBelowBreakpoint,
            }}
        >
            {children}
        </__AppSidebarContext.Provider>
    )
}

AppSidebarProvider.displayName = "AppSidebarProvider"
