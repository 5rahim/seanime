import { Nullish } from "@/api/generated/types"
import { IconButton } from "@/components/ui/button"
import { Modal, ModalProps } from "@/components/ui/modal"
import { Popover, PopoverProps } from "@/components/ui/popover"
import { Separator } from "@/components/ui/separator"
import Image from "next/image"
import React from "react"
import { AiFillWarning } from "react-icons/ai"
import { MdInfo } from "react-icons/md"
import { useWindowSize } from "react-use"

type MediaEpisodeInfoModalProps = {
    title?: Nullish<string>
    image?: Nullish<string>
    episodeTitle?: Nullish<string>
    airDate?: Nullish<string>
    length?: Nullish<number | string>
    summary?: Nullish<string>
    isInvalid?: Nullish<boolean>
    filename?: Nullish<string>
}

function IsomorphicPopover(props: PopoverProps & ModalProps) {
    const { title, children, ...rest } = props
    const { width } = useWindowSize()

    if (width && width > 1024) {
        return <Popover
            {...rest}
            className="max-w-xl !w-full overflow-hidden"
        >
            {children}
        </Popover>
    }

    return <Modal
        {...rest}
        title={title}
        titleClass="text-xl"
        contentClass="max-w-2xl overflow-hidden"
    >
        {children}
    </Modal>
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
        filename,
        ...rest
    } = props

    return (
        <>
            <IsomorphicPopover
                data-media-episode-info-modal
                trigger={<IconButton
                    icon={<MdInfo />}
                    className="opacity-30 hover:opacity-100 transform-opacity"
                    intent="gray-basic"
                    size="xs"
                />}
                title={title || "Episode"}
                // contentClass="max-w-2xl"
                // titleClass="text-xl"
            >

                {image && <div
                    data-media-episode-info-modal-image-container
                    className="h-[8rem] rounded-t-md w-full flex-none object-cover object-center overflow-hidden absolute left-0 top-0 z-[0]"
                >
                    <Image
                        data-media-episode-info-modal-image
                        src={image}
                        alt="banner"
                        fill
                        quality={80}
                        priority
                        sizes="20rem"
                        className="object-cover object-center opacity-30"
                    />
                    <div
                        data-media-episode-info-modal-image-gradient
                        className="z-[5] absolute bottom-0 w-full h-[80%] bg-gradient-to-t from-[--background] to-transparent"
                    />
                </div>}

                <div data-media-episode-info-modal-content className="space-y-4 z-[5] relative">
                    <p data-media-episode-info-modal-content-title className="text-lg line-clamp-2 font-semibold">
                        {episodeTitle?.replaceAll("`", "'")}
                        {isInvalid && <AiFillWarning />}
                    </p>
                    {!(!airDate && !length) && <p className="text-[--muted]">
                        {airDate || "Unknown airing date"} - {length || "N/A"} minutes
                    </p>}
                    <p className="text-gray-300">
                        {summary?.replaceAll("`", "'") || "No summary"}
                    </p>

                    {filename && <>
                        <Separator />
                        <p className="text-[--muted] line-clamp-2">
                            {filename}
                        </p>
                    </>}
                </div>

            </IsomorphicPopover>
        </>
    )
}
