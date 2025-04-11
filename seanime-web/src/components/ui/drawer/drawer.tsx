"use client"

import * as DialogPrimitive from "@radix-ui/react-dialog"
import { VisuallyHidden } from "@radix-ui/react-visually-hidden"
import { cva, VariantProps } from "class-variance-authority"
import { atom } from "jotai/index"
import { useAtom } from "jotai/react"
import * as React from "react"
import { CloseButton } from "../button"
import { cn, ComponentAnatomy, defineStyleAnatomy } from "../core/styling"

export const __openDrawersAtom = atom<string[]>([])

function useDrawerBodyBehavior(id: string, open: boolean | undefined) {
    const [openDrawers, setOpenDrawers] = useAtom(__openDrawersAtom)

    React.useEffect(() => {
        const body = document.querySelector("body")
        if (!body) return

        if (open) {
            setOpenDrawers(prev => [...prev, id])
        } else {
            setOpenDrawers(prev => {
                let next = prev.filter(i => i !== id)
                return next
            })
        }

        return () => {
            setOpenDrawers(prev => prev.filter(i => i !== id))
        }
    }, [open])

}

/* -------------------------------------------------------------------------------------------------
 * Anatomy
 * -----------------------------------------------------------------------------------------------*/

export const DrawerAnatomy = defineStyleAnatomy({
    overlay: cva([
        "UI-Drawer__overlay",
        "fixed inset-0 z-[50] bg-black/80",
        "data-[state=open]:animate-in data-[state=closed]:animate-out",
        "data-[state=closed]:fade-out-0 data-[state=open]:fade-in-0",
        // "transition-opacity duration-300",
    ]),
    content: cva([
        "UI-Drawer__content",
        "fixed z-50 w-full gap-4 bg-[--background] p-6 shadow-lg overflow-y-auto",
        "transition ease-in-out data-[state=open]:animate-in data-[state=closed]:animate-out data-[state=closed]:duration-500 data-[state=open]:duration-500",
        "focus:outline-none focus-visible:outline-none",
        process.env.NEXT_PUBLIC_PLATFORM === "desktop" && "select-none",
    ], {
        variants: {
            side: {
                mangaReader: "w-full inset-x-0 top-0 border data-[state=closed]:slide-out-to-bottom data-[state=open]:slide-in-from-bottom",
                top: "w-full lg:w-[calc(100%_-_20px)] inset-x-0 top-0 border data-[state=closed]:slide-out-to-top data-[state=open]:slide-in-from-top",
                bottom: "w-full lg:w-[calc(100%_-_20px)] inset-x-0 bottom-0 border data-[state=closed]:slide-out-to-bottom data-[state=open]:slide-in-from-bottom",
                left: "inset-y-0 left-0 h-full lg:h-[calc(100%_-_20px)] border data-[state=closed]:slide-out-to-left data-[state=open]:slide-in-from-left",
                right: "inset-y-0 right-0 h-full lg:h-[calc(100%_-_20px)] border data-[state=closed]:slide-out-to-right data-[state=open]:slide-in-from-right",
            },
            size: { sm: null, md: null, lg: null, xl: null, full: null },
        },
        defaultVariants: {
            side: "right",
            size: "md",
        },
        compoundVariants: [
            { size: "sm", side: "left", className: "sm:max-w-sm" },
            { size: "sm", side: "right", className: "sm:max-w-sm" },
            { size: "md", side: "left", className: "sm:max-w-md" },
            { size: "md", side: "right", className: "sm:max-w-md" },
            { size: "lg", side: "left", className: "sm:max-w-2xl" },
            { size: "lg", side: "right", className: "sm:max-w-2xl" },
            { size: "xl", side: "left", className: "sm:max-w-5xl" },
            { size: "xl", side: "right", className: "sm:max-w-5xl" },
            /**/
            { size: "full", side: "top", className: "h-dvh" },
            { size: "full", side: "bottom", className: "h-dvh" },
        ],
    }),
    close: cva([
        "UI-Drawer__close",
        "absolute right-4 top-4",
    ]),
    header: cva([
        "UI-Drawer__header",
        "flex flex-col space-y-1.5 text-center sm:text-left",
    ]),
    footer: cva([
        "UI-Drawer__footer",
        "flex flex-col-reverse sm:flex-row sm:justify-end sm:space-x-2",
    ]),
    title: cva([
        "UI-Drawer__title",
        "text-xl font-semibold leading-none tracking-tight",
    ]),
    description: cva([
        "UI-Drawer__description",
        "text-sm text-[--muted]",
    ]),
})

/* -------------------------------------------------------------------------------------------------
 * Drawer
 * -----------------------------------------------------------------------------------------------*/

export type DrawerProps = Omit<React.ComponentPropsWithoutRef<typeof DialogPrimitive.Root>, "modal"> &
    Pick<React.ComponentPropsWithoutRef<typeof DialogPrimitive.Content>,
        "onOpenAutoFocus" | "onCloseAutoFocus" | "onEscapeKeyDown" | "onPointerDownCapture" | "onInteractOutside"> &
    VariantProps<typeof DrawerAnatomy.content> &
    ComponentAnatomy<typeof DrawerAnatomy> & {
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
    /**
     *  Portal container
     */
    portalContainer?: HTMLElement

    mangaReader?: boolean
}

export function Drawer(props: DrawerProps) {

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
        side = "right",
        size,
        open,
        // Content
        onOpenAutoFocus,
        onCloseAutoFocus,
        onEscapeKeyDown,
        onPointerDownCapture,
        onInteractOutside,
        portalContainer,
        mangaReader,
        ...rest
    } = props

    const id = React.useId()

    useDrawerBodyBehavior(id, open)

    return (
        <DialogPrimitive.Root modal={!allowOutsideInteraction} open={open} {...rest}>

            {trigger && <DialogPrimitive.Trigger asChild>{trigger}</DialogPrimitive.Trigger>}

            <DialogPrimitive.Portal container={portalContainer}>

                <DialogPrimitive.Overlay className={cn(DrawerAnatomy.overlay(), overlayClass)} />

                <DialogPrimitive.Content
                    className={cn(
                        DrawerAnatomy.content({ size, side: mangaReader ? "mangaReader" : side }),
                        // process.env.NEXT_PUBLIC_PLATFORM === "desktop" && "pt-12",
                        !mangaReader && "lg:m-[10px] rounded-[--radius]",
                        contentClass,
                    )}
                    style={{
                        marginTop: (process.env.NEXT_PUBLIC_PLATFORM === "desktop" && !mangaReader) ? "30px" : undefined,
                        height: (
                            process.env.NEXT_PUBLIC_PLATFORM === "desktop"
                            && !mangaReader
                            && (side === "left" || side === "right")
                        ) ? "calc(100dvh - 50px)" : undefined,
                    }}
                    onOpenAutoFocus={onOpenAutoFocus}
                    onCloseAutoFocus={onCloseAutoFocus}
                    onEscapeKeyDown={onEscapeKeyDown}
                    onPointerDownCapture={onPointerDownCapture}
                    onInteractOutside={onInteractOutside}
                    tabIndex={-1}
                >
                    {!title && !description ? (
                        <VisuallyHidden>
                            <DialogPrimitive.Title>Drawer</DialogPrimitive.Title>
                        </VisuallyHidden>
                    ) : (
                        <div className={cn(DrawerAnatomy.header(), headerClass)}>
                            <DialogPrimitive.Title
                                className={cn(
                                    DrawerAnatomy.title(),
                                    process.env.NEXT_PUBLIC_PLATFORM === "desktop" && "relative",
                                    titleClass,
                                )}
                            >
                                {title}
                            </DialogPrimitive.Title>
                            {description && (
                                <DialogPrimitive.Description className={cn(DrawerAnatomy.description(), descriptionClass)}>
                                    {description}
                                </DialogPrimitive.Description>
                            )}
                        </div>
                    )}

                    {children}

                    {footer && <div className={cn(DrawerAnatomy.footer(), footerClass)}>
                        {footer}
                    </div>}

                    {!hideCloseButton && <DialogPrimitive.Close
                        className={cn(
                            DrawerAnatomy.close(),
                            // process.env.NEXT_PUBLIC_PLATFORM === "desktop" && "!top-10 !right-4",
                            closeClass,
                        )}
                        asChild
                    >
                        {closeButton ? closeButton : <CloseButton />}
                    </DialogPrimitive.Close>}

                </DialogPrimitive.Content>

            </DialogPrimitive.Portal>

        </DialogPrimitive.Root>
    )
}

Drawer.displayName = "Drawer"
