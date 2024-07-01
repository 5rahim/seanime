"use client"
import { cva } from "class-variance-authority"
import * as React from "react"
import { useIsomorphicLayoutEffect, useUpdateEffect } from "../core/hooks"
import { cn, ComponentAnatomy, defineStyleAnatomy } from "../core/styling"
import { useDraggableScroll } from "./use-draggable-scroll"

/* -------------------------------------------------------------------------------------------------
 * Anatomy
 * -----------------------------------------------------------------------------------------------*/

const HorizontalDraggableScrollAnatomy = defineStyleAnatomy({
    root: cva([
        "UI-HorizontalDraggableScroll__root",
        "relative flex items-center lg:gap-2",
    ]),
    container: cva([
        "UI-HorizontalDraggableScroll__container",
        "flex max-w-full w-full space-x-3 overflow-x-scroll scrollbar-hide scroll select-none",
    ]),
    chevronOverlay: cva([
        "flex flex-none items-center justify-center cursor-pointer hover:text-[--foreground] absolute bg-gradient-to-r from-[--background] z-40",
        "h-full w-16 opacity-90 hover:opacity-100 transition-opacity",
        "data-[state=hidden]:opacity-0 data-[state=hidden]:pointer-events-none",
        "data-[state=visible]:animate-in data-[state=hidden]:animate-out",
        "data-[state=visible]:fade-in-0 data-[state=hidden]:fade-out-0",
        "data-[state=visible]:duration-600 data-[state=hidden]:duration-600",
        "hidden md:flex",
    ], {
        variants: {
            side: {
                left: "left-0 bg-gradient-to-r rounded-l-xl",
                right: "right-0 bg-gradient-to-l rounded-r-xl",
            },
        },
    }),
    scrollContainer: cva([
        "flex max-w-full w-full space-x-3 overflow-x-scroll scrollbar-hide scroll select-none",
    ]),
    chevronIcon: cva([
        "w-7 h-7 stroke-2 mx-auto",
    ]),

})

/* -------------------------------------------------------------------------------------------------
 * HorizontalDraggableScroll
 * -----------------------------------------------------------------------------------------------*/

export type HorizontalDraggableScrollProps = ComponentAnatomy<typeof HorizontalDraggableScrollAnatomy> & {
    className?: string
    children?: React.ReactNode
    /**
     * Callback fired when the slider has reached the end
     */
    onSlideEnd?: () => void
    /**
     * The amount of pixels to scroll when the chevron is clicked
     * @default 500
     */
    scrollAmount?: number
    /**
     * Decay rate of the inertial effect by using an optional parameter.
     * A value of 0.95 means that at the speed will decay 5% of its current value at every 1/60 seconds.
     */
    decayRate?: number
    /**
     * Control drag sensitivity by specifying the minimum distance in order to distinguish an intentional drag movement from an unwanted one.
     */
    safeDisplacement?: number
    /**
     * Whether to apply a rubber band effect when the slider reaches the end
     */
    applyRubberBandEffect?: boolean
}

export const HorizontalDraggableScroll = React.forwardRef<HTMLDivElement, HorizontalDraggableScrollProps>((props, forwadedRef) => {

    const {
        children,
        onSlideEnd,
        className,
        containerClass,
        scrollContainerClass,
        chevronIconClass,
        chevronOverlayClass,
        decayRate = 0.95,
        safeDisplacement = 20,
        applyRubberBandEffect = true,
        scrollAmount = 500,
        ...rest
    } = props

    const ref = React.useRef<HTMLDivElement>(null) as React.MutableRefObject<HTMLDivElement>
    const { events } = useDraggableScroll(ref, {
        decayRate,
        safeDisplacement,
        applyRubberBandEffect,
    })

    const [isScrolledToLeft, setIsScrolledToLeft] = React.useState(true)
    const [isScrolledToRight, setIsScrolledToRight] = React.useState(false)
    const [showChevronRight, setShowRightChevron] = React.useState(false)

    const handleScroll = React.useCallback(() => {
        const div = ref.current

        if (div) {
            const scrolledToLeft = div.scrollLeft === 0
            const scrolledToRight = div.scrollLeft + div.clientWidth === div.scrollWidth

            setIsScrolledToLeft(scrolledToLeft)
            setIsScrolledToRight(scrolledToRight)
        }
    }, [])

    useUpdateEffect(() => {
        if (!isScrolledToLeft && isScrolledToRight) {
            onSlideEnd && onSlideEnd()
            const t = setTimeout(() => {
                const div = ref.current
                if (div) {
                    div.scrollTo({
                        left: div.scrollLeft + scrollAmount,
                        behavior: "smooth",
                    })
                }
            }, 1000)
            return () => clearTimeout(t)
        }
    }, [isScrolledToLeft, isScrolledToRight])

    const slideLeft = React.useCallback(() => {
        const div = ref.current
        if (div) {
            div.scrollTo({
                left: div.scrollLeft - scrollAmount,
                behavior: "smooth",
            })
        }
    }, [scrollAmount])

    const slideRight = React.useCallback(() => {
        const div = ref.current
        if (div) {
            div.scrollTo({
                left: div.scrollLeft + scrollAmount,
                behavior: "smooth",
            })
        }
    }, [scrollAmount])

    useIsomorphicLayoutEffect(() => {
        if (ref.current.clientWidth < ref.current.scrollWidth) {
            setShowRightChevron(true)
        } else {
            setShowRightChevron(false)
        }
    }, [])

    return (
        <div ref={forwadedRef} className={cn(HorizontalDraggableScrollAnatomy.root(), className)}>
            <div
                onClick={slideLeft}
                className={cn(HorizontalDraggableScrollAnatomy.chevronOverlay({ side: "left" }), chevronOverlayClass)}
                data-state={isScrolledToLeft ? "hidden" : "visible"}
            >
                <svg
                    xmlns="http://www.w3.org/2000/svg"
                    viewBox="0 0 24 24"
                    fill="none"
                    stroke="currentColor"
                    strokeWidth="2"
                    strokeLinecap="round"
                    strokeLinejoin="round"
                    className={cn(HorizontalDraggableScrollAnatomy.chevronIcon(), chevronIconClass)}
                >
                    <path d="m15 18-6-6 6-6" />
                </svg>
            </div>
            <div
                onScroll={handleScroll}
                className={cn(HorizontalDraggableScrollAnatomy.container(), containerClass)}
                {...events}
                ref={ref}
            >
                {children}
            </div>
            <div
                onClick={slideRight}
                className={cn(HorizontalDraggableScrollAnatomy.chevronOverlay({ side: "right" }), chevronOverlayClass)}
                data-state={!isScrolledToRight && showChevronRight ? "visible" : "hidden"}
            >
                <svg
                    xmlns="http://www.w3.org/2000/svg"
                    viewBox="0 0 24 24"
                    fill="none"
                    stroke="currentColor"
                    strokeWidth="2"
                    strokeLinecap="round"
                    strokeLinejoin="round"
                    className={cn(HorizontalDraggableScrollAnatomy.chevronIcon(), chevronIconClass)}
                >
                    <path d="m9 18 6-6-6-6" />
                </svg>
            </div>
        </div>
    )
})
