import { Extension_Extension } from "@/api/generated/types"
import { useFetchExternalExtensionData, useInstallExternalExtension, useInstallExternalExtensionRepository } from "@/api/hooks/extensions.hooks"
import { ExtensionDetails } from "@/app/(main)/extensions/_components/extension-details"
import { MarketplaceExtensionCard } from "@/app/(main)/extensions/_containers/marketplace-extensions"
import { Button } from "@/components/ui/button"
import { Modal } from "@/components/ui/modal"
import { Separator } from "@/components/ui/separator"
import { TextInput } from "@/components/ui/text-input"
import React from "react"
import { FiDownload } from "react-icons/fi"
import { LuSearch } from "react-icons/lu"
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
    const [repositoryURL, setRepositoryURL] = React.useState<string>("")

    const { mutate: fetchExtensionData, data: extensionData, isPending, reset } = useFetchExternalExtensionData(null)
    const {
        mutate: installFromRepository,
        data: repositoryData,
        isPending: isInstallingFromRepo,
        reset: resetRepo,
    } = useInstallExternalExtensionRepository()

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

    function handleInstallFromRepository(install: boolean) {
        if (!repositoryURL) {
            toast.warning("Please provide a valid URL.")
            return
        }

        installFromRepository({
            repositoryUri: repositoryURL,
            install: install,
        }, {
            onSuccess: () => {
                if (install) {
                    toast.success("Extensions installed successfully.")
                    setOpen(false)
                    setRepositoryURL("")
                    resetRepo()
                }
            },
        })
    }

    return (
        <>
            <Modal
                open={open}
                onOpenChange={setOpen}
                trigger={children}
                contentClass="max-w-3xl"
                titleClass="text-center pb-4"
                title="Add extensions"
            >
                <div className="flex gap-4 flex-col lg:flex-row">
                    <div className="lg:w-1/3">
                        <h3 className="text-2xl font-bold">Install from URL</h3>
                        <p className="text-[--muted]">Install an extension by entering the manifest URL.</p>
                    </div>
                    <div className="lg:w-2/3 gap-3 flex flex-col">
                        <TextInput
                            placeholder="https://example.com/extension.json"
                            value={manifestURL}
                            onValueChange={setManifestURL}
                            // label="URL"
                        />
                        <Button
                            leftIcon={<LuSearch />}
                            intent="white"
                            onClick={handleFetchExtensionData}
                            loading={isPending}
                        >Find</Button>
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

                {!extensionData && (
                    <>
                        <Separator />

                        <p className="text-center text-[--muted]">
                            You can also install many extensions at once by importing them from a repository.
                        </p>

                        <div className="flex gap-4 flex-col lg:flex-row-reverse">
                            <div className="lg:w-1/3">
                                <h3 className="text-xl font-bold">Import from repository</h3>
                                <p className="text-[--muted]">Import and automatically install extensions by entering a repository URL.</p>
                            </div>
                            <div className="lg:w-2/3 gap-3 flex flex-col">
                                <TextInput
                                    placeholder={"https://example.com/extensions.json or { \"urls\": [...] }"}
                                    value={repositoryURL}
                                    onValueChange={setRepositoryURL}
                                    // label="URL"
                                />
                                <Button
                                    leftIcon={<FiDownload />}
                                    intent="gray-outline"
                                    onClick={() => handleInstallFromRepository(false)}
                                    loading={isInstallingFromRepo}
                                >Import all</Button>
                            </div>
                        </div>

                        {!!repositoryData && (
                            <>
                                {repositoryData.extensions?.toSorted((a, b) => a.id.toLowerCase().localeCompare(b.id.toLowerCase())).map(ext => (
                                    <div key={ext.id} className="">
                                        <MarketplaceExtensionCard extension={ext} isInstalled={false} hideInstallButton showType />
                                    </div>
                                ))}
                                <Button
                                    leftIcon={<FiDownload />}
                                    intent="white"
                                    onClick={() => handleInstallFromRepository(true)}
                                    loading={isInstallingFromRepo}
                                >Install all</Button>
                            </>
                        )}
                    </>
                )}

            </Modal>
        </>
    )
}
