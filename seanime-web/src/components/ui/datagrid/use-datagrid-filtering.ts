import { Column, ColumnFiltersState, Table } from "@tanstack/react-table"
import * as React from "react"
import { getColumnHelperMeta } from "./helpers"
import { addDays } from "date-fns/addDays"
import { isSameDay } from "date-fns/isSameDay"

interface DataGridFilteringHookProps<T> {
    table: Table<T>,
    columnFilters: ColumnFiltersState,
}

export function useDataGridFiltering<T>(props: DataGridFilteringHookProps<T>) {

    const {
        table,
        columnFilters,
    } = props

    /**
     * Item filtering
     */
    const [filterableColumns, filteredColumns] = React.useMemo(() => {
        return [
            table.getAllLeafColumns().filter(col => col.getCanFilter() && !!getColumnHelperMeta(col, "filteringMeta")),
            table.getAllLeafColumns().filter(col => columnFilters.map(filter => filter.id).includes(col.id)),
        ]
    }, [table.getAllLeafColumns(), columnFilters])
    const unselectedFilterableColumns = filterableColumns.filter(n => !columnFilters.map(c => c.id).includes(n.id))

    // Get the default value for a filter when the user selects it
    const getFilterDefaultValue = React.useCallback((col: Column<any>) => {
        // Since the column is filterable, get options
        const options = getColumnHelperMeta(col, "filteringMeta")
        if (options) {
            if (options.type === "select" || options.type === "radio") {
                return options.options?.[0]?.value ?? ""
            } else if (options.type === "boolean") {
                return true
            } else if (options.type === "checkbox") {
                return options.options?.map(n => n.value) ?? []
            } else if (options.type === "date-range") {
                return { from: new Date(), to: addDays(new Date(), 7) }
            }
        }
        return null
    }, [])

    return {
        getFilterDefaultValue,
        unselectedFilterableColumns,
        filteredColumns,
        filterableColumns,
    }

}

export const dateRangeFilter = (row: any, columnId: string, filterValue: any) => {
    if (!filterValue || !filterValue.start || !filterValue.end) return true
    const value: Date = row.getValue(columnId)
    return (value >= filterValue.start && value <= filterValue.end) || isSameDay(value, filterValue.start) || isSameDay(value, filterValue.end)
}
