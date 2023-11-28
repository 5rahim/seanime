"use client"

import React, { useState } from "react"
import { cn, ComponentWithAnatomy, defineStyleAnatomy } from "../core"
import { cva, VariantProps } from "class-variance-authority"
import type { NavigationMenuProps as NavigationMenuPrimitiveProps } from "@radix-ui/react-navigation-menu"
import * as NavigationMenuPrimitive from "@radix-ui/react-navigation-menu"
import { Drawer } from "../modal"
import { VerticalNav } from "../vertical-nav"


/* -------------------------------------------------------------------------------------------------
 * Anatomy
 * -----------------------------------------------------------------------------------------------*/

export const HorizontalNavAnatomy = defineStyleAnatomy({
    nav: cva([
        "UI-HorizontalNav__nav",
        "relative inline-block z-10 max-w-full"
    ]),
    item: cva([
        "UI-HorizontalNav__item",
        "group/item inline-flex items-center h-full select-none rounded-[--radius] p-3 leading-none no-underline outline-none transition-colors",
        "text-[--muted] hover:bg-[--highlight] hover:text-[--text-color] focus:bg-[--highlight]",
        "data-[selected=true]:text-[--brand]", // Selected
        "text-sm font-[600] leading-none"
    ]),
    icon: cva([
        "UI-VerticalNav__icon",
        "flex-shrink-0 -ml-1 mr-3 h-6 w-6",
        "text-[--muted] group-hover/item:text-[--text-color] data-[selected=true]:text-[--brand] data-[selected=true]:group-hover/item:text-[--brand]"
    ]),
    parentItemChevron: cva([
        "UI-VerticalNav__parentItemChevron",
        "ml-2 w-4 h-4 transition-transform duration-200 group-hover/item:rotate-180",
    ]),
    desktopList: cva([
        "UI-VerticalNav__desktopList",
        "inline-block space-x-1"
    ], {
        variants: {
            switchToDrawerBelow: {
                sm: "hidden sm:flex",
                md: "hidden md:flex",
                lg: "hidden lg:flex",
                never: "flex"
            }
        },
        defaultVariants: {
            switchToDrawerBelow: "md"
        }
    }),
    mobileTrigger: cva([
        "UI-VerticalNav__mobileTrigger",
        "items-center justify-center rounded-[--radius] p-2 text-[--muted] hover:bg-[--highlight] hover:text-[--text-color]",
        "focus:outline-none focus:ring-2 focus:ring-inset focus:ring-[--ring]"
    ], {
        variants: {
            switchToDrawerBelow: {
                sm: "inline-flex sm:hidden",
                md: "inline-flex md:hidden",
                lg: "inline-flex lg:hidden",
                never: "hidden"
            }
        },
        defaultVariants: {
            switchToDrawerBelow: "md"
        }
    }),
    menuContainer: cva([
        "UI-HorizontalNav__menuContainer",
        "left-0 top-0 overflow-hidden p-2 data-[motion^=from-]:animate-in data-[motion^=to-]:animate-out",
        "data-[motion^=from-]:fade-in data-[motion^=to-]:fade-out data-[motion=from-end]:slide-in-from-right-52",
        "data-[motion=from-start]:slide-in-from-left-52 data-[motion=to-end]:slide-out-to-right-52",
        "data-[motion=to-start]:slide-out-to-left-52 md:absolute md:w-full",
    ]),
    viewport: cva([
        "UI-HorizontalNav__viewport",
        "relative mt-1.5 h-[var(--radix-navigation-menu-viewport-height)]",
        "w-full overflow-hidden rounded-[--radius] shadow-lg border border-[--border] bg-[--paper] text-[--text-color]",
        "data-[state=open]:animate-in data-[state=closed]:animate-out data-[state=closed]:zoom-out-95",
        "data-[state=open]:zoom-in-90",
    ]),
})

/* -------------------------------------------------------------------------------------------------
 * HorizontalNav
 * -----------------------------------------------------------------------------------------------*/

export interface HorizontalNavProps extends ComponentWithAnatomy<typeof HorizontalNavAnatomy>,
    NavigationMenuPrimitiveProps,
    VariantProps<typeof HorizontalNavAnatomy.desktopList> {
    children?: React.ReactNode
    items: {
        name: string,
        href?: string | null | undefined,
        icon?: ((props: any) => JSX.Element) | null | undefined,
        isCurrent?: boolean,
        addon?: React.ReactNode
        content?: React.ReactNode
    }[],
    // Add components to the mobile drawer. The content is appended below the menu
    drawerContent?: React.ReactNode
}

export const HorizontalNav = React.forwardRef<HTMLDivElement, HorizontalNavProps>((props, ref) => {

    const {
        children,
        navClassName,
        iconClassName,
        itemClassName,
        desktopListClassName,
        parentItemChevronClassName,
        mobileTriggerClassName,
        menuContainerClassName,
        viewportClassName,
        className,
        switchToDrawerBelow,
        drawerContent,
        items,
        ...rest
    } = props

    const [mobileOpen, setMobileOpen] = useState(false)

    const Icon = ({ item }: { item: HorizontalNavProps["items"][number] }) => item.icon ? <item.icon
        className={cn(
            HorizontalNavAnatomy.icon(),
            iconClassName,
        )}
        aria-hidden="true"
        data-selected={item.isCurrent}
    /> : null

    return (
        <NavigationMenuPrimitive.Root
            ref={ref}
            className={cn(
                HorizontalNavAnatomy.nav(),
                navClassName,
                className
            )}
            {...rest}
        >
            {/*Mobile*/}
            <button
                className={cn(
                    HorizontalNavAnatomy.mobileTrigger({
                        switchToDrawerBelow
                    }),
                    mobileTriggerClassName,
                )}
                onClick={() => setMobileOpen(s => !s)}
            >
                <span className="sr-only">Open main menu</span>
                {mobileOpen ? (
                    <svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none"
                         stroke="currentColor"
                         strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" className="block h-6 w-6">
                        <line x1="18" x2="6" y1="6" y2="18"></line>
                        <line x1="6" x2="18" y1="6" y2="18"></line>
                    </svg>
                ) : (
                    <svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none"
                         stroke="currentColor"
                         strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" className="block h-6 w-6">
                        <line x1="4" x2="20" y1="12" y2="12"></line>
                        <line x1="4" x2="20" y1="6" y2="6"></line>
                        <line x1="4" x2="20" y1="18" y2="18"></line>
                    </svg>
                )}
            </button>
            <Drawer isOpen={mobileOpen} onClose={() => setMobileOpen(false)} placement="left" isClosable>
                <VerticalNav items={items} className="mt-2"/>
                {drawerContent}
            </Drawer>

            {/*Desktop*/}
            <NavigationMenuPrimitive.List
                className={cn(
                    HorizontalNavAnatomy.desktopList({
                        switchToDrawerBelow
                    }),
                    desktopListClassName
                )}
            >
                {items.map(item => {

                    if (item.content) {
                        return (
                            <NavigationMenuPrimitive.Item key={item.name}>
                                <NavigationMenuPrimitive.Trigger
                                    className={cn(
                                        HorizontalNavAnatomy.item(),
                                        itemClassName
                                    )}
                                    data-selected={item.isCurrent}
                                >
                                    <Icon item={item}/>
                                    <span className="flex-none">{item.name}</span>
                                    <svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24"
                                         fill="none"
                                         stroke="currentColor"
                                         strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"
                                         className={cn(HorizontalNavAnatomy.parentItemChevron(), parentItemChevronClassName)}
                                         data-open={`${open}`}
                                    >
                                        <polyline points="6 9 12 15 18 9"/>
                                    </svg>
                                </NavigationMenuPrimitive.Trigger>
                                <NavigationMenuPrimitive.Content
                                    ref={ref}
                                    className={cn(
                                        HorizontalNavAnatomy.menuContainer(),
                                        menuContainerClassName
                                    )}
                                >
                                    <div className={"w-full"}>
                                        {item.content && item.content}
                                    </div>
                                </NavigationMenuPrimitive.Content>
                            </NavigationMenuPrimitive.Item>
                        )
                    } else {
                        return (
                            <NavigationMenuPrimitive.Item key={item.name}>
                                <NavigationMenuPrimitive.NavigationMenuLink asChild>
                                    <a
                                        href={item.href ?? "#"}
                                        className={cn(
                                            HorizontalNavAnatomy.item(),
                                            itemClassName
                                        )}
                                        data-selected={item.isCurrent}
                                    >
                                        <Icon item={item}/>
                                        <span className="flex-none">{item.name}</span>
                                        {item.addon}
                                    </a>
                                </NavigationMenuPrimitive.NavigationMenuLink>
                            </NavigationMenuPrimitive.Item>
                        )
                    }

                })}
            </NavigationMenuPrimitive.List>
            <div className={cn("absolute left-0 top-full w-full flex justify-center")}>
                <NavigationMenuPrimitive.Viewport
                    className={cn(
                        HorizontalNavAnatomy.viewport(),
                        viewportClassName
                    )}
                />
            </div>
        </NavigationMenuPrimitive.Root>
    )

})

HorizontalNav.displayName = "HorizontalNav"
