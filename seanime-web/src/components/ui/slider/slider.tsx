import { cn } from "../core/styling"
import * as SliderPrimitive from "@radix-ui/react-slider"
import { cva } from "class-variance-authority"
import * as React from "react"

/* -------------------------------------------------------------------------------------------------
 * Anatomy
 * -----------------------------------------------------------------------------------------------*/

export const SliderAnatomy = {
    root: cva([
        "UI-Slider__root",
        "relative shrink grow flex items-center",
    ], {
        variants: {
            orientation: {
                horizontal: "w-full h-[20px]",
                vertical: "h-full w-[20px]",
            },
        },
    }),
    track: cva([
        "UI-Slider__track",
        // hide overflow to make range rounded
        "relative grow bg-[--border] h-[3px] rounded overflow-hidden",
    ]),
    thumb: cva([
        "UI-Slider__thumb",
        "block h-[15px] w-[15px] rounded-[15px] bg-white",
    ]),
    range: cva([
        "UI-Slider__range",
        "absolute h-full bg-white",
    ]),
}

/* -------------------------------------------------------------------------------------------------
 * Slider
 * -----------------------------------------------------------------------------------------------*/

export type SliderProps = React.ComponentPropsWithoutRef<typeof SliderPrimitive.Root>

export const Slider = React.forwardRef<HTMLDivElement, SliderProps>((props, ref) => {
    const {
        className,
        orientation = "horizontal",
        ...rest
    } = props

    return (
        <SliderPrimitive.Root
            ref={ref}
            orientation={orientation}
            className={cn(
                SliderAnatomy.root({ orientation }),
                className,
            )}
            {...rest}
        >
            <SliderPrimitive.Track className={SliderAnatomy.track()}>
			    <SliderPrimitive.Range className={SliderAnatomy.range()} />
            </SliderPrimitive.Track>
            <SliderPrimitive.Thumb className={SliderAnatomy.thumb()} />
        </SliderPrimitive.Root>
    )
})

Slider.displayName = "Slider"
