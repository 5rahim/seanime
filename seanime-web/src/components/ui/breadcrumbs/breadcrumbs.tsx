"use client"

import React from "react"
import { cn, ComponentWithAnatomy, defineStyleAnatomy } from "../core"
import { cva } from "class-variance-authority"

/* -------------------------------------------------------------------------------------------------
 * Anatomy
 * -----------------------------------------------------------------------------------------------*/

export const BreadcrumbsAnatomy = defineStyleAnatomy({
    container: cva([
        "UI-Breadcrumbs__container",
        "flex"
    ]),
    list: cva([
        "UI-Breadcrumbs__list",
        "flex items-center space-x-2"
    ]),
    chevronIcon: cva([
        "UI-Breadcrumbs__chevronIcon",
        "h-5 w-5 flex-shrink-0 text-gray-400 mr-4"
    ]),
    item: cva([
        "UI-Breadcrumbs__item",
        "flex items-center",
    ]),
    itemLink: cva([
        "UI-Breadcrumbs__itemLink",
        "text-sm font-medium text-[--muted] hover:text-[--text-color]",
        "data-[selected=true]:pointer-events-none data-[selected=true]:font-semibold data-[selected=true]:text-[--text-color]" // Selected
    ]),
    homeItem: cva([
        "UI-Breadcrumbs__homeItem",
        "text-[--muted] hover:text-[--text-color]"
    ]),
    homeIcon: cva([
        "UI-Breadcrumbs__homeIcon",
        "h-5 w-5 flex-shrink-0"
    ]),
})

/* -------------------------------------------------------------------------------------------------
 * Breadcrumbs
 * -----------------------------------------------------------------------------------------------*/

export interface BreadcrumbsProps extends React.ComponentPropsWithRef<"nav">, ComponentWithAnatomy<typeof BreadcrumbsAnatomy> {
    homeHref?: string
    items: { name: string, href: string | null | undefined, isCurrent: boolean }[]
    showHomeButton?: boolean
    homeIcon?: React.ReactElement
}

export const Breadcrumbs: React.FC<BreadcrumbsProps> = React.forwardRef<HTMLElement, BreadcrumbsProps>((props, ref) => {

    const {
        children,
        containerClassName,
        listClassName,
        itemClassName,
        itemLinkClassName,
        chevronIconClassName,
        homeIconClassName,
        homeItemClassName,
        className,
        items,
        homeHref = "/",
        showHomeButton = true,
        homeIcon,
        ...rest
    } = props

    return (
        <div
        >
            <nav
                className={cn(BreadcrumbsAnatomy.container(), containerClassName, className)}
                {...rest}
                ref={ref}
            >
                <ol role="list" className={cn(BreadcrumbsAnatomy.list(), listClassName)}>
                    {showHomeButton &&
                        <li>
                            <div>
                                <a
                                    href={homeHref}
                                    className={cn(BreadcrumbsAnatomy.homeItem(), homeItemClassName)}
                                >
                                    {homeIcon ? homeIcon :
                                        <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24"
                                             strokeWidth="2" stroke="currentColor"
                                             className={cn(BreadcrumbsAnatomy.homeIcon(), homeIconClassName)}>
                                            <path strokeLinecap="round" strokeLinejoin="round"
                                                  d="M2.25 12l8.954-8.955c.44-.439 1.152-.439 1.591 0L21.75 12M4.5 9.75v10.125c0 .621.504 1.125 1.125 1.125H9.75v-4.875c0-.621.504-1.125 1.125-1.125h2.25c.621 0 1.125.504 1.125 1.125V21h4.125c.621 0 1.125-.504 1.125-1.125V9.75M8.25 21h8.25"/>
                                        </svg>}
                                </a>
                            </div>
                        </li>
                    }
                    {items.map((page, idx) => (
                        <li key={page.name}>
                            <div className={cn(BreadcrumbsAnatomy.item(), itemClassName)}>
                                {(!showHomeButton && idx > 0 || showHomeButton) &&
                                    <svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24"
                                         fill="none"
                                         stroke="currentColor"
                                         strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"
                                         className={cn(BreadcrumbsAnatomy.chevronIcon(), chevronIconClassName)}
                                    >
                                        <polyline points="9 18 15 12 9 6"></polyline>
                                    </svg>
                                }
                                <a
                                    href={page.href ?? "#"}
                                    className={cn(BreadcrumbsAnatomy.itemLink(), itemLinkClassName)}
                                    data-selected={page.isCurrent}
                                    aria-current={page.isCurrent ? "page" : undefined}
                                >
                                    {page.name}
                                </a>
                            </div>
                        </li>
                    ))}
                </ol>
            </nav>
        </div>
    )

})

Breadcrumbs.displayName = "Breadcrumbs"
