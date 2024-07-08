import { LuffyError } from "@/components/shared/luffy-error"
import { cn } from "@/components/ui/core/styling"
import { mergeRefs } from "@/components/ui/core/utils"
import { useInView } from "framer-motion"
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

    const itemRef = React.useRef<HTMLDivElement | null>(null)
    const [itemHeight, setItemHeight] = React.useState<number | null>(null)

    const [initialRenderArr] = React.useState(Array.from(Array(colClasses.find(c => window.innerWidth >= c.min)?.cols ?? 8).keys()))

    // Render the first row of items
    const [indicesToRender, setIndicesToRender] = React.useState<number[]>(initialRenderArr)

    React.useLayoutEffect(() => {
        if (itemRef.current) {
            const itemRect = itemRef.current.getBoundingClientRect()
            const itemHeight = itemRect.height
            setItemHeight(itemHeight)
            setIndicesToRender(Array.from(Array(itemCount).keys()))
        }
    }, [itemRef.current])

    const visibleChildren = indicesToRender.map((index) => (children as any)[index])

    return (
        <div {...rest}>
            <div
                className={cn(gridClass)}
            >
                {visibleChildren.map((child, index) => (
                    <MediaCardLazyGridItem
                        key={!!(child as React.ReactElement)?.key ? (child as React.ReactElement)?.key : index}
                        ref={index === 0 ? itemRef : null}
                        itemHeight={itemHeight}
                        initialRenderCount={initialRenderArr.length}
                        index={index}
                    >
                        {child}
                    </MediaCardLazyGridItem>
                ))}
            </div>
        </div>
    )
}

const MediaCardLazyGridItem = React.forwardRef<HTMLDivElement, React.HTMLAttributes<HTMLDivElement> & {
    itemHeight: number | null,
    index: number,
    initialRenderCount: number
}>(({
    children,
    itemHeight,
    initialRenderCount,
    index,
    ...rest
}, mRef) => {
    const ref = React.useRef<HTMLDivElement | null>(null)
    const isInView = useInView(ref as any, {
        margin: "200px",
        once: true,
    })

    return (
        <div ref={mergeRefs([mRef, ref])} {...rest}>
            {(index < initialRenderCount || isInView) ? children : <div className="w-full" style={{ height: itemHeight || 0 }}></div>}
        </div>

    )
})
