"use client"

import { Dialog, Transition } from "@headlessui/react"
import { cn, ComponentWithAnatomy, defineStyleAnatomy } from "../core"
import { cva, VariantProps } from "class-variance-authority"
import React, { Fragment } from "react"
import { CloseButton, CloseButtonProps } from "../button"

/* -------------------------------------------------------------------------------------------------
 * Anatomy
 * -----------------------------------------------------------------------------------------------*/

export const ModalAnatomy = defineStyleAnatomy({
    title: cva([
        "UI-Modal__title",
        "text-lg font-medium leading-6"
    ]),
    panel: cva([
        "UI-Modal__panel",
        "w-full transform overflow-hidden rounded-none sm:rounded-[--radius] p-6 text-left align-middle shadow-xl transition-all relative",
        "bg-[--paper]"
    ], {
        variants: {
            size: {
                sm: "sm:max-w-md",
                md: "sm:max-w-lg",
                lg: "sm:max-w-xl",
                xl: "sm:max-w-2xl",
                "2xl": "sm:max-w-4xl",
            },
        },
        defaultVariants: {
            size: "md",
        },
    }),
    body: cva("UI-Modal__body mt-2"),
    backdrop: cva([
        "UI-Modal__backdrop",
        "fixed inset-0 bg-black bg-opacity-25 dark:bg-opacity-70"
    ]),
    outsideContainer: cva([
        "UI-Modal__outsideContainer",
        "flex min-h-full justify-center p-0 sm:p-4 text-center"
    ], {
        variants: {
            mobilePlacement: {
                initial: "items-center",
                bottom: "items-end sm:items-center",
                top: "items-start sm:items-center"
            }
        },
        defaultVariants: {
            mobilePlacement: "bottom"
        }
    }),
    closeButton: cva([
        "UI-Modal__closeButton",
        "absolute right-2 top-2"
    ]),
})

/* -------------------------------------------------------------------------------------------------
 * Modal
 * -----------------------------------------------------------------------------------------------*/

export interface ModalProps extends Omit<React.ComponentPropsWithRef<"div">, "title">,
    ComponentWithAnatomy<typeof ModalAnatomy>,
    VariantProps<typeof ModalAnatomy.panel>, VariantProps<typeof ModalAnatomy.outsideContainer> {
    isOpen: boolean,
    onClose?: () => void
    title?: React.ReactNode
    isClosable?: boolean
    closeButtonIntent?: CloseButtonProps["intent"]
}

export const Modal = React.forwardRef<HTMLDivElement, ModalProps>((props, ref) => {

    const {
        children,
        className,
        isOpen,
        onClose,
        title,
        size,
        panelClassName,
        titleClassName,
        closeButtonClassName,
        outsideContainerClassName,
        bodyClassName,
        backdropClassName,
        isClosable,
        mobilePlacement,
        closeButtonIntent = "gray-outline",
        ...rest
    } = props

    return (
        <>
            <Transition appear show={isOpen} as={Fragment}>
                <Dialog as="div" className={cn("relative z-50")} onClose={() => onClose ? onClose() : () => {
                }}>
                    <Transition.Child
                        as={Fragment}
                        enter="ease-out duration-300"
                        enterFrom="opacity-0"
                        enterTo="opacity-100"
                        leave="ease-in duration-200"
                        leaveFrom="opacity-100"
                        leaveTo="opacity-0"
                    >
                        <div className={cn(ModalAnatomy.backdrop(), backdropClassName)}/>
                    </Transition.Child>

                    <div className="fixed inset-0 overflow-y-auto">
                        <div
                            className={cn(ModalAnatomy.outsideContainer({ mobilePlacement }), outsideContainerClassName)}>
                            <Transition.Child
                                as={Fragment}
                                enter="ease-out duration-300"
                                enterFrom="opacity-0 scale-95"
                                enterTo="opacity-100 scale-100"
                                leave="ease-in duration-200"
                                leaveFrom="opacity-100 scale-100"
                                leaveTo="opacity-0 scale-95"
                            >
                                <Dialog.Panel
                                    className={cn(
                                        ModalAnatomy.panel({ size }),
                                        panelClassName,
                                        className,
                                    )}
                                    {...rest}
                                >
                                    {title && <Dialog.Title
                                        as={"div"}
                                        className={cn(ModalAnatomy.title(), titleClassName)}
                                    >
                                        {title}
                                    </Dialog.Title>}

                                    {isClosable && <CloseButton
                                        onClick={onClose}
                                        className={cn(ModalAnatomy.closeButton(), closeButtonClassName)}
                                        intent={closeButtonIntent}
                                    />}

                                    <div className={cn(ModalAnatomy.body(), bodyClassName)}>
                                        {children}
                                    </div>

                                </Dialog.Panel>
                            </Transition.Child>
                        </div>
                    </div>
                </Dialog>
            </Transition>
        </>
    )

})

Modal.displayName = "Modal"
