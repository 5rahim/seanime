"use client"

import { SeaLink } from "@/components/shared/sea-link"
import * as NavigationMenuPrimitive from "@radix-ui/react-navigation-menu"
import { cva, VariantProps } from "class-variance-authority"
import * as React from "react"
import { cn, ComponentAnatomy, defineStyleAnatomy } from "../core/styling"
import { Drawer } from "../drawer"
import { VerticalMenu, VerticalMenuItem } from "../vertical-menu"


/* -------------------------------------------------------------------------------------------------
 * Anatomy
 * -----------------------------------------------------------------------------------------------*/

export const NavigationMenuAnatomy = defineStyleAnatomy({
    root: cva([
        "UI-NavigationMenu__root",
        "relative inline-block z-10 max-w-full",
    ]),
    item: cva([
        "UI-NavigationMenu__item",
        "relative group/navigationMenu_item inline-flex !text-[1.15rem] items-center h-full select-none rounded-[--radius] leading-none no-underline outline-none transition-colors",
        "text-[--muted] hover:text-[--foreground]",
        "data-[current=true]:text-white", // Selected
        "font-[600] leading-none",
        "focus-visible:ring-1 focus-visible:ring-inset focus-visible:ring-[--ring]",
    ], {
        variants: {
            size: {
                sm: "px-3 h-8 text-sm",
                md: "px-2 sm",
                lg: "px-3 h-12 text-base",
            },
        },
        defaultVariants: {
            size: "md",
        },
    }),
    icon: cva([
        "UI-VerticalNav__icon",
        "flex-shrink-0 mr-3",
        "text-[--muted] group-hover/navigationMenu_item:text-[--foreground] data-[current=true]:text-[--brand] data-[current=true]:group-hover/navigationMenu_item:text-[--brand]",
    ], {
        variants: {
            size: {
                sm: "size-4",
                md: "size-5",
                lg: "size-6",
            },
        },
        defaultVariants: {
            size: "md",
        },
    }),
    itemChevron: cva([
        "UI-VerticalNav__itemChevron",
        "ml-2 w-4 h-4 transition-transform duration-200 group-hover/navigationMenu_item:rotate-180",
    ]),
    desktopList: cva([
        "UI-VerticalNav__desktopList",
        "inline-block space-x-2",
    ], {
        variants: {
            switchToDrawerBelow: {
                sm: "hidden sm:flex",
                md: "hidden md:flex",
                lg: "hidden lg:flex",
                never: "flex",
            },
        },
        defaultVariants: {
            switchToDrawerBelow: "md",
        },
    }),
    mobileTrigger: cva([
        "UI-VerticalNav__mobileTrigger",
        "items-center justify-center rounded-[--radius] p-2 text-[--muted] hover:bg-[--subtle] hover:text-[--foreground]",
        "focus:outline-none focus:ring-2 focus:ring-inset focus:ring-[--ring]",
    ], {
        variants: {
            switchToDrawerBelow: {
                sm: "inline-flex sm:hidden",
                md: "inline-flex md:hidden",
                lg: "inline-flex lg:hidden",
                never: "hidden",
            },
        },
        defaultVariants: {
            switchToDrawerBelow: "md",
        },
    }),
    menuContainer: cva([
        "UI-NavigationMenu__menuContainer",
        "absolute left-0 top-0 p-1 data-[motion^=from-]:animate-in data-[motion^=to-]:animate-out",
        "data-[motion^=from-]:fade-in data-[motion^=to-]:fade-out data-[motion=from-end]:slide-in-from-right-52",
        "data-[motion=from-start]:slide-in-from-left-52 data-[motion=to-end]:slide-out-to-right-52",
        "data-[motion=to-start]:slide-out-to-left-52 w-full sm:min-w-full",
    ]),
    viewport: cva([
        "UI-NavigationMenu__viewport",
        "relative mt-1.5 duration-300 h-[var(--radix-navigation-menu-viewport-height)]",
        "w-full min-w-96 rounded-[--radius] shadow-sm border bg-[--paper] text-[--foreground]",
        "data-[state=open]:animate-in data-[state=open]:zoom-in-90 data-[state=open]:fade-in-25",
        "data-[state=closed]:animate-out data-[state=closed]:zoom-out-100 data-[state=closed]:fade-out-0",
    ]),
})

/* -------------------------------------------------------------------------------------------------
 * NavigationMenu
 * -----------------------------------------------------------------------------------------------*/

export type NavigationMenuProps = ComponentAnatomy<typeof NavigationMenuAnatomy> &
    React.ComponentPropsWithoutRef<typeof NavigationMenuPrimitive.Root> &
    VariantProps<typeof NavigationMenuAnatomy.desktopList> &
    VariantProps<typeof NavigationMenuAnatomy.item> & {
    children?: React.ReactNode
    items: VerticalMenuItem[],
    /**
     * Add content to the mobile drawer. The content is appended above the menu
     */
    mobileDrawerHeader?: React.ReactNode
    /**
     * Add content to the mobile drawer. The content is appended below the menu
     */
    mobileDrawerContent?: React.ReactNode
    /**
     * Additional props passed to the mobile drawer
     */
    mobileDrawerProps?: Partial<React.ComponentPropsWithoutRef<typeof Drawer>>
}

export const NavigationMenu = React.forwardRef<HTMLDivElement, NavigationMenuProps>((props, ref) => {

    const {
        children,
        iconClass,
        itemClass,
        desktopListClass,
        itemChevronClass,
        mobileTriggerClass,
        menuContainerClass,
        viewportClass,
        className,
        switchToDrawerBelow,
        mobileDrawerHeader,
        mobileDrawerContent,
        mobileDrawerProps,
        items,
        size,
        ...rest
    } = props

    const [mobileOpen, setMobileOpen] = React.useState(false)

    const Icon = React.useCallback(({ item }: { item: NavigationMenuProps["items"][number] }) => item.iconType ? <item.iconType
        className={cn(
            NavigationMenuAnatomy.icon({ size }),
            iconClass,
        )}
        aria-hidden="true"
        data-current={item.isCurrent}
    /> : null, [iconClass, size])

    return (
        <NavigationMenuPrimitive.Root
            ref={ref}
            className={cn(
                NavigationMenuAnatomy.root(),
                className,
            )}
            {...rest}
        >
            {/*Mobile*/}
            <button
                className={cn(
                    NavigationMenuAnatomy.mobileTrigger({
                        switchToDrawerBelow,
                    }),
                    mobileTriggerClass,
                )}
                onClick={() => setMobileOpen(s => !s)}
            >
                <span className="sr-only">Open main menu</span>
                {mobileOpen ? (
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
            <Drawer
                open={mobileOpen}
                onOpenChange={open => setMobileOpen(open)}
                side="left"
                {...mobileDrawerProps}
            >
                {mobileDrawerHeader}
                <VerticalMenu
                    items={items}
                    className="mt-2"
                    onLinkItemClick={() => setMobileOpen(false)} // Close the drawer when a link item is clicked
                />
                {mobileDrawerContent}
            </Drawer>

            {/*Desktop*/}
            <NavigationMenuPrimitive.List
                className={cn(
                    NavigationMenuAnatomy.desktopList({
                        switchToDrawerBelow,
                    }),
                    desktopListClass,
                )}
            >
                {items.map(item => {

                    if (item.subContent) {
                        return (
                            <NavigationMenuPrimitive.Item key={item.name}>
                                <NavigationMenuPrimitive.Trigger
                                    className={cn(
                                        NavigationMenuAnatomy.item({ size }),
                                        itemClass,
                                    )}
                                    data-current={item.isCurrent}
                                >
                                    <Icon item={item} />
                                    <span className="flex-none">{item.name}</span>
                                    <svg
                                        xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24"
                                        fill="none"
                                        stroke="currentColor"
                                        strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"
                                        className={cn(NavigationMenuAnatomy.itemChevron(), itemChevronClass)}
                                        data-open={`${mobileOpen}`}
                                    >
                                        <polyline points="6 9 12 15 18 9" />
                                    </svg>
                                </NavigationMenuPrimitive.Trigger>
                                <NavigationMenuPrimitive.Content
                                    ref={ref}
                                    className={cn(
                                        NavigationMenuAnatomy.menuContainer(),
                                        menuContainerClass,
                                    )}
                                >
                                    <div className="w-full">
                                        {item.subContent && item.subContent}
                                    </div>
                                </NavigationMenuPrimitive.Content>
                            </NavigationMenuPrimitive.Item>
                        )
                    } else {
                        return (
                            <NavigationMenuPrimitive.Item key={item.name}>
                                <NavigationMenuPrimitive.NavigationMenuLink asChild>
                                    {item.href ? (
                                        <SeaLink
                                            href={item.href}
                                            className={cn(
                                                NavigationMenuAnatomy.item({ size }),
                                                itemClass,
                                            )}
                                            data-current={item.isCurrent}
                                        >
                                            <Icon item={item} />
                                            <span className="flex-none">{item.name}</span>
                                            {item.addon}
                                        </SeaLink>
                                    ) : (
                                        <button
                                            className={cn(
                                                NavigationMenuAnatomy.item({ size }),
                                                itemClass,
                                            )}
                                            data-current={item.isCurrent}
                                        >
                                            <Icon item={item} />
                                            <span className="flex-none">{item.name}</span>
                                            {item.addon}
                                        </button>
                                    )}
                                </NavigationMenuPrimitive.NavigationMenuLink>
                            </NavigationMenuPrimitive.Item>
                        )
                    }

                })}
            </NavigationMenuPrimitive.List>
            <div className={cn("perspective-[2000px] absolute left-0 top-full w-full flex justify-center")}>
                <NavigationMenuPrimitive.Viewport
                    className={cn(
                        NavigationMenuAnatomy.viewport(),
                        viewportClass,
                    )}
                />
            </div>
        </NavigationMenuPrimitive.Root>
    )

})

NavigationMenu.displayName = "NavigationMenu"
