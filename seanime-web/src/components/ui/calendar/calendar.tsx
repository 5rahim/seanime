"use client"

import { cva } from "class-variance-authority"
import * as React from "react"
import { DayPicker } from "react-day-picker"
import { ButtonAnatomy } from "../button"
import { cn, ComponentAnatomy, defineStyleAnatomy } from "../core/styling"

/* -------------------------------------------------------------------------------------------------
 * Anatomy
 * -----------------------------------------------------------------------------------------------*/

export const CalendarAnatomy = defineStyleAnatomy({
    root: cva([
        "UI-Calendar__root",
        "p-3",
    ]),
    months: cva([
        "UI-Calendar__months",
        "flex flex-col sm:flex-row space-y-4 sm:space-x-4 sm:space-y-0",
    ]),
    month: cva([
        "UI-Calendar__month",
        "space-y-4",
    ]),
    caption: cva([
        "UI-Calendar__caption",
        "flex justify-center pt-1 relative items-center",
    ]),
    captionLabel: cva([
        "UI-Calendar__captionLabel",
        "text-sm font-medium",
    ]),
    nav: cva([
        "UI-Calendar__nav",
        "space-x-1 flex items-center",
    ]),
    navButton: cva([
        "UI-Calendar__navButton",
    ]),
    navButtonPrevious: cva([
        "UI-Calendar__navButtonPrevious",
        "absolute left-1",
    ]),
    navButtonNext: cva([
        "UI-Calendar__navButtonNext",
        "absolute right-1",
    ]),
    table: cva([
        "UI-Calendar__table",
        "w-full border-collapse space-y-1",
    ]),
    headRow: cva([
        "UI-Calendar__headRow",
        "flex",
    ]),
    headCell: cva([
        "UI-Calendar__headCell",
        "text-[--muted] rounded-[--radius] w-9 font-normal text-[0.8rem]",
    ]),
    row: cva([
        "UI-Calendar__row",
        "flex w-full mt-2",
    ]),
    cell: cva([
        "UI-Calendar__cell",
        "h-9 w-9 text-center text-sm p-0 relative",
        "[&:has([aria-selected].day-range-end)]:rounded-r-[--radius]",
        "[&:has([aria-selected].day-outside)]:bg-[--subtle]/50",
        "[&:has([aria-selected])]:bg-[--subtle]",
        "first:[&:has([aria-selected])]:rounded-l-[--radius]",
        "last:[&:has([aria-selected])]:rounded-r-[--radius]",
        "focus-within:relative focus-within:z-20",
    ]),
    day: cva([
        "UI-Calendar__day",
        "h-9 w-9 p-0 font-normal aria-selected:opacity-100",
    ]),
    dayRangeEnd: cva([
        "UI-Calendar__dayRangeEnd",
        "day-range-end",
    ]),
    daySelected: cva([
        "UI-Calendar__daySelected",
        "bg-brand text-white hover:bg-brand hover:text-white",
        "focus:bg-brand focus:text-white rounded-[--radius] font-semibold",
    ]),
    dayToday: cva([
        "UI-Calendar__dayToday",
        "bg-[--subtle] text-[--foreground] rounded-[--radius]",
    ]),
    dayOutside: cva([
        "UI-Calendar__dayOutside",
        "day-outside !text-[--muted] opacity-20",
        "aria-selected:bg-transparent",
        "aria-selected:opacity-30",
    ]),
    dayDisabled: cva([
        "UI-Calendar__dayDisabled",
        "text-[--muted] opacity-30",
    ]),
    dayRangeMiddle: cva([
        "UI-Calendar__dayRangeMiddle",
        "aria-selected:bg-[--subtle]",
        "aria-selected:text-[--foreground]",
    ]),
    dayHidden: cva([
        "UI-Calendar__dayHidden",
        "invisible",
    ]),
})

/* -------------------------------------------------------------------------------------------------
 * Calendar
 * -----------------------------------------------------------------------------------------------*/

export type CalendarProps =
    React.ComponentProps<typeof DayPicker> &
    ComponentAnatomy<typeof CalendarAnatomy>

export function Calendar(props: CalendarProps) {

    const {
        className,
        classNames,
        monthsClass,
        monthClass,
        captionClass,
        captionLabelClass,
        navClass,
        navButtonClass,
        navButtonPreviousClass,
        navButtonNextClass,
        tableClass,
        headRowClass,
        headCellClass,
        rowClass,
        cellClass,
        dayClass,
        dayRangeEndClass,
        daySelectedClass,
        dayTodayClass,
        dayOutsideClass,
        dayDisabledClass,
        dayRangeMiddleClass,
        dayHiddenClass,
        ...rest
    } = props

    return (
        <DayPicker
            fixedWeeks
            className={cn(CalendarAnatomy.root(), className)}
            classNames={{
                months: cn(CalendarAnatomy.months(), monthsClass),
                month: cn(CalendarAnatomy.month(), monthClass),
                caption: cn(CalendarAnatomy.caption(), captionClass),
                caption_label: cn(CalendarAnatomy.captionLabel(), captionLabelClass),
                nav: cn(CalendarAnatomy.nav(), navClass),
                nav_button: cn(CalendarAnatomy.navButton(), ButtonAnatomy.root({ size: "sm", intent: "gray-basic" }), navButtonClass),
                nav_button_previous: cn(CalendarAnatomy.navButtonPrevious(), navButtonPreviousClass),
                nav_button_next: cn(CalendarAnatomy.navButtonNext(), navButtonNextClass),
                table: cn(CalendarAnatomy.table(), tableClass),
                head_row: cn(CalendarAnatomy.headRow(), headRowClass),
                head_cell: cn(CalendarAnatomy.headCell(), headCellClass),
                row: cn(CalendarAnatomy.row(), rowClass),
                cell: cn(CalendarAnatomy.cell(), cellClass),
                day: cn(CalendarAnatomy.day(), dayClass),
                day_range_end: cn(CalendarAnatomy.dayRangeEnd(), dayRangeEndClass),
                day_selected: cn(CalendarAnatomy.daySelected(), daySelectedClass),
                day_today: cn(CalendarAnatomy.dayToday(), dayTodayClass),
                day_outside: cn(CalendarAnatomy.dayOutside(), dayOutsideClass),
                day_disabled: cn(CalendarAnatomy.dayDisabled(), dayDisabledClass),
                day_range_middle: cn(CalendarAnatomy.dayRangeMiddle(), dayRangeMiddleClass),
                day_hidden: cn(CalendarAnatomy.dayHidden(), dayHiddenClass),
                ...classNames,
            }}
            components={{
                IconLeft: ({ ...props }) => <svg
                    xmlns="http://www.w3.org/2000/svg"
                    viewBox="0 0 24 24"
                    fill="none"
                    stroke="currentColor"
                    strokeWidth="2"
                    strokeLinecap="round"
                    strokeLinejoin="round"
                    className="w-4 h-4"
                >
                    <path d="m15 18-6-6 6-6" />
                </svg>,
                IconRight: ({ ...props }) => <svg
                    xmlns="http://www.w3.org/2000/svg"
                    viewBox="0 0 24 24"
                    fill="none"
                    stroke="currentColor"
                    strokeWidth="2"
                    strokeLinecap="round"
                    strokeLinejoin="round"
                    className="rotate-180 w-4 h-4"
                >
                    <path d="m15 18-6-6 6-6" />
                </svg>,
            }}
            {...rest}
        />
    )
}

Calendar.displayName = "Calendar"
