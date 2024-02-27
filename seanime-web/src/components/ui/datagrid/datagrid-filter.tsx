"use client"

import { Column } from "@tanstack/react-table"
import { cva } from "class-variance-authority"
import * as React from "react"
import { DataGridAnatomy, DataGridFilteringHelper, getColumnHelperMeta, getValueFormatter } from "."
import { CloseButton } from "../button"
import { CheckboxGroup } from "../checkbox"
import { cn, ComponentAnatomy, defineStyleAnatomy } from "../core/styling"
import { DateRangePicker } from "../date-picker"
import { DropdownMenu, DropdownMenuGroup, DropdownMenuItem } from "../dropdown-menu"
import { RadioGroup } from "../radio-group"
import { Select } from "../select"
import translations, { dateFnsLocales } from "./locales"

/* -------------------------------------------------------------------------------------------------
 * Anatomy
 * -----------------------------------------------------------------------------------------------*/

export const DataGridFilterAnatomy = defineStyleAnatomy({
    root: cva([
        "UI-DataGridFilter__root",
        "flex items-center max-w-full gap-2",
    ]),
})

export const DataGridActiveFilterAnatomy = defineStyleAnatomy({
    root: cva([
        "UI-DataGridActiveFilter__root",
        "py-1 px-2 rounded-[--radius] border flex gap-2 items-center",
    ]),
})

/* -------------------------------------------------------------------------------------------------
 * DataGridFilter
 * -----------------------------------------------------------------------------------------------*/

export type DataGridFilterProps<T extends Record<string, any>> = React.ComponentPropsWithoutRef<"div"> &
    ComponentAnatomy<typeof DataGridFilterAnatomy> & {
    column: Column<T>
    onRemove: () => void
    lng?: string
}

export function DataGridFilter<T extends Record<string, any>>(props: DataGridFilterProps<T>) {

    const {
        children,
        className,
        column,
        onRemove,
        lng = "en",
        ...rest
    } = props

    const filterParams = getColumnHelperMeta(column, "filteringMeta")!
    const filterValue = React.useMemo(() => column.getFilterValue(), [column.getFilterValue()]) as any
    const setFilterValue = React.useMemo(() => column.setFilterValue, [column.setFilterValue])
    const icon = filterParams.icon

    // Value formatter - if undefined, use the default behavior
    const valueFormatter = filterParams.valueFormatter || getValueFormatter(column)

    // Get the options
    const options = filterParams.options ?? []

    // Update handler
    const handleUpdate = React.useCallback((value: any) => {
        setFilterValue(value)
    }, [])

    return (
        <div
            className={cn(DataGridFilterAnatomy.root(), className)}
            {...rest}
        >
            {(filterParams.type === "select" && (!options || options.length === 0)) && (
                <div className="text-red-500">/!\ "Select" filtering option passed without options</div>
            )}
            {/*Select*/}
            {(filterParams.type === "select" && !!options && options.length > 0) && (
                <Select
                    leftIcon={icon ? icon :
                        <svg
                            xmlns="http://www.w3.org/2000/svg" width="18" height="24" viewBox="0 0 24 24" fill="none"
                            stroke="currentColor"
                            strokeWidth="2"
                            strokeLinecap="round" strokeLinejoin="round" className="w-4 h-4"
                        >
                            <polygon points="22 3 2 3 10 12.46 10 19 14 21 14 12.46 22 3" />
                        </svg>}
                    leftAddon={filterParams.name}
                    options={[...options.map(n => ({ value: n.value, label: valueFormatter(n.value) }))]}
                    onValueChange={v => handleUpdate(v.trim().toLowerCase())}
                    size="sm"
                    fieldClass="w-fit"
                    className="sm:w-auto pr-8 md:max-w-sm"
                />
            )}
            {/*Boolean*/}
            {(filterParams.type === "boolean") && (
                <DropdownMenu
                    className="right-[inherit] left"
                    trigger={
                        <DataGridActiveFilter
                            options={filterParams}
                            value={valueFormatter(filterValue)}
                        />
                    }
                >
                    <DropdownMenuGroup>
                        <DropdownMenuItem onClick={() => handleUpdate(true)}>
                            {typeof valueFormatter(true) === "boolean" ? translations["true"][lng] : valueFormatter(true)}
                        </DropdownMenuItem>
                        <DropdownMenuItem onClick={() => handleUpdate(false)}>
                            {typeof valueFormatter(false) === "boolean" ? translations["false"][lng] : valueFormatter(false)}
                        </DropdownMenuItem>
                    </DropdownMenuGroup>
                </DropdownMenu>
            )}
            {/*Checkbox*/}
            {(filterParams.type === "checkbox" && !!options.length) && (
                <DropdownMenu
                    className="right-[inherit] left"
                    trigger={
                        <DataGridActiveFilter
                            options={filterParams}
                            value={Array.isArray(filterValue) ?
                                (filterValue as any).map((n: string) => valueFormatter(n)) :
                                valueFormatter(filterValue)
                            }
                        />}
                >
                    <DropdownMenuGroup className="p-1">
                        {filterParams.options?.length && (
                            <CheckboxGroup
                                options={filterParams.options}
                                value={filterValue}
                                onValueChange={handleUpdate}
                                itemContainerClass="flex flex-row-reverse w-full justify-between"
                                itemLabelClass="cursor-pointer"
                            />
                        )}
                    </DropdownMenuGroup>
                </DropdownMenu>
            )}
            {/*Radio*/}
            {(filterParams.type === "radio" && !!options.length) && (
                <DropdownMenu
                    className="right-[inherit] left"
                    trigger={
                        <DataGridActiveFilter
                            options={filterParams}
                            value={Array.isArray(filterValue) ?
                                (filterValue as any).map((n: string) => valueFormatter(n)) :
                                valueFormatter(filterValue)
                            }
                        />}
                >
                    <DropdownMenuGroup className="p-1">
                        {filterParams.options?.length && (
                            <RadioGroup
                                options={filterParams.options}
                                value={filterValue}
                                onValueChange={handleUpdate}
                                itemContainerClass="flex flex-row-reverse w-full justify-between"
                                itemLabelClass="cursor-pointer"
                            />
                        )}
                    </DropdownMenuGroup>
                </DropdownMenu>
            )}
            {/*Date*/}
            {filterParams.type === "date-range" && (
                <div className={cn(DataGridAnatomy.filterDropdownButton(), "truncate overflow-ellipsis")}>
                    {filterParams.icon && <span>{filterParams.icon}</span>}
                    <span>{filterParams.name}:</span>
                    <DateRangePicker
                        value={filterValue ? {
                            from: filterValue.start,
                            to: filterValue.end,
                        } : undefined}
                        onValueChange={value => handleUpdate({
                            start: value?.from,
                            end: value?.to,
                        })}
                        placeholder={translations["date-range-placeholder"][lng]}
                        intent="unstyled"
                        locale={dateFnsLocales[lng]}
                    />
                </div>
            )}

            <CloseButton
                intent="gray-outline"
                onClick={onRemove}
                size="md"
            />
        </div>
    )

}

DataGridFilter.displayName = "DataGridFilter"


interface DataGridActiveFilterProps extends Omit<React.ComponentPropsWithRef<"button">, "value">,
    ComponentAnatomy<typeof DataGridActiveFilterAnatomy> {
    children?: React.ReactNode
    options: DataGridFilteringHelper<any>
    value: unknown
}

export const DataGridActiveFilter = React.forwardRef<HTMLButtonElement, DataGridActiveFilterProps>((props, ref) => {

    const { children, options, value, ...rest } = props

    // Truncate and join the value to be displayed if it is an array
    const displayedValue = Array.isArray(value) ? (value.length > 2 ? [...value.slice(0, 2), "..."].join(", ") : value.join(", ")) : String(value)

    return (
        <button
            ref={ref}
            className={cn(DataGridAnatomy.filterDropdownButton(), "truncate overflow-ellipsis")} {...rest}
        >
            {options.icon && <span>{options.icon}</span>}
            <span>{options.name}:</span>
            <span className="font-semibold flex flex-none overflow-hidden whitespace-normal">{displayedValue}</span>
        </button>
    )

})

DataGridActiveFilter.displayName = "DataGridActiveFilter"
