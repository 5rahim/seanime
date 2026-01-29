import { Row, Table } from "@tanstack/react-table"
import * as React from "react"

export type DataGridOnRowSelect<T> = (event: DataGridRowSelectedEvent<T>) => void

type DataGridRowSelectionProps<T> = {
    /**
     * Whether the row selection is persistent.
     * If true, the selected rows will be cached and restored when the table is paginated, filtered, sorted or when the data changes.
     */
    persistent: boolean
    /**
     * Callback fired when a row is selected.
     */
    onRowSelect?: DataGridOnRowSelect<T>
    /**
     * The table instance.
     */
    table: Table<T>,
    /**
     * The data passed to the table.
     */
    data: T[] | null
    /**
     * The rows currently displayed in the table.
     */
    displayedRows: Row<T>[]
    /**
     * The primary key of the data. This is used to identify the rows.
     */
    rowSelectionPrimaryKey: string | undefined
    /**
     * Whether row selection is enabled.
     */
    enabled: boolean
}

export type DataGridRowSelectedEvent<T> = {
    data: T[]
}

export function useDataGridRowSelection<T extends Record<string, any>>(props: DataGridRowSelectionProps<T>) {

    const {
        table,
        data,
        onRowSelect,
        persistent,
        rowSelectionPrimaryKey: key,
        displayedRows,
        enabled,
    } = props


    const rowSelection = React.useMemo(() => table.getState().rowSelection, [table.getState().rowSelection])
    const selectedRowsRef = React.useRef<Map<string | number, T>>(new Map())

    //----------------------------------

    const canSelect = React.useRef<boolean>(enabled)

    React.useEffect(() => {
        selectedRowsRef.current.clear()

        if (enabled && !key) {
            console.error(
                "[DataGrid] You've enable row selection without providing a primary key. Make sure to define the `rowSelectionPrimaryKey` prop.")
            canSelect.current = false
        }
    }, [])

    const firstCheckRef = React.useRef<boolean>(false)

    React.useEffect(() => {
        if (enabled && key && !firstCheckRef.current && displayedRows.length > 0 && !displayedRows.some(row => !!row.original[key])) {
            console.error("[DataGrid] The key provided by `rowSelectionPrimaryKey` does not match any property in the data.")
            firstCheckRef.current = true
            canSelect.current = false
        }
    }, [displayedRows])

    /** Client-side row selection **/
    React.useEffect(() => {
        if (data && data?.length > 0 && canSelect.current && !!key) {
            const selectedKeys = new Set<string | number>(Object.keys(rowSelection))

            if (persistent) {
                // Remove the keys that are no longer selected
                selectedRowsRef.current.forEach((_, k) => {
                    if (!selectedKeys.has(k.toString())) {
                        selectedRowsRef.current.delete(k)
                    }
                })

                // Add the selected rows to the selectedRowsRef
                selectedKeys.forEach(n => {
                    const row = data.find((v: any) => v[key] === n)
                    if (row) {
                        selectedRowsRef.current.set(n, row)
                    }
                })

                onRowSelect && onRowSelect({
                    data: Array.from(selectedRowsRef.current.values()).filter((v: any) => selectedKeys.has(v[key])) ?? [],
                })
            } else {
                onRowSelect && onRowSelect({
                    data: data.filter((v: any) => selectedKeys.has(v[key])) ?? [],
                })
            }

        }
    }, [rowSelection])


    return {
        selectedRowCount: Object.keys(rowSelection).length,
    }

}
