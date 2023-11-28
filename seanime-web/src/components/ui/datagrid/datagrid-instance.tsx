import {
    ColumnDef,
    ColumnFiltersState,
    ColumnOrderState,
    FilterFn,
    getCoreRowModel,
    getFilteredRowModel,
    getSortedRowModel,
    OnChangeFn,
    PaginationState,
    RowSelectionState,
    SortingState,
    useReactTable,
    VisibilityState,
} from "@tanstack/react-table"
import { dateRangeFilter } from "./use-datagrid-filtering"
import React, { useCallback, useEffect, useLayoutEffect, useMemo, useState } from "react"
import { Checkbox } from "../checkbox"
import { DataGridOnRowEdit, DataGridOnRowValidationError } from "./use-datagrid-editing"
import { DataGridOnRowSelect } from "./use-datagrid-row-selection"
import { AnyZodObject } from "zod"

export type DataGridInstanceProps<T extends Record<string, any>> = {
    data: T[] | null | undefined
    rowCount: number
    columns: ColumnDef<T>[]
    isLoading?: boolean

    hideColumns?: { below: number, hide: string[] }[]
    columnOrder?: ColumnOrderState | undefined

    /* -------------------------------------------------------------------------------------------------
     * Row selection
     * -----------------------------------------------------------------------------------------------*/

    enableRowSelection?: boolean
    onRowSelect?: DataGridOnRowSelect<T>
    rowSelectionPrimaryKey?: string
    enablePersistentRowSelection?: boolean

    /* -------------------------------------------------------------------------------------------------
     * Sorting
     * -----------------------------------------------------------------------------------------------*/

    enableSorting?: boolean
    enableManualSorting?: boolean

    /* -------------------------------------------------------------------------------------------------
     * Filters
     * -----------------------------------------------------------------------------------------------*/

    enableColumnFilters?: boolean
    enableFilters?: boolean
    enableManualFiltering?: boolean
    enableGlobalFilter?: boolean

    /* -------------------------------------------------------------------------------------------------
     * Pagination
     * -----------------------------------------------------------------------------------------------*/

    enableManualPagination?: boolean

    /* -------------------------------------------------------------------------------------------------
     * Editing
     * -----------------------------------------------------------------------------------------------*/

    enableOptimisticUpdates?: boolean
    optimisticUpdatePrimaryKey?: string
    isDataMutating?: boolean
    validationSchema?: AnyZodObject
    onRowEdit?: DataGridOnRowEdit<T>
    onRowValidationError?: DataGridOnRowValidationError<T>

    initialState?: {
        sorting?: SortingState
        pagination?: PaginationState
        rowSelection?: RowSelectionState
        globalFilter?: string
        columnFilters?: ColumnFiltersState
        columnVisibility?: VisibilityState
    }

    state?: {
        sorting?: SortingState
        pagination?: PaginationState
        rowSelection?: RowSelectionState
        globalFilter?: string
        columnFilters?: ColumnFiltersState
        columnVisibility?: VisibilityState
    },

    onSortingChange?: OnChangeFn<SortingState>
    onPaginationChange?: OnChangeFn<PaginationState>
    onRowSelectionChange?: OnChangeFn<RowSelectionState>
    onGlobalFilterChange?: OnChangeFn<string>
    onColumnFiltersChange?: OnChangeFn<ColumnFiltersState>
    onColumnVisibilityChange?: OnChangeFn<VisibilityState>

    filterFns?: Record<string, FilterFn<T>>
}

export function useDataGrid<T extends Record<string, any>>(props: DataGridInstanceProps<T>) {

    const defaultValues: Required<DataGridInstanceProps<T>["state"]> = {
        globalFilter: "",
        sorting: [],
        pagination: { pageIndex: 0, pageSize: 5 },
        rowSelection: {},
        columnFilters: [],
        columnVisibility: {},
    }

    const {
        data: _actualData,
        rowCount: _initialRowCount,
        columns,
        initialState,
        state,

        onRowValidationError,
        validationSchema,

        columnOrder,

        onSortingChange,
        onPaginationChange,
        onRowSelectionChange,
        onGlobalFilterChange,
        onColumnFiltersChange,
        onColumnVisibilityChange,

        enableManualSorting = false,
        enableManualFiltering = false,
        enableManualPagination = false,
        enableRowSelection = false,
        enablePersistentRowSelection = false,
        enableOptimisticUpdates = false,

        enableColumnFilters = true,
        enableSorting = true,
        enableFilters = true,
        enableGlobalFilter = true,

        filterFns,

        ...rest
    } = props

    const [data, setData] = useState<T[]>(_actualData ?? [])

    const [rowCount, setRowCount] = useState(_initialRowCount)

    useEffect(() => {
        if (_actualData) setData(_actualData)
    }, [_actualData])

    useEffect(() => {
        if (_initialRowCount) setRowCount(_initialRowCount)
    }, [_initialRowCount])

    const [globalFilter, setGlobalFilter] = useState<string>(initialState?.globalFilter ?? defaultValues.globalFilter)
    const [rowSelection, setRowSelection] = useState<RowSelectionState>(initialState?.rowSelection ?? defaultValues.rowSelection)
    const [sorting, setSorting] = useState<SortingState>(initialState?.sorting ?? defaultValues.sorting)
    const [columnFilters, setColumnFilters] = useState<ColumnFiltersState>(initialState?.columnFilters ?? defaultValues.columnFilters)
    const [columnVisibility, setColumnVisibility] = useState<VisibilityState>(initialState?.columnVisibility ?? defaultValues.columnVisibility)
    const [pagination, setPagination] = useState<PaginationState>(initialState?.pagination ?? defaultValues.pagination)

    const pageCount = useMemo(() => Math.ceil(rowCount / pagination.pageSize) ?? -1, [rowCount, pagination.pageSize])

    const columnsWithSelection = useMemo<ColumnDef<T>[]>(() => [{
        id: "_select",
        size: 0,
        maxSize: 0,
        enableSorting: false,
        disableSortBy: true,
        disableGlobalFilter: true,
        header: ({ table }) => {
            return (
                <Checkbox
                    checked={table.getIsSomeRowsSelected() ? "indeterminate" : table.getIsAllRowsSelected()}
                    onChange={() => table.toggleAllRowsSelected()}
                />
            )
        },
        cell: ({ row }) => {
            return (
                <div className="px-1">
                    <Checkbox
                        key={row.id}
                        checked={row.getIsSomeSelected() ? "indeterminate" : row.getIsSelected()}
                        isDisabled={!row.getCanSelect()}
                        onChange={row.getToggleSelectedHandler()}
                    />
                </div>
            )
        },
    }, ...columns], [columns])

    const sortingState = useMemo(() => state?.sorting ?? sorting, [state?.sorting, sorting])
    const paginationState = useMemo(() => state?.pagination ?? pagination, [state?.pagination, pagination])
    const rowSelectionState = useMemo(() => state?.rowSelection ?? rowSelection, [state?.rowSelection, rowSelection])
    const globalFilterState = useMemo(() => state?.globalFilter ?? globalFilter, [state?.globalFilter, globalFilter])
    const columnFiltersState = useMemo(() => state?.columnFilters ?? columnFilters, [state?.columnFilters, columnFilters])
    const columnVisibilityState = useMemo(() => state?.columnVisibility ?? columnVisibility, [state?.columnVisibility, columnVisibility])

    const changeHandler = useCallback((func: any, func2: any) => {
        return ((updaterOrValue) => {
            if (func) func(updaterOrValue)
            if (func2) func2(updaterOrValue)
        }) as OnChangeFn<any>
    }, [])

    const table = useReactTable<T>({
        data: data,
        columns: enableRowSelection ? columnsWithSelection : columns,
        pageCount: pageCount,
        globalFilterFn: (row, columnId, filterValue) => {
            const safeValue: string = ((): string => {
                const value: any = row.getValue(columnId)
                return typeof value === "number" ? String(value) : value
            })()
            return safeValue?.trim().toLowerCase().includes(filterValue.trim().toLowerCase())
        },
        state: {
            sorting: sortingState,
            pagination: paginationState,
            rowSelection: rowSelectionState,
            globalFilter: globalFilterState,
            columnFilters: columnFiltersState,
            columnVisibility: columnVisibilityState,
            columnOrder: columnOrder,
        },
        onSortingChange: changeHandler(onSortingChange, setSorting),
        onPaginationChange: changeHandler(onPaginationChange, setPagination),
        onRowSelectionChange: changeHandler(onRowSelectionChange, setRowSelection),
        onGlobalFilterChange: changeHandler(onGlobalFilterChange, setGlobalFilter),
        onColumnFiltersChange: changeHandler(onColumnFiltersChange, setColumnFilters),
        onColumnVisibilityChange: changeHandler(onColumnVisibilityChange, setColumnVisibility),
        getCoreRowModel: getCoreRowModel(),
        getSortedRowModel: enableManualSorting ? undefined : getSortedRowModel(),
        getFilteredRowModel: enableManualFiltering ? undefined : getFilteredRowModel(),
        filterFns: {
            dateRangeFilter: dateRangeFilter,
            ...filterFns,
        },
        manualPagination: enableManualPagination,
        manualSorting: enableManualSorting,
        manualFiltering: enableManualFiltering,
        enableRowSelection: enableRowSelection,
        enableSorting: enableSorting,
        enableColumnFilters: enableColumnFilters,
        enableFilters: enableFilters,
        enableGlobalFilter: enableGlobalFilter,
    })

    const displayedRows = useMemo(() => {
        const pn = table.getState().pagination
        if (enableManualPagination) {
            return table.getRowModel().rows
        }
        return table.getRowModel().rows.slice(pn.pageIndex * pn.pageSize, (pn.pageIndex + 1) * pn.pageSize)
    }, [table.getRowModel().rows, table.getState().pagination])

    useLayoutEffect(() => {
        table.setPageIndex(0)
    }, [table.getState().globalFilter])

    useEffect(() => {
        if (!enableManualPagination) {
            setRowCount(table.getRowModel().rows.length)
        }
    }, [table.getRowModel().rows])

    return {
        ...rest,

        table,
        displayedRows,
        setData,
        data,
        pageCount,
        rowCount,
        columns,

        sorting: sortingState,
        pagination: paginationState,
        rowSelection: rowSelectionState,
        globalFilter: globalFilterState,
        columnFilters: columnFiltersState,
        columnVisibility: columnVisibilityState,

        enableManualSorting,
        enableManualFiltering,
        enableManualPagination,
        enableRowSelection,
        enablePersistentRowSelection,
        enableOptimisticUpdates,
        enableGlobalFilter,

        validationSchema,
        onRowValidationError,

        handleGlobalFilterChange: onGlobalFilterChange ?? setGlobalFilter,
        handleColumnFiltersChange: onColumnFiltersChange ?? setColumnFilters,

    }

}

export type DataGridApi<T extends Record<string, any>> = ReturnType<typeof useDataGrid<T>>
