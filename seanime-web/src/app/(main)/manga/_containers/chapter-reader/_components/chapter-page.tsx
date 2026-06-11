import { HibikeManga_ChapterPage, Manga_PageContainer } from "@/api/generated/types"
import { useMangaReaderUtils } from "@/app/(main)/manga/_lib/handle-manga-utils"
import { IconButton } from "@/components/ui/button"
import { LoadingSpinner } from "@/components/ui/loading-spinner"
import { logger } from "@/lib/helpers/debug"
import { HIDE_IMAGES } from "@/types/constants.ts"
import React from "react"
import { FaRedo } from "react-icons/fa"
import { useUpdateEffect } from "react-use"

type ChapterPageProps = {
    children?: React.ReactNode
    index: number
    pageContainer: Manga_PageContainer | undefined
    page: HibikeManga_ChapterPage | undefined
    containerClass: string
    imageClass: string
    readingMode: string
    onFinishedLoading?: () => void
    imageWidth?: number | string
    imageMaxWidth?: number | string
    containerMaxWidth?: number | string
    pageZoom?: number
    pageFit?: string
}

export function ChapterPage(props: ChapterPageProps) {

    const {
        index,
        pageContainer,
        page,
        containerClass,
        imageClass,
        children,
        readingMode,
        onFinishedLoading,
        imageWidth,
        imageMaxWidth,
        containerMaxWidth,
        pageZoom = 1,
        pageFit,
        ...rest
    } = props

    const ref = React.useRef<HTMLImageElement>(null)
    const { getChapterPageUrl, isReady } = useMangaReaderUtils()
    const pageUrl = React.useMemo(() => {
        if (!page || !isReady) return undefined
        return HIDE_IMAGES ? "/no-cover.png" : getChapterPageUrl(page.url, pageContainer?.isDownloaded, page.headers)
    }, [getChapterPageUrl, isReady, page, pageContainer?.isDownloaded])

    const { isLoaded, isLoading, hasError, retry } = useImageLoadStatus(ref, isReady, pageUrl)

    useUpdateEffect(() => {
        if (isLoaded && onFinishedLoading) {
            onFinishedLoading()
        }
    }, [isLoaded])

    if (!page) return null

    const containerStyle: React.CSSProperties & { zoom?: number } = {
        maxWidth: pageZoom !== 1 ? "none" : containerMaxWidth,
        minHeight: isLoaded ? "20px" : undefined,
        transformOrigin: "top center",
    }
    if (pageZoom !== 1) {
        containerStyle.zoom = pageZoom
    }

    return (
        <>
            <div
                data-chapter-page-container
                className={containerClass}
                style={containerStyle}
                id={`page-${index}`}
                tabIndex={-1}
            >
                {(isLoading || !isReady) &&
                    <LoadingSpinner
                        data-chapter-page-loading-spinner
                        containerClass="h-full absolute inset-0 z-[1] w-full mx-auto"
                        style={{ zoom: pageZoom !== 1 ? 1 / pageZoom : undefined }}
                        tabIndex={-1}
                    />}
                {hasError &&
                    <div
                        data-chapter-page-retry-container
                        className="h-full w-full flex justify-center items-center absolute inset-0 z-[10]"
                        style={{ zoom: pageZoom !== 1 ? 1 / pageZoom : undefined }}
                        id="retry-container"
                        tabIndex={-1}
                    >
                        <IconButton intent="white" icon={<FaRedo id="retry-icon" />} onClick={retry} id="retry-button" tabIndex={-1} />
                    </div>}
                {isReady && <img
                    data-chapter-page-image
                    data-page-index={index}
                    src={pageUrl}
                    alt={`Page ${index}`}
                    crossOrigin="anonymous"
                    draggable={false}
                    className={imageClass}
                    style={{
                        width: pageZoom !== 1 ? (pageFit === "contain" || pageFit === "true-size" ? "auto" : "100%") : imageWidth,
                        height: pageZoom !== 1 ? (pageFit === "contain" ? "100%" : "auto") : undefined,
                        maxWidth: pageZoom !== 1 ? "none" : imageMaxWidth,
                        maxHeight: pageZoom !== 1 ? "none" : undefined,
                        objectFit: pageZoom !== 1 ? "initial" : undefined,
                        display: pageZoom !== 1 ? "block" : undefined,
                        marginLeft: pageZoom !== 1 ? "auto" : undefined,
                        marginRight: pageZoom !== 1 ? "auto" : undefined,
                    }}
                    ref={ref}
                    tabIndex={-1}
                />}
            </div>
        </>
    )
}

export const IMAGE_STATUS = {
    LOADING: "loading",
    RETRYING: "retrying",
    LOADED: "loaded",
    ERROR: "error",
}

const useImageLoadStatus = (
    imageRef: React.RefObject<HTMLImageElement | null>,
    enabled: boolean,
    src?: string,
) => {
    const [imageStatus, setImageStatus] = React.useState(IMAGE_STATUS.LOADING)
    const retries = React.useRef(0)

    const isRetrying = imageStatus === IMAGE_STATUS.RETRYING
    const isLoaded = imageStatus === IMAGE_STATUS.LOADED
    const isLoading =
        imageStatus === IMAGE_STATUS.LOADING ||
        imageStatus === IMAGE_STATUS.RETRYING
    const hasError = imageStatus === IMAGE_STATUS.ERROR

    const retry = React.useCallback(() => {
        retries.current = 0
        setImageStatus(IMAGE_STATUS.LOADING)
        const image = imageRef.current
        const imgSrc = image?.src
        if (!image || !imgSrc) {
            return
        }
        image.src = imgSrc
    }, [])

    React.useEffect(() => {
        if (!enabled || !src) {
            setImageStatus(IMAGE_STATUS.LOADING)
            return
        }

        retries.current = 0
        setImageStatus(IMAGE_STATUS.LOADING)

        if (!imageRef.current) {
            return
        }

        // Keep a stable reference to the image
        const image = imageRef.current
        if (!image) {
            return
        }
        let timerIds: any[] = []

        if (
            image &&
            image.complete &&
            image.naturalWidth > 0 &&
            timerIds.length === 0
        ) {
            setImageStatus(IMAGE_STATUS.LOADED)
            return
        }

        /**
         * if an image errors retry 3 times
         * @param {*} event
         */
        const handleError = (event: ErrorEvent) => {
            logger("chapter-page").info("retrying")
            if (retries.current >= 3) {
                logger("chapter-page").info("max retries reached", event.error)
                setImageStatus(IMAGE_STATUS.ERROR)
                return
            }

            setImageStatus(IMAGE_STATUS.RETRYING)

            retries.current = retries.current + 1

            const timerId = setTimeout(() => {
                const img = event.target as HTMLImageElement
                if (!img) {
                    return
                }
                const imgSrc = img.src

                img.src = imgSrc

                // Already removes itself from the list of timerIds
                timerIds.splice(timerIds.indexOf(timerId), 1)
            }, 1000)
            timerIds.push(timerId)
        }
        const handleLoad = () => {
            setImageStatus(IMAGE_STATUS.LOADED)
        }

        image.addEventListener("error", handleError)
        image.addEventListener("load", handleLoad, { once: true })

        return () => {
            image.removeEventListener("error", handleError)
            image.removeEventListener("load", handleLoad)
            // Cleanup pending setTimeout's. We use `splice(0)` to clear the list.
            for (const timerId of timerIds.splice(0)) {
                clearTimeout(timerId)
            }
        }
    }, [enabled, imageRef, src])

    return {
        isLoaded,
        isLoading,
        isRetrying,
        hasError,
        retry,
    }
}
