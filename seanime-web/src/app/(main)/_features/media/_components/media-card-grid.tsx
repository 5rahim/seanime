import React from "react"
import { useWindowSize } from "react-use"
import { useDebounce } from "use-debounce"

type MediaCardGridProps = {
    children?: React.ReactNode
} & React.HTMLAttributes<HTMLDivElement>

export function MediaCardGrid(props: MediaCardGridProps) {

    const {
        children,
        ...rest
    } = props

    return (
        <>
            <div
                className="grid grid-cols-2 md:grid-cols-3 lg:grid-cols-4 xl:grid-cols-5 2xl:grid-cols-7 min-[2000px]:grid-cols-8 gap-4"
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
} & React.HTMLAttributes<HTMLDivElement>;

export function MediaCardLazyGrid({
    children,
    itemCount,
    ...rest
}: MediaCardLazyGridProps) {
    const [visibleItems, setVisibleItems] = React.useState<number[]>([])
    const [itemHeight, setItemHeight] = React.useState<number | null>(null)
    const [itemsPerRow, setItemsPerRow] = React.useState<number | null>(null)
    const containerRef = React.useRef<HTMLDivElement | null>(null)
    const itemRef = React.useRef<HTMLDivElement | null>(null)
    const loadMoreRef = React.useRef<HTMLDivElement | null>(null)

    const { width } = useWindowSize()
    const debouncedWidth = useDebounce(width, 500)

    // Initial load of items to measure their height and determine columns per row
    React.useLayoutEffect(() => {
        setVisibleItems(Array.from({ length: Math.min(itemCount, 8) }, (_, i) => i))
    }, [itemCount])

    // Calculate item height and items per row based on screen size
    React.useLayoutEffect(() => {
        if (itemRef.current) {
            const itemRect = itemRef.current.getBoundingClientRect()
            setItemHeight(itemRect.height)

            const containerWidth = containerRef.current?.clientWidth || window.innerWidth
            const colClasses = [
                { min: 2000, cols: 8 },
                { min: 1536, cols: 7 },
                { min: 1280, cols: 5 },
                { min: 1024, cols: 4 },
                { min: 768, cols: 3 },
                { min: 0, cols: 2 },
            ]

            const columns = colClasses.find(c => containerWidth >= c.min)?.cols || 2
            setItemsPerRow(columns)
        }
    }, [itemRef.current, debouncedWidth])

    // Update min-height of the container
    React.useLayoutEffect(() => {
        if (itemHeight && itemsPerRow) {
            const totalRows = Math.ceil(itemCount / itemsPerRow)
            const totalHeight = totalRows * itemHeight

            if (containerRef.current) {
                containerRef.current.style.minHeight = `${totalHeight}px`
            }
        }
    }, [itemHeight, itemsPerRow, itemCount, debouncedWidth])

    // Load more items using IntersectionObserver
    React.useEffect(() => {
        const observer = new IntersectionObserver(
            (entries) => {
                entries.forEach((entry) => {
                    if (entry.isIntersecting) {
                        setVisibleItems((prevVisibleItems) => {
                            const nextItems = Array.from(
                                { length: Math.min(itemCount - prevVisibleItems.length, 10) },
                                (_, i) => i + prevVisibleItems.length,
                            )
                            return [...prevVisibleItems, ...nextItems]
                        })
                    }
                })
            },
            {
                root: null,
                rootMargin: "200px",
                threshold: 0.1,
            },
        )

        if (loadMoreRef.current) {
            observer.observe(loadMoreRef.current)
        }

        return () => {
            if (loadMoreRef.current) {
                observer.unobserve(loadMoreRef.current)
            }
        }
    }, [itemCount])

    // Map visible items to children
    const visibleChildren = visibleItems.map((index) => (children as any)[index])

    return (
        <div ref={containerRef} {...rest}>
            <div
                className="grid grid-cols-2 md:grid-cols-3 lg:grid-cols-4 xl:grid-cols-5 2xl:grid-cols-7 min-[2000px]:grid-cols-8 gap-4"
            >
                {visibleChildren.map((child, index) => (
                    <div key={index} ref={index === 0 ? itemRef : null}>
                        {child}
                    </div>
                ))}
            </div>
            <div ref={loadMoreRef} style={{ height: "1px" }}></div>
        </div>
    )
}
