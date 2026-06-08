import { CloseButton } from "@/components/ui/button"
import { cn, ComponentAnatomy, defineStyleAnatomy } from "@/components/ui/core/styling"
import { __isDesktop__ } from "@/types/constants"
import type * as DialogPrimitive from "@radix-ui/react-dialog"
import { VisuallyHidden } from "@radix-ui/react-visually-hidden"
import { cva, VariantProps } from "class-variance-authority"
import * as React from "react"
import { RemoveScrollBar } from "react-remove-scroll-bar"
import { Drawer as VaulPrimitive } from "vaul"

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

type DrawerProps = Omit<React.ComponentPropsWithoutRef<typeof DialogPrimitive.Root>, "modal"> &
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

    onMiniPlayerClick?: () => void
}

export function VideoCoreDrawer(props: DrawerProps) {

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
        miniPlayer,
        onMiniPlayerClick,
        ...rest
    } = props

    const isControlled = open !== undefined
    const [uncontrolledOpen, setUncontrolledOpen] = React.useState(defaultOpen ?? false)
    const resolvedOpen = isControlled ? open : uncontrolledOpen
    const container = portalContainer ?? (typeof document !== "undefined" ? document.body : undefined)
    const canInteractOutside = allowOutsideInteraction || !!miniPlayer

    const handleOpenChange = React.useCallback((nextOpen: boolean) => {
        if (!isControlled) {
            setUncontrolledOpen(nextOpen)
        }

        onOpenChange?.(nextOpen)
    }, [isControlled, onOpenChange])

    React.useLayoutEffect(() => {
        if (!resolvedOpen || !canInteractOutside || typeof document === "undefined") return

        const body = document.body
        const previousPointerEvents = body.style.pointerEvents

        const unlockBodyPointerEvents = () => {
            if (body.style.pointerEvents !== "auto") {
                body.style.pointerEvents = "auto"
            }
        }

        unlockBodyPointerEvents()

        const frame = window.requestAnimationFrame(unlockBodyPointerEvents)
        const observer = new MutationObserver(unlockBodyPointerEvents)

        observer.observe(body, { attributes: true, attributeFilter: ["style"] })

        return () => {
            window.cancelAnimationFrame(frame)
            observer.disconnect()

            if (body.style.pointerEvents === "auto") {
                body.style.pointerEvents = previousPointerEvents
            }
        }
    }, [resolvedOpen, canInteractOutside])

    React.useEffect(() => {
        const t = setTimeout(() => {
            if (open && size === "full") {
                const v = document.body.getAttribute("data-scroll-locked")
                document.body.setAttribute("data-scroll-locked", v || "1")
            }
        }, 1000)
        return () => clearTimeout(t)
    }, [open, size])

    const prevMiniPlayerRef = React.useRef(miniPlayer)
    const prevRectRef = React.useRef<DOMRect | null>(null)
    const motionAnimationRef = React.useRef<Animation | null>(null)

    // Dragging
    const contentRef = React.useRef<HTMLDivElement>(null)
    const draggableAreaRef = React.useRef<HTMLDivElement>(null)

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

        const handleMouseUp = (e: MouseEvent) => {
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

            // if it's just a click and the target is the draggable area, do nothing
            if (Math.abs(position.x - elementStartPos.current.x) < 10 && Math.abs(position.y - elementStartPos.current.y) < 10) {
                if (e.target === draggableAreaRef.current) {
                    onMiniPlayerClick?.()
                    return
                }
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
        draggableAreaRef.current?.addEventListener("mousedown", handleMouseDown)
        window.addEventListener("mousemove", handleMouseMove)
        window.addEventListener("mouseup", handleMouseUp)

        // Clean up
        return () => {
            draggableAreaRef.current?.removeEventListener("mousedown", handleMouseDown)
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


    React.useLayoutEffect(() => {
        if (!contentRef.current) return

        const element = contentRef.current
        const currentPosition = (position.x === 0 && position.y === 0) ? getInitialPosition() : position
        const miniWidth = window.innerWidth >= 1024 ? 400 : 300
        const miniHeight = miniWidth * (9 / 16)
        const didToggleMiniPlayer = prevMiniPlayerRef.current !== miniPlayer

        element.style.width = ""
        element.style.height = ""
        element.style.borderRadius = ""
        element.style.cursor = ""

        if (!miniPlayer) {
            element.style.position = ""
            element.style.left = ""
            element.style.right = ""
            element.style.top = ""
            element.style.bottom = ""
            element.style.transform = ""
            element.style.zIndex = ""
            element.style.opacity = ""
            element.style.transition = didToggleMiniPlayer ? "none" : ""
            return
        }

        element.style.position = "fixed"
        element.style.left = `${currentPosition.x}px`
        element.style.right = "auto"
        element.style.top = `${currentPosition.y}px`
        element.style.bottom = "auto"
        element.style.width = `${miniWidth}px`
        element.style.height = `${miniHeight}px`
        element.style.borderRadius = "0.5rem"
        element.style.zIndex = isHidden ? "40" : "50"
        element.style.opacity = isHidden ? "0.7" : "1"
        element.style.transform = isHidden ? "scale(0.95)" : "scale(1)"

        if (didToggleMiniPlayer) {
            element.style.transition = "none"
            return
        }

        if (!isDragging) {
            element.style.transition = "left 0.3s cubic-bezier(0.4, 0, 0.2, 1), top 0.3s cubic-bezier(0.4, 0, 0.2, 1), opacity 0.2s ease-out, transform 0.2s ease-out"
        } else {
            element.style.transition = "opacity 0.2s ease-out, transform 0.2s ease-out" // Keep opacity and scale transition during
            // drag
        }
    }, [miniPlayer, position, isDragging, isHidden, getInitialPosition])

    React.useLayoutEffect(() => {
        if (!contentRef.current) return

        const element = contentRef.current
        const previousRect = prevRectRef.current
        const currentRect = element.getBoundingClientRect()
        const didToggleMiniPlayer = prevMiniPlayerRef.current !== miniPlayer
        const finalTransform = miniPlayer ? (isHidden ? "scale(0.95)" : "scale(1)") : "none"
        const browserViewTransitionActive = document.documentElement.hasAttribute("data-vc-miniplayer-view-transition")

        motionAnimationRef.current?.cancel()

        if (!browserViewTransitionActive && previousRect && didToggleMiniPlayer && currentRect.width > 0 && currentRect.height > 0) {
            const deltaX = previousRect.left - currentRect.left
            const deltaY = previousRect.top - currentRect.top
            const scaleX = previousRect.width / currentRect.width
            const scaleY = previousRect.height / currentRect.height

            motionAnimationRef.current = element.animate([
                {
                    transformOrigin: "top left",
                    transform: `translate(${deltaX}px, ${deltaY}px) scale(${scaleX}, ${scaleY})`,
                },
                {
                    transformOrigin: "top left",
                    transform: finalTransform,
                },
            ], {
                duration: 300,
                easing: "cubic-bezier(0.4, 0, 0.2, 1)",
                fill: "both",
            })
        }

        prevMiniPlayerRef.current = miniPlayer

        return () => {
            motionAnimationRef.current?.cancel()
        }
    }, [miniPlayer])

    React.useLayoutEffect(() => {
        if (!contentRef.current) return

        prevRectRef.current = contentRef.current.getBoundingClientRect()
    }, [miniPlayer, position, isHidden])

    return (
        <VaulPrimitive.Root
            modal={!canInteractOutside}
            container={container}
            direction="bottom"
            dismissible={false}
            handleOnly
            noBodyStyles
            autoFocus={false}
            disablePreventScroll={false}
            open={resolvedOpen}
            defaultOpen={defaultOpen}
            onOpenChange={handleOpenChange}
            {...rest}
        >

            {trigger && <VaulPrimitive.Trigger asChild>{trigger}</VaulPrimitive.Trigger>}

            {(resolvedOpen && size === "full") && <RemoveScrollBar />}

            <VaulPrimitive.Portal>

                {/* <VaulPrimitive.Overlay className={cn(DrawerAnatomy.overlay(), overlayClass)} /> */}

                <VaulPrimitive.Content
                    data-vc-element="drawer-content"
                    data-vc-miniplayer-state={miniPlayer}
                    data-vc-dragging-state={isDragging}
                    data-vc-hidden-state={isHidden}
                    className={cn(
                        DrawerAnatomy.content({ size, side: "player" }),
                        contentClass,
                        "w-full h-full transition-all duration-300 overflow-hidden fixed transform-gpu [contain:layout_paint_style]",
                        miniPlayer && "aspect-video w-[300px] lg:w-[400px] h-auto rounded-lg shadow-xl will-change-[transform,opacity]",
                        isHidden && "ring-2 ring-brand-300",
                    )}
                    ref={contentRef}
                    onOpenAutoFocus={e => e.preventDefault()}
                    onCloseAutoFocus={onCloseAutoFocus}
                    onEscapeKeyDown={onEscapeKeyDown}
                    onPointerDownCapture={onPointerDownCapture}
                    onInteractOutside={onInteractOutside}
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

                    {miniPlayer && <div
                        ref={draggableAreaRef}
                        data-vc-element="drawer-miniplayer-draggable-area"
                        className="vc-drawer-draggable-area absolute inset-0 z-[6]"
                    >

                    </div>}


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

VideoCoreDrawer.displayName = "VideoCoreDrawer"
