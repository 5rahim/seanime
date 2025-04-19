import { Extension_Extension, Extension_InvalidExtension } from "@/api/generated/types"
import { useGetExtensionUserConfig, useSaveExtensionUserConfig } from "@/api/hooks/extensions.hooks"
import { LuffyError } from "@/components/shared/luffy-error"
import { Alert } from "@/components/ui/alert"
import { Button } from "@/components/ui/button"
import { cn } from "@/components/ui/core/styling"
import { LoadingSpinner } from "@/components/ui/loading-spinner"
import { Modal } from "@/components/ui/modal"
import { Select } from "@/components/ui/select"
import { Switch } from "@/components/ui/switch"
import { TextInput } from "@/components/ui/text-input"
import { atomWithImmer } from "jotai-immer"
import { useAtom } from "jotai/react"
import React from "react"

type ExtensionUserConfigModalProps = {
    children?: React.ReactElement
    extension: Extension_Extension
    userConfigError?: Extension_InvalidExtension | undefined
}

export function ExtensionUserConfigModal(props: ExtensionUserConfigModalProps) {

    const {
        children,
        extension,
        userConfigError,
        ...rest
    } = props

    return (
        <>
            <Modal
                contentClass="max-w-3xl"
                trigger={children}
                title="Preferences"
                // size="xl"
                // contentClass="space-y-4"
            >
                <Content extension={extension} userConfigError={userConfigError} />
            </Modal>
        </>
    )
}

const userConfigFormValuesAtom = atomWithImmer<Record<string, string>>({})

function Content({ extension, userConfigError }: { extension: Extension_Extension, userConfigError?: Extension_InvalidExtension | undefined }) {

    const { data: extUserConfig, isLoading } = useGetExtensionUserConfig(extension.id)

    const { mutate: saveExtUserConfig, isPending } = useSaveExtensionUserConfig()

    const [userConfigFormValues, setUserConfigFormValues] = useAtom(userConfigFormValuesAtom)

    React.useLayoutEffect(() => {
        if (extUserConfig) {
            for (const field of extUserConfig.userConfig?.fields || []) {
                if (extUserConfig.savedUserConfig?.values && field.name in extUserConfig.savedUserConfig?.values) {
                    setUserConfigFormValues(draft => {
                        draft[field.name] = extUserConfig.savedUserConfig?.values?.[field.name] || field.default || ""
                        return
                    })
                }
            }
        }
    }, [extUserConfig])

    function handleSave() {
        console.log("Saving user config", userConfigFormValues)
        let values: Record<string, string> = {}
        for (const field of extUserConfig?.userConfig?.fields || []) {
            values[field.name] = userConfigFormValues[field.name] || field.default || ""
        }
        saveExtUserConfig({
            id: extension.id,
            version: extUserConfig?.userConfig?.version || 0,
            values: values,
        })
    }

    if (isLoading) return <LoadingSpinner />

    if (!extUserConfig) return <LuffyError />

    return (
        <>
            <div>
                <p>
                    {extension.name}
                </p>
                <div className="text-sm text-[--muted]">
                    You can edit the preferences for this extension here.
                </div>
            </div>

            {userConfigError && (
                <Alert
                    intent="alert-basic"
                    title="Config error"
                    description={userConfigError.reason}
                />
            )}

            {extUserConfig?.userConfig?.fields?.map(field => {
                if (field.type === "text") {
                    return (
                        <TextInput
                            key={field.name}
                            label={field.label}
                            value={userConfigFormValues[field.name] || field.default}
                            onValueChange={v => setUserConfigFormValues(draft => {
                                draft[field.name] = v
                                return
                            })}
                            help={!!field.default ? `Default: ${field.default}` : undefined}
                        />
                    )
                }
                if (field.type === "switch") {
                    return (
                        <Switch
                            key={field.name}
                            label={field.label}
                            value={userConfigFormValues[field.name] ? userConfigFormValues[field.name] === "true" : field.default === "true"}
                            onValueChange={v => setUserConfigFormValues(draft => {
                                draft[field.name] = v ? "true" : "false"
                                return
                            })}
                            help={!!field.default ? `Default: ${field.default}` : undefined}
                        />
                    )
                }
                if (field.type === "select" && field.options) {
                    return (
                        <Select
                            key={field.name}
                            label={field.label}
                            value={userConfigFormValues[field.name] || field.default}
                            onValueChange={v => setUserConfigFormValues(draft => {
                                draft[field.name] = v
                                return
                            })}
                            options={field.options}
                            help={!!field.default ? `Default: ${field.options.find(n => n.value === field.default)?.label ?? "N/A"}` : undefined}
                        />
                    )
                }
            })}

            <div className="flex">
                <Button
                    intent="white"
                    loading={isPending}
                    onClick={handleSave}
                    className={cn(!!userConfigError && "animate-pulse")}
                >
                    Save
                </Button>
                <div className="flex flex-1"></div>
            </div>
        </>
    )
}
