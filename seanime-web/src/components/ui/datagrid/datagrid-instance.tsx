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
    Row,
    RowSelectionState,
    SortingState,
    useReactTable,
    VisibilityState,
} from "@tanstack/react-table"
import * as React from "react"
import { AnyZodObject } from "zod"
import { Checkbox } from "../checkbox"
import { DataGridOnRowEdit, DataGridOnRowValidationError } from "./use-datagrid-editing"
import { dateRangeFilter } from "./use-datagrid-filtering"
import { DataGridOnRowSelect } from "./use-datagrid-row-selection"

export type DataGridInstanceProps<T extends Record<string, any>> = {
    data: T[] | null | undefined
    rowCount: number
    columns: ColumnDef<T>[]
    isLoading?: boolean

    /**
     * Hide columns below a certain breakpoint.
     */
    hideColumns?: { below: number, hide: string[] }[]
    columnOrder?: ColumnOrderState | undefined

    /* -------------------------------------------------------------------------------------------------
     * Row selection
     * -----------------------------------------------------------------------------------------------*/

    /**
     * If true, rows will be selectable.
     * A checkbox will be shown in the first column of each row.
     * - Requires `rowSelectionPrimaryKey` for more accurate selection (default is row index)
     */
    enableRowSelection?: boolean | ((row: Row<T>) => boolean) | undefined
    /**
     * Callback invoked when a row is selected.
     */
    onRowSelect?: DataGridOnRowSelect<T>
    /**
     * The column used to uniquely identify the row.
     */
    rowSelectionPrimaryKey?: string
    /**
     * Requires `rowSelectionPrimaryKey`
     */
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

    /**
     * Requires `enableOptimisticUpdates`
     * NOTE: This will not work if your `validationSchema` contains server-side validation.
     */
    enableOptimisticUpdates?: boolean
    /**
     * The column used to uniquely identify the row.
     */
    optimisticUpdatePrimaryKey?: string
    /**
     * If true, a loading indicator will be shown while the row is being updated.
     */
    isDataMutating?: boolean
    /**
     * Zod validation schema for the columns.
     */
    validationSchema?: AnyZodObject
    /**
     * Callback invoked when a cell is successfully edited.
     */
    onRowEdit?: DataGridOnRowEdit<T>
    /**
     * Callback invoked when a cell fails validation.
     */
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

    const [data, setData] = React.useState<T[]>(_actualData ?? [])

    const [rowCount, setRowCount] = React.useState(_initialRowCount)

    React.useEffect(() => {
        if (_actualData) setData(_actualData)
    }, [_actualData])

    React.useEffect(() => {
        if (_initialRowCount) setRowCount(_initialRowCount)
    }, [_initialRowCount])

    const [globalFilter, setGlobalFilter] = React.useState<string>(initialState?.globalFilter ?? defaultValues.globalFilter)
    const [rowSelection, setRowSelection] = React.useState<RowSelectionState>(initialState?.rowSelection ?? defaultValues.rowSelection)
    const [sorting, setSorting] = React.useState<SortingState>(initialState?.sorting ?? defaultValues.sorting)
    const [columnFilters, setColumnFilters] = React.useState<ColumnFiltersState>(initialState?.columnFilters ?? defaultValues.columnFilters)
    const [columnVisibility, setColumnVisibility] = React.useState<VisibilityState>(initialState?.columnVisibility ?? defaultValues.columnVisibility)
    const [pagination, setPagination] = React.useState<PaginationState>(initialState?.pagination ?? defaultValues.pagination)

    const pageCount = React.useMemo(() => Math.ceil(rowCount / pagination.pageSize) ?? -1, [rowCount, pagination.pageSize])

    const columnsWithSelection = React.useMemo<ColumnDef<T>[]>(() => [{
        id: "_select",
        size: 6,
        maxSize: 6,
        enableSorting: false,
        disableSortBy: true,
        disableGlobalFilter: true,
        header: ({ table }) => {
            return (
                <div className="px-1">
                    <Checkbox
                        value={table.getIsSomeRowsSelected() ? "indeterminate" : table.getIsAllRowsSelected()}
                        onValueChange={() => table.toggleAllRowsSelected()}
                        fieldClass="w-fit"
                    />
                </div>
            )
        },
        cell: ({ row }) => {
            return (
                <div className="">
                    <Checkbox
                        key={row.id}
                        value={row.getIsSomeSelected() ? "indeterminate" : row.getIsSelected()}
                        disabled={!row.getCanSelect()}
                        onValueChange={row.getToggleSelectedHandler()}
                        fieldClass="w-fit"
                    />
                </div>
            )
        },
    }, ...columns], [columns])

    const sortingState = React.useMemo(() => state?.sorting ?? sorting, [state?.sorting, sorting])
    const paginationState = React.useMemo(() => state?.pagination ?? pagination, [state?.pagination, pagination])
    const rowSelectionState = React.useMemo(() => state?.rowSelection ?? rowSelection, [state?.rowSelection, rowSelection])
    const globalFilterState = React.useMemo(() => state?.globalFilter ?? globalFilter, [state?.globalFilter, globalFilter])
    const columnFiltersState = React.useMemo(() => state?.columnFilters ?? columnFilters, [state?.columnFilters, columnFilters])
    const columnVisibilityState = React.useMemo(() => state?.columnVisibility ?? columnVisibility, [state?.columnVisibility, columnVisibility])

    const changeHandler = React.useCallback((func: any, func2: any) => {
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
        getRowId: !!props.rowSelectionPrimaryKey ? (row) => row[props.rowSelectionPrimaryKey!] : undefined,
    })

    const displayedRows = React.useMemo(() => {
        const pn = table.getState().pagination
        if (enableManualPagination) {
            return table.getRowModel().rows
        }
        return table.getRowModel().rows.slice(pn.pageIndex * pn.pageSize, (pn.pageIndex + 1) * pn.pageSize)
    }, [table.getRowModel().rows, table.getState().pagination])

    React.useEffect(() => {
        table.setPageIndex(0)
    }, [table.getState().globalFilter, table.getState().columnFilters])

    React.useEffect(() => {
        if (!enableManualPagination) {
            setRowCount(table.getRowModel().rows.length)
        }
    }, [table.getRowModel().rows.length])

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
