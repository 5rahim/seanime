import { __isDesktop__ } from "@/types/constants"
import { VisuallyHidden } from "@radix-ui/react-visually-hidden"
import { cva, VariantProps } from "class-variance-authority"
import { atom } from "jotai"
import { useAtom } from "jotai/react"
import * as React from "react"
import { Drawer as VaulPrimitive } from "vaul"
import { CloseButton } from "../button"
import { cn, ComponentAnatomy, defineStyleAnatomy } from "../core/styling"

export const __openDrawersAtom = atom<string[]>([])

function useDrawerBodyBehavior(id: string, open: boolean | undefined) {
    const [, setOpenDrawers] = useAtom(__openDrawersAtom)

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
        "fixed z-50 w-full gap-4 bg-[--paper] p-6 shadow-lg overflow-y-auto",
        "transition ease-in-out data-[state=open]:animate-in data-[state=closed]:animate-out data-[state=closed]:duration-500 data-[state=open]:duration-500",
        "focus:outline-none focus-visible:outline-none",
        __isDesktop__ && "select-none",
    ], {
        variants: {
            side: {
                mangaReader: "w-full inset-x-0 top-0 border data-[state=closed]:slide-out-to-bottom data-[state=open]:slide-in-from-bottom !bg-[#000]",
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

export type DrawerProps =
    Pick<React.ComponentPropsWithoutRef<typeof VaulPrimitive.Content>,
        "onOpenAutoFocus" | "onCloseAutoFocus" | "onEscapeKeyDown" | "onPointerDownCapture" | "onInteractOutside"> &
    Pick<React.ComponentPropsWithoutRef<typeof VaulPrimitive.Root>,
        "open" | "defaultOpen" | "onOpenChange"> &
    VariantProps<typeof DrawerAnatomy.content> &
    ComponentAnatomy<typeof DrawerAnatomy> & {
    children?: React.ReactNode
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

    borderToBorder?: boolean
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
        defaultOpen,
        onOpenChange,
        // Content
        onOpenAutoFocus,
        onCloseAutoFocus,
        onEscapeKeyDown,
        onPointerDownCapture,
        onInteractOutside,
        portalContainer,
        borderToBorder: mangaReader,
        ...rest
    } = props

    const id = React.useId()
    const isControlled = open !== undefined
    const [uncontrolledOpen, setUncontrolledOpen] = React.useState(defaultOpen ?? false)
    const drawerSide = mangaReader ? "mangaReader" : (side ?? "right")
    const direction = drawerSide === "mangaReader" ? "bottom" : drawerSide
    const resolvedOpen = isControlled ? open : uncontrolledOpen
    const container = portalContainer ?? (typeof document !== "undefined" ? document.body : undefined)

    const handleOpenChange = React.useCallback((nextOpen: boolean) => {
        if (!isControlled) {
            setUncontrolledOpen(nextOpen)
        }

        onOpenChange?.(nextOpen)
    }, [isControlled, onOpenChange])

    const handleInteractOutside = React.useCallback((event: Parameters<NonNullable<typeof onInteractOutside>>[0]) => {
        onInteractOutside?.(event)

        if (allowOutsideInteraction && !event.defaultPrevented) {
            handleOpenChange(false)
        }
    }, [allowOutsideInteraction, handleOpenChange, onInteractOutside])

    useDrawerBodyBehavior(id, resolvedOpen)

    return (
        <VaulPrimitive.Root
            modal={!allowOutsideInteraction}
            container={container}
            direction={direction}
            handleOnly
            noBodyStyles
            disablePreventScroll={false}
            autoFocus
            open={resolvedOpen}
            defaultOpen={defaultOpen}
            onOpenChange={handleOpenChange}
            {...rest}
        >

            {trigger && <VaulPrimitive.Trigger asChild>{trigger}</VaulPrimitive.Trigger>}

            <VaulPrimitive.Portal>

                <VaulPrimitive.Overlay className={cn(DrawerAnatomy.overlay(), overlayClass)} />

                <VaulPrimitive.Content
                    className={cn(
                        DrawerAnatomy.content({ size, side: drawerSide }),
                        // __isDesktop__ && "pt-12",
                        !mangaReader && "lg:m-[10px] rounded-xl scroll-mt-2",
                        contentClass,
                    )}
                    style={{
                        marginTop: (__isDesktop__ && !mangaReader) ? "30px" : undefined,
                        height: (
                            __isDesktop__
                            && !mangaReader
                            && (side === "left" || side === "right")
                        ) ? "calc(100dvh - 50px)" : undefined,
                    }}
                    onOpenAutoFocus={onOpenAutoFocus}
                    onCloseAutoFocus={onCloseAutoFocus}
                    onEscapeKeyDown={onEscapeKeyDown}
                    onPointerDownCapture={onPointerDownCapture}
                    onInteractOutside={handleInteractOutside}
                    tabIndex={-1}
                >
                    {!title && !description ? (
                        <VisuallyHidden>
                            <VaulPrimitive.Title>Drawer</VaulPrimitive.Title>
                        </VisuallyHidden>
                    ) : (
                        <div className={cn(DrawerAnatomy.header(), headerClass)}>
                            <VaulPrimitive.Title
                                className={cn(
                                    DrawerAnatomy.title(),
                                    __isDesktop__ && "relative",
                                    titleClass,
                                )}
                            >
                                {title}
                            </VaulPrimitive.Title>
                            {description && (
                                <VaulPrimitive.Description className={cn(DrawerAnatomy.description(), descriptionClass)}>
                                    {description}
                                </VaulPrimitive.Description>
                            )}
                        </div>
                    )}

                    {children}

                    {footer && <div className={cn(DrawerAnatomy.footer(), footerClass)}>
                        {footer}
                    </div>}

                    {!hideCloseButton && <VaulPrimitive.Close
                        className={cn(
                            DrawerAnatomy.close(),
                            // __isDesktop__ && "!top-10 !right-4",
                            closeClass,
                        )}
                        asChild
                    >
                        {closeButton ? closeButton : <CloseButton />}
                    </VaulPrimitive.Close>}

                </VaulPrimitive.Content>

            </VaulPrimitive.Portal>

        </VaulPrimitive.Root>
    )
}

Drawer.displayName = "Drawer"
