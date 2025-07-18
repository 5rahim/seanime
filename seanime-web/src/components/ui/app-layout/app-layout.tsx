"use client"
import { ElectronSidebarPaddingMacOS } from "@/app/(main)/_electron/electron-padding"
import { TauriSidebarPaddingMacOS } from "@/app/(main)/_tauri/tauri-padding"
import { __isDesktop__, __isElectronDesktop__, __isTauriDesktop__ } from "@/types/constants"
import { cva, VariantProps } from "class-variance-authority"
import * as React from "react"
import { __AppSidebarContext } from "."
import { cn, ComponentAnatomy, defineStyleAnatomy } from "../core/styling"

/* -------------------------------------------------------------------------------------------------
 * Anatomy
 * -----------------------------------------------------------------------------------------------*/

export const AppLayoutAnatomy = defineStyleAnatomy({
    root: cva([
        "UI-AppLayout__root appLayout",
        "flex w-full group/appLayout",
    ], {
        variants: {
            withSidebar: {
                true: "flex-row with-sidebar",
                false: "flex-col",
            },
            sidebarSize: {
                slim: "sidebar-slim",
                sm: "sidebar-sm",
                md: "sidebar-md",
                lg: "sidebar-lg",
                xl: "sidebar-xl",
            },
        },
        defaultVariants: {
            withSidebar: false,
            sidebarSize: "md",
        },
        compoundVariants: [
            { withSidebar: true, sidebarSize: "slim", className: "lg:[&>.appLayout]:pl-20" },
            { withSidebar: true, sidebarSize: "sm", className: "lg:[&>.appLayout]:pl-48" },
            { withSidebar: true, sidebarSize: "md", className: "lg:[&>.appLayout]:pl-64" },
            { withSidebar: true, sidebarSize: "lg", className: "lg:[&>.appLayout]:pl-[20rem]" },
            { withSidebar: true, sidebarSize: "xl", className: "lg:[&>.appLayout]:pl-[25rem]" },
        ],
    }),
})

export const AppLayoutHeaderAnatomy = defineStyleAnatomy({
    root: cva([
        "UI-AppLayoutHeader__root",
        "relative w-full",
    ]),
})

export const AppLayoutSidebarAnatomy = defineStyleAnatomy({
    root: cva([
        "UI-AppLayoutSidebar__root z-50",
        "hidden lg:fixed lg:inset-y-0 lg:flex lg:flex-col grow-0 shrink-0 basis-0",
        "group-[.sidebar-slim]/appLayout:w-20",
        "group-[.sidebar-sm]/appLayout:w-48",
        "group-[.sidebar-md]/appLayout:w-64",
        "group-[.sidebar-lg]/appLayout:w-[20rem]",
        "group-[.sidebar-xl]/appLayout:w-[25rem]",
    ]),
})

export const AppLayoutContentAnatomy = defineStyleAnatomy({
    root: cva([
        "UI-AppLayoutContent__root",
        "relative",
    ]),
})

export const AppLayoutFooterAnatomy = defineStyleAnatomy({
    root: cva([
        "UI-AppLayoutFooter__root",
        "relative",
    ]),
})

export const AppLayoutStackAnatomy = defineStyleAnatomy({
    root: cva([
        "UI-AppLayoutStack__root",
        "relative",
    ], {
        variants: {
            spacing: {
                sm: "space-y-2",
                md: "space-y-4",
                lg: "space-y-8",
                xl: "space-y-10",
            },
        },
        defaultVariants: {
            spacing: "md",
        },
    }),
})

export const AppLayoutGridAnatomy = defineStyleAnatomy({
    root: cva([
        "UI-AppLayoutGrid__root",
        "relative flex flex-col",
    ], {
        variants: {
            breakBelow: {
                sm: "sm:grid sm:space-y-0",
                md: "md:grid md:space-y-0",
                lg: "lg:grid lg:space-y-0",
                xl: "xl:grid xl:space-y-0",
            },
            spacing: {
                sm: "gap-2",
                md: "gap-4",
                lg: "gap-8",
                xl: "gap-10",
            },
            cols: { 1: null, 2: null, 3: null, 4: null, 5: null, 6: null },
        },
        defaultVariants: {
            breakBelow: "xl",
            spacing: "md",
            cols: 3,
        },
        compoundVariants: [
            { breakBelow: "sm", cols: 1, className: "sm:grid-cols-1" },
            { breakBelow: "sm", cols: 2, className: "sm:grid-cols-2" },
            { breakBelow: "sm", cols: 3, className: "sm:grid-cols-3" },
            { breakBelow: "sm", cols: 4, className: "sm:grid-cols-4" },
            { breakBelow: "sm", cols: 5, className: "sm:grid-cols-5" },
            { breakBelow: "sm", cols: 6, className: "sm:grid-cols-6" },
            { breakBelow: "md", cols: 1, className: "md:grid-cols-1" },
            { breakBelow: "md", cols: 2, className: "md:grid-cols-2" },
            { breakBelow: "md", cols: 3, className: "md:grid-cols-3" },
            { breakBelow: "md", cols: 4, className: "md:grid-cols-4" },
            { breakBelow: "md", cols: 5, className: "md:grid-cols-5" },
            { breakBelow: "md", cols: 6, className: "md:grid-cols-6" },
            { breakBelow: "lg", cols: 1, className: "lg:grid-cols-1" },
            { breakBelow: "lg", cols: 2, className: "lg:grid-cols-2" },
            { breakBelow: "lg", cols: 3, className: "lg:grid-cols-3" },
            { breakBelow: "lg", cols: 4, className: "lg:grid-cols-4" },
            { breakBelow: "lg", cols: 5, className: "lg:grid-cols-5" },
            { breakBelow: "lg", cols: 6, className: "lg:grid-cols-6" },
            { breakBelow: "xl", cols: 1, className: "xl:grid-cols-1" },
            { breakBelow: "xl", cols: 2, className: "xl:grid-cols-2" },
            { breakBelow: "xl", cols: 3, className: "xl:grid-cols-3" },
            { breakBelow: "xl", cols: 4, className: "xl:grid-cols-4" },
            { breakBelow: "xl", cols: 5, className: "xl:grid-cols-5" },
            { breakBelow: "xl", cols: 6, className: "xl:grid-cols-6" },
        ],
    }),
})

/* -------------------------------------------------------------------------------------------------
 * AppLayout
 * -----------------------------------------------------------------------------------------------*/

export type AppLayoutProps = React.ComponentPropsWithRef<"div"> &
    ComponentAnatomy<typeof AppLayoutAnatomy> &
    VariantProps<typeof AppLayoutAnatomy.root>

export const AppLayout = React.forwardRef<HTMLDivElement, AppLayoutProps>((props, ref) => {

    const {
        children,
        className,
        withSidebar = false,
        sidebarSize,
        ...rest
    } = props

    const ctx = React.useContext(__AppSidebarContext)

    return (
        <div
            ref={ref}
            className={cn(
                AppLayoutAnatomy.root({ withSidebar, sidebarSize: ctx.size || sidebarSize }),
                __isDesktop__ && "pt-4 select-none",
                className,
            )}
            {...rest}
        >
            {children}
        </div>
    )

})

AppLayout.displayName = "AppLayout"

/* -------------------------------------------------------------------------------------------------
 * AppLayoutHeader
 * -----------------------------------------------------------------------------------------------*/

export type AppLayoutHeaderProps = React.ComponentPropsWithRef<"header">

export const AppLayoutHeader = React.forwardRef<HTMLElement, AppLayoutHeaderProps>((props, ref) => {

    const {
        children,
        className,
        ...rest
    } = props

    return (
        <header
            ref={ref}
            className={cn(AppLayoutHeaderAnatomy.root(), className)}
            {...rest}
        >
            {children}
        </header>
    )

})

AppLayoutHeader.displayName = "AppLayoutHeader"

/* -------------------------------------------------------------------------------------------------
 * AppLayoutSidebar
 * -----------------------------------------------------------------------------------------------*/

export type AppLayoutSidebarProps = React.ComponentPropsWithRef<"aside">

export const AppLayoutSidebar = React.forwardRef<HTMLElement, AppLayoutSidebarProps>((props, ref) => {

    const {
        children,
        className,
        ...rest
    } = props

    return (
        <aside
            ref={ref}
            className={cn(AppLayoutSidebarAnatomy.root(), className)}
            {...rest}
        >
            {__isTauriDesktop__ && <TauriSidebarPaddingMacOS />}
            {__isElectronDesktop__ && <ElectronSidebarPaddingMacOS />}
            {children}
        </aside>
    )

})

AppLayoutSidebar.displayName = "AppLayoutSidebar"

/* -------------------------------------------------------------------------------------------------
 * AppLayoutContent
 * -----------------------------------------------------------------------------------------------*/

export type AppLayoutContentProps = React.ComponentPropsWithRef<"main">

export const AppLayoutContent = React.forwardRef<HTMLElement, AppLayoutContentProps>((props, ref) => {

    const {
        children,
        className,
        ...rest
    } = props

    return (
        <main
            ref={ref}
            className={cn(AppLayoutContentAnatomy.root(), className)}
            {...rest}
        >
            {children}
        </main>
    )

})

AppLayoutContent.displayName = "AppLayoutContent"

/* -------------------------------------------------------------------------------------------------
 * AppLayoutGrid
 * -----------------------------------------------------------------------------------------------*/

export type AppLayoutGridProps = React.ComponentPropsWithRef<"section"> &
    VariantProps<typeof AppLayoutGridAnatomy.root>

export const AppLayoutGrid = React.forwardRef<HTMLElement, AppLayoutGridProps>((props, ref) => {

    const {
        children,
        className,
        breakBelow,
        cols,
        spacing,
        ...rest
    } = props

    return (
        <section
            ref={ref}
            className={cn(AppLayoutGridAnatomy.root({ breakBelow, cols, spacing }), className)}
            {...rest}
        >
            {children}
        </section>
    )

})

AppLayoutGrid.displayName = "AppLayoutGrid"

/* -------------------------------------------------------------------------------------------------
 * AppLayoutFooter
 * -----------------------------------------------------------------------------------------------*/

export type AppLayoutFooterProps = React.ComponentPropsWithRef<"footer">

export const AppLayoutFooter = React.forwardRef<HTMLElement, AppLayoutFooterProps>((props, ref) => {

    const {
        children,
        className,
        ...rest
    } = props

    return (
        <footer
            ref={ref}
            className={cn(AppLayoutFooterAnatomy.root(), className)}
            {...rest}
        >
            {children}
        </footer>
    )

})

AppLayoutFooter.displayName = "AppLayoutFooter"

/* -------------------------------------------------------------------------------------------------
 * AppLayoutStack
 * -----------------------------------------------------------------------------------------------*/

export type AppLayoutStackProps = React.ComponentPropsWithRef<"div"> &
    VariantProps<typeof AppLayoutStackAnatomy.root>

export const AppLayoutStack = React.forwardRef<HTMLDivElement, AppLayoutStackProps>((props, ref) => {

    const {
        children,
        className,
        spacing,
        ...rest
    } = props

    return (
        <div
            ref={ref}
            className={cn(AppLayoutStackAnatomy.root({ spacing }), className)}
            {...rest}
        >
            {children}
        </div>
    )

})

AppLayoutStack.displayName = "AppLayoutStack"

