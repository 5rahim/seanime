"use client"

import { cva } from "class-variance-authority"
import * as React from "react"
import { Accept, FileError, useDropzone } from "react-dropzone"
import { BasicField, BasicFieldOptions, extractBasicFieldProps } from "../basic-field"
import { CloseButton, IconButton } from "../button"
import { cn, ComponentAnatomy, defineStyleAnatomy } from "../core/styling"
import { hiddenInputStyles } from "../input"

/* -------------------------------------------------------------------------------------------------
 * Anatomy
 * -----------------------------------------------------------------------------------------------*/

export const SimpleDropzoneAnatomy = defineStyleAnatomy({
    root: cva([
        "UI-SimpleDropzone__root",
        "appearance-none w-full mb-2 cursor-pointer hover:text-[--foreground] flex items-center justify-center p-4 border rounded-[--radius] border-dashed",
        "gap-3 text-sm sm:text-base",
        "outline-none ring-[--ring] focus-visible:ring-2",
        "text-[--muted] transition ease-in-out hover:border-[--foreground]",
        "data-[drag-active=true]:border-brand-500",
        "data-[drag-reject=true]:border-[--red]",
    ]),
    list: cva([
        "UI-SimpleDropzone__list",
        "flex rounded-[--radius-md] flex-wrap divide-y divide-[--border]",
    ]),
    listItem: cva([
        "UI-SimpleDropzone__listItem",
        "flex items-center justify-space-between relative p-1 hover:bg-[--subtle] w-full overflow-hidden",
    ]),
    listItemDetailsContainer: cva([
        "UI-SimpleDropzone__listItemDetailsContainer",
        "flex items-center gap-2 truncate w-full",
    ]),
    listItemTitle: cva([
        "UI-SimpleDropzone__listItemTitle",
        "truncate max-w-[180px] text-[.9rem]",
    ]),
    listItemSize: cva([
        "UI-SimpleDropzone__listItemSize",
        "text-xs uppercase text-center font-semibold align-center text-[--muted]",
    ]),
    listItemRemoveButton: cva([
        "UI-SimpleDropzone__listItemRemoveButton",
        "ml-2 rounded-full",
    ]),
    imagePreviewGrid: cva([
        "UI-SimpleDropzone__imagePreviewGrid",
        "flex gap-2 flex-wrap place-content-center pt-4",
    ]),
    imagePreviewContainer: cva([
        "UI-SimpleDropzone__imagePreviewContainer",
        "col-span-1 row-span-1 aspect-square w-36 h-auto",
    ]),
    imagePreview: cva([
        "UI-SimpleDropzone__imagePreview",
        "relative bg-transparent border h-full bg-center bg-no-repeat bg-contain rounded-[--radius-md] overflow-hidden",
        "col-span-1 row-span-1",
    ]),
    imagePreviewRemoveButton: cva([
        "UI-SimpleDropzone__imagePreviewRemoveButton",
        "absolute top-1 right-1",
    ]),
    fileIcon: cva([
        "UI-SimpleDropzone__fileIcon",
        "w-5 h-5 flex-none",
    ]),
    maxSizeText: cva([
        "UI-SimpleDropzone__maxSizeText",
        "text-sm text-[--muted] font-medium",
    ]),
})

/* -------------------------------------------------------------------------------------------------
 * SimpleDropzone
 * -----------------------------------------------------------------------------------------------*/

export type SimpleDropzoneProps = Omit<React.ComponentPropsWithRef<"input">, "size" | "accept" | "type" | "onError" | "onDrop"> &
    ComponentAnatomy<typeof SimpleDropzoneAnatomy> &
    BasicFieldOptions & {
    /**
     * Callback fired when files are selected
     */
    onValueChange?: (files: File[]) => void,
    /**
     * Whether to show a preview of the image(s) under the dropzone
     */
    withImagePreview?: boolean
    /**
     * Whether to allow multiple files
     */
    multiple?: boolean
    /**
     * The accepted file types
     */
    accept?: Accept
    /**
     * The minimum file size
     */
    minSize?: number
    /**
     * The maximum file size
     */
    maxSize?: number
    /**
     * The maximum number of files
     */
    maxFiles?: number
    /**
     * If false, allow dropped items to take over the current browser window
     */
    preventDropOnDocument?: boolean
    /**
     * Whether to prevent click to open file dialog
     */
    noClick?: boolean
    /**
     * Whether to prevent drag and drop
     */
    noDrag?: boolean
    /**
     * Callback fired when an error occurs
     */
    onError?: (err: Error) => void
    /**
     * Custom file validator function
     */
    validator?: <T extends File>(file: T) => FileError | FileError[] | null
    /**
     * The dropzoneText text displayed in the dropzone
     */
    dropzoneText?: string
}

export const SimpleDropzone = React.forwardRef<HTMLInputElement, SimpleDropzoneProps>((props, ref) => {

    const [{
        children,
        className,
        listClass,
        listItemClass,
        listItemDetailsContainerClass,
        listItemRemoveButtonClass,
        listItemSizeClass,
        listItemTitleClass,
        imagePreviewGridClass,
        imagePreviewContainerClass,
        imagePreviewRemoveButtonClass,
        imagePreviewClass,
        maxSizeTextClass,
        fileIconClass,
        onValueChange,
        withImagePreview,
        dropzoneText,
        /**/
        accept,
        minSize,
        maxSize,
        maxFiles,
        preventDropOnDocument,
        noClick,
        noDrag,
        onError,
        validator,
        multiple,
        value, // ignored
        ...rest
    }, basicFieldProps] = extractBasicFieldProps(props, React.useId())

    const buttonRef = React.useRef<HTMLButtonElement>(null)

    const [files, setFiles] = React.useState<File[]>([])

    const onDrop = React.useCallback((acceptedFiles: File[]) => {
        // Update files - add the preview
        setFiles(acceptedFiles.map(file => Object.assign(file, { preview: URL.createObjectURL(file) })))
    }, [])

    const handleRemoveFile = React.useCallback((file: number) => {
        setFiles(p => p.toSpliced(file, 1))
    }, [])

    React.useEffect(() => {
        onValueChange?.(files)
    }, [files])

    React.useEffect(() => () => {
        files.forEach((file: any) => URL.revokeObjectURL(file.preview))
    }, [files])

    const {
        getRootProps,
        getInputProps,
        isDragActive,
        isDragReject,
    } = useDropzone({
        onDrop,
        multiple,
        minSize,
        maxSize,
        maxFiles,
        preventDropOnDocument,
        noClick,
        noDrag,
        validator,
        accept,
        onError,
    })

    return (
        <BasicField {...basicFieldProps}>
            <button
                ref={buttonRef}
                className={cn(
                    SimpleDropzoneAnatomy.root(),
                    className,
                )}
                data-drag-active={isDragActive}
                data-drag-reject={isDragReject}
                {...getRootProps()}
                tabIndex={0}
            >
                <input
                    ref={ref}
                    id={basicFieldProps.id}
                    name={basicFieldProps.name ?? "files"}
                    value=""
                    onFocusCapture={() => buttonRef.current?.focus()}
                    aria-hidden="true"
                    {...getInputProps()}
                    {...rest}
                    className={cn("block", hiddenInputStyles)}
                    style={{ display: "block" }}
                />
                <svg
                    xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none"
                    stroke="currentColor"
                    strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" className="w-5 h-5"
                >
                    <path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4" />
                    <polyline points="7 10 12 15 17 10" />
                    <line x1="12" x2="12" y1="15" y2="3" />
                </svg>
                <span>
                    {dropzoneText ?? "Click or drag file to this area to upload"}
                </span>
            </button>

            {maxSize && <div className={cn(SimpleDropzoneAnatomy.maxSizeText(), maxSizeTextClass)}>{`≤`} {humanFileSize(maxSize, 0)}</div>}

            {!withImagePreview && <div className={cn(SimpleDropzoneAnatomy.list(), listClass)}>
                {files?.map((file: any, index) => {

                    let Icon: React.ReactElement

                    if (["image/jpeg", "image/png", "image/jpg", "image/webm"].includes(file.type)) {
                        Icon = <svg
                            xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24"
                            fill="none" stroke="currentColor"
                            strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"
                            className={cn(SimpleDropzoneAnatomy.fileIcon(), fileIconClass)}
                        >
                            <path d="M14.5 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V7.5L14.5 2z" />
                            <polyline points="14 2 14 8 20 8" />
                            <circle cx="10" cy="13" r="2" />
                            <path d="m20 17-1.09-1.09a2 2 0 0 0-2.82 0L10 22" />
                        </svg>
                    } else if (file.type.includes("video")) {
                        Icon = <svg
                            xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24"
                            fill="none" stroke="currentColor"
                            strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"
                            className={cn(SimpleDropzoneAnatomy.fileIcon(), fileIconClass)}
                        >
                            <path d="M14.5 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V7.5L14.5 2z" />
                            <polyline points="14 2 14 8 20 8" />
                            <path d="m10 11 5 3-5 3v-6Z" />
                        </svg>
                    } else if (file.type.includes("audio")) {
                        Icon = <svg
                            xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24"
                            fill="none" stroke="currentColor"
                            strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"
                            className={cn(SimpleDropzoneAnatomy.fileIcon(), fileIconClass)}
                        >
                            <path
                                d="M17.5 22h.5c.5 0 1-.2 1.4-.6.4-.4.6-.9.6-1.4V7.5L14.5 2H6c-.5 0-1 .2-1.4.6C4.2 3 4 3.5 4 4v3"
                            />
                            <polyline points="14 2 14 8 20 8" />
                            <path d="M10 20v-1a2 2 0 1 1 4 0v1a2 2 0 1 1-4 0Z" />
                            <path d="M6 20v-1a2 2 0 1 0-4 0v1a2 2 0 1 0 4 0Z" />
                            <path d="M2 19v-3a6 6 0 0 1 12 0v3" />
                        </svg>
                    } else if (file.type.includes("pdf") || file.type.includes("document") || file.type.includes("txt") || file.type.includes("text")) {
                        Icon = <svg
                            xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24"
                            fill="none" stroke="currentColor"
                            strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"
                            className={cn(SimpleDropzoneAnatomy.fileIcon(), fileIconClass)}
                        >
                            <path d="M14.5 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V7.5L14.5 2z" />
                            <polyline points="14 2 14 8 20 8" />
                            <line x1="16" x2="8" y1="13" y2="13" />
                            <line x1="16" x2="8" y1="17" y2="17" />
                            <line x1="10" x2="8" y1="9" y2="9" />
                        </svg>
                    } else if (file.type.includes("compressed") || file.type.includes("zip") || file.type.includes("archive")) {
                        Icon = <svg
                            xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24"
                            fill="none" stroke="currentColor"
                            strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"
                            className={cn(SimpleDropzoneAnatomy.fileIcon(), fileIconClass)}
                        >
                            <path
                                d="M22 20V8a2 2 0 0 0-2-2h-7.93a2 2 0 0 1-1.66-.9l-.82-1.2A2 2 0 0 0 7.93 3H4a2 2 0 0 0-2 2v13c0 1.1.9 2 2 2h6"
                            />
                            <circle cx="16" cy="19" r="2" />
                            <path d="M16 11v-1" />
                            <path d="M16 17v-2" />
                        </svg>
                    } else {
                        Icon = <svg
                            xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24"
                            fill="none" stroke="currentColor"
                            strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"
                            className={cn(SimpleDropzoneAnatomy.fileIcon(), fileIconClass)}
                        >
                            <path d="M14.5 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V7.5L14.5 2z" />
                            <polyline points="14 2 14 8 20 8" />
                        </svg>
                    }

                    return (

                        <div
                            key={file.name}
                            className={cn(SimpleDropzoneAnatomy.listItem(), listItemClass)}
                        >
                            <div
                                className={cn(SimpleDropzoneAnatomy.listItemDetailsContainer(), listItemDetailsContainerClass)}
                            >
                                {Icon}
                                <p className={cn(SimpleDropzoneAnatomy.listItemTitle(), listItemTitleClass)}>{file.name}</p>
                                <p className={cn(SimpleDropzoneAnatomy.listItemSize(), listItemSizeClass)}>{humanFileSize(file.size)}</p>
                            </div>
                            <IconButton
                                size="xs"
                                intent="gray-basic"
                                icon={
                                    <svg
                                        xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24"
                                        fill="none"
                                        stroke="currentColor" strokeWidth="2" strokeLinecap="round"
                                        strokeLinejoin="round"
                                        className="w-4 h-4"
                                    >
                                        <path d="M3 6h18" />
                                        <path d="M19 6v14c0 1-1 2-2 2H7c-1 0-2-1-2-2V6" />
                                        <path d="M8 6V4c0-1 1-2 2-2h4c1 0 2 1 2 2v2" />
                                        <line x1="10" x2="10" y1="11" y2="17" />
                                        <line x1="14" x2="14" y1="11" y2="17" />
                                    </svg>
                                }
                                className={cn(SimpleDropzoneAnatomy.listItemRemoveButton(), listItemRemoveButtonClass)}
                                onClick={() => handleRemoveFile(index)}
                            />
                        </div>
                    )
                })}
            </div>}

            {withImagePreview && !!files.length && <div className={cn(SimpleDropzoneAnatomy.imagePreviewGrid(), imagePreviewGridClass)}>
                {files?.map((file, index) => {
                    return (
                        <div
                            key={file.name}
                            className={cn(SimpleDropzoneAnatomy.imagePreviewContainer(), imagePreviewContainerClass)}
                        >
                            <div
                                className={cn(SimpleDropzoneAnatomy.imagePreview(), imagePreviewClass)}
                                style={{ backgroundImage: file ? `url(${(file as File & { preview: string }).preview})` : undefined }}
                            >
                                <CloseButton
                                    intent="alert"
                                    size="xs"
                                    className={cn(SimpleDropzoneAnatomy.imagePreviewRemoveButton(), imagePreviewRemoveButtonClass)}
                                    onClick={() => handleRemoveFile(index)}
                                />
                            </div>
                            <div className={cn(SimpleDropzoneAnatomy.listItemDetailsContainer(), listItemDetailsContainerClass)}>
                                <p className={cn(SimpleDropzoneAnatomy.listItemTitle(), listItemTitleClass)}>{file.name}</p>
                                <p className={cn(SimpleDropzoneAnatomy.listItemSize(), listItemSizeClass)}>{humanFileSize(file.size)}</p>
                            </div>
                        </div>
                    )
                })}
            </div>}

        </BasicField>
    )

})

SimpleDropzone.displayName = "SimpleDropzone"

function humanFileSize(size: number, precision = 2): string {
    const i = Math.floor(Math.log(size) / Math.log(1024))
    return (size / Math.pow(1024, i)).toFixed(precision).toString() + ["bytes", "Kb", "Mb", "Gb", "Tb"][i]
}
