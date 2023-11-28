import { Row, Table } from "@tanstack/react-table"
import React, { startTransition, useCallback, useEffect, useLayoutEffect, useMemo, useState } from "react"
import { DataGridEditingValueUpdater } from "./datagrid-cell-input-field"
import { AnyZodObject, ZodIssue } from "zod"
import { useToast } from "../toast"
import isEqual from "lodash/isEqual"
import flatten from "lodash/flatten"

/**
 * DataGrid Props
 */
export type DataGridRowEditedEvent<T extends Record<string, any>> = {
    row: Row<T>
    originalData: T
    data: T
}

export type DataGridOnRowEdit<T extends Record<string, any>> = (event: DataGridRowEditedEvent<T>) => void

//----

export type DataGridRowValidationError<T extends Record<string, any>> = {
    row: Row<T>
    originalData: T
    data: T
    errors: ZodIssue[]
}

export type DataGridOnRowValidationError<T extends Record<string, any>> = (event: DataGridRowValidationError<T>) => void

//----

export type DataGridValidationRowErrors = Array<{ rowId: string, key: string, message: string }>

/**
 * Hook props
 */
type Props<T extends Record<string, any>> = {
    data: T[]
    table: Table<T>
    rows: Row<T>[]
    onRowEdit?: DataGridOnRowEdit<T>
    isDataMutating: boolean | undefined
    enableOptimisticUpdates: boolean
    onDataChange: React.Dispatch<React.SetStateAction<T[]>>
    optimisticUpdatePrimaryKey: string | undefined
    manualPagination: boolean
    schema: AnyZodObject | undefined
    onRowValidationError: DataGridOnRowValidationError<T> | undefined
}

export function useDataGridEditing<T extends Record<string, any>>(props: Props<T>) {

    const {
        data,
        table,
        rows,
        onRowEdit,
        isDataMutating,
        onDataChange,
        enableOptimisticUpdates,
        optimisticUpdatePrimaryKey,
        manualPagination,
        schema,
        onRowValidationError,
    } = props

    const toast = useToast()

    const leafColumns = table.getAllLeafColumns()
    // Keep track of the state of each editable cell
    const [editableCellStates, setEditableCellStates] = useState<{
        id: string,
        colId: string,
        rowId: string,
        isEditing: boolean
    }[]>([])

    // Track updated value
    const [activeValue, setActiveValue] = useState<unknown>(undefined)
    // Track current row data being updated
    const [rowData, setRowData] = useState<T | undefined>(undefined)
    // Track current row being updated
    const [row, setRow] = useState<Row<T> | undefined>(undefined)

    const [rowErrors, setRowErrors] = useState<DataGridValidationRowErrors>([])

    // Keep track of editable columns (columns defined with the `withEditing` helper)
    const editableColumns = useMemo(() => {
        return leafColumns.filter(n => n.getIsVisible() && !!(n.columnDef.meta as any)?.editingMeta)
    }, [leafColumns])

    useEffect(() => {
        if (manualPagination) {
            setActiveValue(undefined)
            setRowData(undefined)
            setRow(undefined)
            setEditableCellStates([])
        }
    }, [table.getState().pagination.pageIndex, table.getState().pagination.pageSize])

    // Keep track of editable cells (cells whose columns are editable)
    const editableCells = useMemo(() => {
        if (rows.length > 0) {
            return flatten(rows.map(row => row.getVisibleCells().filter(cell => !!editableColumns.find(col => col.id === cell.column.id)?.id)))
        }
        return []
    }, [rows])

    // Set/update editable cells
    useLayoutEffect(() => {
        // Control the states of individual cells that can be edited
        if (editableCells.length > 0) {
            editableCells.map(cell => {
                setEditableCellStates(prev => [...prev, {
                    id: cell.id,
                    colId: cell.column.id,
                    rowId: cell.row.id,
                    isEditing: false,
                }])
            })
        }
    }, [editableCells])

    /**/
    const handleStartEditing = useCallback((cellId: string) => {
        // Manage editing state of cells
        setEditableCellStates(prev => {
            const others = prev.filter(prevCell => prevCell.id !== cellId)
            const cell = prev.find(prevCell => prevCell.id === cellId)

            if (cell && prev.every(prevCell => !prevCell.isEditing)) { // (Event 1) When we select a cell and nothing else is being edited
                return [...others, { ...cell, id: cellId, isEditing: true }]

            } else if (cell && prev.some(prevCell => prevCell.isEditing)) { // (Event 2) When another cell is being edited
                const otherCellBeingEdited = prev.find(prevCell => prevCell.isEditing) // Find the cell being edited

                if (otherCellBeingEdited?.rowId === cell?.rowId) { // Only allow cells on the same row to be edited
                    return [...others, { ...cell, id: cellId, isEditing: true }]
                }
            }
            return prev
        })
    }, [])

    /**/
    const getIsCellActivelyEditing = useCallback((cellId: string) => {
        return editableCellStates.some(cell => cell.id === cellId && cell.isEditing)
    }, [editableCellStates])
    /**/
    const getIsCellEditable = useCallback((cellId: string) => {
        return !!editableCellStates.find(cell => cell.id === cellId)
    }, [editableCellStates])
    /**/
    const getIsCurrentlyEditing = useCallback(() => {
        return editableCellStates.some(cell => cell.isEditing)
    }, [editableCellStates])
    /**/
    const getFirstCellBeingEdited = useCallback(() => {
        return editableCellStates.find(cell => cell.isEditing)
    }, [editableCellStates])
    /**/
    const handleStopEditing = useCallback(() => {
        setEditableCellStates(prev => {
            return prev.map(n => ({ ...n, isEditing: false }))
        })
    }, [])

    const mutationRef = React.useRef<boolean>(false)

    /**
     * When `isDataMutating` is provided to watch mutations,
     * Wait for it to be `false` to cancel editing
     */
    useEffect(() => {
        if (isDataMutating !== undefined && !isDataMutating && mutationRef.current) {
            handleStopEditing()
            mutationRef.current = false
        }
    }, [isDataMutating])

    /**
     * When `isDataMutating` is not provided, immediately cancel editing
     */
    useEffect(() => {
        if (isDataMutating === undefined) {
            handleStopEditing()
        }
    }, [mutationRef.current])

    const saveEdit = useCallback((transformedData?: T) => {
        if (!row || !rowData) return

        startTransition(() => {
            // Compare data
            if (!isEqual(rowData, row.original)) {
                // Return new data
                onRowEdit && onRowEdit({
                    originalData: row.original,
                    data: transformedData || rowData,
                    row: row,
                })

                // Optimistic update
                if (enableOptimisticUpdates && optimisticUpdatePrimaryKey) {
                    let clone = structuredClone(data)
                    const index = clone.findIndex(p => {
                        if (!p[optimisticUpdatePrimaryKey] || !rowData[optimisticUpdatePrimaryKey]) return false
                        return p[optimisticUpdatePrimaryKey] === rowData[optimisticUpdatePrimaryKey]
                    })
                    if (clone[index] && index > -1) {
                        clone[index] = rowData
                        onDataChange(clone) // Emit optimistic update
                    } else {
                        console.error("[DataGrid] Could not perform optimistic update. Make sure `optimisticUpdatePrimaryKey` is a valid property.")
                    }

                } else if (enableOptimisticUpdates) {
                    console.error("[DataGrid] Could not perform optimistic update. Make sure `optimisticUpdatePrimaryKey` is defined.")
                }

                // Immediately stop edit if optimistic updates are enabled
                if (enableOptimisticUpdates) {
                    handleStopEditing()
                } else {
                    // Else, we wait for `isDataMutating` to be false
                    mutationRef.current = true
                }
            } else {
                handleStopEditing()
            }

        })
    }, [row, rowData])

    const handleOnSave = useCallback(async () => {
        if (!row || !rowData) return
        setRowErrors([])

        // Safely parse the schema object when a `validationSchema` is provided
        if (schema) {
            try {
                const parsed = await schema.safeParseAsync(rowData)
                if (parsed.success) {
                    let finalData = structuredClone(rowData)
                    Object.keys(parsed.data).map(key => {
                        // @ts-ignore
                        finalData[key] = parsed.data[key]
                    })
                    saveEdit(finalData)
                } else {

                    parsed.error.errors.map(error => {
                        toast.error(error.message ?? "Error")
                        setRowErrors(prev => [
                            ...prev,
                            { rowId: row.id, key: String(error.path[0]), message: error.message },
                        ])
                    })

                    if (onRowValidationError) {
                        onRowValidationError({
                            data: rowData,
                            originalData: row.original,
                            row: row,
                            errors: parsed.error.errors,
                        })
                    }
                }
            } catch (e) {
                console.error("[DataGrid] Could not perform validation")
            }
        } else {
            saveEdit()
        }

    }, [row, rowData])

    /**
     * This fires every time the user changes an input
     */
    const handleUpdateValue = useCallback<DataGridEditingValueUpdater<T>>((value, _row, cell, zodType) => {
        setActiveValue(value) // Set the updated value (could be anything)
        setRow(_row) // Set the row being updated
        setRowData(prev => ({
            // If we are updating a different row, reset the rowData, else keep the past updates
            ...((row?.id !== _row.id || !rowData) ? _row.original : rowData),
            [cell.column.id]: value,
        }))
    }, [row, rowData])


    return {
        handleStartEditing,
        getIsCellActivelyEditing,
        getIsCellEditable,
        getIsCurrentlyEditing,
        getFirstCellBeingEdited,
        handleStopEditing,
        handleOnSave,
        handleUpdateValue,
        rowErrors,
    }

}
