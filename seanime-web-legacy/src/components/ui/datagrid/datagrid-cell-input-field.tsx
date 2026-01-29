"use client"

import { Cell, Row, Table } from "@tanstack/react-table"
import { cva } from "class-variance-authority"
import * as React from "react"
import { z, ZodTypeAny } from "zod"
import { cn, ComponentAnatomy, defineStyleAnatomy } from "../core/styling"
import { DataGridEditingHelper } from "./helpers"
import { DataGridValidationRowErrors } from "./use-datagrid-editing"

/* -------------------------------------------------------------------------------------------------
 * Anatomy
 * -----------------------------------------------------------------------------------------------*/

export const DataGridCellInputFieldAnatomy = defineStyleAnatomy({
    root: cva([
        "UI-DataGridCellInputField__root",
    ]),
})

/* -------------------------------------------------------------------------------------------------
 * DataGridCellInputField
 * -----------------------------------------------------------------------------------------------*/

/**
 * Context passed to a field in order to render a cell input
 * @example
 * withEditing({ field: (ctx: DataGridCellInputFieldContext) => <></> })
 */
export type DataGridEditingFieldContext<T> = {
    value: T,
    onChange: (value: T) => void
    ref: React.MutableRefObject<any>
}

/**
 * @internal
 */
export type DataGridEditingValueUpdater<T extends Record<string, any>> = (
    value: unknown,
    row: Row<T>,
    cell: Cell<T, unknown>,
    zodType: ZodTypeAny | undefined,
) => void

/**
 * @internal
 */
export type DataGridCellInputFieldProps<T extends Record<string, any>> = ComponentAnatomy<typeof DataGridCellInputFieldAnatomy> & {
    /**
     * Meta information about the field from the column definition
     * - This is defined by the `withEditing` helper
     */
    meta: DataGridEditingHelper
    /** Cell being edited */
    cell: Cell<T, unknown>
    /** Table instance */
    table: Table<T>
    /** Row being edited */
    row: Row<T>
    /** Errors coming from the built-in row validation (useDataGridEditing) */
    rowErrors: DataGridValidationRowErrors
    /** Emits updates to the hook (useDataGridEditing) */
    onValueUpdated: DataGridEditingValueUpdater<T>
    /** Field container class name */
    className?: string
}

export function DataGridCellInputField<Schema extends z.ZodObject<z.ZodRawShape>, T extends Record<string, any>, Key extends keyof z.infer<Schema>>
(props: DataGridCellInputFieldProps<T>) {

    const {
        className,
        cell,
        table,
        row,
        rowErrors,
        onValueUpdated,
        meta: {
            field,
            zodType,
            valueFormatter: _valueFormatter,
        },
    } = props
    const defaultValueFormatter = (value: any) => value
    const valueFormatter = (_valueFormatter ?? defaultValueFormatter) as (value: any) => any

    const cellValue = valueFormatter(cell.getContext().getValue())
    const inputRef = React.useRef<any>(null)

    const [value, setValue] = React.useState<unknown>(cellValue)

    React.useLayoutEffect(() => {
        onValueUpdated(cellValue, row, cell, zodType)
        inputRef.current?.focus()
    }, [])

    return (
        <div className={cn(DataGridCellInputFieldAnatomy.root(), className)}>
            {field({
                value: value,
                onChange: (value => {
                    setValue(value)
                    onValueUpdated(valueFormatter(value), row, cell, zodType)
                }),
                ref: inputRef,
            }, {
                rowErrors: rowErrors,
                table: table,
                row: row,
                cell: cell,
            })}
        </div>
    )

}

DataGridCellInputField.displayName = "DataGridCellInputField"
