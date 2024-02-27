"use client"
import { Button } from "@/components/ui/button/button"
import { cn } from "@/components/ui/core/styling"
import Image from "next/image"
import { useRouter } from "next/navigation"
import React from "react"

interface LuffyErrorProps {
    children?: React.ReactNode
    className?: string
    reset?: () => void
    title?: string | null
    showRefreshButton?: boolean
}

export const LuffyError: React.FC<LuffyErrorProps> = (props) => {

    const { children, reset, className, title = "Oops!", showRefreshButton = false, ...rest } = props

    const router = useRouter()


    return (
        <>
            <div className={cn("w-full flex flex-col items-center mt-10 space-y-4", className)}>
                {<div
                    className="h-[10rem] w-[10rem] mx-auto flex-none rounded-md object-cover object-center relative overflow-hidden">
                    <Image
                        src="/luffy-01.png"
                        alt={""}
                        fill
                        quality={100}
                        priority
                        sizes="10rem"
                        className="object-contain object-top"
                    />
                </div>}
                <div className="text-center space-y-4">
                    {!!title && <h2>{title}</h2>}
                    <p>{children}</p>
                    <div>
                        {(showRefreshButton && !reset) && (
                            <Button intent="warning-subtle" onClick={() => router.refresh()}>Retry</Button>
                        )}
                        {!!reset && (
                            <Button intent="warning-subtle" onClick={reset}>Retry</Button>
                        )}
                    </div>
                </div>
            </div>
        </>
    )

}
