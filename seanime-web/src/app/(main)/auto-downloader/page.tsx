"use client"
import { serverStatusAtom } from "@/atoms/server-status"
import { Button } from "@/components/ui/button"
import { cn } from "@/components/ui/core"
import { Divider } from "@/components/ui/divider"
import { LoadingSpinner } from "@/components/ui/loading-spinner"
import { TabPanels } from "@/components/ui/tabs"
import { createTypesafeFormSchema, Field, TypesafeForm } from "@/components/ui/typesafe-form"
import { SeaEndpoints } from "@/lib/server/endpoints"
import { useSeaMutation, useSeaQuery } from "@/lib/server/queries/utils"
import { AutoDownloaderRule } from "@/lib/server/types"
import { BiPlus } from "@react-icons/all-files/bi/BiPlus"
import { useQueryClient } from "@tanstack/react-query"
import { useAtomValue } from "jotai/react"
import { InferType } from "prop-types"
import React from "react"
import toast from "react-hot-toast"

const settingsSchema = createTypesafeFormSchema(({ z }) => z.object({
    interval: z.number().min(2),
    enabled: z.boolean(),
    downloadAutomatically: z.boolean(),
}))

export default function Page() {
    const serverStatus = useAtomValue(serverStatusAtom)
    const qc = useQueryClient()

    const { mutate: updateSettings, isPending } = useSeaMutation<null, InferType<typeof settingsSchema>>({
        mutationKey: ["auto-downloader-settings"],
        endpoint: SeaEndpoints.AUTO_DOWNLOADER_SETTINGS,
        method: "patch",
        onSuccess: async () => {
            await qc.refetchQueries({ queryKey: ["status"] })
            await qc.refetchQueries({ queryKey: ["auto-downloader-rules"] })
            toast.success("Settings updated")
        },
    })


    const { data, isLoading } = useSeaQuery<AutoDownloaderRule[] | null>({
        queryKey: ["auto-downloader-rules"],
        endpoint: SeaEndpoints.AUTO_DOWNLOADER_RULES,
    })


    return (
        <div className="space-y-4">

            <TabPanels
                navClassName="border-[--border]"
                tabClassName={cn(
                    "text-sm rounded-none border-b border-b-2 data-[selected=true]:text-white data-[selected=true]:border-brand-400",
                    "hover:bg-transparent dark:hover:bg-transparent hover:text-white",
                    "dark:border-transparent dark:hover:border-b-transparent dark:data-[selected=true]:border-brand-400 dark:data-[selected=true]:text-white",
                    "dark:data-[selected=true]:bg-[--highlight]",
                )}
            >
                <TabPanels.Nav>
                    <TabPanels.Tab>Rules</TabPanels.Tab>
                    <TabPanels.Tab>Settings</TabPanels.Tab>
                </TabPanels.Nav>
                <TabPanels.Container>

                    <TabPanels.Panel>
                        <div className="p-4">
                            {isLoading && <LoadingSpinner />}
                            {!isLoading && (
                                <div>
                                    <div className="w-full flex justify-end">
                                        <Button
                                            className="rounded-full"
                                            intent="success-subtle"
                                            leftIcon={<BiPlus />}
                                            onClick={() => {
                                                // openModal("add-rule")
                                            }}
                                        >
                                            Add Rule
                                        </Button>
                                    </div>
                                    {(!data?.length) && <div className="p-4 text-[--muted] text-center">No rules</div>}
                                </div>
                            )}
                        </div>
                    </TabPanels.Panel>


                    <TabPanels.Panel>
                        <div className="p-4">
                            <TypesafeForm
                                schema={settingsSchema}
                                onSubmit={data => {
                                    updateSettings(data)
                                }}
                                defaultValues={{
                                    enabled: serverStatus?.settings?.autoDownloader?.enabled ?? false,
                                    interval: serverStatus?.settings?.autoDownloader?.interval ?? 10,
                                    downloadAutomatically: serverStatus?.settings?.autoDownloader?.downloadAutomatically ?? false,
                                }}
                            >
                                {(f) => (
                                    <>
                                        <Field.Switch
                                            label="Enabled"
                                            name="enabled"
                                        />

                                        <Divider />

                                        {<div
                                            className={cn(
                                                !f.watch("enabled") && "pointer-events-none opacity-50 space-y-4",
                                            )}
                                        >
                                            <Field.Checkbox
                                                label="Download immediately"
                                                name="downloadAutomatically"
                                                help="Download new episodes as soon as they are found."
                                            />
                                            <Field.Number
                                                label="Interval"
                                                help="How often to check for new episodes."
                                                name="interval"
                                                leftAddon="Every"
                                                rightAddon="minutes"
                                                discrete
                                                size="sm"
                                                className="text-center w-20"
                                                min={2}
                                            />
                                        </div>}

                                        <Field.Submit role="save" isLoading={isPending} />
                                    </>
                                )}
                            </TypesafeForm>
                        </div>
                    </TabPanels.Panel>

                </TabPanels.Container>
            </TabPanels>

        </div>
    )

}
