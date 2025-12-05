import { HIDE_IMAGES } from "@/types/constants"
import NextImage, { ImageProps } from "next/image"
import React from "react"

export function SeaImage({ isExternal, ...props }: ImageProps & { isExternal?: boolean }) {
    const [overrideSrc, setOverrideSrc] = React.useState<string | undefined>(undefined)

    if (HIDE_IMAGES) {
        return <NextImage
            {...props}
            src="/no-cover.png"
        />
    }

    const blocked = isExternal && props.src && !(
        (props.src as string).endsWith(".png")
        || (props.src as string).endsWith(".jpg")
        || (props.src as string).endsWith(".jpeg")
        || (props.src as string).endsWith(".avif")
        || (props.src as string).endsWith(".webp")
        || (props.src as string).endsWith(".ico")
    )

    return <NextImage
        {...props}
        overrideSrc={blocked ? "/no-cover.png" : overrideSrc}
        // onError={() => setOverrideSrc("/no-cover.png")} // stops retries?
    />
}
