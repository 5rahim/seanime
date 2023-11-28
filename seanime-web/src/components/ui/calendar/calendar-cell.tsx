"use client"

import { DateValue, getDayOfWeek, isSameDay, isSameMonth } from "@internationalized/date"
import { cn, defineStyleAnatomy, useUILocaleConfig } from "../core"
import { cva } from "class-variance-authority"
import { useRef } from "react"
import { mergeProps, useCalendarCell, useFocusRing } from "react-aria"
import { CalendarState, RangeCalendarState } from "react-stately"

/* -------------------------------------------------------------------------------------------------
 * Anatomy
 * -----------------------------------------------------------------------------------------------*/

export const CalendarCellAnatomy = defineStyleAnatomy({
    cell: cva("UI-CalendarCell__cell relative", {
        variants: {
            isFocusVisible: { true: "z-10", false: "z-0" },
        },
    }),
    date: cva([
        "UI-CalendarCell__date",
        "w-full h-full rounded-full flex items-center justify-center text-gray-600 dark:text-gray-300 font-medium"
    ], {
        variants: {
            isDisabled: { true: "text-gray-400 cursor-default", false: null },
            isUnavailable: { true: "text-red-300 cursor-default", false: null },
            isSelectionStart: { true: "bg-brand-600 text-white hover:bg-brand-700", false: null },
            isSelectionEnd: { true: "bg-brand-600 text-white hover:bg-brand-700", false: null },
            isSelected: { true: null, false: null },
            isFocusVisible: { true: "ring-2 group-focus:z-2 ring-[--ring] ring-offset-2", false: null },
        },
        compoundVariants: [
            { isDisabled: false, isUnavailable: false, className: "cursor-pointer" },
            { isSelected: true, isSelectionStart: false, isSelectionEnd: false, className: "hover:bg-brand-400" },
            { isSelected: false, isDisabled: false, isUnavailable: false, className: "hover:bg-brand-100" },
        ],
    }),
    button: cva("UI-CalendarCell__button w-10 h-10 outline-none group", {
        variants: {
            isRoundedLeft: { true: "rounded-l-full", false: null },
            isRoundedRight: { true: "rounded-r-full", false: null },
            isSelected: { true: "bg-brand-100 dark:bg-opacity-10", false: null },
            isDisabled: { true: "disabled", false: null },
            isUnavailable: { true: "disabled", false: null },
        },
    }),
})

/* -------------------------------------------------------------------------------------------------
 * CalendarCell
 * -----------------------------------------------------------------------------------------------*/

interface CalendarCellProps {
    state: CalendarState | RangeCalendarState
    date: any
    currentMonth: DateValue
    locale?: string
}

export function CalendarCell({ state, date, currentMonth, locale }: CalendarCellProps) {
    const { countryLocale } = useUILocaleConfig()
    const _locale = locale ?? countryLocale

    const ref = useRef<HTMLDivElement>(null)
    const {
        cellProps,
        buttonProps,
        isSelected,
        isDisabled,
        isUnavailable,
        formattedDate,
    } = useCalendarCell({ date }, state, ref)

    const isOutsideMonth = !isSameMonth(currentMonth, date)

    // The start and end date of the selected range will have
    // an emphasized appearance.
    const isSelectionStart = (state as RangeCalendarState).highlightedRange
        ? isSameDay(date, (state as RangeCalendarState).highlightedRange.start)
        : isSelected
    const isSelectionEnd = (state as RangeCalendarState).highlightedRange
        ? isSameDay(date, (state as RangeCalendarState).highlightedRange.end)
        : isSelected

    // We add rounded corners on the left for the first day of the month,
    // the first day of each week, and the start date of the selection.
    // We add rounded corners on the right for the last day of the month,
    // the last day of each week, and the end date of the selection.
    const dayOfWeek = getDayOfWeek(date, _locale)
    const isRoundedLeft =
        isSelected && (isSelectionStart)
    const isRoundedRight =
        isSelected &&
        (isSelectionEnd)

    const { focusProps, isFocusVisible } = useFocusRing()

    return (
        <td
            {...cellProps}
            className={cn(CalendarCellAnatomy.cell({ isFocusVisible }))}
        >
            <div
                {...mergeProps(buttonProps, focusProps)}
                ref={ref}
                hidden={isOutsideMonth}
                className={cn(CalendarCellAnatomy.button({
                    isDisabled,
                    isSelected,
                    isUnavailable,
                    isRoundedLeft,
                    isRoundedRight
                }))}
            >
                <div
                    className={cn(CalendarCellAnatomy.date({
                        isSelected, isSelectionEnd, isSelectionStart, isUnavailable, isDisabled, isFocusVisible,
                    }))}
                >
                    {formattedDate}
                </div>
            </div>
        </td>
    )
}

CalendarCell.displayName = "CalendarCell"
