import { getServerBaseUrl } from "@/api/client/server-url"
import { AL_BaseAnime, Anime_Episode, Anime_LocalFileType } from "@/api/generated/types"
import { useUpdateLocalFileData } from "@/api/hooks/localfiles.hooks"
import { useExternalPlayerLink } from "@/app/(main)/_atoms/playback.atoms"
import { EpisodeGridItem } from "@/app/(main)/_features/anime/_components/episode-grid-item"
import { IconButton } from "@/components/ui/button"
import { DropdownMenu, DropdownMenuItem, DropdownMenuSeparator } from "@/components/ui/dropdown-menu"
import { defineSchema, Field, Form } from "@/components/ui/form"
import { Modal } from "@/components/ui/modal"
import { Separator } from "@/components/ui/separator"
import { getImageUrl } from "@/lib/server/assets"
import { atom } from "jotai"
import { createIsolation } from "jotai-scope"
import Image from "next/image"
import React, { memo } from "react"
import { AiFillWarning } from "react-icons/ai"
import { BiDotsHorizontal, BiLockOpenAlt } from "react-icons/bi"
import { MdInfo, MdOutlineOndemandVideo, MdOutlineRemoveDone } from "react-icons/md"
import { RiEdit2Line } from "react-icons/ri"
import { VscVerified } from "react-icons/vsc"
import { useCopyToClipboard } from "react-use"
import { toast } from "sonner"

export const EpisodeItemIsolation = createIsolation()

const __metadataModalIsOpenAtom = atom(false)

export const EpisodeItem = memo(({ episode, media, isWatched, onPlay, percentageComplete, minutesRemaining }: {
    episode: Anime_Episode,
    media: AL_BaseAnime,
    onPlay?: ({ path, mediaId }: { path: string, mediaId: number }) => void,
    isWatched?: boolean
    percentageComplete?: number
    minutesRemaining?: number
    isOffline?: boolean
}) => {

    const { updateLocalFile, isPending } = useUpdateLocalFileData(media.id)
    const [_, copyToClipboard] = useCopyToClipboard()

    const { encodePath } = useExternalPlayerLink()

    function encodeFilePath(filePath: string) {
        if (encodePath) {
            return Buffer.from(filePath).toString("base64")
        }
        return encodeURIComponent(filePath)
    }

    return (
        <EpisodeItemIsolation.Provider>
            <EpisodeGridItem
                media={media}
                image={episode.episodeMetadata?.image}
                onClick={() => onPlay?.({ path: episode.localFile?.path ?? "", mediaId: media.id })}
                isInvalid={episode.isInvalid}
                title={episode.displayTitle}
                episodeTitle={episode.episodeTitle}
                fileName={episode.localFile?.name}
                isWatched={episode.progressNumber > 0 && isWatched}
                isFiller={episode.episodeMetadata?.isFiller}
                length={episode.episodeMetadata?.length}
                percentageComplete={percentageComplete}
                minutesRemaining={minutesRemaining}
                episodeNumber={episode.episodeNumber}
                progressNumber={episode.progressNumber}
                description={episode.episodeMetadata?.summary || episode.episodeMetadata?.overview}
                action={<>
                    <IconButton
                        icon={episode.localFile?.locked ? <VscVerified /> : <BiLockOpenAlt />}
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

                    <DropdownMenu
                        trigger={
                            <IconButton
                                icon={<BiDotsHorizontal />}
                                intent="gray-basic"
                                size="xs"
                            />
                        }
                    >
                        <MetadataModalButton />
                        {episode.localFile && <DropdownMenuItem
                            onClick={() => {
                                copyToClipboard(getServerBaseUrl() + "/api/v1/mediastream/file/" + encodeFilePath(episode.localFile!.path))
                                toast.info("Stream URL copied")
                            }}
                        >
                            <MdOutlineOndemandVideo />
                            Copy stream URL
                        </DropdownMenuItem>}
                        <DropdownMenuSeparator />
                        <DropdownMenuItem
                            className="text-[--orange]"
                            onClick={() => {
                                if (episode.localFile) {
                                    updateLocalFile(episode.localFile, {
                                        mediaId: 0,
                                        locked: false,
                                        ignored: false,
                                    })
                                }
                            }}
                        >
                            <MdOutlineRemoveDone /> Unmatch
                        </DropdownMenuItem>
                    </DropdownMenu>

                    {(!!episode.episodeMetadata && (episode.type === "main" || episode.type === "special")) && !!episode.episodeMetadata?.anidbId &&
                        <EpisodeItemInfoModalButton episode={episode} />}
                </>}
            />
            <MetadataModal
                episode={episode}
            />
        </EpisodeItemIsolation.Provider>
    )

})


const metadataSchema = defineSchema(({ z }) => z.object({
    episode: z.number().min(0),
    aniDBEpisode: z.string().transform(value => value.toUpperCase()),
    type: z.string().min(0),
}))

function MetadataModal({ episode }: { episode: Anime_Episode }) {

    const [isOpen, setIsOpen] = EpisodeItemIsolation.useAtom(__metadataModalIsOpenAtom)

    const { updateLocalFile, isPending } = useUpdateLocalFileData(episode.baseAnime?.id)

    return (
        <Modal
            open={isOpen}
            onOpenChange={() => setIsOpen(false)}

            title={episode.displayTitle}
            titleClass="text-center"
            contentClass="max-w-xl"
        >
            <p className="w-full line-clamp-2 text-sm px-4 text-center py-2 flex-none">{episode.localFile?.name}</p>
            <Form
                schema={metadataSchema}
                onSubmit={(data) => {
                    if (episode.localFile) {
                        updateLocalFile(episode.localFile, {
                            metadata: {
                                ...episode.localFile?.metadata,
                                type: data.type as Anime_LocalFileType,
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
                    help="Relative episode number. If movie, episode number = 1"
                    required
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
                    <Field.Submit role="save" intent="success" loading={isPending}>Save</Field.Submit>
                </div>
            </Form>
        </Modal>
    )
}

function MetadataModalButton() {
    const [, setIsOpen] = EpisodeItemIsolation.useAtom(__metadataModalIsOpenAtom)
    return <DropdownMenuItem onClick={() => setIsOpen(true)}>
        <RiEdit2Line />
        Update metadata
    </DropdownMenuItem>
}

export function EpisodeItemInfoModalButton({ episode }: { episode: Anime_Episode }) {
    return <Modal
        title={episode.displayTitle}
        contentClass="max-w-2xl overflow-hidden"
        titleClass="text-xl"
        trigger={<IconButton
            icon={<MdInfo />}
            className="opacity-30 hover:opacity-100 transform-opacity"
            intent="gray-basic"
            size="xs"
        />}
    >

        {episode.episodeMetadata?.image && <div
            className="h-[8rem] w-full flex-none object-cover object-center overflow-hidden absolute left-0 top-0 z-[-1]"
        >
            <Image
                src={getImageUrl(episode.episodeMetadata?.image)}
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
                {episode.episodeTitle?.replaceAll("`", "'")}
                {episode.isInvalid && <AiFillWarning />}
            </p>
            <p className="text-[--muted]">
                {episode.episodeMetadata?.airDate || "Unknown airing date"} - {episode.episodeMetadata?.length || "N/A"} minutes
            </p>
            <p className="text-gray-300">
                {(episode.episodeMetadata?.summary || episode.episodeMetadata?.overview)?.replaceAll("`", "'")?.replace(/source:.*/gi, "") || "No summary"}
            </p>
            <Separator />
            <p className="text-[--muted] line-clamp-2">
                {episode.localFile?.parsedInfo?.original}
            </p>
            {
                (!!episode.episodeMetadata?.anidbId) && <>
                    <div className="w-full flex gap-2">
                        <p>AniDB Episode: {episode.fileMetadata?.aniDBEpisode}</p>
                        <a
                            href={"https://anidb.net/episode/" + episode.episodeMetadata?.anidbId + "#layout-footer"}
                            target="_blank"
                            className="hover:underline text-[--muted]"
                        >Open on AniDB
                        </a>
                    </div>
                </>
            }

        </div>

    </Modal>
}
