import { HIDE_IMAGES } from "@/types/constants"
import NextImage, { ImageProps } from "next/image"

export function SeaImage(props: ImageProps) {

    if (HIDE_IMAGES) {
        return <NextImage
            {...props}
            src="/no-cover.png"
        />
    }

    return <NextImage {...props} />
}
