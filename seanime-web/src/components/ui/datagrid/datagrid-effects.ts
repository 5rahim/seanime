import { Table } from "@tanstack/react-table"
import { startTransition, useEffect } from "react"

export const useDataGridEffects = <T extends Record<string, any>>(
    table: Table<T>,
) => {

    useEffect(() => {
        startTransition(() => {
            table.setPageIndex(0)
        })
    }, [table.getState().globalFilter])

}
