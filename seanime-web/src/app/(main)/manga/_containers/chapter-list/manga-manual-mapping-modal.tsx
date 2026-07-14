import { HibikeManga_SearchResult, Manga_Entry } from "@/api/generated/types"
import {
    useGetMangaMapping,
    useMangaManualMapping,
    useMangaManualSearch,
    usePreviewMangaMapping,
    useRemoveMangaMapping,
} from "@/api/hooks/manga.hooks"
import { useSelectedMangaProvider } from "@/app/(main)/manga/_lib/handle-manga-selected-provider"
import { useMangaReaderUtils } from "@/app/(main)/manga/_lib/handle-manga-utils"
import { ConfirmationDialog, useConfirmationDialog } from "@/components/shared/confirmation-dialog"
import { imageShimmer } from "@/components/shared/image-helpers"
import { SeaImage } from "@/components/shared/sea-image"
import { AppLayoutStack } from "@/components/ui/app-layout"
import { Button } from "@/components/ui/button"
import { cn } from "@/components/ui/core/styling"
import { defineSchema, Field, Form, InferType } from "@/components/ui/form"
import { LoadingSpinner } from "@/components/ui/loading-spinner"
import { Modal } from "@/components/ui/modal"
import { Separator } from "@/components/ui/separator"
import { Tooltip } from "@/components/ui/tooltip"
import { useRouter } from "@/lib/navigation"
import React from "react"
import { FiSearch } from "react-icons/fi"
import { toast } from "sonner"

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
                description="Match this manga to a search result"
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
    const { getChapterPageUrl, isReady: imageProxyReady } = useMangaReaderUtils()

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

    const getSearchResultImageUrl = React.useCallback((image: string | undefined, headers?: Record<string, string>) => {
        if (!image) return "/no-cover.png"
        return imageProxyReady ? getChapterPageUrl(image, false, headers) : "/no-cover.png"
    }, [getChapterPageUrl, imageProxyReady])

    // Match
    const { mutate: match, isPending: isMatching } = useMangaManualMapping()
    const {
        mutate: previewMapping,
        data: mappingPreview,
        isPending: previewLoading,
        reset: resetPreview,
    } = usePreviewMangaMapping()

    // Unmatch
    const { mutate: unmatch, isPending: isUnmatching } = useRemoveMangaMapping()

    const [selectedResult, setSelectedResult] = React.useState<HibikeManga_SearchResult | null>(null)
    const confirmMatch = useConfirmationDialog({
        title: "Manual match",
        description: selectedResult && mappingPreview ? (
            <div className="space-y-3 text-left">
                <p>Match this entry to <span className="font-medium">{selectedResult.title}</span>?</p>
                <div className="grid grid-cols-2 gap-3 rounded-[--radius-md] border p-3 text-sm">
                    <div>
                        <p className="text-[--muted]">Distinct chapters</p>
                        <p className="font-medium">{mappingPreview.chapterCount}</p>
                    </div>
                    <div>
                        <p className="text-[--muted]">Latest chapter</p>
                        <p className="font-medium">{mappingPreview.latest || "Unknown"}</p>
                    </div>
                </div>
                {!!mappingPreview.languages?.length && (
                    <p className="text-sm"><span className="text-[--muted]">Languages:</span> {mappingPreview.languages.join(", ")}</p>
                )}
                {!!mappingPreview.scanlators?.length && (
                    <p className="text-sm"><span className="text-[--muted]">Scanlators:</span> {mappingPreview.scanlators.join(", ")}</p>
                )}
            </div>
        ) : "Review the provider result before saving this match.",
        actionText: "Confirm",
        actionIntent: "success",
        onCancel: () => {
            resetPreview()
            setSelectedResult(null)
        },
        onConfirm: () => {
            if (selectedResult && selectedProvider) {
                match({
                    provider: selectedProvider,
                    mediaId: entry.mediaId,
                    mangaId: selectedResult.id,
                })
                reset()
                resetPreview()
                setSelectedResult(null)
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

                    <Form schema={searchSchema} onSubmit={handleSearch}>
                        <div className="flex gap-2 items-center">
                            <Field.Text
                                name="query"
                                placeholder="Enter a title..."
                                leftIcon={<FiSearch className="text-xl text-[--muted]" />}
                                fieldClass="w-full"
                            />

                            <Field.Submit
                                intent="white"
                                loading={isMatching || previewLoading || searchLoading || mappingLoading}
                                className=""
                            >Search</Field.Submit>
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
                                            if (!selectedProvider || previewLoading) return
                                            setSelectedResult(item)
                                            previewMapping({ provider: selectedProvider, mangaId: item.id }, {
                                                onSuccess: preview => {
                                                    if (!preview?.chapterCount) {
                                                        toast.error("No chapters were found for this result")
                                                        setSelectedResult(null)
                                                        resetPreview()
                                                        return
                                                    }
                                                    React.startTransition(() => confirmMatch.open())
                                                },
                                                onError: () => setSelectedResult(null),
                                            })
                                        }}
                                    >

                                        {<SeaImage
                                            src={getSearchResultImageUrl(item.image, item.imageHeaders)}
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
                                            className="z-[150]"
                                        >
                                            <p>
                                                {item.title} {item.year && `(${item.year})`}
                                            </p>
                                        </Tooltip>
                                        <div
                                            className="z-[5] absolute rounded-br-md rounded-bl-md bottom-0 w-full h-[80%] bg-gradient-to-t from-[--background] to-transparent"
                                        />
                                        {previewLoading && selectedResult?.id === item.id && (
                                            <LoadingSpinner containerClass="absolute inset-0 z-[20] bg-[--background]/70" />
                                        )}
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
