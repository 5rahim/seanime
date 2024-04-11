"use client"
import { serverStatusAtom } from "@/atoms/server-status"
import { PageWrapper } from "@/components/shared/styling/page-wrapper"
import { useAtomValue } from "jotai"

export default function Page() {

    const status = useAtomValue(serverStatusAtom)
    const settings = status?.settings

    if (!settings) return null

    return (
        <PageWrapper
            className="p-4 sm:p-8 pt-4 relative"
        >
            <div
                className="w-[80%] h-[calc(100vh-15rem)] rounded-xl border  overflow-hidden mx-auto mt-10 ring-1 ring-[--border] ring-offset-2"
            >

            </div>
        </PageWrapper>
    )
}
