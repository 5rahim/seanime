import { CustomBackgroundImage } from "@/components/shared/custom-ui/custom-background-image"
import React from "react"

export default function Layout({ children }: { children: React.ReactNode }) {

    return (
        <>
            {/*[CUSTOM UI]*/}
            <CustomBackgroundImage />
            {children}
        </>
    )

}
