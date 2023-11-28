import { startTransition, useEffect, useLayoutEffect, useMemo, useRef, useState } from "react"
import { Row, Table } from "@tanstack/react-table"
import deepEquals from "fast-deep-equal"

/**
 * DataGrid Prop
 */
export type DataGridOnRowSelect<T> = (event: DataGridRowSelectedEvent<T>) => void

/**
 * Hook props
 */
type Props<T> = {
    persistent: boolean
    onRowSelect?: DataGridOnRowSelect<T>
    table: Table<T>,
    data: T[] | null
    displayedRows: Row<T>[]
    rowSelectionPrimaryKey: string | undefined
    enabled: boolean
}

/**
 * Event
 */
export type DataGridRowSelectedEvent<T> = {
    data: T[]
}

export function useDataGridRowSelection<T extends Record<string, any>>(props: Props<T>) {

    const {
        table,
        data,
        onRowSelect,
        persistent,
        rowSelectionPrimaryKey: key,
        displayedRows,
        enabled,
    } = props

    const canSelect = useRef<boolean>(enabled)

    // Server mode
    const pageIndex = useMemo(() => table.getState().pagination.pageIndex, [table.getState().pagination.pageIndex])
    const pageSize = useMemo(() => table.getState().pagination.pageSize, [table.getState().pagination.pageSize])
    const globalFilter = useMemo(() => table.getState().globalFilter, [table.getState().globalFilter])
    const columnFilters = useMemo(() => table.getState().globalFilter, [table.getState().columnFilters])
    const sorting = useMemo(() => table.getState().sorting, [table.getState().sorting])
    // Server mode
    const displayedRowsRef = useRef<Row<T>[]>(displayedRows)
    // Server mode
    const previousSelectionEvent = useRef<DataGridRowSelectedEvent<T>>({ data: [] })
    // Server mode
    const [nonexistentSelectedRows, setNonexistentSelectedRows] = useState<{ id: string, row: Row<T> }[]>([])

    const rowSelection = useMemo(() => table.getState().rowSelection, [table.getState().rowSelection])
    const rows = useMemo(() => table.getRowModel().rows, [table.getRowModel().rows])

    // Warnings
    useEffect(() => {
        if (enabled && !key) {
            console.error("[DataGrid] You've enable row selection without providing a primary key. Make sure to define the `rowSelectionPrimaryKey` prop.")
            canSelect.current = false
        }
    }, [])

    const firstCheckRef = useRef<boolean>(false)
    useEffect(() => {
        if (enabled && key && !firstCheckRef.current && displayedRows.length > 0 && !displayedRows.some(row => !!row.original[key])) {
            console.error("[DataGrid] The key provided by `rowSelectionPrimaryKey` does not match any property in the data.")
            firstCheckRef.current = true
            canSelect.current = false
        }
    }, [displayedRows])

    /** Server-side row selection **/
    useLayoutEffect(() => {
        // When the table is paginated
        if (displayedRows.length > 0 && persistent && !!key && canSelect.current) {
            startTransition(() => {
                table.resetRowSelection()
                // Refresh nonexistent rows
                setNonexistentSelectedRows(prev => {
                    // Find the rows that were selected on the previous page
                    const rowIsSelected = (row: Row<T>) => Object.keys(rowSelection).map(v => parseInt(v)).includes(row.index)
                    const rowDoesntAlreadyExist = (row: Row<T>) => !prev.find(sr => sr.id === row.original[key])

                    const selectedRows = displayedRowsRef.current.filter(rowIsSelected).filter(rowDoesntAlreadyExist)

                    if (selectedRows.length > 0) {
                        return [...prev, ...selectedRows.map(row => ({ id: row.original[key], row: row }))]
                    }

                    return prev
                })
                displayedRowsRef.current = displayedRows // Refresh displayed row
            })

            startTransition(() => {
                displayedRows.map(displayedRow => {
                    if (nonexistentSelectedRows.some(row => row.id === displayedRow.original[key])) {
                        // If the currently displayed row is in the nonexistent array but isn't selected, select it
                        if (displayedRow.getCanSelect() && !displayedRow.getIsSelected()) {
                            displayedRow.toggleSelected(true)
                            // Then remove it from nonexistent array
                            setNonexistentSelectedRows(prev => {
                                return [...prev.filter(row => row.id !== displayedRow.original[key])]
                            })
                        }
                    }
                })
            })
        }

    }, [pageIndex, pageSize, displayedRows, globalFilter, columnFilters, sorting])

    /** Client-side row selection **/
    useEffect(() => {
        if (!persistent && data && data?.length > 0 && canSelect.current) {
            const selectedIndices = Object.keys(rowSelection).map(v => parseInt(v))

            if (selectedIndices.length > 0) {

                onRowSelect && onRowSelect({
                    data: data.filter((v: any, i: number) => selectedIndices.includes(i)) ?? [],
                })

            }
        }
    }, [rowSelection, rows])

    useEffect(() => {
        /** Server-side row selection **/
        if (persistent && data && data?.length > 0 && canSelect.current && key) {
            const selectedIndices = new Set<number>(Object.keys(rowSelection).map(v => parseInt(v)))
            startTransition(() => {
                const result = {
                    data: [
                        ...data.filter((v: any, i: number) => selectedIndices.has(i)),
                        ...nonexistentSelectedRows.map(nr => nr.row.original),
                    ],
                }
                // Compare current selection with previous
                if (!isArrayEqual(result.data, previousSelectionEvent.current.data)) {
                    onRowSelect && onRowSelect(result)
                    previousSelectionEvent.current = result
                }
            })
        }
    }, [rowSelection, previousSelectionEvent.current])


    return {
        // On client-side row selection, the count is simply what is visibly selected. On server-side row selection, the count is what is visible+nonexistent rows
        selectedRowCount: persistent ? +(Object.keys(rowSelection).length) + (nonexistentSelectedRows.length) : Object.keys(rowSelection).length,
    }

}


const isArrayEqual = function (x: Array<Record<string, any>>, y: Array<Record<string, any>>) {
    return deepEquals(x, y)
}
