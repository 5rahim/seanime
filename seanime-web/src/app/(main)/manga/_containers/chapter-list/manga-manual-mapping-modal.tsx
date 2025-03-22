import { Manga_Entry } from "@/api/generated/types"
import { useGetMangaMapping, useMangaManualMapping, useMangaManualSearch, useRemoveMangaMapping } from "@/api/hooks/manga.hooks"
import { useSelectedMangaProvider } from "@/app/(main)/manga/_lib/handle-manga-selected-provider"
import { ConfirmationDialog, useConfirmationDialog } from "@/components/shared/confirmation-dialog"
import { imageShimmer } from "@/components/shared/image-helpers"
import { AppLayoutStack } from "@/components/ui/app-layout"
import { Button } from "@/components/ui/button"
import { cn } from "@/components/ui/core/styling"
import { defineSchema, Field, Form, InferType } from "@/components/ui/form"
import { LoadingSpinner } from "@/components/ui/loading-spinner"
import { Modal } from "@/components/ui/modal"
import { Separator } from "@/components/ui/separator"
import { Tooltip } from "@/components/ui/tooltip"
import Image from "next/image"
import { useRouter } from "next/navigation"
import React from "react"
import { FiSearch } from "react-icons/fi"

type MangaManualMappingModalProps = {
    entry: Manga_Entry
    children: React.ReactElement
}

export function MangaManualMappingModal(props: MangaManualMappingModalProps) {

    const {
        children,
        entry,
        ...rest
    } = props

    return (
        <>
            <Modal
                data-manga-manual-mapping-modal
                title="Manual match"
                trigger={children}
                contentClass="max-w-4xl"
            >
                <Content entry={entry} />
            </Modal>
        </>
    )
}

const searchSchema = defineSchema(({ z }) => z.object({
    query: z.string().min(1),
}))

function Content({ entry }: { entry: Manga_Entry }) {
    const router = useRouter()
    const { selectedProvider } = useSelectedMangaProvider(entry.mediaId)

    // Get current mapping
    const { data: existingMapping, isLoading: mappingLoading } = useGetMangaMapping({
        provider: selectedProvider || undefined,
        mediaId: entry.mediaId,
    })

    // Search
    const { mutate: search, data: searchResults, isPending: searchLoading, reset } = useMangaManualSearch(entry.mediaId, selectedProvider)

    function handleSearch(data: InferType<typeof searchSchema>) {
        if (selectedProvider) {
            search({
                provider: selectedProvider,
                query: data.query,
            })
        }
    }

    // Match
    const { mutate: match, isPending: isMatching } = useMangaManualMapping()

    // Unmatch
    const { mutate: unmatch, isPending: isUnmatching } = useRemoveMangaMapping()

    const [mangaId, setMangaId] = React.useState<string | null>(null)
    const confirmMatch = useConfirmationDialog({
        title: "Manual match",
        description: "Are you sure you want to match this manga to the search result?",
        actionText: "Confirm",
        actionIntent: "success",
        onConfirm: () => {
            if (mangaId && selectedProvider) {
                match({
                    provider: selectedProvider,
                    mediaId: entry.mediaId,
                    mangaId: mangaId,
                })
                reset()
                setMangaId(null)
            }
        },
    })

    return (
        <>
            {mappingLoading ? (
                <LoadingSpinner />
            ) : (
                <AppLayoutStack>
                    <div className="text-center">
                        {!!existingMapping?.mangaId ? (
                            <AppLayoutStack>
                                <p>
                                    Current mapping: <span>{existingMapping.mangaId}</span>
                                </p>
                                <Button
                                    intent="alert-subtle" loading={isUnmatching} onClick={() => {
                                    if (selectedProvider) {
                                        unmatch({
                                            provider: selectedProvider,
                                            mediaId: entry.mediaId,
                                        })
                                    }
                                }}
                                >
                                    Remove mapping
                                </Button>
                            </AppLayoutStack>
                        ) : (
                            <p className="text-[--muted] italic">No manual match</p>
                        )}
                    </div>

                    <Separator />

                    <div>
                        <p>Search from provider</p>
                    </div>

                    <Form schema={searchSchema} onSubmit={handleSearch}>
                        <div className="flex gap-2 items-center">
                            <Field.Text
                                name="query"
                                placeholder="Search"
                                leftIcon={<FiSearch className="text-xl text-[--muted]" />}
                                fieldClass="w-full"
                            />

                            <Field.Submit intent="white" loading={isMatching || searchLoading || mappingLoading} className="">Search</Field.Submit>
                        </div>
                    </Form>

                    {searchLoading ? <LoadingSpinner /> : (
                        <>
                            <div className="grid grid-cols-2 sm:grid-cols-3 md:grid-cols-5 gap-2">
                                {searchResults?.map(item => (
                                    <div
                                        key={item.id}
                                        className={cn(
                                            "group/sr-item col-span-1 aspect-[6/7] rounded-[--radius-md] relative bg-[--background] cursor-pointer transition-opacity",
                                        )}
                                        onClick={() => {
                                            setMangaId(item.id)
                                            React.startTransition(() => {
                                                confirmMatch.open()
                                            })
                                        }}
                                    >

                                        {<Image
                                            src={item.image || "/no-cover.png"}
                                            placeholder={imageShimmer(700, 475)}
                                            sizes="10rem"
                                            fill
                                            alt=""
                                            className={cn(
                                                "object-center object-cover lg:opacity-50 rounded-[--radius-md] transition-opacity lg:group-hover/sr-item:opacity-100",
                                            )}
                                        />}
                                        {/*<Badge intent="gray-solid" size="sm" className="absolute text-sm top-1 left-1">*/}
                                        {/*    {item.id}*/}
                                        {/*</Badge>*/}
                                        <Tooltip
                                            trigger={<p className="line-clamp-2 text-sm absolute m-2 bottom-0 font-semibold z-[10]">
                                                {item.title} {item.year && `(${item.year})`}
                                            </p>}
                                            className="z-[10]"
                                        >
                                            <p>
                                                {item.title} {item.year && `(${item.year})`}
                                            </p>
                                        </Tooltip>
                                        <div
                                            className="z-[5] absolute rounded-br-md rounded-bl-md bottom-0 w-full h-[80%] bg-gradient-to-t from-[--background] to-transparent"
                                        />
                                        {/*<div*/}
                                        {/*    className={cn(*/}
                                        {/*        "z-[5] absolute top-0 w-full h-[80%] bg-gradient-to-b from-[--background] to-transparent transition-opacity",*/}
                                        {/*    )}*/}
                                        {/*/>*/}
                                    </div>
                                ))}
                            </div>
                        </>
                    )}

                </AppLayoutStack>
            )}

            <ConfirmationDialog {...confirmMatch} />
        </>
    )
}
