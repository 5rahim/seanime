"use client"

import React from "react"
import { cn, ComponentWithAnatomy, createPolymorphicComponent, defineStyleAnatomy } from "../core"
import { cva, VariantProps } from "class-variance-authority"

/* -------------------------------------------------------------------------------------------------
 * Anatomy
 * -----------------------------------------------------------------------------------------------*/

export const NavbarAnatomy = defineStyleAnatomy({
    nav: cva([
        "UI-Navbar__nav",
        "w-full h-16",
        "bg-[--paper] border-b border-[--border]"
    ]),
    container: cva([
        "UI-Navbar__container",
        "container max-w-7xl h-full",
    ], {
        variants: {
            fullWidth: {
                true: "max-w-full w-full",
                false: ""
            }
        },
        defaultVariants: {
            fullWidth: false
        }
    }),
})

export const NavbarLayoutAnatomy = defineStyleAnatomy({
    layout: cva([
        "UI-NavbarLayout__content",
        "flex h-full items-center"
    ], {
        variants: {
            spacing: {
                apart: "justify-between",
                around: "justify-around"
            }
        },
        defaultVariants: {
            spacing: "apart"
        }
    }),
})

export const NavbarNavigationAnatomy = defineStyleAnatomy({
    root: cva([
        "UI-NavbarNavigation__root",
        "flex h-full items-center gap-8 flex-row-reverse md:flex-row"
    ])
})

/* -------------------------------------------------------------------------------------------------
 * Navbar
 * -----------------------------------------------------------------------------------------------*/

export interface NavbarProps extends React.ComponentPropsWithRef<"nav">, ComponentWithAnatomy<typeof NavbarAnatomy>,
    VariantProps<typeof NavbarAnatomy.container> {
}

export const _Navbar = (props: NavbarProps) => {

    const {
        children,
        navClassName,
        containerClassName,
        className,
        fullWidth,
        ref,
        ...rest
    } = props

    return (
        <nav
            className={cn(NavbarAnatomy.nav(), navClassName, className)}
            {...rest}
            ref={ref}
        >
            <div
                className={cn(NavbarAnatomy.container({ fullWidth }), containerClassName)}
            >
                {children}
            </div>
        </nav>
    )

}

_Navbar.displayName = "Navbar"

/* -------------------------------------------------------------------------------------------------
 * Navbar.Layout
 * -----------------------------------------------------------------------------------------------*/

export interface NavbarLayoutProps extends React.ComponentPropsWithRef<"div">,
    ComponentWithAnatomy<typeof NavbarLayoutAnatomy>,
    VariantProps<typeof NavbarLayoutAnatomy.layout> {
}

export const NavbarLayout: React.FC<NavbarLayoutProps> = React.forwardRef<HTMLDivElement, NavbarLayoutProps>((props, ref) => {

    const {
        children,
        className,
        layoutClassName,
        spacing,
        ...rest
    } = props

    return (
        <div
            className={cn(NavbarLayoutAnatomy.layout({ spacing }), layoutClassName, className)}
            {...rest}
            ref={ref}
        >
            {children}
        </div>
    )

})

NavbarLayout.displayName = "NavbarLayout"

/* -------------------------------------------------------------------------------------------------
 * Navbar.Navigation
 * -----------------------------------------------------------------------------------------------*/

export interface NavbarNavigationProps extends React.ComponentPropsWithRef<"div">,
    ComponentWithAnatomy<typeof NavbarNavigationAnatomy>,
    VariantProps<typeof NavbarNavigationAnatomy.root> {
}

export const NavbarNavigation: React.FC<NavbarNavigationProps> = React.forwardRef<HTMLDivElement, NavbarNavigationProps>((props, ref) => {

    const {
        children,
        className,
        rootClassName,
        ...rest
    } = props

    return (
        <div
            className={cn(NavbarNavigationAnatomy.root(), rootClassName, className)}
            {...rest}
            ref={ref}
        >
            {children}
        </div>
    )

})

NavbarNavigation.displayName = "NavbarNavigation"

/* -------------------------------------------------------------------------------------------------
 * Component
 * -----------------------------------------------------------------------------------------------*/

_Navbar.Layout = NavbarLayout
_Navbar.Navigation = NavbarNavigation

export const Navbar = createPolymorphicComponent<"nav", NavbarProps, {
    Layout: typeof NavbarLayout
    Navigation: typeof NavbarNavigation
}>(_Navbar)

Navbar.displayName = "Navbar"
