import { cva, VariantProps } from "class-variance-authority"
import React from "react"
import { cn, ComponentWithAnatomy, createPolymorphicComponent, defineStyleAnatomy } from "../core"

/* -------------------------------------------------------------------------------------------------
 * Anatomy
 * -----------------------------------------------------------------------------------------------*/

export const AppLayoutAnatomy = defineStyleAnatomy({
    root: cva([
        "UI-AppLayout__root",
        "flex w-full min-h-screen group",
        "group-[.with-sidebar]:group-[.sidebar-slim]:md:pl-20",
        "group-[.with-sidebar]:group-[.sidebar-sm]:md:pl-48",
        "group-[.with-sidebar]:group-[.sidebar-md]:md:pl-64",
        "group-[.with-sidebar]:group-[.sidebar-lg]:md:pl-[20rem]",
        "group-[.with-sidebar]:group-[.sidebar-xl]:md:pl-[25rem]",
    ], {
        variants: {
            withSidebar: {
                true: "flex-row with-sidebar",
                false: "flex-col"
            },
            sidebarSize: {
                slim: "sidebar-slim",
                sm: "sidebar-sm",
                md: "sidebar-md",
                lg: "sidebar-lg",
                xl: "sidebar-xl",
            }
        },
        defaultVariants: {
            withSidebar: false,
            sidebarSize: "md"
        }
    })
})

export const AppLayoutHeaderAnatomy = defineStyleAnatomy({
    root: cva([
        "UI-AppLayoutHeader__root",
        "block w-full"
    ])
})

export const AppLayoutSidebarAnatomy = defineStyleAnatomy({
    root: cva([
        "UI-AppLayoutSidebar__root",
        "hidden md:fixed md:inset-y-0 md:flex md:flex-col grow-0 shrink-0 basis-0 z-[50]",
        "group-[.sidebar-slim]:md:w-20",
        "group-[.sidebar-sm]:md:w-48",
        "group-[.sidebar-md]:md:w-64",
        "group-[.sidebar-lg]:md:w-[20rem]",
        "group-[.sidebar-xl]:md:w-[25rem]",
    ])
})

export const AppLayoutContentAnatomy = defineStyleAnatomy({
    root: cva([
        "UI-AppLayoutContent__root",
    ])
})

export const AppLayoutFooterAnatomy = defineStyleAnatomy({
    root: cva([
        "UI-AppLayoutFooter__root",
    ])
})

export const AppLayoutStackAnatomy = defineStyleAnatomy({
    root: cva([
        "UI-AppLayoutStack__root",
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
        }
    })
})

export const AppLayoutGridAnatomy = defineStyleAnatomy({
    root: cva([
        "UI-AppLayoutGrid__root",
        "block"
    ], {
        variants: {
            breakBelow: {
                sm: "sm:grid sm:space-y-0",
                md: "md:grid md:space-y-0",
                lg: "lg:grid lg:space-y-0",
                xl: "xl:grid xl:space-y-0",
            },
            spacing: {
                sm: "space-y-2 gap-2",
                md: "space-y-4 gap-4",
                lg: "space-y-8 gap-8",
                xl: "space-y-10 gap-10",
            },
            cols: {
                1: "grid-cols-1",
                2: "grid-cols-2",
                3: "grid-cols-3",
                4: "grid-cols-4",
                5: "grid-cols-5",
                6: "grid-cols-6",
            }
        },
        defaultVariants: {
            breakBelow: "xl",
            spacing: "md",
            cols: 3
        }
    })
})

/* -------------------------------------------------------------------------------------------------
 * AppLayout
 * -----------------------------------------------------------------------------------------------*/

export interface AppLayoutProps extends React.ComponentPropsWithRef<"div">,
    ComponentWithAnatomy<typeof AppLayoutAnatomy>,
    VariantProps<typeof AppLayoutAnatomy.root> {
}

const _AppLayout = (props: AppLayoutProps) => {

    const {
        children,
        rootClassName,
        className,
        ref,
        withSidebar,
        sidebarSize,
        ...rest
    } = props

    return (
        <div
            className={cn(AppLayoutAnatomy.root({ withSidebar, sidebarSize }), rootClassName, className)}
            {...rest}
            ref={ref}
        >
            {children}
        </div>
    )

}

_AppLayout.displayName = "AppLayout"

/* -------------------------------------------------------------------------------------------------
 * AppLayout.Header
 * -----------------------------------------------------------------------------------------------*/

export interface AppLayoutHeaderProps extends React.ComponentPropsWithRef<"header">, ComponentWithAnatomy<typeof AppLayoutHeaderAnatomy> {
}

export const AppLayoutHeader: React.FC<AppLayoutHeaderProps> = React.forwardRef<HTMLElement, AppLayoutHeaderProps>((props, ref) => {

    const {
        children,
        rootClassName,
        className,
        ...rest
    } = props

    return (
        <header
            className={cn(AppLayoutHeaderAnatomy.root(), rootClassName, className)}
            {...rest}
            ref={ref}
        >
            {children}
        </header>
    )

})

AppLayoutHeader.displayName = "AppLayoutHeader"

/* -------------------------------------------------------------------------------------------------
 * AppLayout.Sidebar
 * -----------------------------------------------------------------------------------------------*/

export interface AppLayoutSidebarProps extends React.ComponentPropsWithRef<"aside">, ComponentWithAnatomy<typeof AppLayoutSidebarAnatomy> {
}

export const AppLayoutSidebar: React.FC<AppLayoutSidebarProps> = React.forwardRef<HTMLElement, AppLayoutSidebarProps>((props, ref) => {

    const {
        children,
        rootClassName,
        className,
        ...rest
    } = props

    return (
        <aside
            className={cn(AppLayoutSidebarAnatomy.root(), rootClassName, className)}
            {...rest}
            ref={ref}
        >
            {children}
        </aside>
    )

})

AppLayoutSidebar.displayName = "AppLayoutSidebar"

/* -------------------------------------------------------------------------------------------------
 * AppLayout.Content
 * -----------------------------------------------------------------------------------------------*/

export interface AppLayoutContentProps extends React.ComponentPropsWithRef<"main">, ComponentWithAnatomy<typeof AppLayoutContentAnatomy> {
}

export const AppLayoutContent: React.FC<AppLayoutContentProps> = React.forwardRef<HTMLElement, AppLayoutContentProps>((props, ref) => {

    const {
        children,
        rootClassName,
        className,
        ...rest
    } = props

    return (
        <main
            className={cn(AppLayoutContentAnatomy.root(), rootClassName, className)}
            {...rest}
            ref={ref}
        >
            {children}
        </main>
    )

})

AppLayoutContent.displayName = "AppLayoutContent"

/* -------------------------------------------------------------------------------------------------
 * AppLayout.Grid
 * -----------------------------------------------------------------------------------------------*/

export interface AppLayoutGridProps extends React.ComponentPropsWithRef<"section">,
    ComponentWithAnatomy<typeof AppLayoutGridAnatomy>,
    VariantProps<typeof AppLayoutGridAnatomy.root> {
}

export const AppLayoutGrid: React.FC<AppLayoutGridProps> = React.forwardRef<HTMLElement, AppLayoutGridProps>((props, ref) => {

    const {
        children,
        rootClassName,
        className,
        breakBelow,
        cols,
        spacing,
        ...rest
    } = props

    return (
        <section
            className={cn(AppLayoutGridAnatomy.root({ breakBelow, cols, spacing }), rootClassName, className)}
            {...rest}
            ref={ref}
        >
            {children}
        </section>
    )

})

AppLayoutGrid.displayName = "AppLayoutGrid"

/* -------------------------------------------------------------------------------------------------
 * AppLayout.Footer
 * -----------------------------------------------------------------------------------------------*/

export interface AppLayoutFooterProps extends React.ComponentPropsWithRef<"footer">, ComponentWithAnatomy<typeof AppLayoutFooterAnatomy> {
}

export const AppLayoutFooter: React.FC<AppLayoutFooterProps> = React.forwardRef<HTMLElement, AppLayoutFooterProps>((props, ref) => {

    const {
        children,
        rootClassName,
        className,
        ...rest
    } = props

    return (
        <footer
            className={cn(AppLayoutFooterAnatomy.root(), rootClassName, className)}
            {...rest}
            ref={ref}
        >
            {children}
        </footer>
    )

})

AppLayoutFooter.displayName = "AppLayoutFooter"

/* -------------------------------------------------------------------------------------------------
 * AppLayout.Stack
 * -----------------------------------------------------------------------------------------------*/

export interface AppLayoutStackProps extends React.ComponentPropsWithRef<"div">,
    ComponentWithAnatomy<typeof AppLayoutStackAnatomy>,
    VariantProps<typeof AppLayoutStackAnatomy.root> {
}

export const AppLayoutStack: React.FC<AppLayoutStackProps> = React.forwardRef<HTMLDivElement, AppLayoutStackProps>((props, ref) => {

    const {
        children,
        rootClassName,
        className,
        spacing,
        ...rest
    } = props

    return (
        <div
            className={cn(AppLayoutStackAnatomy.root({ spacing }), rootClassName, className)}
            {...rest}
            ref={ref}
        >
            {children}
        </div>
    )

})

AppLayoutStack.displayName = "AppLayoutStack"

/* -------------------------------------------------------------------------------------------------
 * Component
 * -----------------------------------------------------------------------------------------------*/

_AppLayout.Header = AppLayoutHeader
_AppLayout.Sidebar = AppLayoutSidebar
_AppLayout.Content = AppLayoutContent
_AppLayout.Footer = AppLayoutFooter
_AppLayout.Grid = AppLayoutGrid
_AppLayout.Stack = AppLayoutStack

export const AppLayout = createPolymorphicComponent<"div", AppLayoutProps, {
    Header: typeof AppLayoutHeader
    Sidebar: typeof AppLayoutSidebar
    Content: typeof AppLayoutContent
    Footer: typeof AppLayoutFooter
    Grid: typeof AppLayoutGrid
    Stack: typeof AppLayoutStack
}>(_AppLayout)

AppLayout.displayName = "AppLayout"
