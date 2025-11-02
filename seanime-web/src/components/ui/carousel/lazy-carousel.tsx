"use client"

import { cn } from "@/components/ui/core/styling"
import { Skeleton } from "@/components/ui/skeleton"
import * as React from "react"
import { Carousel, CarouselAnatomy, CarouselContent, CarouselDotButtons, CarouselProps, useCarousel } from "./carousel"

/* -------------------------------------------------------------------------------------------------
 * LazyCarousel
 * -----------------------------------------------------------------------------------------------*/

export type LazyCarouselProps = CarouselProps & {
    children: React.ReactNode
    itemCount: number
    threshold?: number // Minimum number of items before lazy loading kicks in
}

export const LazyCarousel = React.forwardRef<HTMLDivElement, React.HTMLAttributes<HTMLDivElement> & LazyCarouselProps>(({
    children,
    itemCount,
    threshold = 10,
    ...carouselProps
}, ref) => {

    // If item count is below threshold, use regular carousel
    if (itemCount <= threshold) {
        return (
            <Carousel {...carouselProps} ref={ref}>
                <CarouselContent className="px-6">
                    {children}
                </CarouselContent>
            </Carousel>
        )
    }

    return (
        <Carousel {...carouselProps} ref={ref}>
            <LazyCarouselContent itemCount={itemCount}>
                {children}
            </LazyCarouselContent>
        </Carousel>
    )
})
LazyCarousel.displayName = "LazyCarousel"

/* -------------------------------------------------------------------------------------------------
 * LazyCarouselContent
 * -----------------------------------------------------------------------------------------------*/

type LazyCarouselContentProps = {
    children: React.ReactNode
    itemCount: number
    className?: string
}

export const LazyCarouselContent = React.forwardRef<HTMLDivElement, LazyCarouselContentProps>(({
    children,
    itemCount,
    className = "px-6",
}, ref) => {
    const { carouselRef, orientation, gap, api } = useCarousel()
    const [visibleIndices, setVisibleIndices] = React.useState<Set<number>>(new Set())
    const [itemWidths, setItemWidths] = React.useState<Map<number, number>>(new Map())
    const itemRefs = React.useRef<(HTMLDivElement | null)[]>([])
    const observerRef = React.useRef<IntersectionObserver | null>(null)
    const localRef = React.useRef<HTMLDivElement>(null)

    // Combine local ref with forwarded ref
    React.useImperativeHandle(ref, () => localRef.current!, [])

    // Initial visible items - start with first few items
    const initialVisibleCount = React.useMemo(() => {
        // Estimate how many items can fit initially
        if (typeof window === "undefined") return 5
        const viewportWidth = window.innerWidth
        const estimatedItemWidth = 250 // Default estimated width
        return Math.min(Math.ceil(viewportWidth / estimatedItemWidth) + 2, itemCount)
    }, [itemCount])

    // Initialize visible indices with first few items
    React.useEffect(() => {
        const initialVisibleIndices = new Set(
            Array.from(Array(Math.min(initialVisibleCount, itemCount)).keys()),
        )
        setVisibleIndices(initialVisibleIndices)

        return () => {
            setItemWidths(new Map())
        }
    }, [initialVisibleCount, itemCount])

    // Setup intersection observer
    React.useEffect(() => {
        if (!localRef.current || !api) return

        const observerOptions = {
            root: api.rootNode(), // Use embla's root node
            rootMargin: "100% 0px", // Load items when they're about to come into view
            threshold: 0,
        }

        observerRef.current = new IntersectionObserver((entries) => {
            entries.forEach(entry => {
                const index = parseInt(entry.target.getAttribute("data-index") ?? "-1")

                if (entry.isIntersecting) {
                    setVisibleIndices(prev => {
                        const updated = new Set(prev)
                        updated.add(index)
                        return updated
                    })
                } else {
                    setVisibleIndices(prev => {
                        const updated = new Set(prev)
                        // Keep initial items always visible to maintain width estimation
                        if (index >= initialVisibleCount) {
                            updated.delete(index)
                        }
                        return updated
                    })
                }
            })
        }, observerOptions)

        // Observe all item containers
        itemRefs.current.forEach(ref => {
            if (ref) observerRef.current?.observe(ref)
        })

        return () => {
            observerRef.current?.disconnect()
        }
    }, [itemCount, initialVisibleCount, api])

    // Function to update item widths
    const updateItemWidth = React.useCallback((index: number, width: number) => {
        setItemWidths(prev => {
            const updated = new Map(prev)
            updated.set(index, width)
            return updated
        })
    }, [])

    // Calculate estimated width based on visible items
    const estimatedWidth = React.useMemo(() => {
        const widths = Array.from(itemWidths.values())
        if (widths.length === 0) return 250 // Default fallback
        return widths.reduce((sum, width) => sum + width, 0) / widths.length
    }, [itemWidths])

    return (
        <div ref={carouselRef} className={cn(CarouselAnatomy.content())}>
            <div
                ref={localRef}
                className={cn(CarouselAnatomy.innerContent({ orientation, gap }), className)}
            >
                {React.Children.map(children, (child, index) => {
                    const isVisible = visibleIndices.has(index)
                    const storedWidth = itemWidths.get(index)

                    return (
                        <div
                            ref={el => {
                                itemRefs.current[index] = el
                                if (el && !storedWidth && isVisible) {
                                    const width = el.offsetWidth
                                    if (width > 0) {
                                        updateItemWidth(index, width)
                                    }
                                }
                            }}
                            data-index={index}
                            key={!!(child as React.ReactElement)?.key ? (child as React.ReactElement)?.key : index}
                            className={cn(
                                CarouselAnatomy.item({ orientation, gap }),
                                isVisible && React.isValidElement(child) && child.props.containerClassName
                                    ? child.props.containerClassName.split(" ").filter((cls: string) => cls.includes("basis-")).join(" ")
                                    : "",
                            )}
                            style={{
                                ...(isVisible ? {} : {
                                    flexBasis: "auto",
                                    // width: storedWidth || estimatedWidth,
                                }),
                            }}
                        >
                            {isVisible ? (
                                React.isValidElement(child) && child.props.containerClassName ? (
                                    React.cloneElement(child as React.ReactElement, {
                                        containerClassName: child.props.containerClassName
                                            .split(" ")
                                            .filter((cls: string) => !cls.includes("basis-"))
                                            .join(" "),
                                    })
                                ) : child
                            ) : (
                                <LazyCarouselItemSkeleton
                                    width={storedWidth || estimatedWidth}
                                />
                            )}
                        </div>
                    )
                })}
            </div>
        </div>
    )
})
LazyCarouselContent.displayName = "LazyCarouselContent"

/* -------------------------------------------------------------------------------------------------
 * LazyCarouselItemSkeleton
 * -----------------------------------------------------------------------------------------------*/

type LazyCarouselItemSkeletonProps = {
    width: number
}

function LazyCarouselItemSkeleton({ width }: LazyCarouselItemSkeletonProps) {
    return (
        <div
            className="animate-pulse"
            style={{ width }}
        >
            <Skeleton
                className="w-full aspect-[2/3]"
            />
        </div>
    )
}

/* -------------------------------------------------------------------------------------------------
 * LazyCarouselWithDotButtons - Convenience component
 * -----------------------------------------------------------------------------------------------*/

export type LazyCarouselWithDotButtonsProps = LazyCarouselProps & {
    showDotButtons?: boolean
}

export function LazyCarouselWithDotButtons({
    showDotButtons = true,
    ...props
}: LazyCarouselWithDotButtonsProps) {
    return (
        <LazyCarousel {...props}>
            {showDotButtons && <CarouselDotButtons />}
            {props.children}
        </LazyCarousel>
    )
}
