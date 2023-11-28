"use client"

import React, { useCallback, useMemo } from "react"
import { cn, ComponentWithAnatomy, defineStyleAnatomy, useUILocaleConfig } from "../core"
import { cva } from "class-variance-authority"
import { DataGridAnatomy, DataGridFilteringHelper, getColumnHelperMeta, getValueFormatter } from "."
import { Select } from "../select"
import { Column } from "@tanstack/react-table"
import { CloseButton } from "../button"
import { DropdownMenu } from "../dropdown-menu"
import { CheckboxGroup } from "../checkbox"
import { RadioGroup } from "../radio-group"
import { getLocalTimeZone, parseAbsoluteToLocal } from "@internationalized/date"
import { DateRangePicker } from "../date-time"
import locales from "./locales.json"

/* -------------------------------------------------------------------------------------------------
 * Anatomy
 * -----------------------------------------------------------------------------------------------*/

export const DataGridFilterAnatomy = defineStyleAnatomy({
    root: cva([
        "UI-DataGridFilter__root",
        "flex gap-2 items-center",
    ]),
})

export const DataGridActiveFilterAnatomy = defineStyleAnatomy({
    root: cva([
        "UI-DataGridActiveFilter__root",
        "py-1 px-2 rounded-[--radius] border border-[--border] flex gap-2 items-center",
    ]),
})

/* -------------------------------------------------------------------------------------------------
 * DataGridFilter
 * -----------------------------------------------------------------------------------------------*/

export interface DataGridFilterProps<T extends Record<string, any>> extends React.ComponentPropsWithoutRef<"div">,
    ComponentWithAnatomy<typeof DataGridFilterAnatomy> {
    column: Column<T>
    onRemove: () => void
}

export function DataGridFilter<T extends Record<string, any>>(props: DataGridFilterProps<T>) {

    const { locale } = useUILocaleConfig()

    const {
        children,
        rootClassName,
        className,
        column,
        onRemove,
        ...rest
    } = props

    const filterParams = getColumnHelperMeta(column, "filteringMeta")!
    const filterValue = useMemo(() => column.getFilterValue(), [column.getFilterValue()]) as any
    const setFilterValue = useMemo(() => column.setFilterValue, [column.setFilterValue])
    const icon = filterParams.icon

    // Value formatter - if undefined, use the default behavior
    const valueFormatter = filterParams.valueFormatter || getValueFormatter(column)

    // Get the options
    const options = filterParams.options ?? []

    // Update handler
    const handleUpdate = useCallback((value: any) => {
        setFilterValue(value)
    }, [])

    return (
        <div
            className={cn(DataGridFilterAnatomy.root(), rootClassName, className)}
            {...rest}
        >
            {(filterParams.type === "select" && (!options || options.length === 0)) && (
                <div className={"text-red-500"}>/!\ "Select" filtering option passed without options</div>
            )}
            {/*Select*/}
            {(filterParams.type === "select" && !!options && options.length > 0) && (
                <Select
                    leftIcon={icon ? icon :
                        <svg xmlns="http://www.w3.org/2000/svg" width="18" height="24" viewBox="0 0 24 24" fill="none"
                             stroke="currentColor"
                             strokeWidth="2"
                             strokeLinecap="round" strokeLinejoin="round" className="w-4 h-4">
                            <polygon points="22 3 2 3 10 12.46 10 19 14 21 14 12.46 22 3"/>
                        </svg>}
                    leftAddon={filterParams.name}
                    options={[...options.map(n => ({ value: n.value, label: valueFormatter(n.value) }))]}
                    onChange={e => handleUpdate(e.target.value.trim().toLowerCase())}
                    size={"sm"}
                    fieldClassName={"w-fit"}
                    className="sm:w-auto pr-8 md:max-w-sm"
                />
            )}
            {/*Boolean*/}
            {(filterParams.type === "boolean") && (
                <DropdownMenu
                    dropdownClassName={"right-[inherit] left"}
                    trigger={
                        <DataGridActiveFilter
                            options={filterParams}
                            value={valueFormatter(filterValue)}
                        />
                    }>
                    <DropdownMenu.Group>
                        <DropdownMenu.Item onClick={() => handleUpdate(true)}>
                            {typeof valueFormatter(true) === "boolean" ? locales["true"][locale] : valueFormatter(true)}
                        </DropdownMenu.Item>
                        <DropdownMenu.Item onClick={() => handleUpdate(false)}>
                            {typeof valueFormatter(false) === "boolean" ? locales["false"][locale] : valueFormatter(false)}
                        </DropdownMenu.Item>
                    </DropdownMenu.Group>
                </DropdownMenu>
            )}
            {/*Checkbox*/}
            {(filterParams.type === "checkbox" && !!options.length) && (
                <DropdownMenu
                    dropdownClassName={"right-[inherit] left"}
                    trigger={
                        <DataGridActiveFilter
                            options={filterParams}
                            value={Array.isArray(filterValue) ? (filterValue as any).map((n: string) => valueFormatter(n)) : valueFormatter(filterValue)}
                        />}
                >
                    <DropdownMenu.Group className={"p-1"}>
                        {filterParams.options?.length && (
                            <CheckboxGroup
                                options={filterParams.options}
                                value={filterValue}
                                onChange={handleUpdate}
                                checkboxContainerClassName={"flex flex-row-reverse w-full justify-between"}
                                checkboxLabelClassName={"cursor-pointer"}
                            />
                        )}
                    </DropdownMenu.Group>
                </DropdownMenu>
            )}
            {/*Radio*/}
            {(filterParams.type === "radio" && !!options.length) && (
                <DropdownMenu
                    dropdownClassName={"right-[inherit] left"}
                    trigger={
                        <DataGridActiveFilter
                            options={filterParams}
                            value={Array.isArray(filterValue) ? (filterValue as any).map((n: string) => valueFormatter(n)) : valueFormatter(filterValue)}
                        />}
                >
                    <DropdownMenu.Group className={"p-1"}>
                        {filterParams.options?.length && (
                            <RadioGroup
                                options={filterParams.options}
                                value={filterValue}
                                onChange={handleUpdate}
                                radioContainerClassName={"flex flex-row-reverse w-full justify-between"}
                                radioLabelClassName={"cursor-pointer"}
                            />
                        )}
                    </DropdownMenu.Group>
                </DropdownMenu>
            )}
            {/*Date*/}
            {filterParams.type === "date-range" && (
                <div className={cn(DataGridAnatomy.filterDropdownButton(), "truncate overflow-ellipsis")}>
                    {filterParams.icon && <span>{filterParams.icon}</span>}
                    <span>{filterParams.name}:</span>
                    <DateRangePicker
                        value={filterValue ? {
                            start: parseAbsoluteToLocal(filterValue.start.toISOString()),
                            end: parseAbsoluteToLocal(filterValue.end.toISOString()),
                        } : undefined}
                        onChange={value => handleUpdate({
                            start: value?.start.toDate(getLocalTimeZone()),
                            end: value?.end.toDate(getLocalTimeZone()),
                        })}
                        intent={"unstyled"}
                        locale={locale}
                        hideTimeZone
                        granularity={"day"}
                    />
                </div>
            )}

            <CloseButton onClick={onRemove} size={"sm"}/>
        </div>
    )

}

DataGridFilter.displayName = "DataGridFilter"


interface DataGridActiveFilterProps extends Omit<React.ComponentPropsWithRef<"button">, "value">,
    ComponentWithAnatomy<typeof DataGridActiveFilterAnatomy> {
    children?: React.ReactNode
    options: DataGridFilteringHelper<any>
    value: unknown
}

export const DataGridActiveFilter: React.FC<DataGridActiveFilterProps> = React.forwardRef((props, ref) => {

    const { children, options, value, ...rest } = props

    // Truncate and join the value to be displayed if it is an array
    const displayedValue = Array.isArray(value) ? (value.length > 2 ? [...value.slice(0, 2), "..."].join(", ") : value.join(", ")) : String(value)

    return (
        <button className={cn(DataGridAnatomy.filterDropdownButton(), "truncate overflow-ellipsis")} {...rest}
                ref={ref}>
            {options.icon && <span>{options.icon}</span>}
            <span>{options.name}:</span>
            <span className={"font-semibold flex flex-none overflow-hidden whitespace-normal"}>{displayedValue}</span>
        </button>
    )

})

DataGridActiveFilter.displayName = "DataGridActiveFilter"
