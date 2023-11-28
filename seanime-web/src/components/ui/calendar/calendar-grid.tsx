"use client"

import { DateDuration, endOfMonth, getWeeksInMonth } from "@internationalized/date"
import { cn, defineStyleAnatomy, useUILocaleConfig } from "../core"
import { cva } from "class-variance-authority"
import { useMemo } from "react"
import { useCalendarGrid } from "react-aria"
import { CalendarState, RangeCalendarState } from "react-stately"
import { CalendarCell } from "./calendar-cell"
import { getShortenedWeekDays } from "./locale"

/* -------------------------------------------------------------------------------------------------
 * Anatomy
 * -----------------------------------------------------------------------------------------------*/

export const CalendarGridAnatomy = defineStyleAnatomy({
    table: cva("UI-CalendarGrid__table flex-1"),
    head: cva("UI-CalendarGrid__head text-gray-600 text-center"),
})

/* -------------------------------------------------------------------------------------------------
 * CalendarGrid
 * -----------------------------------------------------------------------------------------------*/

interface CalendarGridProps {
    state: CalendarState | RangeCalendarState
    offset: DateDuration
    locale?: string
}

export function CalendarGrid({ locale, state, offset = {} }: CalendarGridProps) {
    const { countryLocale } = useUILocaleConfig()
    const _locale = locale ?? countryLocale

    const startDate = state.visibleRange.start.add(offset)
    const endDate = endOfMonth(startDate)

    let { gridProps, headerProps, weekDays } = useCalendarGrid(
        { startDate, endDate },
        state,
    )

    // Get the number of weeks in the month, so we can render the proper number of rows.
    const weeksInMonth = useMemo(() => getWeeksInMonth(startDate, _locale), [startDate, _locale])

    weekDays = useMemo(() => getShortenedWeekDays(weekDays, _locale), [_locale])

    return (
        <table {...gridProps} cellPadding="0" className={cn(CalendarGridAnatomy.table())}>
            <thead {...headerProps} className={cn(CalendarGridAnatomy.head())}>
            <tr>
                {weekDays.map((day, index) => (
                    <th key={index}>{day}</th>
                ))}
            </tr>
            </thead>
            <tbody>
            {[...new Array(weeksInMonth).keys()].map((weekIndex) => (
                <tr key={weekIndex}>
                    {state
                        .getDatesInWeek(weekIndex, startDate)
                        .map((date, i) =>
                            date ? (
                                <CalendarCell
                                    key={i}
                                    state={state}
                                    date={date}
                                    currentMonth={startDate}
                                />
                            ) : (
                                <td key={i}/>
                            ),
                        )}
                </tr>
            ))}
            </tbody>
        </table>
    )
}

CalendarGrid.displayName = "CalendarGrid"
