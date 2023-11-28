"use client"

import { createCalendar } from "@internationalized/date"
import { useRef } from "react"
import { useRangeCalendar } from "react-aria"
import { RangeCalendarStateOptions, useRangeCalendarState } from "react-stately"
import { useUILocaleConfig } from "../core"
import { CalendarGrid } from "./calendar-grid"
import { CalendarHeader } from "./calendar-header"

/* -------------------------------------------------------------------------------------------------
 * RangeCalendar
 * -----------------------------------------------------------------------------------------------*/

export function RangeCalendar({ locale, ...props }: Omit<RangeCalendarStateOptions, "createCalendar" | "locale"> & {
    locale?: string
}) {
    const { countryLocale } = useUILocaleConfig()
    const _locale = locale ?? countryLocale

    const state = useRangeCalendarState({
        ...props,
        visibleDuration: { months: 2 },
        locale: _locale,
        createCalendar,
    })

    const ref = useRef<HTMLDivElement>(null)
    const {
        calendarProps,
        prevButtonProps,
        nextButtonProps,
    } = useRangeCalendar(
        props,
        state,
        ref,
    )

    return (
        <div {...calendarProps} ref={ref} className="inline-block">
            <CalendarHeader
                state={state}
                calendarProps={calendarProps}
                prevButtonProps={prevButtonProps}
                nextButtonProps={nextButtonProps}
                locale={_locale}
            />
            <div className="flex items-center gap-2 pb-4 w-fit">
                <div className="flex flex-col md:flex-row gap-8">
                    <CalendarGrid state={state} offset={{}} locale={_locale}/>
                    <CalendarGrid state={state} offset={{ months: 1 }} locale={_locale}/>
                </div>
            </div>
        </div>
    )
}

RangeCalendar.displayName = "RangeCalendar"
