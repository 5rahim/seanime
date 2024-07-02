import { Nullish } from "@/api/generated/types"
import { IconButton } from "@/components/ui/button"
import { Modal } from "@/components/ui/modal"
import Image from "next/image"
import React from "react"
import { AiFillWarning } from "react-icons/ai"
import { MdInfo } from "react-icons/md"

type MediaEpisodeInfoModalProps = {
    title?: Nullish<string>
    image?: Nullish<string>
    episodeTitle?: Nullish<string>
    airDate?: Nullish<string>
    length?: Nullish<number | string>
    summary?: Nullish<string>
    isInvalid?: Nullish<boolean>
}

export function MediaEpisodeInfoModal(props: MediaEpisodeInfoModalProps) {

    const {
        title,
        image,
        episodeTitle,
        airDate,
        length,
        summary,
        isInvalid,
        ...rest
    } = props

    return (
        <>
            <Modal
                trigger={<IconButton
                    icon={<MdInfo />}
                    className="opacity-30 hover:opacity-100 transform-opacity"
                    intent="gray-basic"
                    size="xs"
                />}
                title={title}
                contentClass="max-w-2xl"
                titleClass="text-xl"
            >

                {image && <div
                    className="h-[8rem] rounded-t-md w-full flex-none object-cover object-center overflow-hidden absolute left-0 top-0 z-[-1]"
                >
                    <Image
                        src={image}
                        alt="banner"
                        fill
                        quality={80}
                        priority
                        sizes="20rem"
                        className="object-cover object-center opacity-30"
                    />
                    <div
                        className="z-[5] absolute bottom-0 w-full h-[80%] bg-gradient-to-t from-[--background] to-transparent"
                    />
                </div>}

                <div className="space-y-4">
                    <p className="text-lg line-clamp-2 font-semibold">
                        {episodeTitle?.replaceAll("`", "'")}
                        {isInvalid && <AiFillWarning />}
                    </p>
                    {!(!airDate && !length) && <p className="text-[--muted]">
                        {airDate || "Unknown airing date"} - {length || "N/A"} minutes
                    </p>}
                    <p className="text-gray-300">
                        {summary?.replaceAll("`", "'") || "No summary"}
                    </p>
                </div>

            </Modal>
        </>
    )
}
