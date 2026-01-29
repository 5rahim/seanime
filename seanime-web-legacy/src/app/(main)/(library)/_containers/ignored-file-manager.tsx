import { Anime_LocalFile } from "@/api/generated/types"
import { useOpenInExplorer } from "@/api/hooks/explorer.hooks"
import { useUpdateLocalFiles } from "@/api/hooks/localfiles.hooks"
import { LuffyError } from "@/components/shared/luffy-error"
import { AppLayoutStack } from "@/components/ui/app-layout"
import { Button } from "@/components/ui/button"
import { Checkbox } from "@/components/ui/checkbox"
import { Drawer } from "@/components/ui/drawer"
import { upath } from "@/lib/helpers/upath"
import { atom } from "jotai"
import { useAtom } from "jotai/react"
import React from "react"
import { TbFileSad } from "react-icons/tb"
import { toast } from "sonner"

export const __ignoredFileManagerIsOpen = atom(false)

type IgnoredFileManagerProps = {
    files: Anime_LocalFile[]
}

export function IgnoredFileManager(props: IgnoredFileManagerProps) {

    const { files } = props

    const [isOpen, setIsOpen] = useAtom(__ignoredFileManagerIsOpen)

    const { mutate: openInExplorer } = useOpenInExplorer()

    const { mutate: updateLocalFiles, isPending: isUpdating } = useUpdateLocalFiles()

    const [selectedPaths, setSelectedPaths] = React.useState<string[]>([])

    React.useLayoutEffect(() => {
        setSelectedPaths(files?.map(lf => lf.path) ?? [])
    }, [files])

    function handleUnIgnoreSelected() {
        if (selectedPaths.length > 0) {
            updateLocalFiles({
                paths: selectedPaths,
                action: "unignore",
            }, {
                onSuccess: () => {
                    toast.success("Files un-ignored")
                },
            })
        }
    }


    return (
        <Drawer
            open={isOpen}
            onOpenChange={() => setIsOpen(false)}
            // contentClass="max-w-5xl"
            size="xl"
            title="Ignored files"
        >
            <AppLayoutStack className="mt-4">

                {files.length > 0 && <div className="flex flex-wrap items-center gap-2">
                    <div className="flex flex-1"></div>
                    <Button
                        leftIcon={<TbFileSad className="text-lg" />}
                        intent="white"
                        size="sm"
                        rounded
                        loading={isUpdating}
                        onClick={handleUnIgnoreSelected}
                    >
                        Un-ignore selection
                    </Button>
                </div>}

                {files.length === 0 && <LuffyError title={null}>
                    No ignored files
                </LuffyError>}

                {files.length > 0 &&
                    <div className="bg-gray-950 border p-2 px-2 divide-y divide-[--border] rounded-[--radius-md] max-h-[85vh] max-w-full overflow-x-auto overflow-y-auto text-sm">

                        <div className="p-2">
                            <Checkbox
                                label={`Select all files`}
                                value={(selectedPaths.length === files?.length) ? true : (selectedPaths.length === 0
                                    ? false
                                    : "indeterminate")}
                                onValueChange={checked => {
                                    if (typeof checked === "boolean") {
                                        setSelectedPaths(draft => {
                                            if (draft.length === files?.length) {
                                                return []
                                            } else {
                                                return files?.map(lf => lf.path) ?? []
                                            }
                                        })
                                    }
                                }}
                                fieldClass="w-[fit-content]"
                            />
                        </div>

                        {files.map((lf, index) => (
                            <div
                                key={`${lf.path}-${index}`}
                                className="p-2 "
                            >
                                <div className="flex items-center">
                                    <Checkbox
                                        label={`${upath.basename(lf.path)}`}
                                        value={selectedPaths.includes(lf.path)}
                                        onValueChange={checked => {
                                            if (typeof checked === "boolean") {
                                                setSelectedPaths(draft => {
                                                    if (checked) {
                                                        return [...draft, lf.path]
                                                    } else {
                                                        return draft.filter(p => p !== lf.path)
                                                    }
                                                })
                                            }
                                        }}
                                        fieldClass="w-[fit-content]"
                                    />
                                </div>
                            </div>
                        ))}
                    </div>}

            </AppLayoutStack>
        </Drawer>
    )

}
