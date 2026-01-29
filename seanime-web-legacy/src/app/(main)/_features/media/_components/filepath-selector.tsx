import { Checkbox } from "@/components/ui/checkbox"
import { cn } from "@/components/ui/core/styling"
import { Separator } from "@/components/ui/separator"
import { upath } from "@/lib/helpers/upath"
import React from "react"

type FilepathSelectorProps = {
    filepaths: string[]
    allFilepaths: string[]
    onFilepathSelected: React.Dispatch<React.SetStateAction<string[]>>
    showFullPath?: boolean
} & React.ComponentPropsWithoutRef<"div">

export function FilepathSelector(props: FilepathSelectorProps) {

    const {
        filepaths,
        allFilepaths,
        onFilepathSelected,
        showFullPath,
        className,
        ...rest
    } = props

    const allFilesChecked = filepaths.length === allFilepaths.length

    return (
        <>
            <div
                className={cn(
                    "overflow-y-auto px-2 space-y-1",
                    className,
                )} {...rest}>

                <div className="">
                    <Checkbox
                        label="Select all files"
                        value={allFilesChecked ? true : filepaths.length === 0 ? false : "indeterminate"}
                        onValueChange={checked => {
                            if (typeof checked === "boolean") {
                                onFilepathSelected(checked ? allFilepaths : [])
                            }
                        }}
                        fieldClass="w-[fit-content]"
                    />
                </div>

                <Separator />

                <div className="divide-[--border] divide-y">
                    {allFilepaths?.toSorted((a, b) => a.localeCompare(b)).map((path, index) => (
                        <div
                            key={`${path}-${index}`}
                            className="py-2"
                        >
                            <div className="flex items-center">
                                <Checkbox
                                    label={<span className={cn("", showFullPath && "text-[--muted]")}>
                                        {showFullPath ? path.replace(upath.basename(path), "") : upath.basename(path)}{showFullPath &&
                                        <span className="text-[--foreground]">{upath.basename(path)}</span>}
                                    </span>}
                                    value={filepaths.includes(path)}
                                    onValueChange={checked => {
                                        if (typeof checked === "boolean") {
                                            onFilepathSelected(prev => checked
                                                ? [...prev, path]
                                                : prev.filter(p => p !== path),
                                            )
                                        }
                                    }}
                                    labelClass="break-all tracking-wide text-sm"
                                    fieldLabelClass="break-all"
                                    fieldClass="w-[fit-content]"
                                />
                            </div>
                        </div>
                    ))}
                </div>
            </div>
        </>
    )
}
