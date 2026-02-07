import { HIDE_IMAGES } from "@/types/constants"
import React, { forwardRef, useEffect, useState } from "react"

type ImageProps = React.ImgHTMLAttributes<HTMLImageElement> & {
    fill?: boolean
    priority?: boolean
    overrideSrc?: string
    quality?: number | string
    placeholder?: string
    blurDataURL?: string
    sizes?: string
}

export const SeaImage = forwardRef<HTMLImageElement, ImageProps & { isExternal?: boolean }>(
    ({ isExternal, fill, priority, quality, placeholder, sizes, ...props }, ref) => {
        const [hasError, setHasError] = useState(false)

        useEffect(() => {
            setHasError(false)
        }, [props.src])

        if (HIDE_IMAGES) {
            return <Image
                ref={ref}
                {...props}
                src="/no-cover.png"
                className={props.className}
                alt={props.alt || "cover"}
                fill={fill}
            />
        }

        const blocked = isExternal && props.src && typeof props.src === "string" && !(
            props.src.endsWith(".png")
            || props.src.endsWith(".jpg")
            || props.src.endsWith(".jpeg")
            || props.src.endsWith(".avif")
            || props.src.endsWith(".webp")
            || props.src.endsWith(".ico")
        )

        const effectiveOverride = (blocked || hasError) ? "/no-cover.png" : props.overrideSrc

        return <Image
            ref={ref}
            {...props}
            src={props.src || ""}
            alt={props.alt || ""}
            fill={fill}
            priority={priority}
            placeholder={placeholder}
            overrideSrc={effectiveOverride}
            onError={() => setHasError(true)}
        />
    },
)

SeaImage.displayName = "SeaImage"

interface _ImageProps extends React.ImgHTMLAttributes<HTMLImageElement> {
    src: string | any
    alt: string
    width?: number | string
    height?: number | string
    fill?: boolean
    quality?: number | string
    priority?: boolean
    loader?: any
    placeholder?: string
    blurDataURL?: string
    unoptimized?: boolean
    onLoadingComplete?: (img: HTMLImageElement) => void
    layout?: string
    objectFit?: string
    overrideSrc?: string
}

const Image = forwardRef<HTMLImageElement, _ImageProps>((
    {
        src,
        alt,
        width,
        height,
        fill,
        style,
        className,
        quality,
        priority,
        loader,
        placeholder,
        blurDataURL,
        unoptimized,
        onLoadingComplete,
        layout,
        objectFit,
        overrideSrc,
        onLoad,
        ...props
    },
    ref,
) => {
    const [isLoaded, setIsLoaded] = useState(false)

    const isStaticImport = typeof src === "object" && src !== null && "src" in src
    const imageSrc = overrideSrc || (isStaticImport ? src.src : src)

    const staticBlur = isStaticImport ? src.blurDataURL : undefined

    useEffect(() => {
        setIsLoaded(false)
    }, [imageSrc])

    const blurUrl = (placeholder && placeholder !== "blur" && placeholder !== "empty")
        ? placeholder
        : (placeholder === "blur" ? (blurDataURL || staticBlur) : undefined)

    const fillStyle: React.CSSProperties = fill ? {
        position: "absolute",
        height: "100%",
        width: "100%",
        left: 0,
        top: 0,
        right: 0,
        bottom: 0,
        color: "transparent",
    } : {}

    const placeholderStyle: React.CSSProperties = (blurUrl && !isLoaded) ? {
        backgroundImage: `url("${blurUrl}")`,
        backgroundSize: objectFit === "contain" ? "contain" : "cover",
        backgroundPosition: "center",
        backgroundRepeat: "no-repeat",
    } : {}

    const imageWidth = fill ? undefined : (width || (isStaticImport ? src.width : undefined))
    const imageHeight = fill ? undefined : (height || (isStaticImport ? src.height : undefined))

    return (
        <img
            ref={ref}
            src={imageSrc}
            alt={alt}
            width={imageWidth}
            height={imageHeight}
            decoding="async"
            loading={priority ? "eager" : "lazy"}
            className={className}
            style={{
                ...fillStyle,
                ...placeholderStyle,
                ...(objectFit ? { objectFit: objectFit as any } : {}),
                ...style,
            }}
            onLoad={(e) => {
                setIsLoaded(true)
                onLoad?.(e)
            }}
            {...props}
        />
    )
})

Image.displayName = "Image"
