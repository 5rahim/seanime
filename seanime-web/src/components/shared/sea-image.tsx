import NextImage, { ImageProps } from "next/image"

const HIDE_IMAGES = false

export function SeaImage(props: ImageProps) {

    if (HIDE_IMAGES) {
        return <NextImage
            {...props}
            src="/no-cover.png"
        />
    }

    return <NextImage {...props} />
}
