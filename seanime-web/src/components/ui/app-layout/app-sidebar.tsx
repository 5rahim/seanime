"use client"

import React, { useEffect, useState } from "react"
import { cn, ComponentWithAnatomy, defineStyleAnatomy } from "../core"
import { cva } from "class-variance-authority"
import { Drawer, DrawerProps } from "../modal"

/* -------------------------------------------------------------------------------------------------
 * Context
 * -----------------------------------------------------------------------------------------------*/

const __AppSidebarContext = React.createContext<{
    open: boolean,
    setOpen: React.Dispatch<React.SetStateAction<boolean>>
}>({
    open: false,
    setOpen: () => {
    }
})

const useAppSidebarContext = () => {
    return React.useContext(__AppSidebarContext)
}

/* -------------------------------------------------------------------------------------------------
 * Anatomy
 * -----------------------------------------------------------------------------------------------*/

export const AppSidebarAnatomy = defineStyleAnatomy({
    sidebar: cva([
        "UI-AppSidebar__sidebar",
        "flex flex-grow flex-col",
    ])
})

export const AppSidebarTriggerAnatomy = defineStyleAnatomy({
    trigger: cva([
        "UI-AppSidebarTrigger__trigger",
        "block md:hidden",
        "items-center justify-center rounded-[--radius] p-2 text-[--muted] hover:bg-[--highlight] hover:text-[--text-color]",
        "focus:outline-none focus:ring-2 focus:ring-inset focus:ring-[--ring]"
    ])
})

/* -------------------------------------------------------------------------------------------------
 * AppSidebar
 * -----------------------------------------------------------------------------------------------*/

export interface AppSidebarProps extends React.ComponentPropsWithRef<"div">, ComponentWithAnatomy<typeof AppSidebarAnatomy> {
    mobileDrawerProps?: Partial<DrawerProps>
}

export const AppSidebar: React.FC<AppSidebarProps> = React.forwardRef<HTMLDivElement, AppSidebarProps>((props, ref) => {

    const {
        children,
        sidebarClassName,
        className,
        ...rest
    } = props

    const ctx = useAppSidebarContext()

    return (
        <>
            <div
                className={cn(AppSidebarAnatomy.sidebar(), sidebarClassName)}
                {...rest}
                ref={ref}
            >
                <div className={cn(className)}>
                    {children}
                </div>
            </div>
            <Drawer
                isOpen={ctx.open}
                onClose={() => ctx.setOpen(false)}
                placement="left"
                isClosable
                className="md:hidden"
                containerClassName="w-[85%]"
                bodyClassName={cn("p-0 md:p-0", className)}
                headerClassName="absolute p-2 sm:p-2 md:p-2 lg:p-2 right-0"
                closeButtonIntent="white-outline"
            >
                {children}
            </Drawer>
        </>
    )

})

AppSidebar.displayName = "AppSidebar"

/* -------------------------------------------------------------------------------------------------
 * AppSidebarTrigger
 * -----------------------------------------------------------------------------------------------*/

export interface AppSidebarTriggerProps extends React.ComponentPropsWithRef<"button">, ComponentWithAnatomy<typeof AppSidebarTriggerAnatomy> {
}

export const AppSidebarTrigger: React.FC<AppSidebarTriggerProps> = React.forwardRef<HTMLButtonElement, AppSidebarTriggerProps>((props, ref) => {

    const {
        children,
        triggerClassName,
        className,
        ...rest
    } = props

    const ctx = useAppSidebarContext()

    return (
        <button
            className={cn(AppSidebarTriggerAnatomy.trigger(), triggerClassName, className)}
            onClick={() => ctx.setOpen(s => !s)}
            {...rest}
            ref={ref}
        >
            <span className="sr-only">Open main menu</span>
            {ctx.open ? (
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
    )

})

AppSidebarTrigger.displayName = "AppSidebarTrigger"


/* -------------------------------------------------------------------------------------------------
 * Provider
 * -----------------------------------------------------------------------------------------------*/

export const AppSidebarProvider: React.FC<{ children?: React.ReactNode, open?: boolean }> = ({
                                                                                                 children,
                                                                                                 open: _open
                                                                                             }) => {

    const [open, setOpen] = useState(_open ?? false)

    useEffect(() => {
        if (_open !== undefined)
            setOpen(_open)
    }, [_open])

    return (
        <__AppSidebarContext.Provider value={{ open, setOpen }}>
            {children}
        </__AppSidebarContext.Provider>
    )
}
