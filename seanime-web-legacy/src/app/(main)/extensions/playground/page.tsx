"use client"

import { Extension_Language, Extension_Type } from "@/api/generated/types"
import { CustomLibraryBanner } from "@/app/(main)/(library)/_containers/custom-library-banner"
import { ExtensionPlayground } from "@/app/(main)/extensions/playground/_containers/extension-playground"
import { PageWrapper } from "@/components/shared/page-wrapper"
import React from "react"

export default function Page() {

    const [extensionLanguage, setExtensionLanguage] = React.useState<Extension_Language>("typescript")
    const [extensionType, setExtensionType] = React.useState<Extension_Type>("anime-torrent-provider")
    const [code, setCode] = React.useState<string>("")

    return (
        <>
            <CustomLibraryBanner discrete />
            <PageWrapper className="p-4 sm:p-8 pt-0 space-y-8 relative z-[4]">
                <ExtensionPlayground
                    language={extensionLanguage}
                    type={extensionType}
                    onLanguageChange={setExtensionLanguage}
                    onTypeChange={setExtensionType}
                    code={code}
                    onCodeChange={setCode}
                />
            </PageWrapper>
        </>
    )

}
