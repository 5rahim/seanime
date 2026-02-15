import { cn } from "@/components/ui/core/styling"
import { Link } from "@tanstack/react-router"
import React from "react"

type SeaLinkProps = React.ComponentPropsWithRef<"a"> & { href: string | undefined, resetScroll?: boolean }

export const SeaLink = React.forwardRef<HTMLAnchorElement, SeaLinkProps>((props, ref) => {
    const {
        href,
        children,
        className,
        onClick,
        resetScroll = true,
        ...rest
    } = props

    // const navigate = useNavigate()

    const isExternal = href?.startsWith("http") || href?.startsWith("mailto")

    if (!href || isExternal) {
        return (
            <a
                ref={ref}
                href={href}
                className={cn("cursor-pointer", className)}
                onClick={onClick}
                {...rest}
            >
                {children}
            </a>
        )
    }

    const [pathname, searchString] = href.split("?")
    const searchParams: Record<string, any> = {}

    if (searchString) {
        const urlSearchParams = new URLSearchParams(searchString)
        urlSearchParams.forEach((value, key) => {
            const numValue = Number(value)
            const isNumeric = !isNaN(numValue) && value.trim() !== ""
            searchParams[key] = isNumeric ? numValue : value
        })
    }

    return (
        <Link
            to={pathname}
            search={Object.keys(searchParams).length > 0 ? () => searchParams : undefined}
            className={cn("cursor-pointer", className)}
            resetScroll={resetScroll}
            onClick={onClick}
            {...rest}
        >
            {children}
        </Link>
    )

    // return (
    //     <a
    //         ref={ref}
    //         href={href}
    //         className={cn("cursor-pointer", className)}
    //         {...rest}
    //         onClick={(e) => {
    //             if (e.metaKey || e.altKey || e.ctrlKey || e.shiftKey || e.button !== 0) {
    //                 if (onClick) onClick(e)
    //                 return
    //             }
    //
    //             e.preventDefault()
    //
    //             if (onClick) onClick(e)
    //
    //             navigate({
    //                 to: pathname,
    //                 search: () => searchParams,
    //                 replace: false,
    //             }).then(() => {
    //                 if (resetScroll) window.scrollTo(0, 0)
    //             })
    //         }}
    //     >
    //         {children}
    //     </a>
    // )
})
