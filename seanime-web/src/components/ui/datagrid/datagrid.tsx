"use client"

import React, { startTransition, useEffect, useState } from "react"
import { cn, ComponentWithAnatomy, defineStyleAnatomy, UIIcons, useUILocaleConfig } from "../core"
import { flexRender } from "@tanstack/react-table"
import { cva } from "class-variance-authority"
import { TextInput, TextInputProps } from "../text-input"
import { Select } from "../select"
import { NumberInput } from "../number-input"
import { Pagination } from "../pagination"
import { DataGridFilter } from "./datagrid-filter"
import { DropdownMenu } from "../dropdown-menu"
import { Button, IconButton } from "../button"
import { Tooltip } from "../tooltip"
import locales from "./locales.json"
import { useDataGridFiltering } from "./use-datagrid-filtering"
import { useDataGridResponsiveness } from "./use-datagrid-responsiveness"
import { useDataGridRowSelection } from "./use-datagrid-row-selection"
import { useDataGridEditing } from "./use-datagrid-editing"
import { DataGridCellInputField } from "./datagrid-cell-input-field"
import { Transition } from "@headlessui/react"
import { getColumnHelperMeta, getValueFormatter } from "./helpers"
import { Skeleton } from "../skeleton"
import { DataGridApi, DataGridInstanceProps, useDataGrid } from "./datagrid-instance"
import { LoadingOverlay } from "../loading-spinner"

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
    tableWrapper: cva([
        "UI-DataGrid__tableWrapper",
        "flex flex-col",
    ]),
    tableContainer: cva([
        "UI-DataGrid__tableContainer",
        "align-middle inline-block min-w-full relative",
    ]),
    table: cva([
        "UI-DataGrid__table",
        "w-full overflow-x-auto relative table-auto md:table-fixed",
    ]),
    tableHead: cva([
        "UI-DataGrid__tableHead",
        "border-b border-[--border]",
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
        "bg-[--paper] divide-y divide-[--border] w-full relative",
    ]),
    td: cva([
        "UI-DataGrid__td",
        "px-2 py-2 w-full whitespace-nowrap text-base font-normal text-[--text-color]",
        "data-[is-selection-col=true]:px-2 data-[is-selection-col=true]:sm:px-0 data-[is-selection-col=true]:text-center",
        "data-[action-col=false]:truncate data-[action-col=false]:overflow-ellipsis",
        "data-[row-selected=true]:bg-brand-50 dark:data-[row-selected=true]:bg-gray-700",
        "data-[editing=true]:ring-1 data-[editing=true]:ring-[--ring] ring-inset",
        "data-[editable=true]:hover:bg-[--highlight] md:data-[editable=true]:focus:ring-2 md:data-[editable=true]:focus:ring-[--slate]",
        "focus:outline-none",
    ]),
    tr: cva([
        "UI-DataGrid__tr",
        "hover:bg-[--highlight] truncate",
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
        "flex gap-2 items-center bg-[--paper] border border-[--border] rounded-[--radius] py-1 px-2 cursor-pointer hover:bg-[--highlight]",
        "select-none focus-visible:ring-2 outline-none ring-[--ring]",
    ]),
})

/* -------------------------------------------------------------------------------------------------
 * DataGrid
 * -----------------------------------------------------------------------------------------------*/

export interface DataGridProps<T extends Record<string, any>> extends ComponentWithAnatomy<typeof DataGridAnatomy>, DataGridInstanceProps<T> {
    tableApi?: DataGridApi<T>,
    globalSearchInputProps?: Partial<DataGridSearchInputProps & TextInputProps>
    hideGlobalSearchInput?: boolean
}

export function DataGrid<T extends Record<string, any>>(props: DataGridProps<T>) {

    const { locale: lng } = useUILocaleConfig()

    const {
        rootClassName,
        headerClassName,
        toolbarClassName,
        tableWrapperClassName,
        tableContainerClassName,
        tableHeadClassName,
        tableClassName,
        thClassName,
        titleChevronClassName,
        titleChevronContainerClassName,
        tableBodyClassName,
        trClassName,
        tdClassName,
        footerClassName,
        footerPageDisplayContainerClassName,
        footerPaginationInputContainerClassName,
        filterDropdownButtonClassName,
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
        enabled: enableRowSelection,
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
        <div className={cn(DataGridAnatomy.root(), rootClassName)}>
            <div className={cn(DataGridAnatomy.header(), headerClassName)}>

                <div className={cn(DataGridAnatomy.toolbar(), toolbarClassName)}>
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
                                    className={cn(DataGridAnatomy.filterDropdownButton(), filterDropdownButtonClassName)}>
                                    <svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24"
                                         fill="none"
                                         stroke="currentColor"
                                         strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"
                                         className="w-4 h-4">
                                        <polygon points="22 3 2 3 10 12.46 10 19 14 21 14 12.46 22 3"/>
                                    </svg>
                                    <span>{locales["filters"][lng]} ({unselectedFilterableColumns.length})</span>
                                </button>
                            }>
                            {/*Filter list*/}
                            {unselectedFilterableColumns.map(col => {
                                const defaultValue = getFilterDefaultValue(col)
                                // const icon = (col.columnDef.meta as any)?.filteringMeta?.icon
                                const icon = getColumnHelperMeta(col, "filteringMeta")?.icon
                                const name = getColumnHelperMeta(col, "filteringMeta")?.name
                                return (
                                    <DropdownMenu.Item
                                        key={col.id}
                                        onClick={() => handleColumnFiltersChange(p => [...p, {
                                            id: col.id,
                                            value: defaultValue,
                                        }])}
                                    >
                                        {icon && <span className={"text-lg"}>{icon}</span>}
                                        <span>{name}</span>
                                    </DropdownMenu.Item>
                                )
                            })}
                        </DropdownMenu>
                    )}
                    {/*Remove filters button*/}
                    {unselectedFilterableColumns.length !== filterableColumns.length && (
                        <Tooltip
                            trigger={<IconButton icon={UIIcons.undo()} intent={"gray-outline"} size={"sm"}
                                                 onClick={() => handleColumnFiltersChange([])}/>}>
                            {locales["remove-filters"][lng]}
                        </Tooltip>
                    )}
                    {/*Selected row count*/}
                    {(selectedRowCount > 0) && <div className={"text-sm"}>
                        {selectedRowCount} {locales[`row${selectedRowCount > 1 ? "s" : ""}-selected`][lng]}
                    </div>}
                </div>

                {/*Display filters*/}
                {(filteredColumns.length > 0) && <div className={cn(DataGridAnatomy.toolbar(), toolbarClassName)}>
                    {/*Display selected filters*/}
                    {filteredColumns.map(col => {
                        return (
                            <DataGridFilter
                                key={col.id}
                                column={col}
                                onRemove={() => handleColumnFiltersChange(filters => [...filters.filter(filter => filter.id !== col.id)])}
                            />
                        )
                    })}
                </div>}

                {/*Manage editing*/}
                <Transition
                    appear
                    show={getIsCurrentlyEditing()}
                    className={"fixed top-2 left-0 right-0 flex justify-center z-20"}
                    enter="transition-all duration-150"
                    enterFrom="opacity-0 scale-50"
                    enterTo="opacity-100 scale-100"
                    leave="transition-all duration-150"
                    leaveFrom="opacity-100 scale-100"
                    leaveTo="opacity-0 scale-75"
                >
                    <div
                        className={"flex items-center gap-2 rounded-md p-4 bg-[--paper] border border-[--brand] shadow-sm z-20"}>
                        <span className={"font-semibold"}>{locales["updating"][lng]}</span>
                        <Button size={"sm"} onClick={handleOnSave}
                                isDisabled={isDataMutating}>
                            {locales["save"][lng]}
                        </Button>
                        <Button size={"sm"} onClick={handleStopEditing} intent={"gray-outline"}
                                isDisabled={isDataMutating}>
                            {locales["cancel"][lng]}
                        </Button>
                    </div>
                </Transition>

            </div>

            {/* Table */}
            <div
                className={cn(DataGridAnatomy.tableWrapper(), tableWrapperClassName)}
                ref={tableRef}
            >
                <div className="relative">
                    <div className={cn(DataGridAnatomy.tableContainer(), tableContainerClassName)}>

                        <table className={cn(DataGridAnatomy.table(), tableClassName)}>

                            {/*Head*/}

                            <thead className={cn(DataGridAnatomy.tableHead(), tableHeadClassName)}>
                            {table.getHeaderGroups().map((headerGroup) => (
                                <tr key={headerGroup.id}>
                                    {headerGroup.headers.map((header, index) => (
                                        <th
                                            key={header.id}
                                            colSpan={header.colSpan}
                                            scope="col"
                                            className={cn(DataGridAnatomy.th(), thClassName)}
                                            data-is-selection-col={`${index === 0 && enableRowSelection}`}
                                            style={{ width: header.getSize() }}
                                        >
                                            {((index !== 0 && enableRowSelection) || !enableRowSelection) ? <div
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
                                                            className={cn(DataGridAnatomy.titleChevronContainer(), titleChevronContainerClassName)}>
                                                            {header.column.getIsSorted() === "asc" &&
                                                                <svg xmlns="http://www.w3.org/2000/svg" width="24"
                                                                     height="24" viewBox="0 0 24 24"
                                                                     fill="none" stroke="currentColor" strokeWidth="2"
                                                                     strokeLinecap="round"
                                                                     strokeLinejoin="round"
                                                                     className={cn(DataGridAnatomy.titleChevron(), titleChevronClassName)}>
                                                                    <polyline points="18 15 12 9 6 15"/>
                                                                </svg>
                                                            }
                                                            {header.column.getIsSorted() === "desc" &&
                                                                <svg xmlns="http://www.w3.org/2000/svg" width="24"
                                                                     height="24" viewBox="0 0 24 24"
                                                                     fill="none" stroke="currentColor" strokeWidth="2"
                                                                     strokeLinecap="round"
                                                                     strokeLinejoin="round"
                                                                     className={cn(DataGridAnatomy.titleChevron(), titleChevronClassName)}>
                                                                    <polyline points="6 9 12 15 18 9"/>
                                                                </svg>
                                                            }
                                                            {(header.column.getIsSorted() === false && header.column.getCanSort()) &&
                                                                <svg xmlns="http://www.w3.org/2000/svg" width="24"
                                                                     height="24" viewBox="0 0 24 24"
                                                                     fill="none" stroke="currentColor" strokeWidth="2"
                                                                     strokeLinecap="round"
                                                                     strokeLinejoin="round"
                                                                     className={cn(
                                                                         DataGridAnatomy.titleChevron(),
                                                                         "w-4 h-4 opacity-0 transition-opacity group-hover/th:opacity-100",
                                                                         titleChevronClassName,
                                                                     )}>
                                                                    <path d="m7 15 5 5 5-5"/>
                                                                    <path d="m7 9 5-5 5 5"/>
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

                            <tbody className={cn(DataGridAnatomy.tableBody(), tableBodyClassName)}>

                            {displayedRows.map((row) => {
                                return (
                                    <tr key={row.id} className={cn(DataGridAnatomy.tr(), trClassName)}>
                                        {row.getVisibleCells().map((cell, index) => {

                                            // If cell is editable and cell's row is being edited
                                            const isCurrentlyEditable = getIsCellEditable(cell.id) && !getIsCellActivelyEditing(cell.id)
                                                && (!getIsCurrentlyEditing() || getFirstCellBeingEdited()?.rowId === cell.row.id)

                                            return (
                                                <td
                                                    key={cell.id}
                                                    className={cn(DataGridAnatomy.td(), tdClassName)}
                                                    data-is-selection-col={`${index === 0 && enableRowSelection}`} // If cell is in the selection column
                                                    data-action-col={`${cell.column.id === "_actions"}`} // If cell is in the action column
                                                    data-row-selected={cell.getContext().row.getIsSelected()} // If cell's row is currently selected
                                                    data-editing={getIsCellActivelyEditing(cell.id)} // If cell is being edited
                                                    data-editable={isCurrentlyEditable} // If cell is editable
                                                    data-row-editing={getFirstCellBeingEdited()?.rowId === cell.row.id} // If cell's row is being edited
                                                    style={{
                                                        width: cell.column.getSize(),
                                                        maxWidth: cell.column.columnDef.maxSize,
                                                    }}
                                                    onDoubleClick={() => startTransition(() => {
                                                        handleStartEditing(cell.id)
                                                    })}
                                                    onKeyUp={event => {
                                                        if (event.key === "Enter") startTransition(() => handleStartEditing(cell.id))
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
                            <LoadingOverlay className={"backdrop-blur-[1px] bg-opacity-40 pt-0"}/>
                        )}

                        {/*Skeleton*/}
                        {(isInLoadingState && displayedRows.length === 0) && [...Array(5).keys()].map((i, idx) => (
                            <Skeleton key={idx} className={"rounded-none h-12"}/>
                        ))}

                        {/*No rows*/}
                        {(displayedRows.length === 0 && !isInLoadingState && filteredColumns.length === 0) && (
                            <p className={"flex w-full justify-center py-4"}>
                                <svg xmlns="http://www.w3.org/2000/svg" width="48" height="48" viewBox="0 0 48 48">
                                    <path fill="#D1C4E9"
                                          d="M38 7H10c-1.1 0-2 .9-2 2v6c0 1.1.9 2 2 2h28c1.1 0 2-.9 2-2V9c0-1.1-.9-2-2-2zm0 12H10c-1.1 0-2 .9-2 2v6c0 1.1.9 2 2 2h28c1.1 0 2-.9 2-2v-6c0-1.1-.9-2-2-2zm0 12H10c-1.1 0-2 .9-2 2v6c0 1.1.9 2 2 2h28c1.1 0 2-.9 2-2v-6c0-1.1-.9-2-2-2z"/>
                                    <circle cx="38" cy="38" r="10" fill="#F44336"/>
                                    <g fill="#fff">
                                        <path d="m43.31 41.181l-2.12 2.122l-8.485-8.484l2.121-2.122z"/>
                                        <path d="m34.819 43.31l-2.122-2.12l8.484-8.485l2.122 2.121z"/>
                                    </g>
                                </svg>
                            </p>
                        )}

                        {/*No results with filters*/}
                        {(displayedRows.length === 0 && !isInLoadingState && filteredColumns.length > 0) && (
                            <div className={"w-full text-center py-4"}>
                                <p className={"flex w-full justify-center mb-4"}>
                                    <svg xmlns="http://www.w3.org/2000/svg" width="48" height="48" viewBox="0 0 48 48">
                                        <path fill="#D1C4E9"
                                              d="M38 7H10c-1.1 0-2 .9-2 2v6c0 1.1.9 2 2 2h28c1.1 0 2-.9 2-2V9c0-1.1-.9-2-2-2zm0 12H10c-1.1 0-2 .9-2 2v6c0 1.1.9 2 2 2h28c1.1 0 2-.9 2-2v-6c0-1.1-.9-2-2-2zm0 12H10c-1.1 0-2 .9-2 2v6c0 1.1.9 2 2 2h28c1.1 0 2-.9 2-2v-6c0-1.1-.9-2-2-2z"/>
                                        <circle cx="38" cy="38" r="10" fill="#F44336"/>
                                        <g fill="#fff">
                                            <path d="m43.31 41.181l-2.12 2.122l-8.485-8.484l2.121-2.122z"/>
                                            <path d="m34.819 43.31l-2.122-2.12l8.484-8.485l2.122 2.121z"/>
                                        </g>
                                    </svg>
                                </p>
                                <p>{locales["no-matching-result"][lng]}</p>
                            </div>
                        )}
                    </div>
                </div>

            </div>

            <div className={cn(DataGridAnatomy.footer(), footerClassName)}>

                <Pagination>
                    <Pagination.Trigger
                        direction={"left"}
                        isChevrons
                        onClick={() => table.setPageIndex(0)}
                        isDisabled={!table.getCanPreviousPage() || isInLoadingState}
                    />
                    <Pagination.Trigger
                        direction={"left"}
                        onClick={() => table.previousPage()}
                        isDisabled={!table.getCanPreviousPage() || isInLoadingState}
                    />
                    <Pagination.Trigger
                        direction={"right"}
                        onClick={() => table.nextPage()}
                        isDisabled={!table.getCanNextPage() || isInLoadingState}
                    />
                    <Pagination.Trigger
                        direction={"right"}
                        isChevrons
                        onClick={() => table.setPageIndex(table.getPageCount() - 1)}
                        isDisabled={!table.getCanNextPage() || isInLoadingState}
                    />
                </Pagination>

                <div className={cn(DataGridAnatomy.footerPageDisplayContainer(), footerPageDisplayContainerClassName)}>
                    {table.getPageCount() > 0 && (
                        <>
                            <div>{locales["page"][lng]}</div>
                            <strong>
                                {table.getState().pagination.pageIndex + 1} / {table.getPageCount()}
                            </strong>
                        </>
                    )}
                </div>

                <div
                    className={cn(DataGridAnatomy.footerPaginationInputContainer(), footerPaginationInputContainerClassName)}>
                    {(data.length > 0) && <NumberInput
                        discrete
                        value={table.getState().pagination.pageIndex + 1}
                        min={1}
                        onChange={v => {
                            const page = v ? v - 1 : 0
                            startTransition(() => {
                                if (v <= table.getPageCount()) {
                                    table.setPageIndex(page)
                                }
                            })
                        }}
                        className={"inline-flex flex-none items-center w-[3rem]"}
                        size={"sm"}
                    />}
                    <Select
                        value={table.getState().pagination.pageSize}
                        onChange={e => {
                            table.setPageSize(Number(e.target.value))
                        }}
                        options={[Number(table.getState().pagination.pageSize), ...[5, 10, 20, 30, 40, 50].filter(n => n !== Number(table.getState().pagination.pageSize))].map(pageSize => ({
                            value: pageSize,
                            label: `${pageSize}`,
                        }))}
                        fieldClassName="w-auto"
                        className="w-auto pr-8"
                        isDisabled={isInLoadingState}
                        size={"sm"}
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

interface DataGridSearchInputProps {
    value: string,
    onChange: (value: string) => void
    debounce?: number
}

export function DataGridSearchInput(props: DataGridSearchInputProps & Omit<TextInputProps, "onChange">) {

    const { value: initialValue, onChange, debounce = 500, ...rest } = props

    const [value, setValue] = useState(initialValue)

    useEffect(() => {
        setValue(initialValue)
    }, [initialValue])

    useEffect(() => {
        const timeout = setTimeout(() => {
            onChange(value)
        }, debounce)

        return () => clearTimeout(timeout)
    }, [value])

    return (
        <TextInput
            size={"md"}
            fieldClassName={"md:max-w-[30rem]"}
            {...rest}
            value={value}
            onChange={e => setValue(e.target.value)}
            leftIcon={<svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none"
                           stroke="currentColor"
                           strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"
                           className="w-5 h-5 text-[--muted]">
                <circle cx="11" cy="11" r="8"/>
                <path d="m21 21-4.3-4.3"/>
            </svg>}
        />
    )
}

/* -------------------------------------------------------------------------------------------------
 * DataGridWithApi
 * -----------------------------------------------------------------------------------------------*/

export interface DataGridWithApiProps<T extends Record<string, any>> extends ComponentWithAnatomy<typeof DataGridAnatomy> {
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
        ...apiRest
    } = api

    return <DataGrid data={data} rowCount={rowCount} columns={columns} tableApi={api} {...rest} />

}