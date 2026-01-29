"use client"

import { cn, ComponentAnatomy, defineStyleAnatomy } from "../core/styling"
import * as React from "react"
import { cva } from "class-variance-authority"

/* -------------------------------------------------------------------------------------------------
 * Anatomy
 * -----------------------------------------------------------------------------------------------*/

export const PaginationAnatomy = defineStyleAnatomy({
    root: cva([
        "UI-Pagination__root",
        "flex gap-1 text-xs font-medium",
    ]),
    item: cva([
        "UI-Pagination__item",
        "bg-transparent dark:bg-transparent text-sm text-[--muted] inline-flex h-8 w-8 items-center justify-center rounded-[--radius] border cursor-pointer",
        "hover:bg-[--subtle] dark:hover:bg-[--subtle] hover:border-[--subtle] select-none",
        "data-[selected=true]:bg-brand-500 data-[selected=true]:border-transparent data-[selected=true]:text-white data-[selected=true]:hover:bg-brand data-[selected=true]:pointer-events-none", // Selected
        "data-[disabled=true]:opacity-50 data-[disabled=true]:pointer-events-none data-[disabled=true]:cursor-not-allowed", // Disabled
        "outline-none ring-[--ring] focus-visible:ring-2",
    ]),
    ellipsis: cva([
        "UI-Pagination__ellipsis",
        "flex p-2 items-center text-[1.05rem]",
    ]),
})

/* -------------------------------------------------------------------------------------------------
 * Pagination
 * -----------------------------------------------------------------------------------------------*/

const __PaginationAnatomyContext = React.createContext<ComponentAnatomy<typeof PaginationAnatomy>>({})

export type PaginationProps = React.ComponentPropsWithRef<"ul"> & ComponentAnatomy<typeof PaginationAnatomy>

export const Pagination = React.forwardRef<HTMLUListElement, PaginationProps>((props, ref) => {

    const {
        children,
        itemClass,
        className,
        ellipsisClass,
        ...rest
    } = props

    return (
        <__PaginationAnatomyContext.Provider
            value={{
                itemClass,
                ellipsisClass,
            }}
        >
            <ul
                ref={ref}
                className={cn(PaginationAnatomy.root(), className)}
                role="navigation"
                {...rest}
            >
                {children}
            </ul>
        </__PaginationAnatomyContext.Provider>
    )

})

Pagination.displayName = "Pagination"


/* -------------------------------------------------------------------------------------------------
 * PaginationItem
 * -----------------------------------------------------------------------------------------------*/

export type PaginationItemProps = Omit<React.ComponentPropsWithRef<"button">, "children"> & {
    value: string | number
}

export const PaginationItem = React.forwardRef<HTMLButtonElement, PaginationItemProps>((props, ref) => {

    const {
        value,
        className,
        ...rest
    } = props

    const { itemClass } = React.useContext(__PaginationAnatomyContext)

    return (
        <li>
            <button
                className={cn(PaginationAnatomy.item(), itemClass, className)}
                {...rest}
                ref={ref}
            >
                {value}
            </button>
        </li>
    )

})

PaginationItem.displayName = "PaginationItem"

/* -------------------------------------------------------------------------------------------------
 * PaginationTrigger
 * -----------------------------------------------------------------------------------------------*/

export type PaginationTriggerProps = Omit<React.ComponentPropsWithRef<"button">, "children"> & {
    direction: "previous" | "next"
    isChevrons?: boolean
    isDisabled?: boolean
}

export const PaginationTrigger = React.forwardRef<HTMLButtonElement, PaginationTriggerProps>((props, ref) => {

    const {
        isChevrons = false,
        isDisabled = false,
        direction,
        className,
        ...rest
    } = props

    const { itemClass } = React.useContext(__PaginationAnatomyContext)

    return (
        <li>
            <button
                className={cn(PaginationAnatomy.item(), itemClass, className)}
                data-disabled={isDisabled}
                tabIndex={isDisabled ? -1 : undefined}
                {...rest}
                ref={ref}
            >
                {direction === "previous" ? (
                    <svg
                        xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none"
                        stroke="currentColor"
                        strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"
                        className="h-4 w-4"
                    >
                        {!isChevrons ? <polyline points="15 18 9 12 15 6"></polyline> : <>
                            <polyline points="11 17 6 12 11 7" />
                            <polyline points="18 17 13 12 18 7" />
                        </>}
                    </svg>
                ) : (
                    <svg
                        xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none"
                        stroke="currentColor"
                        strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"
                        className="h-4 w-4"
                    >
                        {!isChevrons ? <polyline points="9 18 15 12 9 6"></polyline> : <>
                            <polyline points="13 17 18 12 13 7" />
                            <polyline points="6 17 11 12 6 7" />
                        </>}
                    </svg>

                )}
            </button>
        </li>
    )

})

PaginationTrigger.displayName = "PaginationTrigger"

/* -------------------------------------------------------------------------------------------------
 * PaginationEllipsis
 * -----------------------------------------------------------------------------------------------*/

export type PaginationEllipsisProps = Omit<React.ComponentPropsWithRef<"span">, "children">

export const PaginationEllipsis = React.forwardRef<HTMLSpanElement, PaginationEllipsisProps>((props, ref) => {

    const {
        className,
        ...rest
    } = props

    const { ellipsisClass } = React.useContext(__PaginationAnatomyContext)

    return (
        <li className={cn(PaginationAnatomy.ellipsis(), ellipsisClass, className)}>
            <span
                {...rest}
                ref={ref}
            >
                &#8230;
            </span>
        </li>
    )

})

PaginationEllipsis.displayName = "PaginationEllipsis"

