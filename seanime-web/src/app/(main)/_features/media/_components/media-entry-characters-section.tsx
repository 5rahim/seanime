import { AL_AnimeDetailsById_Media, AL_MangaDetailsById_Media } from "@/api/generated/types"
import { imageShimmer } from "@/components/shared/image-helpers"
import { SeaLink } from "@/components/shared/sea-link"
import { cn } from "@/components/ui/core/styling"
import { useThemeSettings } from "@/lib/theme/hooks"
import Image from "next/image"
import React from "react"
import { BiSolidHeart } from "react-icons/bi"

type RelationsRecommendationsSectionProps = {
    details: AL_AnimeDetailsById_Media | AL_MangaDetailsById_Media | undefined
    isMangaPage?: boolean
}

export function MediaEntryCharactersSection(props: RelationsRecommendationsSectionProps) {

    const {
        details,
        isMangaPage,
        ...rest
    } = props

    const ts = useThemeSettings()

    const characters = React.useMemo(() => {
        return details?.characters?.edges?.filter(n => n.role === "MAIN" || n.role === "SUPPORTING") || []
    }, [details?.characters?.edges])

    if (characters.length === 0) return null

    return (
        <>
            {/*{!isMangaPage && <Separator />}*/}

            <h2 data-media-entry-characters-section-title>Characters</h2>

            <div
                data-media-entry-characters-section-grid
                className={cn(
                    "grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 2xl:grid-cols-5 gap-4",
                    isMangaPage && "grid-cols-1 md:grid-col-2 lg:grid-cols-3 xl:grid-cols-2 2xl:grid-cols-2",
                )}
            >
                {characters?.slice(0, 10).map(edge => {
                    return <div key={edge?.node?.id} className="col-span-1" data-media-entry-characters-section-grid-item>
                        <div
                            data-media-entry-characters-section-grid-item-container
                            className={cn(
                                "max-w-full flex gap-4",
                                "rounded-lg relative transition group/episode-list-item select-none",
                                !!ts.libraryScreenCustomBackgroundImage && ts.libraryScreenCustomBackgroundOpacity > 5
                                    ? "bg-[--background] p-3"
                                    : "py-3",
                                "pr-12",
                            )}
                            {...rest}
                        >

                            <div
                                data-media-entry-characters-section-grid-item-image-container
                                className={cn(
                                    "size-20 flex-none rounded-[--radius-md] object-cover object-center relative overflow-hidden",
                                    "group/ep-item-img-container",
                                )}
                            >
                                <div
                                    data-media-entry-characters-section-grid-item-image-overlay
                                    className="absolute z-[1] rounded-[--radius-md] w-full h-full"
                                ></div>
                                <div
                                    data-media-entry-characters-section-grid-item-image-background
                                    className="bg-[--background] absolute z-[0] rounded-[--radius-md] w-full h-full"
                                ></div>
                                {(edge?.node?.image?.large) && <Image
                                    data-media-entry-characters-section-grid-item-image
                                    src={edge?.node?.image?.large || ""}
                                    alt="episode image"
                                    fill
                                    quality={60}
                                    placeholder={imageShimmer(700, 475)}
                                    sizes="10rem"
                                    className={cn("object-cover object-center transition select-none")}
                                    data-src={edge?.node?.image?.large}
                                />}
                            </div>

                            <div data-media-entry-characters-section-grid-item-content>
                                <SeaLink href={edge?.node?.siteUrl || "#"} target="_blank" data-media-entry-characters-section-grid-item-content-link>
                                    <p
                                        className={cn(
                                            "text-lg font-semibold transition line-clamp-2 leading-5 hover:text-brand-100",
                                        )}
                                    >
                                        {edge?.node?.name?.full}
                                    </p>
                                </SeaLink>

                                {edge?.node?.age && <p data-media-entry-characters-section-grid-item-content-age className="text-sm">
                                    {edge?.node?.age} years old
                                </p>}

                                <p data-media-entry-characters-section-grid-item-content-role className="text-[--muted] text-xs">
                                    {edge?.role}
                                </p>

                                {edge?.node?.isFavourite && <div data-media-entry-characters-section-grid-item-content-favourite>
                                    <BiSolidHeart className="text-pink-600 text-lg block" />
                                </div>}
                            </div>
                        </div>
                    </div>
                })}
            </div>
        </>
    )
}
