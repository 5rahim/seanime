import { getServerBaseUrl } from "@/api/client/server-url"
import { useImportLocalFiles } from "@/api/hooks/localfiles.hooks"
import { SeaLink } from "@/components/shared/sea-link"
import { Button } from "@/components/ui/button"
import { Modal } from "@/components/ui/modal"
import { TextInput } from "@/components/ui/text-input"
import React from "react"
import { CgImport } from "react-icons/cg"
import { TbDatabaseExport } from "react-icons/tb"

type DataSettingsProps = {
    children?: React.ReactNode
}

export function DataSettings(props: DataSettingsProps) {

    const {
        children,
        ...rest
    } = props

    const { mutate: importLocalFiles, isPending: isImportingLocalFiles } = useImportLocalFiles()
    const [localFileDataPath, setLocalFileDataPath] = React.useState("")

    function handleImportLocalFiles() {
        if (!localFileDataPath) return

        importLocalFiles({ dataFilePath: localFileDataPath }, {
            onSuccess: () => {
                setLocalFileDataPath("")
            },
        })
    }

    return (
        <div className="space-y-4">

            <div>
                <h5>Local files</h5>

                <p className="text-[--muted]">
                    Scanned local file data.
                </p>
            </div>

            <div className="flex flex-wrap gap-2">
                <SeaLink href={`${getServerBaseUrl()}/api/v1/library/local-files/dump`} target="_blank" className="block">
                    <Button
                        intent="primary-subtle"
                        leftIcon={<TbDatabaseExport className="text-xl" />}
                        size="md"
                        disabled={isImportingLocalFiles}
                    >
                        Export local file data
                    </Button>
                </SeaLink>

                <Modal
                    title="Import local files"
                    trigger={
                        <Button
                            intent="white-subtle"
                            leftIcon={<CgImport className="text-xl" />}
                            size="md"
                            disabled={isImportingLocalFiles}
                        >
                            Import local files
                        </Button>
                    }
                >

                    <p>
                        This will overwrite your existing library data, make sure you have a backup.
                    </p>

                    <TextInput
                        label="Data file path"
                        help="The path to the JSON file containing the local file data."
                        value={localFileDataPath}
                        onValueChange={setLocalFileDataPath}
                    />

                    <Button
                        intent="white"
                        rounded
                        onClick={handleImportLocalFiles}
                        disabled={isImportingLocalFiles}
                    >Import</Button>

                </Modal>
            </div>
        </div>
    )
}
