"use client"
import { useRouter } from "next/navigation"
import React from "react"

export default function Layout({ children }: { children: React.ReactNode }) {

    const router = useRouter()

    React.useEffect(() => {
        router.push("/")
    }, [])

    return null

    // return (
    //     <>
    //         {/*[CUSTOM UI]*/}
    //         <CustomBackgroundImage />
    //         {children}
    //     </>
    // )

}

export const dynamic = "force-static"
