import { imageShimmer } from "@/components/shared/styling/image-helpers"
import { cn } from "@/components/ui/core/styling"
import Image from "next/image"
import React from "react"
import { AiFillPlayCircle } from "react-icons/ai"

type LargeEpisodeListItemProps = {
    title: React.ReactNode
    actionIcon?: React.ReactElement | null
    image?: string | null
    onClick?: () => void
    topTitle?: string | null
    meta?: string | null
    larger?: boolean
    isInvalid?: boolean
}

export const LargeEpisodeListItem: React.FC<LargeEpisodeListItemProps & Omit<React.ComponentPropsWithoutRef<"div">, "title">> = (props) => {

    const {
        children,
        actionIcon = props.actionIcon !== null ? <AiFillPlayCircle className="opacity-40" /> : undefined,
        image,
        onClick,
        topTitle,
        meta,
        title,
        larger = false,
        isInvalid,
        ...rest
    } = props

    return <>
        <div
            className={cn(
                "rounded-md border overflow-hidden aspect-[4/2] w-96 relative flex items-end flex-none group/missed-episode-item cursor-pointer",
                "w-72 lg:w-96",
                "user-select-none",
                {
                    "w-72 lg:w-[30rem]": larger,
                },
            )}
            onClick={onClick}
            {...rest}
        >
            <div className="absolute w-full h-full overflow-hidden z-[1]">
                {!!image ? <Image
                    src={image}
                    alt={""}
                    fill
                    quality={100}
                    placeholder={imageShimmer(700, 475)}
                    sizes="20rem"
                    className="object-cover object-center transition group-hover/missed-episode-item:scale-110"
                /> : <div
                    className="h-full block absolute w-full bg-gradient-to-t from-gray-800 to-transparent z-[2]"
                ></div>}
                <div
                    className="z-[1] absolute bottom-0 w-full h-full md:h-[80%] bg-gradient-to-t from-[--background] to-transparent"
                />
            </div>
            <div className={cn(
                "group-hover/missed-episode-item:opacity-100 text-6xl text-gray-200",
                "cursor-pointer opacity-0 transition-opacity bg-gray-900 bg-opacity-50 backdrop-blur-md z-[2] absolute w-full h-full items-center justify-center",
                "hidden md:flex",
            )}>
                {actionIcon && actionIcon}
            </div>
            <div className="relative z-[3] w-full p-4 space-y-1">
                {topTitle && <p className="w-[80%] line-clamp-1 text-[--muted] font-semibold">{topTitle}</p>}
                <div className="w-full justify-between flex items-center">
                    <p className="text-base md:text-xl font-semibold line-clamp-2">{title}</p>
                    {(!!meta) &&
                        <p className="text-[--muted] text-sm md:text-base">{meta}</p>}
                </div>
                {isInvalid && <p className="text-red-300">No metadata found</p>}
            </div>
        </div>
    </>

}
