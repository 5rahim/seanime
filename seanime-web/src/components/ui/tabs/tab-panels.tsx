"use client"

import React, { Fragment } from "react"
import { cn, ComponentWithAnatomy, createPolymorphicComponent, defineStyleAnatomy, getChildDisplayName } from "../core"
import { cva } from "class-variance-authority"
import type { TabListProps as TabPrimitiveListProps, TabProps as TabPrimitiveProps } from "@headlessui/react"
import { Tab as TabPrimitive } from "@headlessui/react"


/* -------------------------------------------------------------------------------------------------
 * Anatomy
 * -----------------------------------------------------------------------------------------------*/

export const TabPanelsAnatomy = defineStyleAnatomy({
    panels: cva([
        "UI-TabPanels__panels",
    ])
})

export const TabNavAnatomy = defineStyleAnatomy({
    nav: cva([
        "UI-TabNav__nav",
        "isolate flex border-b"
    ])
})

export const TabAnatomy = defineStyleAnatomy({
    tab: cva([
        "UI-Tab__tab",
        "relative min-w-0 flex-1 overflow-hidden py-4 px-4 text-sm font-medium text-center focus:z-10",
        "flex items-center justify-center gap-2 border-b-2 -mb-px",
        "text-[--muted] data-[selected=true]:text-[--brand] data-[selected=true]:border-brand dark:data-[selected=true]:border-brand-200",
        "border-[--border] hover:border-gray-300 dark:hover:border-gray-600",
        "focus-visible:bg-[--highlight] outline-none",
        "cursor-pointer"
    ])
})

/* -------------------------------------------------------------------------------------------------
 * TabPanels
 * -----------------------------------------------------------------------------------------------*/

export interface TabPanelsProps extends React.ComponentPropsWithRef<"div">,
    ComponentWithAnatomy<typeof TabPanelsAnatomy>,
    ComponentWithAnatomy<typeof TabNavAnatomy>,
    ComponentWithAnatomy<typeof TabAnatomy> {
    selectedIndex?: number
    onIndexChange?: (index: number) => void
}

const _TabPanels = (props: TabPanelsProps) => {

    const {
        children,
        panelsClassName,
        navClassName,
        tabClassName,
        className,
        selectedIndex,
        onIndexChange,
        ref,
        ...rest
    } = props

    const itemsWithProps = React.Children.map(children, (child) => {
        if (React.isValidElement(child) && getChildDisplayName(child) === "TabNav") {
            return React.cloneElement(child, { tabClassName, navClassName } as any)
        }
        return child
    })

    return (
        <TabPrimitive.Group
            selectedIndex={selectedIndex}
            onChange={onIndexChange}
        >
            <div
                className={cn(TabPanelsAnatomy.panels(), panelsClassName)}
                {...rest}
                ref={ref}
            >
                {itemsWithProps}
            </div>
        </TabPrimitive.Group>
    )

}

_TabPanels.displayName = "TabPanels"

/* -------------------------------------------------------------------------------------------------
 * TabNav
 * -----------------------------------------------------------------------------------------------*/

interface TabNavProps extends TabPrimitiveListProps<"div">,
    ComponentWithAnatomy<typeof TabNavAnatomy>,
    ComponentWithAnatomy<typeof TabAnatomy> {
    children?: React.ReactNode
}

export const TabNav: React.FC<TabNavProps> = React.forwardRef<HTMLDivElement, TabNavProps>((props, ref) => {

    const {
        children,
        className,
        navClassName,
        tabClassName,
        ...rest
    } = props

    const itemsWithProps = React.Children.map(children, (child) => {
        if (React.isValidElement(child) && getChildDisplayName(child) === "Tab") {
            return React.cloneElement(child, { tabClassName } as any)
        }
        return child
    })

    return (
        <TabPrimitive.List
            className={cn(TabNavAnatomy.nav(), navClassName, className)}
            {...rest}
            ref={ref}
        >
            {itemsWithProps}
        </TabPrimitive.List>
    )

})

TabNav.displayName = "TabNav"

/* -------------------------------------------------------------------------------------------------
 * Tab
 * -----------------------------------------------------------------------------------------------*/

interface TabProps extends TabPrimitiveProps<"div">, ComponentWithAnatomy<typeof TabAnatomy> {
    children?: React.ReactNode
}

export const Tab: React.FC<TabProps> = React.forwardRef<HTMLDivElement, TabProps>((props, ref) => {

    const {
        children,
        className,
        tabClassName,
        ...rest
    } = props

    return (
        <TabPrimitive
            as={Fragment}
        >
            {({ selected }) => (
                <div
                    className={cn(TabAnatomy.tab(), tabClassName, className)}
                    {...rest}
                    ref={ref}
                    data-selected={selected}
                >
                    {children}
                </div>
            )}
        </TabPrimitive>
    )

})

Tab.displayName = "Tab"

/* -------------------------------------------------------------------------------------------------
 * Component
 * -----------------------------------------------------------------------------------------------*/

_TabPanels.Tab = Tab
_TabPanels.Nav = TabNav
_TabPanels.Container = React.memo(TabPrimitive.Panels)
export const TabContainer = React.memo(TabPrimitive.Panels)
_TabPanels.Panel = TabPrimitive.Panel
export const TabPanel = TabPrimitive.Panel

_TabPanels.Container.displayName = "TabContainer"
_TabPanels.Panel.displayName = "TabPanel"

export const TabPanels = createPolymorphicComponent<"div", TabPanelsProps, {
    Tab: typeof Tab,
    Nav: typeof TabNav,
    Container: typeof TabPrimitive.Panels
    Panel: typeof TabPrimitive.Panel
}>(_TabPanels)
