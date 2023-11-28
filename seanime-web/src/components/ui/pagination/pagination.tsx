"use client"

import React from "react"
import { cn, ComponentWithAnatomy, createPolymorphicComponent, defineStyleAnatomy, getChildDisplayName } from "../core"
import { cva } from "class-variance-authority"

/* -------------------------------------------------------------------------------------------------
 * Anatomy
 * -----------------------------------------------------------------------------------------------*/

export const PaginationAnatomy = defineStyleAnatomy({
    root: cva([
        "UI-Pagination__root",
        "flex gap-1 text-xs font-medium"
    ]),
    item: cva([
        "UI-Pagination__item",
        "bg-transparent dark:bg-transparent text-base inline-flex h-8 w-8 items-center justify-center rounded border border-[--border] cursor-pointer",
        "hover:bg-[--highlight] dark:hover:bg-[--highlight] hover:border-[--highlight] select-none",
        "data-[selected=true]:bg-brand-500 data-[selected=true]:border-transparent data-[selected=true]:text-white data-[selected=true]:hover:bg-brand-500 data-[selected=true]:pointer-events-none", // Selected
        "data-[disabled=true]:opacity-50 data-[disabled=true]:pointer-events-none data-[disabled=true]:cursor-not-allowed", // Disabled
        "outline-none ring-[--ring] focus-visible:ring-2"
    ]),
    ellipsis: cva([
        "UI-Pagination__ellipsis",
        "flex p-2 items-center text-[1.05rem]"
    ]),
})

/* -------------------------------------------------------------------------------------------------
 * Pagination
 * -----------------------------------------------------------------------------------------------*/

export interface PaginationProps extends React.ComponentPropsWithRef<"ul">,
    Omit<ComponentWithAnatomy<typeof PaginationAnatomy>, "ellipsisClassName"> {
    children?: React.ReactNode
}

const _Pagination = (props: PaginationProps) => {

    const {
        children,
        rootClassName,
        itemClassName,
        className,
        ref,
        ...rest
    } = props

    const itemsWithProps = React.useMemo(() => React.Children.map(children, (child) => {
        if (React.isValidElement(child) && (getChildDisplayName(child) === "PaginationItem")) {
            return React.cloneElement(child, { itemClassName } as any)
        }
        return child
    }), [children])

    return (
        <ul
            className={cn(PaginationAnatomy.root(), rootClassName, className)}
            role="list"
            {...rest}
            ref={ref}
        >
            {itemsWithProps}
        </ul>
    )

}

_Pagination.displayName = "Pagination"

/* -------------------------------------------------------------------------------------------------
 * Pagination.Item
 * -----------------------------------------------------------------------------------------------*/

export interface PaginationItemProps extends Omit<React.ComponentPropsWithRef<"button">, "children">, ComponentWithAnatomy<typeof PaginationAnatomy> {
    value: string | number
}

const PaginationItem: React.FC<PaginationItemProps> = React.forwardRef<HTMLButtonElement, PaginationItemProps>((props, ref) => {

    const {
        value,
        className,
        itemClassName,
        ellipsisClassName, // Ignore
        ...rest
    } = props

    return (
        <li>
            <button
                className={cn(PaginationAnatomy.item(), itemClassName, className)}
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
 * Pagination.Trigger
 * -----------------------------------------------------------------------------------------------*/

export interface PaginationTriggerProps extends Omit<React.ComponentPropsWithRef<"button">, "children">,
    ComponentWithAnatomy<typeof PaginationAnatomy> {
    direction: "left" | "right"
    isChevrons?: boolean
    isDisabled?: boolean
}

const PaginationTrigger: React.FC<PaginationTriggerProps> = React.forwardRef<HTMLButtonElement, PaginationTriggerProps>((props, ref) => {

    const {
        isChevrons = false,
        isDisabled = false,
        direction,
        className,
        itemClassName,
        ellipsisClassName, // Ignore
        ...rest
    } = props

    return (
        <li>
            <button
                className={cn(PaginationAnatomy.item(), itemClassName, className)}
                data-disabled={`${isDisabled}`}
                {...rest}
                ref={ref}
            >
                {direction === "left" ? (
                    <svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none"
                         stroke="currentColor"
                         strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"
                         className="h-4 w-4"
                    >
                        {!isChevrons ? <polyline points="15 18 9 12 15 6"></polyline> : <>
                            <polyline points="11 17 6 12 11 7"/>
                            <polyline points="18 17 13 12 18 7"/>
                        </>}
                    </svg>
                ) : (
                    <svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none"
                         stroke="currentColor"
                         strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"
                         className="h-4 w-4"
                    >
                        {!isChevrons ? <polyline points="9 18 15 12 9 6"></polyline> : <>
                            <polyline points="13 17 18 12 13 7"/>
                            <polyline points="6 17 11 12 6 7"/>
                        </>}
                    </svg>

                )}
            </button>
        </li>
    )

})

PaginationTrigger.displayName = "PaginationTrigger"

/* -------------------------------------------------------------------------------------------------
 * Pagination.Ellipsis
 * -----------------------------------------------------------------------------------------------*/

export interface PaginationEllipsisProps extends Omit<React.ComponentPropsWithRef<"span">, "children">,
    ComponentWithAnatomy<typeof PaginationAnatomy> {
}

const PaginationEllipsis: React.FC<PaginationEllipsisProps> = React.forwardRef<HTMLSpanElement, PaginationEllipsisProps>((props, ref) => {

    const {
        className,
        ellipsisClassName,
        itemClassName, // Ignore
        ...rest
    } = props

    return (
        <li className={cn(PaginationAnatomy.ellipsis(), ellipsisClassName, className)}>
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

/* -------------------------------------------------------------------------------------------------
 * Component
 * -----------------------------------------------------------------------------------------------*/

_Pagination.Item = PaginationItem
_Pagination.Ellipsis = PaginationEllipsis
_Pagination.Trigger = PaginationTrigger

export const Pagination = createPolymorphicComponent<"div", PaginationProps, {
    Item: typeof PaginationItem
    Ellipsis: typeof PaginationEllipsis
    Trigger: typeof PaginationTrigger
}>(_Pagination)

Pagination.displayName = "Pagination"
