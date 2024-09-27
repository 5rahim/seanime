import { cn } from "@/components/ui/core/styling"
import Link, { LinkProps } from "next/link"
import { useRouter } from "next/navigation"
import React from "react"

type SeaLinkProps = {} & LinkProps & React.ComponentPropsWithRef<"a">

export const SeaLink = React.forwardRef((props: SeaLinkProps, _) => {

    const {
        href,
        children,
        className,
        ...rest
    } = props

    const router = useRouter()

    if (process.env.NEXT_PUBLIC_PLATFORM === "desktop" && rest.target !== "_blank") {
        return (
            <div
                className={cn(
                    "inline-block cursor-pointer",
                    className,
                )}
                onClick={() => {
                    router.push(href as string)
                }}
            >
                {children}
            </div>
        )
    }

    return (
        <Link href={href} className={className} {...rest}>
            {children}
        </Link>
    )
})
