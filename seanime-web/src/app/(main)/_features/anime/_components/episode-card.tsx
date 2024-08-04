import { EpisodeItemBottomGradient } from "@/app/(main)/_features/custom-ui/item-bottom-gradients"
import { imageShimmer } from "@/components/shared/image-helpers"
import { cn } from "@/components/ui/core/styling"
import Image from "next/image"
import React from "react"
import { AiFillPlayCircle } from "react-icons/ai"

type EpisodeCardProps = {
    title: React.ReactNode
    actionIcon?: React.ReactElement | null
    image?: string
    onClick?: () => void
    topTitle?: string
    meta?: string
    type?: "carousel" | "grid"
    isInvalid?: boolean
    containerClass?: string
    episodeNumber?: number
    progressNumber?: number
    progressTotal?: number
    mRef?: React.RefObject<HTMLDivElement>
    hasDiscrepancy?: boolean
    length?: string | number | null
    imageClass?: string
} & Omit<React.ComponentPropsWithoutRef<"div">, "title">

export function EpisodeCard(props: EpisodeCardProps) {

    const {
        children,
        actionIcon = props.actionIcon !== null ? <AiFillPlayCircle className="opacity-50" /> : undefined,
        image,
        onClick,
        topTitle,
        meta,
        title,
        type = "carousel",
        isInvalid,
        className,
        containerClass,
        mRef,
        episodeNumber,
        progressTotal,
        progressNumber,
        hasDiscrepancy,
        length,
        imageClass,
        ...rest
    } = props

    const showTotalEpisodes = React.useMemo(() => !!progressTotal && progressTotal > 1, [progressTotal])
    const offset = React.useMemo(() => hasDiscrepancy ? 1 : 0, [hasDiscrepancy])

    return (
        <div
            ref={mRef}
            className={cn(
                "rounded-lg overflow-hidden space-y-2 flex-none group/episode-card cursor-pointer",
                "select-none",
                type === "carousel" && "w-full",
                type === "grid" && "w-72 lg:w-[26rem]",
                className,
                containerClass,
            )}
            onClick={onClick}
            {...rest}
        >
            <div className="w-full h-full rounded-lg overflow-hidden z-[1] aspect-[4/2] relative">
                {!!image ? <Image
                    src={image}
                    alt={""}
                    fill
                    quality={100}
                    placeholder={imageShimmer(700, 475)}
                    sizes="20rem"
                    className={cn(
                        "object-cover rounded-lg object-center transition",
                        imageClass,
                    )}
                /> : <div
                    className="h-full block rounded-lg absolute w-full bg-gradient-to-t from-gray-800 to-transparent z-[2]"
                ></div>}
                {/*[CUSTOM UI] BOTTOM GRADIENT*/}
                <EpisodeItemBottomGradient />
                <div
                    className={cn(
                        "group-hover/episode-card:opacity-100 text-6xl text-gray-200",
                        "cursor-pointer opacity-0 transition-opacity bg-gray-950 bg-opacity-60 z-[2] absolute w-[105%] h-[105%] items-center justify-center",
                        "hidden md:flex",
                    )}
                >
                    {actionIcon && actionIcon}
                </div>

                {isInvalid && <p className="text-red-300 opacity-50 absolute left-2 bottom-2 z-[2]">No metadata found</p>}
            </div>
            <div className="relative z-[3] w-full space-y-0">
                <p className="w-[80%] line-clamp-1 text-lg transition-colors duration-200 text-[--foreground] font-semibold">{topTitle?.replaceAll(
                    "`",
                    "'")}</p>
                <div className="w-full justify-between flex flex-none items-center">
                    <p className="text-base md:text-xl font-medium line-clamp-1">
                        <span>{title}{showTotalEpisodes ?
                            <span className="opacity-40">{` / `}{progressTotal! - offset}</span>
                            : ``}</span>
                    </p>
                    {(!!meta || !!length) && <p className="text-[--muted] flex-none ml-2 text-sm md:text-base line-clamp-2 text-right">
                        {meta}{!!meta && !!length && `  • `}{length ? `${length}m` : ""}
                    </p>}
                </div>
            </div>
        </div>
    )

}


// export function EpisodeCard(props: EpisodeCardProps) {
//
//     const {
//         children,
//         actionIcon = props.actionIcon !== null ? <AiFillPlayCircle className="opacity-50" /> : undefined,
//         image,
//         onClick,
//         topTitle,
//         meta,
//         title,
//         type = "carousel",
//         isInvalid,
//         className,
//         containerClass,
//         mRef,
//         episodeNumber,
//         progressTotal,
//         progressNumber,
//         hasDiscrepancy,
//         length,
//         imageClass,
//         ...rest
//     } = props
//
//     const showTotalEpisodes = React.useMemo(() => !!progressTotal && progressTotal > 1, [progressTotal])
//     const offset = React.useMemo(() => hasDiscrepancy ? 1 : 0, [hasDiscrepancy])
//
//     return (
//         <div
//             ref={mRef}
//             className={cn(
//                 "rounded-lg overflow-hidden aspect-[4/2] relative flex items-end flex-none group/episode-card cursor-pointer",
//                 "select-none",
//                 type === "carousel" && "w-full",
//                 type === "grid" && "w-72 lg:w-[26rem]",
//                 className,
//                 containerClass,
//             )}
//             onClick={onClick}
//             {...rest}
//         >
//             <div className="absolute w-full h-full rounded-lg overflow-hidden z-[1]">
//                 {!!image ? <Image
//                     src={image}
//                     alt={""}
//                     fill
//                     quality={100}
//                     placeholder={imageShimmer(700, 475)}
//                     sizes="20rem"
//                     className={cn(
//                         "object-cover rounded-lg object-center transition",
//                         imageClass,
//                     )}
//                 /> : <div
//                     className="h-full block rounded-lg absolute w-full bg-gradient-to-t from-gray-800 to-transparent z-[2]"
//                 ></div>}
//                 {/*[CUSTOM UI] BOTTOM GRADIENT*/}
//                 <EpisodeItemBottomGradient />
//             </div>
//             <div
//                 className={cn(
//                     "group-hover/episode-card:opacity-100 text-6xl text-gray-200",
//                     "cursor-pointer opacity-0 transition-opacity bg-gray-950 bg-opacity-60 z-[2] absolute w-[105%] h-[105%] items-center
// justify-center", "hidden md:flex", )} > {actionIcon && actionIcon} </div> <div className="relative z-[3] w-full p-4 space-y-0"> <p
// className="w-[80%] line-clamp-1 text-[--muted] transition-colors duration-200 group-hover/episode-card:text-[--foreground]
// font-semibold">{topTitle?.replaceAll( "`", "'")}</p> <div className="w-full justify-between flex flex-none items-center"> <p className="text-base
// md:text-xl font-semibold line-clamp-1"> <span>{title}{showTotalEpisodes ? <span className="opacity-40">{` / `}{progressTotal! - offset}</span> :
// ``}</span> </p> {(!!meta || !!length) && <p className="text-[--muted] flex-none ml-2 text-sm md:text-base line-clamp-2 text-right"> {meta}{!!meta
// && !!length && `  • `}{length ? `${length}m` : ""} </p>} </div> {isInvalid && <p className="text-red-300">No metadata found</p>} </div> </div> )
// }
