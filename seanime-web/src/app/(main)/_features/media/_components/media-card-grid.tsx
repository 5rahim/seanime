import { LuffyError } from "@/components/shared/luffy-error"
import { cn } from "@/components/ui/core/styling"
import { Skeleton } from "@/components/ui/skeleton"
import React from "react"

const gridClass = cn(
    "grid grid-cols-2 min-[768px]:grid-cols-3 min-[1080px]:grid-cols-4 min-[1320px]:grid-cols-5 min-[1750px]:grid-cols-6 min-[1850px]:grid-cols-7 min-[2000px]:grid-cols-8 gap-4",
)

type MediaCardGridProps = {
    children?: React.ReactNode
} & React.HTMLAttributes<HTMLDivElement>

export function MediaCardGrid(props: MediaCardGridProps) {

    const {
        children,
        ...rest
    } = props

    if (React.Children.toArray(children).length === 0) {
        return <LuffyError title={null}>
            <p>Nothing to see</p>
        </LuffyError>
    }

    return (
        <>
            <div
                data-media-card-grid
                className={cn(gridClass)}
                {...rest}
            >
                {children}
            </div>
        </>
    )
}

type MediaCardLazyGridProps = {
    children: React.ReactNode
    itemCount: number
    containerRef?: React.RefObject<HTMLElement>
} & React.HTMLAttributes<HTMLDivElement>;

export function MediaCardLazyGrid({
    children,
    itemCount,
    ...rest
}: MediaCardLazyGridProps) {
    if (itemCount === 0) {
        return <LuffyError title={null}>
            <p>Nothing to see</p>
        </LuffyError>
    }

    if (itemCount <= 48) {
        return (
            <MediaCardGrid {...rest}>
                {children}
            </MediaCardGrid>
        )
    }

    return (
        <MediaCardLazyGridRenderer itemCount={itemCount} {...rest}>
            {children}
        </MediaCardLazyGridRenderer>
    )
}

const colClasses = [
    { min: 0, cols: 2 },
    { min: 768, cols: 3 },
    { min: 1080, cols: 4 },
    { min: 1320, cols: 5 },
    { min: 1750, cols: 6 },
    { min: 1850, cols: 7 },
    { min: 2000, cols: 8 },
]

export function MediaCardLazyGridRenderer({
    children,
    itemCount,
    ...rest
}: MediaCardLazyGridProps) {
    const [visibleIndices, setVisibleIndices] = React.useState<Set<number>>(new Set())
    const [itemHeights, setItemHeights] = React.useState<Map<number, number>>(new Map())
    const gridRef = React.useRef<HTMLDivElement>(null)
    const itemRefs = React.useRef<(HTMLDivElement | null)[]>([])
    const observerRef = React.useRef<IntersectionObserver | null>(null)

    // Determine initial columns based on window width
    const initialColumns = React.useMemo(() =>
            colClasses.find(c => window.innerWidth >= c.min)?.cols ?? 8,
        [],
    )

    // Initialize visible indices with first row
    React.useEffect(() => {
        const initialVisibleIndices = new Set(
            Array.from(Array(Math.min(initialColumns, itemCount)).keys()),
        )
        setVisibleIndices(initialVisibleIndices)

        // Clear heights when component unmounts
        return () => {
            setItemHeights(new Map())
        }
    }, [initialColumns, itemCount])

    // Intersection Observer to track which items become visible
    React.useEffect(() => {
        if (!gridRef.current) return

        const observerOptions = {
            root: null,
            rootMargin: "200px 0px",
            threshold: 0,
        }

        observerRef.current = new IntersectionObserver((entries) => {
            entries.forEach(entry => {
                const index = parseInt(entry.target.getAttribute("data-index") ?? "-1")

                if (entry.isIntersecting) {
                    // Add to visible indices
                    setVisibleIndices(prev => {
                        const updated = new Set(prev)
                        updated.add(index)
                        return updated
                    })
                } else {
                    // Remove from visible indices when scrolled out
                    setVisibleIndices(prev => {
                        const updated = new Set(prev)
                        // Keep initial row always visible
                        if (index >= initialColumns) {
                            updated.delete(index)
                        }
                        return updated
                    })
                }
            })
        }, observerOptions)

        // Observe all items
        itemRefs.current.forEach(ref => {
            if (ref) observerRef.current?.observe(ref)
        })

        return () => {
            observerRef.current?.disconnect()
        }
    }, [itemCount, initialColumns])

    // Function to update item heights
    const updateItemHeight = React.useCallback((index: number, height: number) => {
        setItemHeights(prev => {
            const updated = new Map(prev)
            updated.set(index, height)
            return updated
        })
    }, [])

    return (
        <div data-media-card-lazy-grid-renderer {...rest}>
            <div data-media-card-lazy-grid className={cn(gridClass)} ref={gridRef}>
                {React.Children.map(children, (child, index) => {
                    const isVisible = visibleIndices.has(index)
                    const storedHeight = itemHeights.get(index)

                    return (
                        <div
                            data-media-card-lazy-grid-item
                            ref={el => itemRefs.current[index] = el}
                            data-index={index}
                            key={!!(child as React.ReactElement)?.key ? (child as React.ReactElement)?.key : index}
                            className="transition-all duration-300 ease-in-out"
                        >
                            {isVisible ? (
                                <div
                                    data-media-card-lazy-grid-item-content
                                    ref={(el) => {
                                        // Measure and store height when first rendered
                                        if (el && !storedHeight) {
                                            updateItemHeight(index, el.offsetHeight)
                                        }
                                    }}
                                >
                                    {child}
                                </div>
                            ) : (
                                <Skeleton
                                    data-media-card-lazy-grid-item-skeleton
                                    className="w-full"
                                    style={{
                                        height: storedHeight || "300px",
                                    }}
                                ></Skeleton>
                            )}
                        </div>
                    )
                })}
            </div>
        </div>
    )
}


// type MediaCardLazyGridProps = {
//     children: React.ReactNode
//     itemCount: number
// } & React.HTMLAttributes<HTMLDivElement>;
//
// export function MediaCardLazyGrid({
//     children,
//     itemCount,
//     ...rest
// }: MediaCardLazyGridProps) {
//
//     if (itemCount === 0) {
//         return <LuffyError title={null}>
//             <p>Nothing to see</p>
//         </LuffyError>
//     }
//
//     if (itemCount <= 48) {
//         return (
//             <MediaCardGrid {...rest}>
//                 {children}
//             </MediaCardGrid>
//         )
//     }
//
//     return (
//         <MediaCardLazyGridRenderer itemCount={itemCount} {...rest}>
//             {children}
//         </MediaCardLazyGridRenderer>
//     )
// }
//
// const colClasses = [
//     { min: 0, cols: 2 },
//     { min: 768, cols: 3 },
//     { min: 1080, cols: 4 },
//     { min: 1320, cols: 5 },
//     { min: 1750, cols: 6 },
//     { min: 1850, cols: 7 },
//     { min: 2000, cols: 8 },
// ]
//
// export function MediaCardLazyGridRenderer({
//     children,
//     itemCount,
//     ...rest
// }: MediaCardLazyGridProps) {
//
//     const itemRef = React.useRef<HTMLDivElement | null>(null)
//     const [itemHeight, setItemHeight] = React.useState<number | null>(null)
//
//     const [initialRenderArr] = React.useState(Array.from(Array(colClasses.find(c => window.innerWidth >= c.min)?.cols ?? 8).keys()))
//
//     // Render the first row of items
//     const [indicesToRender, setIndicesToRender] = React.useState<number[]>(initialRenderArr)
//
//     React.useLayoutEffect(() => {
//         if (itemRef.current) {
//             const itemRect = itemRef.current.getBoundingClientRect()
//             const itemHeight = itemRect.height
//             setItemHeight(itemHeight)
//             setIndicesToRender(Array.from(Array(itemCount).keys()))
//         }
//     }, [itemRef.current])
//
//     const visibleChildren = indicesToRender.map((index) => (children as any)[index])
//
//     return (
//         <div {...rest}>
//             <div
//                 className={cn(gridClass)}
//             >
//                 {visibleChildren.map((child, index) => (
//                     <MediaCardLazyGridItem
//                         key={!!(child as React.ReactElement)?.key ? (child as React.ReactElement)?.key : index}
//                         ref={index === 0 ? itemRef : null}
//                         itemHeight={itemHeight}
//                         initialRenderCount={initialRenderArr.length}
//                         index={index}
//                     >
//                         {child}
//                     </MediaCardLazyGridItem>
//                 ))}
//             </div>
//         </div>
//     )
// }
//
// const MediaCardLazyGridItem = React.forwardRef<HTMLDivElement, React.HTMLAttributes<HTMLDivElement> & {
//     itemHeight: number | null,
//     index: number,
//     initialRenderCount: number
// }>(({
//     children,
//     itemHeight,
//     initialRenderCount,
//     index,
//     ...rest
// }, mRef) => {
//     const ref = React.useRef<HTMLDivElement | null>(null)
//     const isInView = useInView(ref as any, {
//         margin: "200px",
//         once: true,
//     })
//
//     return (
//         <div ref={mergeRefs([mRef, ref])} {...rest}>
//             {(index < initialRenderCount || isInView) ? children : <div className="w-full" style={{ height: itemHeight || 0 }}></div>}
//         </div>
//
//     )
// })
