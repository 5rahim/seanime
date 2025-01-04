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
            <a
                className={cn(
                    "cursor-pointer",
                    className,
                )}
                onClick={() => {
                    router.push(href as string)
                }}
                data-current={(rest as any)["data-current"]}
            >
                {children}
            </a>
        )
    }

    return (
        <Link href={href} className={cn("cursor-pointer", className)} {...rest}>
            {children}
        </Link>
    )
})
