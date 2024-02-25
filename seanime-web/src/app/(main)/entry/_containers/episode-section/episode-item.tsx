import { useUpdateLocalFile } from "@/app/(main)/entry/_lib/update-local-file"
import { EpisodeListItem } from "@/components/shared/episode-list-item"
import { IconButton } from "@/components/ui/button"
import { DropdownMenu } from "@/components/ui/dropdown-menu"
import { defineSchema, Field, Form } from "@/components/ui/form"
import { Modal } from "@/components/ui/modal"
import { Separator } from "@/components/ui/separator"
import { BaseMediaFragment } from "@/lib/anilist/gql/graphql"
import { LocalFileType, MediaEntryEpisode } from "@/lib/server/types"
import { atom } from "jotai"
import { createIsolation } from "jotai-scope"
import Image from "next/image"
import React, { memo } from "react"
import { AiFillWarning } from "react-icons/ai"
import { BiDotsHorizontal, BiLockOpenAlt } from "react-icons/bi"
import { MdInfo } from "react-icons/md"
import { VscVerified } from "react-icons/vsc"
import { toast } from "sonner"

export const EpisodeItemIsolation = createIsolation()

const __metadataModalIsOpenAtom = atom(false)
export const __episodeItem_infoModalIsOpenAtom = atom(false)


export const EpisodeItem = memo(({ episode, media, isWatched, onPlay }: {
    episode: MediaEntryEpisode,
    media: BaseMediaFragment,
    onPlay: ({ path }: { path: string }) => void,
    isWatched?: boolean
}) => {

    const { updateLocalFile, isPending } = useUpdateLocalFile(media.id)

    return (
        <EpisodeItemIsolation.Provider>
            <EpisodeListItem
                media={media}
                image={episode.episodeMetadata?.image}
                onClick={() => onPlay({ path: episode.localFile?.path ?? "" })}
                isInvalid={episode.isInvalid}
                title={episode.displayTitle}
                episodeTitle={episode.episodeTitle}
                fileName={episode.localFile?.name}
                isWatched={episode.progressNumber > 0 && isWatched}
                action={<>
                    <IconButton
                        icon={episode.localFile?.locked ? <VscVerified/> : <BiLockOpenAlt/>}
                        intent={episode.localFile?.locked ? "success-basic" : "warning-basic"}
                        size="md"
                        className="hover:opacity-60"
                        loading={isPending}
                        onClick={() => {
                            if (episode.localFile) {
                                updateLocalFile(episode.localFile, {
                                    locked: !episode.localFile?.locked,
                                })
                            }
                        }}
                    />

                    <DropdownMenu trigger={
                        <IconButton
                            icon={<BiDotsHorizontal/>}
                            intent="gray-basic"
                            size="xs"
                        />
                    }>
                        <MetadataModalButton/>
                        <DropdownMenu.Separator />
                        <DropdownMenu.Item
                            className="!text-red-300 !dark:text-red-200"
                            onClick={() => {
                                if (episode.localFile) {
                                    updateLocalFile(episode.localFile, {
                                        mediaId: 0,
                                    })
                                }
                            }}
                        >Unmatch</DropdownMenu.Item>
                    </DropdownMenu>

                    {(!!episode.episodeMetadata && (episode.type === "main" || episode.type === "special")) && !!episode.episodeMetadata?.aniDBId &&
                        <EpisodeItemInfoModalButton/>}
                </>}
            />
            <MetadataModal
                episode={episode}
            />
            {episode.episodeMetadata?.aniDBId && <EpisodeItemInfoModal
                episode={episode}
            />}
        </EpisodeItemIsolation.Provider>
    )

})


const metadataSchema = defineSchema(({ z }) => z.object({
    episode: z.number().min(0),
    aniDBEpisode: z.string().transform(value => value.toUpperCase()),
    type: z.string().min(0),
}))

export function MetadataModal({ episode }: { episode: MediaEntryEpisode }) {

    const [isOpen, setIsOpen] = EpisodeItemIsolation.useAtom(__metadataModalIsOpenAtom)

    const { updateLocalFile, isPending } = useUpdateLocalFile(episode.basicMedia?.id)

    return (
        <Modal
            open={isOpen}
            onOpenChange={() => setIsOpen(false)}

            title={episode.displayTitle}
            titleClass="text-center"
            size="lg"
        >
            <p className="w-full line-clamp-2 text-sm text-[--muted] px-4 text-center py-2 flex-none">{episode.localFile?.name}</p>
            <Form
                schema={metadataSchema}
                onSubmit={(data) => {
                    if (episode.localFile) {
                        updateLocalFile(episode.localFile, {
                            metadata: {
                                ...episode.localFile?.metadata,
                                type: data.type as LocalFileType,
                                episode: data.episode,
                                aniDBEpisode: data.aniDBEpisode,
                            },
                        }, () => {
                            setIsOpen(false)
                            toast.success("Metadata saved")
                        })
                    }
                }}
                onError={console.log}
                //@ts-ignore
                defaultValues={{ ...episode.fileMetadata }}
            >
                <Field.Number
                    label="Episode number" name="episode"
                    help="Relative episode number. If movie, episode number = 1" discrete isRequired
                />
                <Field.Text
                    label="AniDB episode"
                    name="aniDBEpisode"
                    help="Specials typically contain the letter S"
                />
                <Field.Select
                    label="Type"
                    name="type"
                    options={[
                        { label: "Main", value: "main" },
                        { label: "Special", value: "special" },
                        { label: "NC/Other", value: "nc" },
                    ]}
                />
                <div className="w-full flex justify-end">
                    <Field.Submit role="save" intent="success" loading={isPending} />
                </div>
            </Form>
        </Modal>
    )
}

function MetadataModalButton() {
    const [, setIsOpen] = EpisodeItemIsolation.useAtom(__metadataModalIsOpenAtom)
    return <DropdownMenu.Item onClick={() => setIsOpen(true)}>Update metadata</DropdownMenu.Item>
}

function EpisodeItemInfoModalButton() {
    const [, setIsOpen] = EpisodeItemIsolation.useAtom(__episodeItem_infoModalIsOpenAtom)
    return <IconButton
        icon={<MdInfo />}
        className="opacity-30 hover:opacity-100 transform-opacity"
        intent="gray-basic"
        size="xs"
        onClick={() => setIsOpen(true)}
    />
}

export function EpisodeItemInfoModal(props: { episode: MediaEntryEpisode, }) {

    const {
        episode,
    } = props

    const [isOpen, setIsOpen] = EpisodeItemIsolation.useAtom(__episodeItem_infoModalIsOpenAtom)

    return (
        <>
            <Modal
                open={isOpen}
                onOpenChange={() => setIsOpen(false)}
                title={episode.displayTitle}

                size="xl"
                titleClass="text-xl"
            >

                {episode.episodeMetadata?.image && <div
                    className="h-[8rem] w-full flex-none object-cover object-center overflow-hidden absolute left-0 top-0 z-[-1]">
                    <Image
                        src={episode.episodeMetadata?.image}
                        alt="banner"
                        fill
                        quality={80}
                        priority
                        sizes="20rem"
                        className="object-cover object-center opacity-30"
                    />
                    <div
                        className={"z-[5] absolute bottom-0 w-full h-[80%] bg-gradient-to-t from-gray-900 to-transparent"}
                    />
                </div>}

                <div className="space-y-4">
                    <p className="text-lg line-clamp-2 font-semibold">
                        {episode.episodeTitle}
                        {episode.isInvalid && <AiFillWarning/>}
                    </p>
                    <p className="text-[--muted]">
                        {episode.episodeMetadata?.airDate || "Unknown airing date"} - {episode.episodeMetadata?.length || "N/A"} minutes
                    </p>
                    <p className="text-[--muted]">
                        {episode.episodeMetadata?.summary || "No summary"}
                    </p>
                    {
                        (!!episode.episodeMetadata?.aniDBId) && <>
                            <Separator />
                            <div className="w-full flex justify-between">
                                <p>AniDB Episode: {episode.fileMetadata?.aniDBEpisode}</p>
                                <a href={"https://anidb.net/episode/" + episode.episodeMetadata?.aniDBId + "#layout-footer"}
                                   target="_blank"
                                   className="text-brand-200"
                                >Open on AniDB
                                </a>
                            </div>
                        </>
                    }
                </div>

            </Modal>
        </>
    )

}
