"use client"

import React from "react"
import { cn, ComponentWithAnatomy, defineStyleAnatomy } from "../core"
import { cva } from "class-variance-authority"
import { DataGridEditingHelper } from "./helpers"
import { z, ZodTypeAny } from "zod"
import { Cell, Row, Table } from "@tanstack/react-table"
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
 * withEditing({ field: (ctx: DataGridCellInputFieldContext) => <></> })
 */
export type DataGridEditingFieldContext<T> = {
    value: T,
    onChange: (value: T) => void
    ref: React.MutableRefObject<any>
}

export type DataGridEditingValueUpdater<T extends Record<string, any>> = (
    value: unknown,
    row: Row<T>,
    cell: Cell<T, unknown>,
    zodType: ZodTypeAny | undefined,
) => void

export interface DataGridCellInputFieldProps<
    Schema extends z.ZodObject<z.ZodRawShape>,
    T extends Record<string, any>,
    Key extends keyof z.infer<Schema>
>
    extends ComponentWithAnatomy<typeof DataGridCellInputFieldAnatomy> {
    meta: DataGridEditingHelper
    cell: Cell<T, unknown>
    table: Table<T>
    row: Row<T>
    rowErrors: DataGridValidationRowErrors
    onValueUpdated: DataGridEditingValueUpdater<T>
}

export function DataGridCellInputField<
    Schema extends z.ZodObject<z.ZodRawShape>,
    T extends Record<string, any>,
    Key extends keyof z.infer<Schema>
>(props: DataGridCellInputFieldProps<Schema, T, Key>) {

    const {
        rootClassName,
        cell,
        table,
        row,
        rowErrors,
        onValueUpdated, // Emits updates to the hook
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
        <div
            className={cn(DataGridCellInputFieldAnatomy.root(), rootClassName)}
        >
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
