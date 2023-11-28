import React from "react"
import { cn, ComponentWithAnatomy, defineStyleAnatomy } from "../core"
import { cva } from "class-variance-authority"
import Link from "next/link"

/* -------------------------------------------------------------------------------------------------
 * Anatomy
 * -----------------------------------------------------------------------------------------------*/

export const NavigationTabsAnatomy = defineStyleAnatomy({
    nav: cva([
        "UI-NavigationTabs__nav",
        "flex w-full overflow-hidden overflow-x-auto"
    ]),
    tab: cva([
        "UI-NavigationTabs__tab",
        "group/navtabs inline-flex flex-none shrink-0 basis-auto items-center py-4 px-2 font-normal text-sm transition outline-none px-4 min-w-0 justify-center relative",
        "focus-visible:bg-[--highlight]",
        "text-[--muted]",
        "hover:text-[--text-color] hover:bg-[--highlight] rounded-md",
        "data-[selected=true]:border-[--brand] data-[selected=true]:font-semibold data-[selected=true]:text-white",
    ]),
    icon: cva([
        "UI-NavigationTabs__icon",
        "-ml-0.5 mr-2 h-5 w-5",
    ]),
})

/* -------------------------------------------------------------------------------------------------
 * NavigationTabs
 * -----------------------------------------------------------------------------------------------*/

export interface NavigationTabsProps extends React.ComponentPropsWithRef<"nav">, ComponentWithAnatomy<typeof NavigationTabsAnatomy> {
    items: {
        name: string,
        href: string | null | undefined,
        icon?: ((props: any) => JSX.Element) | null | undefined,
        isCurrent: boolean
        addon?: React.ReactNode
    }[]
}

export const NavigationTabs = React.forwardRef<HTMLElement, NavigationTabsProps>((props, ref) => {

    const {
        children,
        className,
        navClassName,
        tabClassName,
        iconClassName,
        items,
        ...rest
    } = props

    return (
        <nav
            ref={ref}
            className={cn(NavigationTabsAnatomy.nav(), navClassName, className)}
            {...rest}
        >
            {items.map((tab) => (
                <Link
                    key={tab.name}
                    href={tab.href ?? "#"}
                    className={cn(
                        NavigationTabsAnatomy.tab(),
                        tabClassName,
                    )}
                    aria-current={tab.isCurrent ? "page" : undefined}
                    data-selected={tab.isCurrent}
                >
                    {tab.icon && <tab.icon
                        className={cn(
                            NavigationTabsAnatomy.icon(),
                            iconClassName,
                        )}
                        aria-hidden="true"
                        data-selected={tab.isCurrent}
                    />}
                    <span>{tab.name}</span>
                    {tab.addon}
                </Link>
            ))}
        </nav>
    )

})

NavigationTabs.displayName = "NavigationTabs"
