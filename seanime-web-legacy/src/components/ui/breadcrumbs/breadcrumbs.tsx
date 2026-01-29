"use client"

import { SeaLink } from "@/components/shared/sea-link"
import { cva } from "class-variance-authority"
import * as React from "react"
import { cn, ComponentAnatomy, defineStyleAnatomy } from "../core/styling"

/* -------------------------------------------------------------------------------------------------
 * Anatomy
 * -----------------------------------------------------------------------------------------------*/

export const BreadcrumbsAnatomy = defineStyleAnatomy({
    root: cva([
        "UI-Breadcrumbs__root",
        "flex",
    ]),
    list: cva([
        "UI-Breadcrumbs__list",
        "flex items-center space-x-2",
    ]),
    chevronIcon: cva([
        "UI-Breadcrumbs__chevronIcon",
        "h-5 w-5 flex-shrink-0 text-gray-400 mr-4",
    ]),
    item: cva([
        "UI-Breadcrumbs__item",
        "flex items-center",
    ]),
    itemLink: cva([
        "UI-Breadcrumbs__itemLink",
        "text-sm font-medium text-[--muted] hover:text-[--foreground]",
        "data-[selected=true]:pointer-events-none data-[selected=true]:font-semibold data-[selected=true]:text-[--foreground]", // Selected
    ]),
    homeItem: cva([
        "UI-Breadcrumbs__homeItem",
        "text-[--muted] hover:text-[--foreground]",
    ]),
    homeIcon: cva([
        "UI-Breadcrumbs__homeIcon",
        "h-5 w-5 flex-shrink-0",
    ]),
})

/* -------------------------------------------------------------------------------------------------
 * Breadcrumbs
 * -----------------------------------------------------------------------------------------------*/

export type BreadcrumbsOption = { name: string, href: string | null | undefined, isCurrent: boolean }

export type BreadcrumbsProps = React.ComponentPropsWithRef<"nav"> &
    ComponentAnatomy<typeof BreadcrumbsAnatomy> & {
    rootHref?: string
    items: BreadcrumbsOption[]
    showHomeButton?: boolean
    homeIcon?: React.ReactElement
}

export const Breadcrumbs = React.forwardRef<HTMLElement, BreadcrumbsProps>((props, ref) => {

    const {
        children,
        listClass,
        itemClass,
        itemLinkClass,
        chevronIconClass,
        homeIconClass,
        homeItemClass,
        className,
        items,
        rootHref = "/",
        showHomeButton = true,
        homeIcon,
        ...rest
    } = props

    return (
        <nav
            className={cn(BreadcrumbsAnatomy.root(), className)}
            {...rest}
            ref={ref}
        >
            <ol role="list" className={cn(BreadcrumbsAnatomy.list(), listClass)}>
                {showHomeButton &&
                    <li>
                        <div>
                            <SeaLink
                                href={rootHref}
                                className={cn(BreadcrumbsAnatomy.homeItem(), homeItemClass)}
                            >
                                {homeIcon ? homeIcon :
                                    <svg
                                        xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24"
                                        strokeWidth="2" stroke="currentColor"
                                        className={cn(BreadcrumbsAnatomy.homeIcon(), homeIconClass)}
                                    >
                                        <path
                                            strokeLinecap="round" strokeLinejoin="round"
                                            d="M2.25 12l8.954-8.955c.44-.439 1.152-.439 1.591 0L21.75 12M4.5 9.75v10.125c0 .621.504 1.125 1.125 1.125H9.75v-4.875c0-.621.504-1.125 1.125-1.125h2.25c.621 0 1.125.504 1.125 1.125V21h4.125c.621 0 1.125-.504 1.125-1.125V9.75M8.25 21h8.25"
                                        />
                                    </svg>}
                            </SeaLink>
                        </div>
                    </li>
                }
                {items.map((page, idx) => (
                    <li key={page.name}>
                        <div className={cn(BreadcrumbsAnatomy.item(), itemClass)}>
                            {(!showHomeButton && idx > 0 || showHomeButton) &&
                                <svg
                                    xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24"
                                    fill="none"
                                    stroke="currentColor"
                                    strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"
                                    className={cn(BreadcrumbsAnatomy.chevronIcon(), chevronIconClass)}
                                >
                                    <polyline points="9 18 15 12 9 6"></polyline>
                                </svg>
                            }
                            <SeaLink
                                href={page.href ?? "#"}
                                className={cn(BreadcrumbsAnatomy.itemLink(), itemLinkClass)}
                                data-selected={page.isCurrent}
                                aria-current={page.isCurrent ? "page" : undefined}
                            >
                                {page.name}
                            </SeaLink>
                        </div>
                    </li>
                ))}
            </ol>
        </nav>
    )

})

Breadcrumbs.displayName = "Breadcrumbs"
