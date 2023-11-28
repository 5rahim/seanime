"use client"

import _capitalize from "lodash/capitalize"
import React from "react"
import { AriaButtonProps, VisuallyHidden } from "react-aria"
import { CalendarState, RangeCalendarState } from "react-stately"
import { IconButton } from "../button"
import { useUILocaleConfig } from "../core"

/* -------------------------------------------------------------------------------------------------
 * CalendarHeader
 * -----------------------------------------------------------------------------------------------*/

interface CalendarHeaderProps {
    state: CalendarState | RangeCalendarState
    calendarProps: any
    prevButtonProps: AriaButtonProps
    nextButtonProps: AriaButtonProps
    locale?: string
}

export function CalendarHeader(
    {
        state,
        calendarProps,
        prevButtonProps,
        nextButtonProps,
        locale,
    }: CalendarHeaderProps,
) {
    const { countryLocale } = useUILocaleConfig()
    const _locale = locale ?? countryLocale

    const { onPress: prevButtonOnPress, ...prevButtonRest } = prevButtonProps
    const { onPress: nextButtonOnPress, ...nextButtonRest } = nextButtonProps

    return (
        <div className="flex items-center py-4 text-center">
            {/* Add a screen reader only description of the entire visible range rather than
          * a separate heading above each month grid. This is placed first in the DOM order
          * so that it is the first thing a touch screen reader user encounters.
          * In addition, VoiceOver on iOS does not announce the aria-label of the grid
          * elements, so the aria-label of the Calendar is included here as well. */}
            <VisuallyHidden>
                <h2>{calendarProps["aria-label"]}</h2>
            </VisuallyHidden>
            <IconButton
                size="sm"
                intent="primary-subtle"
                icon={(
                    <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 16 16" fill="currentColor" className="h-6 w-6">
                        <path
                            d="M9.78 12.78a.75.75 0 0 1-1.06 0L4.47 8.53a.75.75 0 0 1 0-1.06l4.25-4.25a.751.751 0 0 1 1.042.018.751.751 0 0 1 .018 1.042L6.06 8l3.72 3.72a.75.75 0 0 1 0 1.06Z"></path>
                    </svg>)}
                rounded {...prevButtonRest} onClick={e => {
                e.preventDefault()
                prevButtonOnPress && prevButtonOnPress(e as any)
            }}
            />
            <h4
                // We have a visually hidden heading describing the entire visible range,
                // and the calendar itself describes the individual month
                // so we don't need to repeat that here for screen reader users.
                aria-hidden
                className="flex-1 align-center font-bold text-md text-center"
            >
                {_capitalize(Intl.DateTimeFormat((_locale), {
                    month: "long", year: "numeric",
                }).format(state.visibleRange.start.toDate(state.timeZone)))}
            </h4>
            <h4
                aria-hidden
                className="flex-1 align-center font-bold text-md text-center"
            >
                {_capitalize(Intl.DateTimeFormat((_locale), {
                    month: "long", year: "numeric",
                }).format(state.visibleRange.start.add({ months: 1 }).toDate(state.timeZone)))}
            </h4>
            <IconButton
                size="sm"
                intent="primary-subtle"
                icon={(
                    <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 16 16" fill="currentColor" className="h-6 w-6">
                        <path
                            d="M6.22 3.22a.75.75 0 0 1 1.06 0l4.25 4.25a.75.75 0 0 1 0 1.06l-4.25 4.25a.751.751 0 0 1-1.042-.018.751.751 0 0 1-.018-1.042L9.94 8 6.22 4.28a.75.75 0 0 1 0-1.06Z"></path>
                    </svg>)}
                rounded {...nextButtonRest} onClick={e => {
                e.preventDefault()
                nextButtonOnPress && nextButtonOnPress(e as any)
            }}
            />
        </div>
    )
}

CalendarHeader.displayName = "CalendarHeader"
