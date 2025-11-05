"use client"

import { useThemeSettings } from "@/lib/theme/hooks"
import { cva } from "class-variance-authority"
import { EmblaCarouselType } from "embla-carousel"
import AutoScroll from "embla-carousel-autoplay"
import useEmblaCarousel, { UseEmblaCarouselType } from "embla-carousel-react"
import * as React from "react"
import { IconButton } from "../button"
import { cn, defineStyleAnatomy } from "../core/styling"

/* -------------------------------------------------------------------------------------------------
 * Anatomy
 * -----------------------------------------------------------------------------------------------*/

export const CarouselAnatomy = defineStyleAnatomy({
    root: cva([
        "UI-Carousel__root",
        "relative",
    ]),
    content: cva([
        "UI-Carousel__content",
        "overflow-hidden",
    ]),
    innerContent: cva([
        "UI-Carousel__innerContent",
        "flex",
    ], {
        variants: {
            gap: { none: null, sm: null, md: null, lg: null, xl: null },
            orientation: { horizontal: null, vertical: null },
        },
        compoundVariants: [
            { gap: "none", orientation: "horizontal", className: "ml-0" },
            { gap: "sm", orientation: "horizontal", className: "-ml-2" },
            { gap: "md", orientation: "horizontal", className: "-ml-4" },
            { gap: "lg", orientation: "horizontal", className: "-ml-6" },
            { gap: "xl", orientation: "horizontal", className: "-ml-8" },
            /**/
            { gap: "none", orientation: "vertical", className: "-mt-0 flex-col" },
            { gap: "sm", orientation: "vertical", className: "-mt-2 flex-col" },
            { gap: "md", orientation: "vertical", className: "-mt-4 flex-col" },
            { gap: "lg", orientation: "vertical", className: "-mt-6 flex-col" },
            { gap: "xl", orientation: "vertical", className: "-mt-8 flex-col" },
        ],
    }),
    item: cva([
        "UI-Carousel__item",
        "min-w-0 shrink-0 grow-0 basis-full",
    ], {
        variants: {
            gap: { none: null, sm: null, md: null, lg: null, xl: null },
            orientation: { horizontal: null, vertical: null },
        },
        compoundVariants: [
            { gap: "none", orientation: "horizontal", className: "pl-0" },
            { gap: "sm", orientation: "horizontal", className: "pl-2" },
            { gap: "md", orientation: "horizontal", className: "pl-4" },
            { gap: "lg", orientation: "horizontal", className: "pl-6" },
            { gap: "xl", orientation: "horizontal", className: "pl-8" },
            /**/
            { gap: "none", orientation: "vertical", className: "pt-0" },
            { gap: "sm", orientation: "vertical", className: "pt-2" },
            { gap: "md", orientation: "vertical", className: "pt-4" },
            { gap: "lg", orientation: "vertical", className: "pt-6" },
            { gap: "xl", orientation: "vertical", className: "pt-8" },
        ],
    }),
    button: cva([
        "UI-Carousel__button",
        "rounded-full z-[10]",
    ], {
        variants: {
            placement: { previous: null, next: null },
            orientation: { horizontal: null, vertical: null },
        },
        compoundVariants: [
            { placement: "previous", orientation: "horizontal", className: "" },
            { placement: "next", orientation: "horizontal", className: "" },

            { placement: "previous", orientation: "vertical", className: "-top-12 left-1/2 -translate-x-1/2 rotate-90" },
            { placement: "next", orientation: "vertical", className: "-bottom-12 left-1/2 -translate-x-1/2 rotate-90" },
        ],
    }),
    chevronIcon: cva([
        "UI-Carousel__chevronIcon",
        "size-6",
    ]),
})

/* -------------------------------------------------------------------------------------------------
 * Carousel
 * -----------------------------------------------------------------------------------------------*/

export const __CarouselContext = React.createContext<CarouselContextProps | null>(null)

export function useCarousel() {
    const context = React.useContext(__CarouselContext)

    if (!context) {
        throw new Error("useCarousel must be used within a <Carousel />")
    }

    return context
}

export type CarouselApi = UseEmblaCarouselType[1]
export type UseCarouselParameters = Parameters<typeof useEmblaCarousel>
export type CarouselOptions = UseCarouselParameters[0]
export type CarouselPlugin = UseCarouselParameters[1]


export type CarouselProps = {
    opts?: CarouselOptions
    plugins?: CarouselPlugin
    orientation?: "horizontal" | "vertical"
    gap?: "none" | "sm" | "md" | "lg" | "xl"
    setApi?: (api: EmblaCarouselType) => void
    autoScroll?: boolean
    autoScrollDelay?: number
}

type CarouselContextProps = {
    carouselRef: ReturnType<typeof useEmblaCarousel>[0]
    api: ReturnType<typeof useEmblaCarousel>[1]
    scrollPrev: () => void
    scrollNext: () => void
    canScrollPrev: boolean
    canScrollNext: boolean
} & CarouselProps

export const Carousel = React.forwardRef<HTMLDivElement, React.HTMLAttributes<HTMLDivElement> & CarouselProps>((props, ref) => {

    const {
        orientation = "horizontal",
        opts,
        gap = "md",
        setApi,
        plugins,
        className,
        children,
        autoScroll,
        autoScrollDelay = 5000,
        ...rest
    } = props

    const ts = useThemeSettings()

    const _plugins = React.useMemo(() => {
        return [
            ...(plugins || []),
            ...((autoScroll && !ts.disableCarouselAutoScroll) ? [AutoScroll({
                delay: autoScrollDelay,
                stopOnMouseEnter: true,
                stopOnInteraction: false,
            })] : []),
        ]
    }, [plugins, autoScroll, ts.disableCarouselAutoScroll])

    const [carouselRef, api] = useEmblaCarousel({ ...opts, axis: orientation === "horizontal" ? "x" : "y" }, _plugins)
    const [canScrollPrev, setCanScrollPrev] = React.useState(false)
    const [canScrollNext, setCanScrollNext] = React.useState(false)

    const onSelect = React.useCallback((api: EmblaCarouselType) => {
        if (!api) return

        setCanScrollPrev(api.canScrollPrev())
        setCanScrollNext(api.canScrollNext())
    }, [])

    const scrollPrev = React.useCallback(() => {
        api?.scrollPrev()
    }, [api])

    const scrollNext = React.useCallback(() => {
        api?.scrollNext()
    }, [api])

    const handleKeyDown = React.useCallback(
        (event: React.KeyboardEvent<HTMLDivElement>) => {
            if (event.key === "ArrowLeft") {
                event.preventDefault()
                scrollPrev()
            } else if (event.key === "ArrowRight") {
                event.preventDefault()
                scrollNext()
            }
        },
        [scrollPrev, scrollNext],
    )

    React.useEffect(() => {
        if (!api || !setApi) return

        setApi(api)
    }, [api, setApi])

    React.useEffect(() => {
        if (!api) return

        onSelect(api)
        api.on("reInit", onSelect)
        api.on("select", onSelect)

        return () => {
            api?.off("select", onSelect)
        }
    }, [api, onSelect])

    return (
        <__CarouselContext.Provider
            value={{
                carouselRef,
                api: api,
                opts,
                gap,
                orientation: orientation || (opts?.axis === "y" ? "vertical" : "horizontal"),
                scrollPrev,
                scrollNext,
                canScrollPrev,
                canScrollNext,
            }}
        >
            <div
                ref={ref}
                onKeyDownCapture={handleKeyDown}
                className={cn(CarouselAnatomy.root(), className)}
                role="region"
                aria-roledescription="carousel"
                {...rest}
            >
                {children}
                <CarouselButtons />
            </div>
        </__CarouselContext.Provider>
    )
})
Carousel.displayName = "Carousel"

type CarouselButtonsProps = {
    children?: React.ReactNode
}

export function CarouselButtons(props: CarouselButtonsProps) {

    const {
        children,
        ...rest
    } = props

    const { scrollSnaps } = useDotButton()

    return (
        <>
            {scrollSnaps.length > 30 && <div className="flex gap-2 absolute top-[-3.5rem] right-0">
                <CarouselPrevious />
                <CarouselNext />
            </div>}
        </>
    )
}


/* -------------------------------------------------------------------------------------------------
 * CarouselContent
 * -----------------------------------------------------------------------------------------------*/

export type CarouselContentProps = React.ComponentPropsWithoutRef<"div"> & {
    contentClass?: string
}

export const CarouselContent = React.forwardRef<HTMLDivElement, CarouselContentProps>((props, ref) => {
    const { className, contentClass, ...rest } = props
    const { carouselRef, orientation, gap } = useCarousel()

    return (
        <div ref={carouselRef} className={cn(CarouselAnatomy.content(), contentClass)}>
            <div
                ref={ref}
                className={cn(CarouselAnatomy.innerContent({ orientation, gap }), className)}
                {...rest}
            />
        </div>
    )
})
CarouselContent.displayName = "CarouselContent"

/* -------------------------------------------------------------------------------------------------
 * CarouselItem
 * -----------------------------------------------------------------------------------------------*/

export type CarouselItemProps = React.ComponentPropsWithoutRef<"div">

export const CarouselItem = React.forwardRef<HTMLDivElement, CarouselItemProps>((props, ref) => {
    const { className, ...rest } = props
    const { orientation, gap } = useCarousel()

    return (
        <div
            ref={ref}
            role="group"
            aria-roledescription="slide"
            className={cn(CarouselAnatomy.item({ orientation, gap }), className)}
            {...rest}
        />
    )
})
CarouselItem.displayName = "CarouselItem"

/* -------------------------------------------------------------------------------------------------
 * CarouselPrevious
 * -----------------------------------------------------------------------------------------------*/

export type CarouselButtonProps = React.ComponentProps<typeof IconButton> & { chevronIconClass?: string }

export const CarouselPrevious = React.forwardRef<HTMLButtonElement, CarouselButtonProps>((props, ref) => {
    const { className, chevronIconClass, intent = "gray-outline", ...rest } = props
    const { orientation, scrollPrev, canScrollPrev } = useCarousel()

    return (
        <IconButton
            ref={ref}
            intent={intent}
            className={CarouselAnatomy.button({ orientation, placement: "previous" })}
            disabled={!canScrollPrev}
            onClick={scrollPrev}
            icon={<svg
                xmlns="http://www.w3.org/2000/svg"
                viewBox="0 0 24 24"
                fill="none"
                stroke="currentColor"
                strokeWidth="2"
                strokeLinecap="round"
                strokeLinejoin="round"
                className={cn(CarouselAnatomy.chevronIcon(), chevronIconClass)}
            >
                <path d="m15 18-6-6 6-6" />
            </svg>}
            {...rest}
        />
    )
})
CarouselPrevious.displayName = "CarouselPrevious"

/* -------------------------------------------------------------------------------------------------
 * CarouselNext
 * -----------------------------------------------------------------------------------------------*/

export const CarouselNext = React.forwardRef<HTMLButtonElement, CarouselButtonProps>((props, ref) => {
    const { className, chevronIconClass, intent = "gray-outline", ...rest } = props
    const { orientation, scrollNext, canScrollNext } = useCarousel()

    return (
        <IconButton
            ref={ref}
            intent={intent}
            className={CarouselAnatomy.button({ orientation, placement: "next" })}
            disabled={!canScrollNext}
            onClick={scrollNext}
            icon={<svg
                xmlns="http://www.w3.org/2000/svg"
                viewBox="0 0 24 24"
                fill="none"
                stroke="currentColor"
                strokeWidth="2"
                strokeLinecap="round"
                strokeLinejoin="round"
                className={cn(CarouselAnatomy.chevronIcon(), chevronIconClass)}
            >
                <path d="m9 18 6-6-6-6" />
            </svg>}
            {...rest}
        />
    )
})
CarouselNext.displayName = "CarouselNext"

/* -------------------------------------------------------------------------------------------------
 *
 * -----------------------------------------------------------------------------------------------*/

type UseDotButtonType = {
    selectedIndex: number
    scrollSnaps: number[]
    onDotButtonClick: (index: number) => void
}


export const useDotButton = (): UseDotButtonType => {
    const { api: emblaApi } = useCarousel()
    const [selectedIndex, setSelectedIndex] = React.useState(0)
    const [scrollSnaps, setScrollSnaps] = React.useState<number[]>([])

    const onDotButtonClick = React.useCallback(
        (index: number) => {
            if (!emblaApi) return
            emblaApi.scrollTo(index)
        },
        [emblaApi],
    )

    const onInit = React.useCallback((emblaApi: EmblaCarouselType) => {
        setScrollSnaps(emblaApi.scrollSnapList())
    }, [])

    const onSelect = React.useCallback((emblaApi: EmblaCarouselType) => {
        setSelectedIndex(emblaApi.selectedScrollSnap())
    }, [])

    React.useEffect(() => {
        if (!emblaApi) return

        onInit(emblaApi)
        onSelect(emblaApi)
        emblaApi.on("reInit", onInit)
        emblaApi.on("reInit", onSelect)
        emblaApi.on("select", onSelect)
    }, [emblaApi, onInit, onSelect])

    return {
        selectedIndex,
        scrollSnaps,
        onDotButtonClick,
    }
}

const DotButton = (props: React.ComponentPropsWithoutRef<"div">) => {
    const { children, className, onClick, ...rest } = props

    return (
        <div
            className={cn(
                "rounded-full px-[2.5px] size-5 lg:w-5 lg:h-4 group/dot-button flex items-center justify-center cursor-pointer",
            )}
            onClick={onClick}
        >
            {/* {children} */}
            <div
                className={cn(
                    "cursor-pointer w-full lg:h-2 lg:group-hover/dot-button:h-3 transition-all duration-300 rounded-full bg-gray-600 dark:bg-opacity-50",
                    className)} {...rest} />
        </div>
    )
}


export const CarouselDotButtons = (props: { className?: string, flag?: any }) => {

    const { selectedIndex, scrollSnaps, onDotButtonClick } = useDotButton()

    React.useEffect(() => {
        onDotButtonClick(0)
    }, [onDotButtonClick, scrollSnaps, props.flag])

    if (scrollSnaps.length > 30) return null

    return (
        <div className={cn("absolute -top-8 right-0 hidden md:flex items-center z-[10]", props.className)}>
            {scrollSnaps.map((_, index) => (
                <DotButton
                    key={index}
                    onClick={() => onDotButtonClick(index)}
                    className={cn(
                        { "bg-white": index === selectedIndex },
                        { "size-4 [&:nth-child(odd)]:hidden": scrollSnaps.length > 30 },
                    )}
                />
            ))}
        </div>
    )
}


export const CarouselMasks = () => {

    const ts = useThemeSettings()

    if (!!ts.libraryScreenCustomBackgroundImage) return null

    return (
        <>
            <div className="absolute hidden md:block left-0 h-full w-8 bg-gradient-to-r from-[--background] to-transparent z-[1]" />
            <div className="absolute hidden md:block right-0 h-full w-8 bg-gradient-to-l from-[--background] to-transparent z-[1]" />
        </>
    )
}
