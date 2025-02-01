"use client"

import { flexRender } from "@tanstack/react-table"
import { cva } from "class-variance-authority"
import * as React from "react"
import { Button, IconButton } from "../button"
import { Card } from "../card"
import { cn, ComponentAnatomy, defineStyleAnatomy } from "../core/styling"
import { DropdownMenu, DropdownMenuItem } from "../dropdown-menu"
import { LoadingOverlay } from "../loading-spinner"
import { NumberInput } from "../number-input"
import { Pagination, PaginationTrigger } from "../pagination"
import { Select } from "../select"
import { Skeleton } from "../skeleton"
import { TextInput, TextInputProps } from "../text-input"
import { Tooltip } from "../tooltip"
import { DataGridCellInputField } from "./datagrid-cell-input-field"
import { DataGridFilter } from "./datagrid-filter"
import { DataGridApi, DataGridInstanceProps, useDataGrid } from "./datagrid-instance"
import { getColumnHelperMeta, getValueFormatter } from "./helpers"
import translations from "./locales"
import { useDataGridEditing } from "./use-datagrid-editing"
import { useDataGridFiltering } from "./use-datagrid-filtering"
import { useDataGridResponsiveness } from "./use-datagrid-responsiveness"
import { useDataGridRowSelection } from "./use-datagrid-row-selection"

/* -------------------------------------------------------------------------------------------------
 * Anatomy
 * -----------------------------------------------------------------------------------------------*/

export const DataGridAnatomy = defineStyleAnatomy({
    root: cva([
        "UI-DataGrid__root",
    ]),
    header: cva([
        "UI-DataGrid__header",
        "block space-y-4 w-full mb-4",
    ]),
    toolbar: cva([
        "UI-DataGrid__toolbar",
        "flex w-full items-center gap-4 flex-wrap",
    ]),
    tableContainer: cva([
        "UI-DataGrid__tableContainer",
        "align-middle inline-block min-w-full max-w-full overflow-x-auto relative",
    ]),
    table: cva([
        "UI-DataGrid__table",
        "w-full relative table-fixed",
    ]),
    tableHead: cva([
        "UI-DataGrid__tableHead",
        "",
    ]),
    th: cva([
        "UI-DataGrid__th group/th",
        "px-3 h-12 text-left text-sm font-bold",
        "data-[is-selection-col=true]:px-3 data-[is-selection-col=true]:sm:px-1 data-[is-selection-col=true]:text-center",
    ]),
    titleChevronContainer: cva([
        "UI-DataGrid__titleChevronContainer",
        "absolute flex items-center inset-y-0 top-1 -right-9 group",
    ]),
    titleChevron: cva([
        "UI-DataGrid__titleChevron",
        "mr-3 h-4 w-4 text-gray-400 group-hover:text-gray-500 relative bottom-0.5",
    ]),
    tableBody: cva([
        "UI-DataGrid__tableBody",
        "w-full relative border-b",
    ]),
    td: cva([
        "UI-DataGrid__td",
        "px-2 py-2 w-full whitespace-nowrap text-base font-normal text-[--foreground]",
        "data-[is-selection-col=true]:px-2 data-[is-selection-col=true]:sm:px-0 data-[is-selection-col=true]:text-center",
        "data-[action-col=false]:truncate data-[action-col=false]:overflow-ellipsis",
        "data-[row-selected=true]:bg-brand-50 dark:data-[row-selected=true]:bg-gray-800",
        "data-[editing=true]:ring-1 data-[editing=true]:ring-[--ring] ring-inset",
        "data-[editable=true]:hover:bg-[--subtle] md:data-[editable=true]:focus:ring-2 md:data-[editable=true]:focus:ring-[--slate]",
        "focus:outline-none",
        "border-b",
    ]),
    tr: cva([
        "UI-DataGrid__tr",
        "hover:bg-[--subtle] truncate",
    ]),
    footer: cva([
        "UI-DataGrid__footer",
        "flex flex-col sm:flex-row w-full items-center gap-2 justify-between p-2 mt-2 overflow-x-auto max-w-full",
    ]),
    footerPageDisplayContainer: cva([
        "UI-DataGrid__footerPageDisplayContainer",
        "flex flex-none items-center gap-1 ml-2 text-sm",
    ]),
    footerPaginationInputContainer: cva([
        "UI-DataGrid__footerPaginationInputContainer",
        "flex flex-none items-center gap-2",
    ]),
    filterDropdownButton: cva([
        "UI-DataGrid__filterDropdownButton",
        "flex gap-2 items-center bg-[--paper] border rounded-[--radius] h-10 py-1 px-3 cursor-pointer hover:bg-[--subtle]",
        "select-none focus-visible:ring-2 outline-none ring-[--ring]",
    ]),
    editingCard: cva([
        "UI-DataGrid__editingCard",
        "flex items-center gap-2 rounded-[--radius-md] px-3 py-2",
    ]),
})

/* -------------------------------------------------------------------------------------------------
 * DataGrid
 * -----------------------------------------------------------------------------------------------*/

export type DataGridProps<T extends Record<string, any>> = ComponentAnatomy<typeof DataGridAnatomy> & DataGridInstanceProps<T> & {
    tableApi?: DataGridApi<T>,
    globalSearchInputProps?: Partial<DataGridSearchInputProps>
    hideGlobalSearchInput?: boolean
    className?: string
    lng?: string
}

export function DataGrid<T extends Record<string, any>>(props: DataGridProps<T>) {

    const {
        lng = "en",
        className,
        headerClass,
        toolbarClass,
        tableContainerClass,
        tableHeadClass,
        tableClass,
        thClass,
        titleChevronClass,
        titleChevronContainerClass,
        tableBodyClass,
        trClass,
        tdClass,
        footerClass,
        footerPageDisplayContainerClass,
        footerPaginationInputContainerClass,
        filterDropdownButtonClass,
        editingCardClass,
        tableApi,
        globalSearchInputProps,
        hideGlobalSearchInput,
        ...rest
    } = props

    const {
        table,
        data,
        setData,
        displayedRows,
        globalFilter,
        columnFilters,
        handleGlobalFilterChange,
        handleColumnFiltersChange,
        isLoading,
        isDataMutating,
        hideColumns,
        enablePersistentRowSelection,
        onRowEdit,
        onRowSelect,
        rowSelectionPrimaryKey,
        enableRowSelection,
        enableOptimisticUpdates,
        optimisticUpdatePrimaryKey,
        enableManualPagination,
        enableGlobalFilter,
        validationSchema,
        onRowValidationError,
    } = (tableApi ?? useDataGrid<T>({ ...rest })) as DataGridApi<T>

    const isInLoadingState = isLoading || (!enableOptimisticUpdates && isDataMutating)
    const { tableRef } = useDataGridResponsiveness({ table, hideColumns })

    const {
        selectedRowCount,
    } = useDataGridRowSelection({
        table: table,
        data: data,
        displayedRows: displayedRows,
        persistent: enablePersistentRowSelection,
        onRowSelect: onRowSelect,
        rowSelectionPrimaryKey: rowSelectionPrimaryKey,
        enabled: !!enableRowSelection,
    })

    const {
        getFilterDefaultValue,
        unselectedFilterableColumns,
        filteredColumns,
        filterableColumns,
    } = useDataGridFiltering({
        table: table,
        columnFilters: columnFilters,
    })

    const {
        handleStartEditing,
        getIsCellActivelyEditing,
        getIsCellEditable,
        getIsCurrentlyEditing,
        getFirstCellBeingEdited,
        handleStopEditing,
        handleOnSave,
        handleUpdateValue,
        rowErrors,
    } = useDataGridEditing({
        table: table,
        data: data,
        rows: displayedRows,
        onRowEdit: onRowEdit,
        isDataMutating: isDataMutating,
        enableOptimisticUpdates: enableOptimisticUpdates,
        optimisticUpdatePrimaryKey: optimisticUpdatePrimaryKey,
        manualPagination: enableManualPagination,
        onDataChange: setData,
        schema: validationSchema,
        onRowValidationError: onRowValidationError,
    })


    return (
        <div className={cn(DataGridAnatomy.root(), className)}>
            <div className={cn(DataGridAnatomy.header(), headerClass)}>

                <div className={cn(DataGridAnatomy.toolbar(), toolbarClass)}>
                    {/* Search Box */}
                    {(enableGlobalFilter && !hideGlobalSearchInput) && (
                        <DataGridSearchInput
                            value={globalFilter ?? ""}
                            onChange={value => handleGlobalFilterChange(String(value))}
                            {...globalSearchInputProps}
                        />
                    )}
                    {/* Filter dropdown */}
                    {(unselectedFilterableColumns.length > 0) && (
                        <DropdownMenu
                            trigger={
                                <button
                                    className={cn(DataGridAnatomy.filterDropdownButton(), filterDropdownButtonClass)}
                                >
                                    <svg
                                        xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24"
                                        fill="none"
                                        stroke="currentColor"
                                        strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"
                                        className="w-4 h-4"
                                    >
                                        <polygon points="22 3 2 3 10 12.46 10 19 14 21 14 12.46 22 3" />
                                    </svg>
                                    <span>{translations["filters"][lng]} ({unselectedFilterableColumns.length})</span>
                                </button>
                            }
                        >
                            {/*Filter list*/}
                            {unselectedFilterableColumns.map(col => {
                                const defaultValue = getFilterDefaultValue(col)
                                const icon = getColumnHelperMeta(col, "filteringMeta")?.icon
                                const name = getColumnHelperMeta(col, "filteringMeta")?.name
                                return (
                                    <DropdownMenuItem
                                        key={col.id}
                                        onClick={() => handleColumnFiltersChange(p => [...p, {
                                            id: col.id,
                                            value: defaultValue,
                                        }])}
                                    >
                                        {icon && <span className="text-md mr-2">{icon}</span>}
                                        <span>{name}</span>
                                    </DropdownMenuItem>
                                )
                            })}
                        </DropdownMenu>
                    )}
                    {/*Remove filters button*/}
                    {unselectedFilterableColumns.length !== filterableColumns.length && (
                        <Tooltip
                            trigger={<IconButton
                                icon={
                                    <svg
                                        xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="none"
                                        stroke="currentColor" strokeWidth="2"
                                        strokeLinecap="round" strokeLinejoin="round" className="h-4 w-4"
                                    >
                                        <path d="M9 14 4 9l5-5" />
                                        <path d="M4 9h10.5a5.5 5.5 0 0 1 5.5 5.5v0a5.5 5.5 0 0 1-5.5 5.5H11" />
                                    </svg>
                                }
                                intent="gray-outline"
                                size="md"
                                onClick={() => handleColumnFiltersChange([])}
                            />}
                        >
                            {translations["remove-filters"][lng]}
                        </Tooltip>
                    )}
                    {/*Selected row count*/}
                    {(selectedRowCount > 0) && <div className="text-sm">
                        {selectedRowCount} {translations[`row${selectedRowCount > 1 ? "s" : ""}-selected`][lng]}
                    </div>}
                </div>

                {/*Display filters*/}
                {(filteredColumns.length > 0) && <div className={cn(DataGridAnatomy.toolbar(), toolbarClass)}>
                    {/*Display selected filters*/}
                    {filteredColumns.map(col => {
                        return (
                            <DataGridFilter
                                key={col.id}
                                column={col}
                                onRemove={() => handleColumnFiltersChange(filters => [...filters.filter(filter => filter.id !== col.id)])}
                                lng={lng}
                            />
                        )
                    })}
                </div>}

                {/*Manage editing*/}
                {getIsCurrentlyEditing() &&
                    <Card className={cn(DataGridAnatomy.editingCard(), editingCardClass)}>
                        <Button size="sm" onClick={handleOnSave} loading={isDataMutating}>
                            {translations["save"][lng]}
                        </Button>
                        <Button
                            size="sm"
                            onClick={handleStopEditing}
                            intent="gray-outline"
                            disabled={isDataMutating}
                        >
                            {translations["cancel"][lng]}
                        </Button>
                    </Card>}

            </div>

            {/* Table */}
            <div ref={tableRef} className={cn(DataGridAnatomy.tableContainer(), tableContainerClass)}>

                <table className={cn(DataGridAnatomy.table(), tableClass)}>

                    {/*Head*/}

                    <thead className={cn(DataGridAnatomy.tableHead(), tableHeadClass)}>
                    {table.getHeaderGroups().map((headerGroup) => (
                        <tr key={headerGroup.id}>
                            {headerGroup.headers.map((header, index) => (
                                <th
                                    key={header.id}
                                    colSpan={header.colSpan}
                                    scope="col"
                                    className={cn(DataGridAnatomy.th(), thClass)}
                                    data-is-selection-col={`${index === 0 && !!enableRowSelection}`}
                                    style={{ width: header.getSize() }}
                                >
                                    {((index !== 0 && !!enableRowSelection) || !enableRowSelection) ? <div
                                        className={cn(
                                            "flex items-center justify-between",
                                            {
                                                "cursor-pointer": header.column.getCanSort(),
                                            },
                                        )}
                                    >
                                        {header.isPlaceholder ? null : (
                                            <div
                                                className="flex relative items-center"
                                                {...{
                                                    onClick: header.column.getToggleSortingHandler(),
                                                }}
                                            >
                                                {flexRender(
                                                    header.column.columnDef.header,
                                                    header.getContext(),
                                                )}
                                                <span
                                                    className={cn(DataGridAnatomy.titleChevronContainer(), titleChevronContainerClass)}
                                                >
                                                    {header.column.getIsSorted() === "asc" &&
                                                        <svg
                                                            xmlns="http://www.w3.org/2000/svg" width="24"
                                                            height="24" viewBox="0 0 24 24"
                                                            fill="none" stroke="currentColor" strokeWidth="2"
                                                            strokeLinecap="round"
                                                            strokeLinejoin="round"
                                                            className={cn(DataGridAnatomy.titleChevron(), titleChevronClass)}
                                                        >
                                                            <polyline points="18 15 12 9 6 15" />
                                                        </svg>
                                                    }
                                                    {header.column.getIsSorted() === "desc" &&
                                                        <svg
                                                            xmlns="http://www.w3.org/2000/svg" width="24"
                                                            height="24" viewBox="0 0 24 24"
                                                            fill="none" stroke="currentColor" strokeWidth="2"
                                                            strokeLinecap="round"
                                                            strokeLinejoin="round"
                                                            className={cn(DataGridAnatomy.titleChevron(), titleChevronClass)}
                                                        >
                                                            <polyline points="6 9 12 15 18 9" />
                                                        </svg>
                                                    }
                                                    {(header.column.getIsSorted() === false && header.column.getCanSort()) &&
                                                        <svg
                                                            xmlns="http://www.w3.org/2000/svg" width="24"
                                                            height="24" viewBox="0 0 24 24"
                                                            fill="none" stroke="currentColor" strokeWidth="2"
                                                            strokeLinecap="round"
                                                            strokeLinejoin="round"
                                                            className={cn(
                                                                DataGridAnatomy.titleChevron(),
                                                                "w-4 h-4 opacity-0 transition-opacity group-hover/th:opacity-100",
                                                                titleChevronClass,
                                                            )}
                                                        >
                                                            <path d="m7 15 5 5 5-5" />
                                                            <path d="m7 9 5-5 5 5" />
                                                        </svg>
                                                    }
                                                </span>
                                            </div>
                                        )}
                                    </div> : flexRender(
                                        header.column.columnDef.header,
                                        header.getContext(),
                                    )}
                                </th>
                            ))}
                        </tr>
                    ))}
                    </thead>

                    {/*Body*/}

                    <tbody className={cn(DataGridAnatomy.tableBody(), tableBodyClass)}>

                    {displayedRows.map((row) => {
                        return (
                            <tr key={row.id} className={cn(DataGridAnatomy.tr(), trClass)}>
                                {row.getVisibleCells().map((cell, index) => {

                                    // If cell is editable and cell's row is being edited
                                    const isCurrentlyEditable = getIsCellEditable(cell.id) && !getIsCellActivelyEditing(cell.id)
                                        && (!getIsCurrentlyEditing() || getFirstCellBeingEdited()?.rowId === cell.row.id)

                                    return (
                                        <td
                                            key={cell.id}
                                            className={cn(DataGridAnatomy.td(), tdClass)}
                                            data-is-selection-col={`${index === 0 && enableRowSelection}`} // If cell is in the selection
                                            // column
                                            data-action-col={`${cell.column.id === "_actions"}`} // If cell is in the action column
                                            data-row-selected={cell.getContext().row.getIsSelected()} // If cell's row is currently selected
                                            data-editing={getIsCellActivelyEditing(cell.id)} // If cell is being edited
                                            data-editable={isCurrentlyEditable} // If cell is editable
                                            data-row-editing={getFirstCellBeingEdited()?.rowId === cell.row.id} // If cell's row is being edited
                                            style={{
                                                width: cell.column.getSize(),
                                                maxWidth: cell.column.columnDef.maxSize,
                                            }}
                                            onDoubleClick={() => React.startTransition(() => {
                                                handleStartEditing(cell.id)
                                            })}
                                            onKeyUp={event => {
                                                if (event.key === "Enter") React.startTransition(() => handleStartEditing(cell.id))
                                            }}
                                            tabIndex={isCurrentlyEditable ? 0 : undefined} // Is focusable if it can be edited
                                        >
                                            {((!getIsCellEditable(cell.id) || !getIsCellActivelyEditing(cell.id))) && flexRender(
                                                cell.column.columnDef.cell,
                                                {
                                                    ...cell.getContext(),
                                                    renderValue: () => getValueFormatter(cell.column)(cell.getContext().getValue()),
                                                },
                                            )}
                                            {getIsCellActivelyEditing(cell.id) && (
                                                <DataGridCellInputField
                                                    cell={cell}
                                                    row={cell.row}
                                                    table={table}
                                                    rowErrors={rowErrors}
                                                    meta={getColumnHelperMeta(cell.column, "editingMeta")!}
                                                    onValueUpdated={handleUpdateValue}
                                                />
                                            )}
                                        </td>
                                    )
                                })}
                            </tr>
                        )
                    })}
                    </tbody>
                </table>

                {(isInLoadingState && displayedRows.length > 0) && (
                    <LoadingOverlay className="backdrop-blur-[1px] bg-opacity-40 pt-0" />
                )}

                {/*Skeleton*/}
                {(isInLoadingState && displayedRows.length === 0) && [...Array(5).keys()].map((i, idx) => (
                    <Skeleton key={idx} className="rounded-none h-12" />
                ))}

                {/*No rows*/}
                {(displayedRows.length === 0 && !isInLoadingState && filteredColumns.length === 0) && (
                    <p className="flex w-full justify-center py-4">
                        <svg xmlns="http://www.w3.org/2000/svg" width="48" height="48" viewBox="0 0 48 48">
                            <path
                                fill="#D1C4E9"
                                d="M38 7H10c-1.1 0-2 .9-2 2v6c0 1.1.9 2 2 2h28c1.1 0 2-.9 2-2V9c0-1.1-.9-2-2-2zm0 12H10c-1.1 0-2 .9-2 2v6c0 1.1.9 2 2 2h28c1.1 0 2-.9 2-2v-6c0-1.1-.9-2-2-2zm0 12H10c-1.1 0-2 .9-2 2v6c0 1.1.9 2 2 2h28c1.1 0 2-.9 2-2v-6c0-1.1-.9-2-2-2z"
                            />
                            <circle cx="38" cy="38" r="10" fill="#F44336" />
                            <g fill="#fff">
                                <path d="m43.31 41.181l-2.12 2.122l-8.485-8.484l2.121-2.122z" />
                                <path d="m34.819 43.31l-2.122-2.12l8.484-8.485l2.122 2.121z" />
                            </g>
                        </svg>
                    </p>
                )}

                {/*No results with filters*/}
                {(displayedRows.length === 0 && !isInLoadingState && filteredColumns.length > 0) && (
                    <div className="w-full text-center py-4">
                        <p className="flex w-full justify-center mb-4">
                            <svg xmlns="http://www.w3.org/2000/svg" width="48" height="48" viewBox="0 0 48 48">
                                <path
                                    fill="#D1C4E9"
                                    d="M38 7H10c-1.1 0-2 .9-2 2v6c0 1.1.9 2 2 2h28c1.1 0 2-.9 2-2V9c0-1.1-.9-2-2-2zm0 12H10c-1.1 0-2 .9-2 2v6c0 1.1.9 2 2 2h28c1.1 0 2-.9 2-2v-6c0-1.1-.9-2-2-2zm0 12H10c-1.1 0-2 .9-2 2v6c0 1.1.9 2 2 2h28c1.1 0 2-.9 2-2v-6c0-1.1-.9-2-2-2z"
                                />
                                <circle cx="38" cy="38" r="10" fill="#F44336" />
                                <g fill="#fff">
                                    <path d="m43.31 41.181l-2.12 2.122l-8.485-8.484l2.121-2.122z" />
                                    <path d="m34.819 43.31l-2.122-2.12l8.484-8.485l2.122 2.121z" />
                                </g>
                            </svg>
                        </p>
                        <p>{translations["no-matching-result"][lng]}</p>
                    </div>
                )}
            </div>

            <div className={cn(DataGridAnatomy.footer(), footerClass)}>

                <Pagination>
                    <PaginationTrigger
                        direction="previous"
                        isChevrons
                        onClick={() => table.setPageIndex(0)}
                        disabled={!table.getCanPreviousPage() || isInLoadingState}
                    />
                    <PaginationTrigger
                        direction="previous"
                        onClick={() => table.previousPage()}
                        disabled={!table.getCanPreviousPage() || isInLoadingState}
                    />
                    <PaginationTrigger
                        direction="next"
                        onClick={() => table.nextPage()}
                        disabled={!table.getCanNextPage() || isInLoadingState}
                    />
                    <PaginationTrigger
                        direction="next"
                        isChevrons
                        onClick={() => table.setPageIndex(table.getPageCount() - 1)}
                        disabled={!table.getCanNextPage() || isInLoadingState}
                    />
                </Pagination>

                <div className={cn(DataGridAnatomy.footerPageDisplayContainer(), footerPageDisplayContainerClass)}>
                    {table.getPageCount() > 0 && (
                        <>
                            <div>{translations["page"][lng]}</div>
                            <strong>
                                {table.getState().pagination.pageIndex + 1} / {table.getPageCount()}
                            </strong>
                        </>
                    )}
                </div>

                <div className={cn(DataGridAnatomy.footerPaginationInputContainer(), footerPaginationInputContainerClass)}>
                    {(data.length > 0) && <NumberInput
                        hideControls
                        value={table.getState().pagination.pageIndex + 1}
                        min={1}
                        onValueChange={v => {
                            const page = v ? v - 1 : 0
                            React.startTransition(() => {
                                if (v <= table.getPageCount()) {
                                    table.setPageIndex(page)
                                }
                            })
                        }}
                        className="inline-flex flex-none items-center w-[3rem]"
                        size="sm"
                    />}
                    <Select
                        value={String(table.getState().pagination.pageSize)}
                        onValueChange={v => {
                            table.setPageSize(Number(v))
                        }}
                        options={[Number(table.getState().pagination.pageSize),
                            ...[5, 10, 20, 30, 40, 50].filter(n => n !== Number(table.getState().pagination.pageSize))].map(pageSize => ({
                            value: String(pageSize),
                            label: String(pageSize),
                        }))}
                        fieldClass="w-auto"
                        className="w-auto"
                        disabled={isInLoadingState}
                        size="sm"
                    />
                </div>

            </div>

        </div>
    )

}

DataGrid.displayName = "DataGrid"

/* -------------------------------------------------------------------------------------------------
 * DataGridSearchInput
 * -----------------------------------------------------------------------------------------------*/

type DataGridSearchInputProps = Omit<TextInputProps, "onChange"> & {
    value: string,
    onChange: (value: string) => void
    debounce?: number
}

export function DataGridSearchInput(props: DataGridSearchInputProps) {

    const { value: initialValue, onChange, debounce = 500, ...rest } = props

    const [value, setValue] = React.useState(initialValue)

    React.useEffect(() => {
        setValue(initialValue)
    }, [initialValue])

    React.useEffect(() => {
        const timeout = setTimeout(() => {
            onChange(value)
        }, debounce)

        return () => clearTimeout(timeout)
    }, [value])

    return (
        <TextInput
            size="md"
            fieldClass="md:max-w-[30rem]"
            {...rest}
            value={value}
            onChange={e => setValue(e.target.value)}
            leftIcon={<svg
                xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none"
                stroke="currentColor"
                strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"
                className="w-5 h-5 text-[--muted]"
            >
                <circle cx="11" cy="11" r="8" />
                <path d="m21 21-4.3-4.3" />
            </svg>}
        />
    )
}

/* -------------------------------------------------------------------------------------------------
 * DataGridWithApi
 * -----------------------------------------------------------------------------------------------*/

export type DataGridWithApiProps<T extends Record<string, any>> = ComponentAnatomy<typeof DataGridAnatomy> & {
    api: DataGridApi<T>
}

export function DataGridWithApi<T extends Record<string, any>>(props: DataGridWithApiProps<T>) {

    const {
        api,
        ...rest
    } = props

    const {
        data,
        rowCount,
        columns,
    } = api

    return <DataGrid
        data={data}
        rowCount={rowCount}
        columns={columns}
        tableApi={api}
        {...rest}
    />

}
