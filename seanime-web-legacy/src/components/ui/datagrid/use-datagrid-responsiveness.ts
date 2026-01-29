import { useDataGridSize } from "./use-datagrid-size"
import * as React from "react"
import { Table } from "@tanstack/react-table"

interface DataGridResponsivenessHookProps<T extends Record<string, any>> {
    hideColumns: { below: number, hide: string[] }[] | undefined,
    table: Table<T>
}

export function useDataGridResponsiveness<T extends Record<string, any>>(props: DataGridResponsivenessHookProps<T>) {

    const {
        hideColumns = [],
        table,
    } = props

    const [tableRef, { width: tableWidth }] = useDataGridSize<HTMLDivElement>()
    const deferredTableWidth = React.useDeferredValue(tableWidth)

    React.useLayoutEffect(() => {
        hideColumns.map(({ below, hide }) => {
            table.getAllLeafColumns().map(column => {
                if (hide.includes(column.id)) {
                    if (tableWidth !== 0 && tableWidth < below) {
                        if (column.getIsVisible()) column.toggleVisibility(false)
                    } else {
                        if (!column.getIsVisible()) column.toggleVisibility(true)
                    }
                }
            })
        })
    }, [hideColumns, deferredTableWidth])

    return {
        tableRef,
        tableWidth,
    }

}
