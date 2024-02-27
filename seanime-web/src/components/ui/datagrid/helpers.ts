import { BuiltInFilterFn, Cell, Column, ColumnDef, Row, Table } from "@tanstack/react-table"
import React from "react"
import { AnyZodObject, z, ZodAny, ZodTypeAny } from "zod"
import { DataGridEditingFieldContext } from "./datagrid-cell-input-field"
import { DataGridValidationRowErrors } from "./use-datagrid-editing"

/* -------------------------------------------------------------------------------------------------
 * Editing
 * -----------------------------------------------------------------------------------------------*/

export type DataGridEditingHelper<T extends any = unknown, ZodType extends ZodTypeAny = ZodAny> = {
    zodType?: ZodType
    field: (
        context: DataGridEditingFieldContext<ZodType extends ZodAny ? T : z.infer<ZodType>>,
        options: {
            rowErrors: DataGridValidationRowErrors
            table: Table<any>
            row: Row<any>
            cell: Cell<any, unknown>
        },
    ) => React.ReactElement
    valueFormatter?: <K = z.infer<ZodType>, R = z.infer<ZodType>>(value: K) => R
}

function withEditing<T extends any = unknown, ZodType extends ZodTypeAny = ZodAny>(params: DataGridEditingHelper<T, ZodType>) {
    return {
        editingMeta: {
            ...params,
        },
    }
}

/* -------------------------------------------------------------------------------------------------
 * Filtering
 * -----------------------------------------------------------------------------------------------*/

export type DataGridFilteringType = "select" | "radio" | "checkbox" | "boolean" | "date-range"

export interface FilterFns {
    dateRangeFilter: any
}

type _DefaultFilteringProps = {
    type: DataGridFilteringType
    name: string,
    icon?: React.ReactElement
    options?: { value: string, label?: any }[]
    valueFormatter?: (value: any) => any
}

type DefaultFilteringProps<T extends DataGridFilteringType> = {
    type: T
    name: string,
    icon?: React.ReactElement
    options: { value: string, label?: T extends "select" ? string : React.ReactNode }[]
    valueFormatter?: (value: any) => any
}

// Improve type safety by removing "options" when the type doesn't need it
export type DataGridFilteringHelper<T extends DataGridFilteringType = "select"> =
    T extends Extract<DataGridFilteringType, "select" | "radio" | "checkbox">
        ? DefaultFilteringProps<T>
        : Omit<DefaultFilteringProps<T>, "options">

/**
 * Built-in filter functions supported DataGrid
 */
export type DataGridSupportedFilterFn =
    Extract<BuiltInFilterFn, "equals" | "equalsString" | "arrIncludesSome" | "inNumberRange">
    | "dateRangeFilter"

function withFiltering<T extends DataGridFilteringType>(params: DataGridFilteringHelper<T>) {
    return {
        filteringMeta: {
            ...params,
        },
    }
}

const getFilterFn = (type: DataGridFilteringType) => {
    const fns: { [key: string]: DataGridSupportedFilterFn } = {
        select: "equalsString",
        boolean: "equals",
        checkbox: "arrIncludesSome",
        radio: "equalsString",
        "date-range": "dateRangeFilter",
    }
    return fns[type] as any
}

/* -------------------------------------------------------------------------------------------------
 * Value formatter
 * -----------------------------------------------------------------------------------------------*/

function withValueFormatter<T extends any, R extends any = any>(callback: (value: T) => R) {
    return {
        valueFormatter: callback,
    }
}

export function getValueFormatter<T>(column: Column<T>): (value: any) => any {
    return (column.columnDef.meta as any)?.valueFormatter || ((value: any) => value)
}

/* -------------------------------------------------------------------------------------------------
 * Column Def Helpers
 * -----------------------------------------------------------------------------------------------*/

export type DataGridHelpers = "filteringMeta" | "editingMeta" | "valueFormatter"

export type DataGridColumnDefHelpers<T extends Record<string, any>> = {
    withFiltering: typeof withFiltering
    getFilterFn: typeof getFilterFn
    withEditing: typeof withEditing
    withValueFormatter: typeof withValueFormatter
}

/**
 * Return
 * @example
 * const columns = useMemo(() => defineDataGridColumns<T>(() => [
 *  ...
 * ]), [])
 * @param callback
 */
export function defineDataGridColumns<T extends Record<string, any>, Schema extends AnyZodObject = any>(
    callback: (helpers: DataGridColumnDefHelpers<T>, schema?: Schema) => Array<ColumnDef<T>>,
) {
    return callback({
        withFiltering,
        getFilterFn,
        withEditing,
        withValueFormatter,
    })
}


export function getColumnHelperMeta<T, K extends DataGridHelpers>(column: Column<T>, helper: K) {
    return (column.columnDef.meta as any)?.[helper] as (
        K extends "filteringMeta" ? _DefaultFilteringProps :
            K extends "editingMeta" ? DataGridEditingHelper :
                K extends "valueFormatter" ? ReturnType<typeof withValueFormatter> :
                    never
        ) | undefined
}
