import { CustomBackgroundImage } from "@/app/(main)/_features/custom-ui/custom-background-image"
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
