"use client"

import { createCalendar } from "@internationalized/date"
import { cn, defineStyleAnatomy, useUILocaleConfig } from "../core"
import { cva } from "class-variance-authority"
import { useRef } from "react"
import { useDateField, useDateSegment } from "react-aria"
import { DateFieldState, DateFieldStateOptions, DateSegment, useDateFieldState } from "react-stately"

/* -------------------------------------------------------------------------------------------------
 * Anatomy
 * -----------------------------------------------------------------------------------------------*/

export const DateSegmentAnatomy = defineStyleAnatomy({
    segment: cva([
        "UI-DateSegment__segment",
        "box-content tabular-nums text-right outline-none rounded-sm",
        "group shadow-none",
        "focus:font-bold focus:text-[--brand] dark:focus:text-[--brand]",
    ], {
        variants: {
            isEditable: {
                false: "text-gray-500",
                true: "text-gray-800 dark:text-gray-200",
            },
        },
    }),
    input: cva([
        "UI-DateSegment__input",
        "block w-full text-center italic text-gray-500 group-focus:text-brand-500 dark:group-focus:text-white group-focus:font-semibold"
    ]),
})

/* -------------------------------------------------------------------------------------------------
 * DateField
 * -----------------------------------------------------------------------------------------------*/

export function DateField({ locale, ...props }: Omit<DateFieldStateOptions, "locale" | "createCalendar"> & {
    locale?: string
}) {
    const { countryLocale } = useUILocaleConfig()

    const state = useDateFieldState({
        ...props,
        locale: locale ?? countryLocale,
        createCalendar,
    })

    const ref = useRef<HTMLDivElement>(null)
    const { fieldProps } = useDateField(props, state, ref)

    return (
        <div {...fieldProps} ref={ref} className="flex">
            {state.segments.map((segment, i) => (
                <DateSegmentComponent key={i} segment={segment} state={state}/>
            ))}
        </div>
    )
}

DateField.displayName = "DateField"


/* -------------------------------------------------------------------------------------------------
 * DateSegmentComponent
 * -----------------------------------------------------------------------------------------------*/

export function DateSegmentComponent({ segment, state }: { segment: DateSegment, state: DateFieldState }) {
    const ref = useRef<HTMLDivElement>(null)
    const { segmentProps } = useDateSegment(segment, state, ref)

    return (
        <div
            {...segmentProps}
            ref={ref}
            style={{
                ...segmentProps.style,
            }}
            className={cn(DateSegmentAnatomy.segment({ isEditable: segment.isEditable }))}
            suppressHydrationWarning
        >
            <span
                aria-hidden="true" className={cn(DateSegmentAnatomy.input())} style={{
                display: segment.isPlaceholder ? undefined : "none", height: segment.isPlaceholder ? undefined : 0,
                pointerEvents: "none",
            }}
            >
                {segment.placeholder}
            </span>
            {segment.isPlaceholder ? null : segment.text}
        </div>
    )
}

DateSegmentComponent.displayName = "DateSegmentComponent"
