"use client"

import React, { Fragment } from "react"
import { cn, ComponentWithAnatomy, defineStyleAnatomy } from "../core"
import { cva } from "class-variance-authority"
import { Disclosure } from "@headlessui/react"
import Link from "next/link"
import { Tooltip } from "@/components/ui/tooltip"

/* -------------------------------------------------------------------------------------------------
 * Anatomy
 * -----------------------------------------------------------------------------------------------*/

export const VerticalNavAnatomy = defineStyleAnatomy({
    nav: cva([
        "UI-VerticalNav__nav",
        "block space-y-1",
    ]),
    item: cva([
        "UI-VerticalNav__tab",
        "group/vnav flex flex-none truncate items-center px-4 py-2 text-sm font-[600] rounded-[--radius] transition cursor-pointer",
        "hover:bg-[--highlight] hover:text-[--text-color]",
        "focus-visible:ring-2 ring-[--ring] outline-none",
        "text-[--muted]",
        "data-[selected=true]:bg-[--highlight]",
    ]),
    parentItem: cva([
        "UI-VerticalNav__parentItem",
        "cursor-pointer",
    ]),
    parentItemChevron: cva([
        "UI-VerticalNav__parentItemChevron",
        "w-5 h-5 transition-transform data-[open=true]:rotate-90",
    ]),
    icon: cva([
        "UI-VerticalNav__icon",
        "flex-shrink-0 -ml-1 mr-3 h-6 w-6",
        "text-[--muted]",
        "group-hover/vnav:text-[--text-color] data-[selected=true]:text-white data-[selected=true]:group-hover/vnav:text-white",
    ]),
    subList: cva([
        "UI-VerticalNav__subList",
        "pl-2",
    ]),
})

/* -------------------------------------------------------------------------------------------------
 * VerticalNav
 * -----------------------------------------------------------------------------------------------*/

export interface VerticalNavProps extends React.ComponentPropsWithRef<"div">, ComponentWithAnatomy<typeof VerticalNavAnatomy> {
    children?: React.ReactNode
    items: {
        name: string
        href?: string | null | undefined
        icon?: ((props: any) => JSX.Element) | null | undefined
        isCurrent?: boolean
        onClick?: React.MouseEventHandler<HTMLElement>
        addon?: React.ReactNode
        content?: React.ReactNode
    }[]
}

export const VerticalNav = React.forwardRef<HTMLDivElement, VerticalNavProps>((props, ref) => {

    const {
        children,
        navClassName,
        itemClassName,
        iconClassName,
        parentItemClassName,
        subListClassName,
        parentItemChevronClassName,
        className,
        items,
        ...rest
    } = props

    return (
        <nav
            ref={ref}
            className={cn(VerticalNavAnatomy.nav(), navClassName, className)}
            {...rest}
        >
            {items.map((item, idx) => !item.content ? (
                <Tooltip
                    side={"right"}
                    sideOffset={4}
                    align={"start"}
                    key={item.name}
                    trigger={<Link
                        href={item.href ?? "#"}
                        className={cn(
                            VerticalNavAnatomy.item(),
                            itemClassName,
                        )}
                        aria-current={item.isCurrent ? "page" : undefined}
                        data-selected={item.isCurrent}
                        onClick={item.onClick}
                    >
                        {item.icon && <item.icon
                            className={cn(
                                VerticalNavAnatomy.icon(),
                                iconClassName,
                            )}
                            aria-hidden="true"
                            data-selected={item.isCurrent}
                        />}
                        <span>{item.name}</span>
                        {item.addon}
                    </Link>}>
                    {item.name}
                </Tooltip>
            ) : (
                <Disclosure as={Fragment} key={item.name}>
                    {({ open }) => (
                        <>
                            <Disclosure.Button
                                as="div"
                                key={item.name}
                                tabIndex={idx}
                                className={cn(
                                    VerticalNavAnatomy.item(),
                                    VerticalNavAnatomy.parentItem(),
                                    itemClassName,
                                    parentItemClassName,
                                )}
                                aria-current={item.isCurrent ? "page" : undefined}
                                data-selected={item.isCurrent}
                                onClick={item.onClick}
                            >
                                <div className="w-full flex items-center">
                                    {item.icon && <item.icon
                                        className={cn(
                                            VerticalNavAnatomy.icon(),
                                            iconClassName,
                                        )}
                                        aria-hidden="true"
                                        data-selected={item.isCurrent}
                                    />}
                                    <span>{item.name}</span>
                                </div>
                                <svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24"
                                     fill="none" stroke="currentColor"
                                     strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"
                                     className={cn(VerticalNavAnatomy.parentItemChevron(), parentItemChevronClassName)}
                                     data-open={`${open}`}
                                >
                                    <polyline points="9 18 15 12 9 6"></polyline>
                                </svg>
                            </Disclosure.Button>
                            <Disclosure.Panel
                                className={cn(VerticalNavAnatomy.subList(), subListClassName)}
                            >
                                {item.content && item.content}
                            </Disclosure.Panel>
                        </>
                    )}
                </Disclosure>
            ))}
        </nav>
    )

})

VerticalNav.displayName = "VerticalNav"
