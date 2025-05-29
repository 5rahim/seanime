"use client"

import { CloseButton } from "@/components/ui/button"
import { cn, ComponentAnatomy, defineStyleAnatomy } from "@/components/ui/core/styling"
import { __isDesktop__ } from "@/types/constants"
import * as DialogPrimitive from "@radix-ui/react-dialog"
import { VisuallyHidden } from "@radix-ui/react-visually-hidden"
import { cva, VariantProps } from "class-variance-authority"
import * as React from "react"

/* -------------------------------------------------------------------------------------------------
 * Anatomy
 * -----------------------------------------------------------------------------------------------*/

export const DrawerAnatomy = defineStyleAnatomy({
    overlay: cva([
        "UI-Drawer__overlay",
        // "transition-opacity duration-300",
    ]),
    content: cva([
        "UI-Drawer__content",
        "fixed z-50 w-full",
        "transition ease-in-out data-[state=open]:animate-in data-[state=closed]:animate-out data-[state=closed]:duration-500 data-[state=open]:duration-500",
        "focus:outline-none focus-visible:outline-none outline-none",
        __isDesktop__ && "select-none",
    ], {
        variants: {
            side: {
                player: "w-full inset-x-0 top-0 data-[state=closed]:slide-out-to-bottom data-[state=open]:slide-in-from-bottom",
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

    borderToBorder?: boolean

    miniPlayer?: boolean
}

export function NativePlayerDrawer(props: DrawerProps) {

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
        miniPlayer,
        ...rest
    } = props

    React.useEffect(() => {
        if (open && size === "full") {
            document.body.setAttribute("data-scroll-locked", "1")
        } else if (size === "full") {
            document.body.removeAttribute("data-scroll-locked")
        }
    }, [open, size])

    const isMiniPlayerRef = React.useRef(miniPlayer)

    React.useEffect(() => {
        setTimeout(() => {
            isMiniPlayerRef.current = miniPlayer
        }, 500)
    }, [miniPlayer])

    // Dragging
    const contentRef = React.useRef<HTMLDivElement>(null)

    // Calculate initial position immediately based on known dimensions
    const getInitialPosition = React.useCallback(() => {
        // Use the known CSS dimensions from the className
        const width = window.innerWidth >= 1024 ? 400 : 300 // lg:w-[400px] w-[300px]
        const height = width * (9 / 16) // aspect-video

        const rightBoundary = window.innerWidth - width - PADDING
        const bottomBoundary = window.innerHeight - height - PADDING

        return { x: rightBoundary, y: bottomBoundary }
    }, [])

    // Dragging functionality
    const [position, setPosition] = React.useState({ x: 0, y: 0 })
    const [isDragging, setIsDragging] = React.useState(false)
    const [isHidden, setIsHidden] = React.useState(false)
    const dragStartPos = React.useRef({ x: 0, y: 0 })
    const elementStartPos = React.useRef({ x: 0, y: 0 })
    const PADDING = 20 // Define padding constant
    const AUTO_HIDE_THRESHOLD = 0.5 // Hide when 50% is overflowing

    // Calculate boundaries helper function
    const calculateBoundaries = React.useCallback(() => {
        if (!contentRef.current) return null

        const width = contentRef.current.offsetWidth || 0
        const height = contentRef.current.offsetHeight || 0

        return {
            leftBoundary: 80 + PADDING,
            rightBoundary: window.innerWidth - width - PADDING,
            topBoundary: PADDING,
            bottomBoundary: window.innerHeight - height - PADDING,
            width,
            height,
        }
    }, [])

    React.useLayoutEffect(() => {
        if (miniPlayer) {
            setPosition(getInitialPosition())
            setIsHidden(false)
        }
    }, [miniPlayer, getInitialPosition])

    // Handle dragging only when in mini player mode
    React.useEffect(() => {
        if (!miniPlayer || !contentRef.current) return

        const handleMouseDown = (e: MouseEvent) => {
            if (!contentRef.current) return
            if ((e.target as HTMLElement).tagName === "MEDIA-TIME-RANGE" ||
                (e.target as HTMLElement).tagName === "MEDIA-VOLUME-RANGE" ||
                (e.target as HTMLElement).tagName === "MEDIA-PLAYBACK-RATE-RANGE") {
                return
            }
            setIsDragging(true)
            dragStartPos.current = { x: e.clientX, y: e.clientY }
            elementStartPos.current = { x: position.x, y: position.y }
        }

        const handleMouseMove = (e: MouseEvent) => {
            if (!isDragging || !contentRef.current) return

            const deltaX = e.clientX - dragStartPos.current.x
            const deltaY = e.clientY - dragStartPos.current.y

            const newX = elementStartPos.current.x + deltaX
            const newY = elementStartPos.current.y + deltaY

            const boundaries = calculateBoundaries()
            if (!boundaries) return

            // Check for auto-hide (when dragged to the right edge)
            const hideThresholdX = window.innerWidth - (boundaries.width * AUTO_HIDE_THRESHOLD)
            if (newX >= hideThresholdX) {
                if (!isHidden) {
                    setIsHidden(true)
                }
                setPosition({ x: window.innerWidth - boundaries.width * 0.15, y: newY }) // Show just a sliver
                return
            } else {
                if (isHidden) {
                    setIsHidden(false)
                }
            }

            // Apply boundaries for normal dragging
            const boundedX = Math.max(boundaries.leftBoundary, Math.min(newX, boundaries.rightBoundary * 1.25))
            const boundedY = Math.max(boundaries.topBoundary, Math.min(newY, boundaries.bottomBoundary))

            setPosition({ x: boundedX, y: boundedY })
        }

        const handleMouseUp = () => {
            setIsDragging(false)

            // If hidden, snap to hidden position or reveal based on drag behavior
            if (isHidden) {
                const boundaries = calculateBoundaries()
                if (!boundaries) return

                // If dragged back far enough to the left, show it again
                if (position.x < window.innerWidth - boundaries.width * 0.5) {
                    setIsHidden(false)
                    setPosition({ x: boundaries.rightBoundary, y: position.y })
                } else {
                    // Keep it hidden at the edge
                    setPosition({ x: window.innerWidth - boundaries.width * 0.1, y: position.y })
                }
                return
            }

            // Snap to the nearest corner when dragging stops
            const boundaries = calculateBoundaries()
            if (!boundaries) return

            const corners = [
                { x: boundaries.leftBoundary, y: boundaries.topBoundary }, // Top-left
                { x: boundaries.rightBoundary, y: boundaries.topBoundary }, // Top-right
                { x: boundaries.leftBoundary, y: boundaries.bottomBoundary }, // Bottom-left
                { x: boundaries.rightBoundary, y: boundaries.bottomBoundary }, // Bottom-right
            ]

            // Find the nearest corner
            let nearestCorner = corners[0]
            let minDistance = Number.MAX_VALUE

            corners.forEach(corner => {
                const distance = Math.sqrt(
                    Math.pow(position.x - corner.x, 2) +
                    Math.pow(position.y - corner.y, 2),
                )

                if (distance < minDistance) {
                    minDistance = distance
                    nearestCorner = corner
                }
            })

            // Snap to the nearest corner
            setPosition({ x: nearestCorner.x, y: nearestCorner.y })
        }

        // Add event listeners
        contentRef.current.addEventListener("mousedown", handleMouseDown)
        window.addEventListener("mousemove", handleMouseMove)
        window.addEventListener("mouseup", handleMouseUp)

        // Clean up
        return () => {
            contentRef.current?.removeEventListener("mousedown", handleMouseDown)
            window.removeEventListener("mousemove", handleMouseMove)
            window.removeEventListener("mouseup", handleMouseUp)
        }
    }, [miniPlayer, isDragging, position, calculateBoundaries, isHidden])

    // Handle window resize to maintain proper positioning
    React.useEffect(() => {
        if (!miniPlayer) return

        const handleResize = () => {
            const boundaries = calculateBoundaries()
            if (!boundaries) return

            // Adjust position if it's now outside boundaries
            setPosition(prevPosition => {
                let newX = prevPosition.x
                let newY = prevPosition.y

                // If hidden, maintain hidden state but adjust position
                if (isHidden) {
                    newX = window.innerWidth - boundaries.width * 0.1
                } else {
                    // Ensure position is within new boundaries
                    newX = Math.max(boundaries.leftBoundary, Math.min(prevPosition.x, boundaries.rightBoundary))
                    newY = Math.max(boundaries.topBoundary, Math.min(prevPosition.y, boundaries.bottomBoundary))
                }

                return { x: newX, y: newY }
            })
        }

        window.addEventListener("resize", handleResize)
        return () => window.removeEventListener("resize", handleResize)
    }, [miniPlayer, calculateBoundaries, isHidden])



    // Apply position styles when in mini player mode
    React.useEffect(() => {
        if (!contentRef.current || !miniPlayer) return

        // Use current position or calculate initial position if not set
        const currentPosition = (position.x === 0 && position.y === 0) ? getInitialPosition() : position

        contentRef.current.style.position = "fixed"
        contentRef.current.style.left = `${currentPosition.x}px`
        contentRef.current.style.top = `${currentPosition.y}px`
        contentRef.current.style.cursor = "move"
        contentRef.current.style.zIndex = isHidden ? "40" : "50" // Lower z-index when hidden

        // Handle opacity and scale for hiding/showing
        contentRef.current.style.opacity = isHidden ? "0.7" : "1"
        contentRef.current.style.transform = isHidden ? "scale(0.95)" : "scale(1)"

        // Add transition for smooth snapping and hiding/showing, remove it during dragging
        if (isMiniPlayerRef.current) {
            if (!isDragging) {
                contentRef.current.style.transition = "left 0.3s cubic-bezier(0.4, 0, 0.2, 1), top 0.3s cubic-bezier(0.4, 0, 0.2, 1), opacity 0.2s ease-out, transform 0.2s ease-out"
            } else {
                contentRef.current.style.transition = "opacity 0.2s ease-out, transform 0.2s ease-out" // Keep opacity and scale transition during
                                                                                                       // drag
            }
        }

        return () => {
            if (contentRef.current) {
                contentRef.current.style.position = ""
                contentRef.current.style.left = ""
                contentRef.current.style.top = ""
                contentRef.current.style.transform = ""
                contentRef.current.style.cursor = ""
                contentRef.current.style.transition = ""
                contentRef.current.style.zIndex = ""
                contentRef.current.style.opacity = ""
            }
        }
    }, [miniPlayer, position, isDragging, isHidden, getInitialPosition])

    return (
        <DialogPrimitive.Root modal={!allowOutsideInteraction} open={open} {...rest}>

            {trigger && <DialogPrimitive.Trigger asChild>{trigger}</DialogPrimitive.Trigger>}

            <DialogPrimitive.Portal container={portalContainer}>

                {/* <DialogPrimitive.Overlay className={cn(DrawerAnatomy.overlay(), overlayClass)} /> */}

                <DialogPrimitive.Content
                    className={cn(
                        DrawerAnatomy.content({ size, side: "player" }),
                        contentClass,
                        "w-full h-full transition-all duration-300 overflow-hidden fixed",
                        miniPlayer && "aspect-video w-[300px] lg:w-[400px] h-auto rounded-lg shadow-xl",
                        isHidden && "ring-2 ring-brand-300",
                    )}
                    ref={contentRef}
                    onOpenAutoFocus={e => e.preventDefault()}
                    onCloseAutoFocus={onCloseAutoFocus}
                    onEscapeKeyDown={onEscapeKeyDown}
                    onPointerDownCapture={onPointerDownCapture}
                    onInteractOutside={e => e.preventDefault()}
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
                                    __isDesktop__ && "relative",
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

                    <div
                        className={cn(
                            "h-full w-full",
                            miniPlayer && isDragging && "pointer-events-none",
                        )}
                    >
                        {children}
                    </div>

                    {footer && <div className={cn(DrawerAnatomy.footer(), footerClass)}>
                        {footer}
                    </div>}

                    {!hideCloseButton && <DialogPrimitive.Close
                        className={cn(
                            DrawerAnatomy.close(),
                            // __isDesktop__ && "!top-10 !right-4",
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

NativePlayerDrawer.displayName = "NativePlayerDrawer"
