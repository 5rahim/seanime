"use client"

import { Transition } from "@headlessui/react"
import React from "react"
import toast, { resolveValue, Toast as ToastType, Toaster } from "react-hot-toast"
import { cn, ComponentWithAnatomy, defineStyleAnatomy } from "../core"
import { CloseButton } from "../button"
import { cva } from "class-variance-authority"
import { Spinner } from "../loading-spinner"

/* -------------------------------------------------------------------------------------------------
 * Anatomy
 * -----------------------------------------------------------------------------------------------*/

export const ToastAnatomy = defineStyleAnatomy({
    toast: cva([
        "UI-Toast__toast",
        "relative transform py-2 px-5 flex rounded-lg shadow-md",
        "data-[type=error]:bg-red-500 data-[type=error]:text-white",
        "data-[type=success]:bg-green-500 data-[type=success]:text-white",
        "data-[type=loading]:bg-gray-50 dark:data-[type=loading]:bg-gray-800 data-[type=error]:text-white",
    ]),
    content: cva([
        "UI-Toast__content",
        "h-full flex items-center"
    ]),
    icon: cva([
        "UI-Toast__icon",
        "w-5 h-5"
    ]),
    closeButton: cva([
        "UI-Toast__closeButton",
        "absolute top-1 right-1"
    ]),
})

/* -------------------------------------------------------------------------------------------------
 * Toast
 * -----------------------------------------------------------------------------------------------*/

export interface ToastProps extends ComponentWithAnatomy<typeof ToastAnatomy> {
    t: ToastType
}

export const Toast = React.forwardRef<HTMLDivElement, ToastProps>((props, ref) => {

    const {
        t,
        toastClassName,
        closeButtonClassName,
        contentClassName,
        iconClassName,
    } = props


    return (
        <Transition
            appear
            show={t.visible}
            className={cn(ToastAnatomy.toast(), toastClassName)}
            data-type={t.type}
            enter="transition-all duration-150"
            enterFrom="opacity-0 scale-50"
            enterTo="opacity-100 scale-100"
            leave="transition-all duration-150"
            leaveFrom="opacity-100 scale-100"
            leaveTo="opacity-0 scale-75"
        >
            <div className={cn(ToastAnatomy.content(), contentClassName)}>
                {t.type !== "loading" &&
                    <svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none"
                         stroke="currentColor"
                         strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"
                         className={cn(ToastAnatomy.icon(), iconClassName)}>
                        {t.type === "error" && <>
                            <circle cx="12" cy="12" r="10"></circle>
                            <line x1="15" x2="9" y1="9" y2="15"></line>
                            <line x1="9" x2="15" y1="9" y2="15"></line>
                        </>}
                        {t.type === "success" && <>
                            <path d="M12 22c5.523 0 10-4.477 10-10S17.523 2 12 2 2 6.477 2 12s4.477 10 10 10z"></path>
                            <path d="m9 12 2 2 4-4"></path>
                        </>}
                    </svg>}
                {t.type === "loading" && <>
                    <Spinner className={cn(ToastAnatomy.icon(), iconClassName)}/>
                </>}
            </div>
            <p className="py-1 px-2 pr-6">{resolveValue(t.message, t)}</p>
            <CloseButton
                className={cn(ToastAnatomy.closeButton(), closeButtonClassName)}
                size="sm"
                intent={t.type !== "loading" ? "white-basic" : "gray-basic"}
                onClick={() => toast.dismiss(t.id)}
            />
        </Transition>
    )

})

/* -------------------------------------------------------------------------------------------------
 * ToastProvider
 * -----------------------------------------------------------------------------------------------*/

export const ToastProvider = () => {
    return (
        <Toaster>
            {(t) => <Toast t={t}/>}
        </Toaster>
    )
}
