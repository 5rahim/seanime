"use client"

import { createCalendar } from "@internationalized/date"
import { cn, defineStyleAnatomy, useUILocaleConfig } from "../core"
import { cva } from "class-variance-authority"
import _capitalize from "lodash/capitalize"
import { useRef } from "react"
import { useCalendar } from "react-aria"
import { CalendarStateOptions, useCalendarState } from "react-stately"
import { IconButton } from "../button"
import { CalendarGrid } from "./calendar-grid"

/* -------------------------------------------------------------------------------------------------
 * Anatomy
 * -----------------------------------------------------------------------------------------------*/

export const CalendarAnatomy = defineStyleAnatomy({
    container: cva("UI-Calendar__container inline-block text-gray-800 dark:text-gray-200"),
    header: cva("UI-Calendar__header flex items-center gap-2 pb-4 w-full justify-between"),
    title: cva("UI-Calendar__title flex-none font-bold text-lg ml-2 w-fit"),
})

/* -------------------------------------------------------------------------------------------------
 * Calendar
 * -----------------------------------------------------------------------------------------------*/

export function Calendar({ locale, ...props }: Omit<CalendarStateOptions, "createCalendar" | "locale"> & {
    locale?: string
}) {

    const { countryLocale } = useUILocaleConfig()
    const _locale = locale ?? countryLocale

    const state = useCalendarState({
        ...props,
        locale: _locale,
        createCalendar,
    })

    const ref = useRef<HTMLDivElement>(null)
    const {
        calendarProps,
        prevButtonProps: { onPress: prevButtonOnPress },
        nextButtonProps: { onPress: nextButtonOnPress },
    } = useCalendar(
        props,
        state,
    )

    return (
        <div {...calendarProps} ref={ref} className={cn(CalendarAnatomy.container())} tabIndex={0}>
            <div className={cn(CalendarAnatomy.header())}>
                <IconButton
                    size="sm"
                    intent="primary-subtle"
                    icon={(<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 16 16" fill="currentColor"
                                className="h-6 w-6">
                        <path
                            d="M9.78 12.78a.75.75 0 0 1-1.06 0L4.47 8.53a.75.75 0 0 1 0-1.06l4.25-4.25a.751.751 0 0 1 1.042.018.751.751 0 0 1 .018 1.042L6.06 8l3.72 3.72a.75.75 0 0 1 0 1.06Z"></path>
                    </svg>)}
                    rounded
                    onClick={e => {
                        e.preventDefault()
                        prevButtonOnPress && prevButtonOnPress(e as any)
                    }}
                />
                <h4 className={cn(CalendarAnatomy.title())}>
                    {_capitalize(
                        Intl.DateTimeFormat(_locale, { month: "long", year: "numeric", })
                            .format(state.visibleRange.start.toDate(state.timeZone))
                    )}
                </h4>
                <IconButton
                    size="sm"
                    intent="primary-subtle"
                    icon={(<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 16 16" fill="currentColor"
                                className="h-6 w-6">
                        <path
                            d="M6.22 3.22a.75.75 0 0 1 1.06 0l4.25 4.25a.75.75 0 0 1 0 1.06l-4.25 4.25a.751.751 0 0 1-1.042-.018.751.751 0 0 1-.018-1.042L9.94 8 6.22 4.28a.75.75 0 0 1 0-1.06Z"></path>
                    </svg>)}
                    rounded
                    onClick={e => {
                        e.preventDefault()
                        nextButtonOnPress && nextButtonOnPress(e as any)
                    }}
                />
            </div>
            <CalendarGrid state={state} locale={_locale} offset={{}}/>
        </div>
    )
}

Calendar.displayName = "Calendar"
