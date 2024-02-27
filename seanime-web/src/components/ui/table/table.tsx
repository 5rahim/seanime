"use client"

import { cva } from "class-variance-authority"
import * as React from "react"
import { cn, defineStyleAnatomy } from "../core/styling"

/* -------------------------------------------------------------------------------------------------
 * Anatomy
 * -----------------------------------------------------------------------------------------------*/

export const TableAnatomy = defineStyleAnatomy({
    table: cva([
        "UI-Table__table",
        "w-full caption-bottom text-sm",
    ]),
    tableHeader: cva([
        "UI-Table__tableHeader",
        "[&_tr]:border-b",
    ]),
    tableBody: cva([
        "UI-Table__tableBody",
        "[&_tr:last-child]:border-0",
    ]),
    tableFooter: cva([
        "UI-Table__tableFooter",
        "border-t bg-gray-100 dark:bg-gray-900 bg-opacity-40 font-medium [&>tr]:last:border-b-0",
    ]),
    tableRow: cva([
        "UI-Table__tableRow",
        "border-b transition-colors hover:bg-[--subtle] data-[state=selected]:bg-[--subtle]",
    ]),
    tableHead: cva([
        "UI-Table__tableHead",
        "h-12 px-4 text-left align-middle font-medium",
        "[&:has([role=checkbox])]:pr-0",
    ]),
    tableCell: cva([
        "UI-Table__tableCell",
        "p-4 align-middle [&:has([role=checkbox])]:pr-0",
    ]),
    tableCaption: cva([
        "UI-Table__tableCaption",
        "mt-4 text-sm",
    ]),
})

/* -------------------------------------------------------------------------------------------------
 * Table
 * -----------------------------------------------------------------------------------------------*/

export type TableProps = React.ComponentPropsWithoutRef<"table">

export const Table = React.forwardRef<HTMLTableElement, TableProps>((props, ref) => {
    const { className, ...rest } = props

    return (
        <div className="relative w-full overflow-auto">
            <table
                ref={ref}
                className={cn(TableAnatomy.table(), className)}
                {...rest}
            />
        </div>
    )
})
Table.displayName = "Table"

/* -------------------------------------------------------------------------------------------------
 * TableHeader
 * -----------------------------------------------------------------------------------------------*/

export type TableHeaderProps = React.ComponentPropsWithoutRef<"thead">

export const TableHeader = React.forwardRef<HTMLTableSectionElement, TableHeaderProps>((props, ref) => {
    const { className, ...rest } = props

    return <thead ref={ref} className={cn(TableAnatomy.tableHeader(), className)} {...rest} />
})
TableHeader.displayName = "TableHeader"

/* -------------------------------------------------------------------------------------------------
 * TableBody
 * -----------------------------------------------------------------------------------------------*/

export type TableBodyProps = React.ComponentPropsWithoutRef<"tbody">

export const TableBody = React.forwardRef<HTMLTableSectionElement, TableBodyProps>((props, ref) => {
    const { className, ...rest } = props

    return <tbody ref={ref} className={cn(TableAnatomy.tableBody(), className)} {...rest} />
})
TableBody.displayName = "TableBody"

/* -------------------------------------------------------------------------------------------------
 * TableFooter
 * -----------------------------------------------------------------------------------------------*/

export type TableFooterProps = React.ComponentPropsWithoutRef<"tfoot">

export const TableFooter = React.forwardRef<HTMLTableSectionElement, TableFooterProps>((props, ref) => {
    const { className, ...rest } = props

    return <tfoot ref={ref} className={cn(TableAnatomy.tableFooter(), className)} {...rest} />
})
TableFooter.displayName = "TableFooter"

/* -------------------------------------------------------------------------------------------------
 * TableRow
 * -----------------------------------------------------------------------------------------------*/

export type TableRowProps = React.ComponentPropsWithoutRef<"tr">

export const TableRow = React.forwardRef<HTMLTableRowElement, TableRowProps>((props, ref) => {
    const { className, ...rest } = props

    return <tr ref={ref} className={cn(TableAnatomy.tableRow(), className)} {...rest} />
})
TableRow.displayName = "TableRow"

/* -------------------------------------------------------------------------------------------------
 * TableHead
 * -----------------------------------------------------------------------------------------------*/

export type TableHeadProps = React.ComponentPropsWithoutRef<"th">

export const TableHead = React.forwardRef<HTMLTableCellElement, TableHeadProps>((props, ref) => {
    const { className, ...rest } = props

    return <th ref={ref} className={cn(TableAnatomy.tableHead(), className)} {...rest} />
})
TableHead.displayName = "TableHead"

/* -------------------------------------------------------------------------------------------------
 * TableCell
 * -----------------------------------------------------------------------------------------------*/

export type TableCellProps = React.ComponentPropsWithoutRef<"td">

export const TableCell = React.forwardRef<HTMLTableCellElement, TableCellProps>((props, ref) => {
    const { className, ...rest } = props

    return <td ref={ref} className={cn(TableAnatomy.tableCell(), className)} {...rest} />
})
TableCell.displayName = "TableCell"

/* -------------------------------------------------------------------------------------------------
 * TableCaption
 * -----------------------------------------------------------------------------------------------*/

export type TableCaptionProps = React.ComponentPropsWithoutRef<"caption">

export const TableCaption = React.forwardRef<HTMLTableCaptionElement, TableCaptionProps>((props, ref) => {
    const { className, ...rest } = props

    return <caption ref={ref} className={cn(TableAnatomy.tableCaption(), className)} {...rest} />
})
TableCaption.displayName = "TableCaption"
