import { cn } from "@/components/ui/core/styling"
import { __isDesktop__ } from "@/types/constants"
import Link, { LinkProps } from "next/link"
import { useRouter } from "next/navigation"
import React from "react"

type SeaLinkProps = { href: string | undefined } & Omit<LinkProps, "href"> & React.ComponentPropsWithRef<"a">

export const SeaLink = React.forwardRef((props: SeaLinkProps, _) => {

    const {
        href,
        children,
        className,
        onClick,
        ...rest
    } = props

    const router = useRouter()

    if (!href) return (
        <a
            className={cn(
                "cursor-pointer",
                className,
            )}
            onClick={e => {
                if (onClick) {
                    onClick(e)
                } else {
                    router.push(href as string)
                }
            }}
            data-current={(rest as any)["data-current"]}
            {...rest}
        >
            {children}
        </a>
    )

    if (__isDesktop__ && rest.target !== "_blank") {
        return (
            <a
                className={cn(
                    "cursor-pointer",
                    className,
                )}
                onClick={e => {
                    router.push(href as string)
                    if (onClick) {
                        onClick(e)
                    }
                }}
                data-current={(rest as any)["data-current"]}
            >
                {children}
            </a>
        )
    }

    return (
        <Link href={href} className={cn("cursor-pointer", className)} onClick={onClick} {...rest}>
            {children}
        </Link>
    )
})
