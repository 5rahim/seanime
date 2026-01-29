"use client"
import { cn } from "@/components/ui/core/styling"
import { useDraggableScroll } from "@/hooks/use-draggable-scroll"
import { MdChevronLeft } from "react-icons/md"
import { MdChevronRight } from "react-icons/md"
import React, { useRef, useState } from "react"
import { useIsomorphicLayoutEffect, useUpdateEffect } from "react-use"

interface SliderProps {
    children?: React.ReactNode
    sliderClassName?: string
    containerClassName?: string
    onSlideEnd?: () => void
}

export const Slider: React.FC<SliderProps> = (props) => {

    const { children, onSlideEnd, ...rest } = props

    const ref = useRef<HTMLDivElement>() as React.MutableRefObject<HTMLInputElement>
    const { events } = useDraggableScroll(ref, {
        decayRate: 0.96,
        safeDisplacement: 15,
        applyRubberBandEffect: true,
    })

    const [isScrolledToLeft, setIsScrolledToLeft] = useState(true)
    const [isScrolledToRight, setIsScrolledToRight] = useState(false)
    const [showChevronRight, setShowChevronRight] = useState(false)

    const handleScroll = () => {
        const div = ref.current

        if (div) {
            const scrolledToLeft = div.scrollLeft === 0
            const scrolledToRight = div.scrollLeft + div.clientWidth === div.scrollWidth

            setIsScrolledToLeft(scrolledToLeft)
            setIsScrolledToRight(scrolledToRight)
        }
    }

    useUpdateEffect(() => {
        if (!isScrolledToLeft && isScrolledToRight) {
            onSlideEnd && onSlideEnd()
            const t = setTimeout(() => {
                const div = ref.current
                if (div) {
                    div.scrollTo({
                        left: div.scrollLeft + 500,
                        behavior: "smooth",
                    })
                }
            }, 1000)
            return () => clearTimeout(t)
        }
    }, [isScrolledToLeft, isScrolledToRight])

    function slideLeft() {
        const div = ref.current
        if (div) {
            div.scrollTo({
                left: div.scrollLeft - 500,
                behavior: "smooth",
            })
        }
    }

    function slideRight() {
        const div = ref.current
        if (div) {
            div.scrollTo({
                left: div.scrollLeft + 500,
                behavior: "smooth",
            })
        }
    }

    useIsomorphicLayoutEffect(() => {
        if (ref.current.clientWidth < ref.current.scrollWidth) {
            setShowChevronRight(true)
        } else {
            setShowChevronRight(false)
        }
    }, [ref.current])

    return (
        <div className={cn(
            "relative flex items-center lg:gap-2",
            props.containerClassName,
        )}>
            <div
                onClick={slideLeft}
                className={`flex items-center cursor-pointer hover:text-action absolute left-0 bg-gradient-to-r from-[--background] z-40 h-full w-16 hover:opacity-100 ${
                    !isScrolledToLeft ? "lg:visible" : "invisible"
                }`}
            >
                <MdChevronLeft className="w-7 h-7 stroke-2 mx-auto"/>
            </div>
            <div
                onScroll={handleScroll}
                className="flex max-w-full w-full space-x-3 overflow-x-scroll scrollbar-hide scroll"
                {...events}
                ref={ref}
            >
                {children}
            </div>
            <div
                onClick={slideRight}
                className={cn(
                    "flex items-center invisible cursor-pointer hover:text-action absolute right-0 bg-gradient-to-l from-[--background] z-40 h-full w-16 hover:opacity-100",
                    {
                        "lg:visible": !isScrolledToRight && showChevronRight,
                    })}
            >
                <MdChevronRight className="w-7 h-7 stroke-2 mx-auto"/>
            </div>
        </div>
    )
}
