"use client"
import { serverStatusAtom } from "@/atoms/server-status"
import { PageWrapper } from "@/components/shared/styling/page-wrapper"
import { IconButton } from "@/components/ui/button"
import { Form } from "@/components/ui/form"
import { settingsSchema } from "@/lib/server/settings"
import { useAtom } from "jotai/react"
import Link from "next/link"
import React from "react"
import { AiOutlineArrowLeft } from "react-icons/ai"


export default function Page() {
    const [status, setServerStatus] = useAtom(serverStatusAtom)

    // const { mutate, data, isPending } = useSeaMutation<ServerStatus, Settings>({
    //     endpoint: SeaEndpoints.SETTINGS,
    //     mutationKey: ["patch-settings"],
    //     method: "patch",
    //     onSuccess: () => {
    //         toast.success("Settings updated")
    //     },
    // })

    return (
        <PageWrapper className="p-4 sm:p-8 space-y-4">
            <div className="flex gap-4 items-center">
                <Link href={`/settings`}>
                    <IconButton icon={<AiOutlineArrowLeft />} rounded intent="white-outline" size="sm" />
                </Link>
                <div className="space-y-1">
                    <h2>Theme</h2>
                    <p className="text-[--muted]">
                        Change the look and feel of Seanime
                    </p>
                </div>
            </div>
            {/*<Separator/>*/}
            <Form
                schema={settingsSchema}
                onSubmit={data => {

                }}
                defaultValues={{}}
                stackClass="space-y-4"
            >


            </Form>
        </PageWrapper>
    )

}
