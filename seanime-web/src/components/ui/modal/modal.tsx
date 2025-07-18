"use client"

import { __isDesktop__ } from "@/types/constants"
import * as DialogPrimitive from "@radix-ui/react-dialog"
import { VisuallyHidden } from "@radix-ui/react-visually-hidden"
import { cva } from "class-variance-authority"
import * as React from "react"
import { CloseButton } from "../button"
import { cn, ComponentAnatomy, defineStyleAnatomy } from "../core/styling"

/* -------------------------------------------------------------------------------------------------
 * Anatomy
 * -----------------------------------------------------------------------------------------------*/

export const ModalAnatomy = defineStyleAnatomy({
    overlay: cva([
        "UI-Modal__overlay",
        "fixed inset-0 z-50 bg-black/80",
        "data-[state=open]:animate-in data-[state=closed]:animate-out",
        "data-[state=closed]:fade-out-0 data-[state=open]:fade-in-0",
        // "overflow-y-auto p-0 md:p-4 grid place-items-center",
    ]),
    content: cva([
        "UI-Modal__content",
        "z-50 grid relative w-full w-full shadow-xl border border-[rgb(255_255_255_/_5%)] max-w-lg gap-4 bg-[--background] p-6 shadow-xl duration-200",
        "data-[state=open]:animate-in data-[state=closed]:animate-out",
        "data-[state=closed]:fade-out-0 data-[state=open]:fade-in-0",
        // "data-[state=open]:slide-in-from-top-[40%] data-[state=closed]:slide-out-to-bottom-[40%]",
        // "data-[state=open]:slide-in-from-left-1/2 data-[state=open]:slide-in-from-top-[48%]",
        "data-[state=closed]:zoom-out-95 data-[state=open]:zoom-in-95",
        // __isDesktop__ && "mt-10",
        // __isDesktop__ && "select-none",
        "sm:rounded-xl",
    ]),
    close: cva([
        "UI-Modal__close",
        "absolute right-4 top-4 !mt-0",
    ]),
    header: cva([
        "UI-Modal__header",
        "flex flex-col space-y-1.5 text-center sm:text-left",
    ]),
    footer: cva([
        "UI-Modal__footer",
        "flex flex-col-reverse sm:flex-row sm:justify-end sm:space-x-2",
    ]),
    title: cva([
        "UI-Modal__title",
        "text-xl font-semibold leading-none tracking-tight",
    ]),
    description: cva([
        "UI-Modal__description",
        "text-sm text-[--muted]",
    ]),
})

/* -------------------------------------------------------------------------------------------------
 * Modal
 * -----------------------------------------------------------------------------------------------*/

export type ModalProps =
    Omit<React.ComponentPropsWithoutRef<typeof DialogPrimitive.Root>, "modal">
    &
    Pick<React.ComponentPropsWithoutRef<typeof DialogPrimitive.Content>, "onOpenAutoFocus" | "onCloseAutoFocus" | "onEscapeKeyDown" | "onPointerDownCapture" | "onInteractOutside">
    &
    ComponentAnatomy<typeof ModalAnatomy>
    & {
    /**
     * Interaction with outside elements will be enabled and other elements will be visible to screen readers.
     */
    allowOutsideInteraction?: boolean
    /**
     * The button that opens the modal
     */
    trigger?: React.ReactElement
    /**
     * Title of the modal
     */
    title?: React.ReactNode
    /**
     * An optional accessible description to be announced when the dialog is opened.
     */
    description?: React.ReactNode
    /**
     * Footer of the modal
     */
    footer?: React.ReactNode
    /**
     * Optional replacement for the default close button
     */
    closeButton?: React.ReactElement
    /**
     * Whether to hide the close button
     */
    hideCloseButton?: boolean
}

export function Modal(props: ModalProps) {

    const {
        allowOutsideInteraction = false,
        trigger,
        title,
        footer,
        description,
        children,
        closeButton,
        overlayClass,
        contentClass,
        closeClass,
        headerClass,
        footerClass,
        titleClass,
        descriptionClass,
        hideCloseButton,
        // Content
        onOpenAutoFocus,
        onCloseAutoFocus,
        onEscapeKeyDown,
        onPointerDownCapture,
        onInteractOutside,
        ...rest
    } = props

    return <DialogPrimitive.Root modal={!allowOutsideInteraction} {...rest}>

        {trigger && <DialogPrimitive.Trigger asChild>{trigger}</DialogPrimitive.Trigger>}

        <DialogPrimitive.Portal>
            <DialogPrimitive.Overlay className={cn(ModalAnatomy.overlay(), overlayClass)}>
                <div
                    className={cn(
                        "overflow-y-auto absolute inset-0 grid place-items-center p-0 md:p-4",
                        __isDesktop__ && "md:p-8",
                    )}
                >
                    <DialogPrimitive.Content
                        className={cn(ModalAnatomy.content(), contentClass)}
                        onOpenAutoFocus={onOpenAutoFocus}
                        onCloseAutoFocus={onCloseAutoFocus}
                        onEscapeKeyDown={onEscapeKeyDown}
                        onPointerDownCapture={onPointerDownCapture}
                        onInteractOutside={onInteractOutside}
                    >
                        {!title && !description ? (
                            <VisuallyHidden>
                                <DialogPrimitive.Title>Dialog</DialogPrimitive.Title>
                            </VisuallyHidden>
                        ) : (
                            <div className={cn(ModalAnatomy.header(), headerClass)}>
                                <DialogPrimitive.Title className={cn(ModalAnatomy.title(), titleClass)}>
                                    {title}
                                </DialogPrimitive.Title>
                                {description && (
                                    <DialogPrimitive.Description className={cn(ModalAnatomy.description(), descriptionClass)}>
                                        {description}
                                    </DialogPrimitive.Description>
                                )}
                            </div>
                        )}

                        {children}

                        {footer && <div className={cn(ModalAnatomy.footer(), footerClass)}>
                            {footer}
                        </div>}

                        {!hideCloseButton && <DialogPrimitive.Close className={cn(ModalAnatomy.close(), closeClass)} asChild>
                            {closeButton ? closeButton : <CloseButton />}
                        </DialogPrimitive.Close>}

                    </DialogPrimitive.Content>
                </div>


            </DialogPrimitive.Overlay>
        </DialogPrimitive.Portal>

    </DialogPrimitive.Root>
}

Modal.displayName = "Modal"
