import Link, { LinkProps } from "next/link"
import { useRouter } from "next/navigation"
import React from "react"

type SeaLinkProps = {} & LinkProps & React.ComponentPropsWithoutRef<"a">

export function SeaLink(props: SeaLinkProps) {

    const {
        href,
        children,
        ...rest
    } = props

    const router = useRouter()

    if (process.env.NEXT_PUBLIC_PLATFORM === "desktop" && rest.target !== "_blank") {
        return (
            <div
                className="cursor-pointer"
                onClick={() => {
                    router.push(href as string)
                }}
            >
                {children}
            </div>
        )
    }

    return (
        <Link href={href} {...rest}>
            {children}
        </Link>
    )
}
