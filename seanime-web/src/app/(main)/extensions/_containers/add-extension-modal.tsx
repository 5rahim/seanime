import { Extension_Extension } from "@/api/generated/types"
import { useFetchExternalExtensionData, useInstallExternalExtension } from "@/api/hooks/extensions.hooks"
import { ExtensionDetails } from "@/app/(main)/extensions/_components/extension-details"
import { Button } from "@/components/ui/button"
import { Modal } from "@/components/ui/modal"
import { Separator } from "@/components/ui/separator"
import { TextInput } from "@/components/ui/text-input"
import React from "react"
import { FiSearch } from "react-icons/fi"
import { toast } from "sonner"

type AddExtensionModalProps = {
    extensions: Extension_Extension[] | undefined
    children?: React.ReactElement
}

export function AddExtensionModal(props: AddExtensionModalProps) {

    const {
        extensions,
        children,
        ...rest
    } = props

    const [open, setOpen] = React.useState(false)
    const [manifestURL, setManifestURL] = React.useState<string>("")

    const { mutate: fetchExtensionData, data: extensionData, isPending, reset } = useFetchExternalExtensionData(null)

    const {
        mutate: installExtension,
        data: installResponse,
        isPending: isInstalling,
    } = useInstallExternalExtension()

    React.useEffect(() => {
        if (installResponse) {
            toast.success(installResponse.message)
            setOpen(false)
            reset()
        }
    }, [installResponse])

    function handleFetchExtensionData() {
        if (!manifestURL) {
            toast.warning("Please provide a valid URL.")
            return
        }

        fetchExtensionData({
            manifestUri: manifestURL,
        })
    }

    return (
        <>
            <Modal
                open={open}
                onOpenChange={setOpen}
                trigger={children}
                contentClass="max-w-3xl"
            >
                <div className="flex gap-4 flex-col lg:flex-row">
                    <div className="lg:w-1/3">
                        <h3 className="text-2xl font-bold">Install from URL</h3>
                        <p className="text-[--muted]">Install an extension by entering URL of the manifest file.</p>
                    </div>
                    <div className="lg:w-2/3 gap-3 flex flex-col">
                        <TextInput
                            placeholder="https://example.com/extension.json"
                            value={manifestURL}
                            onValueChange={setManifestURL}
                            label="URL"
                        />
                        <Button
                            leftIcon={<FiSearch />}
                            intent="gray-outline"
                            onClick={handleFetchExtensionData}
                            loading={isPending}
                        >Check</Button>
                    </div>
                </div>

                {!!extensionData && (
                    <>
                        <Separator />

                        <ExtensionDetails extension={extensionData} />

                        {extensions?.find(n => n.id === extensionData.id) ? (
                            <p className="text-center">
                                This extension is already installed.
                            </p>
                        ) : (
                            <Button
                                intent="white"
                                loading={isInstalling}
                                onClick={() => {
                                    installExtension({
                                        manifestUri: extensionData?.manifestURI,
                                    })
                                }}
                            >Install</Button>
                        )}
                    </>
                )}

            </Modal>
        </>
    )
}
