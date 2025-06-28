import { AL_BaseAnime, Anime_AutoDownloaderRule } from "@/api/generated/types"
import { AutoDownloaderRuleForm } from "@/app/(main)/auto-downloader/_containers/autodownloader-rule-form"
import { IconButton } from "@/components/ui/button"
import { cn } from "@/components/ui/core/styling"
import { Modal } from "@/components/ui/modal"
import { useBoolean } from "@/hooks/use-disclosure"
import React from "react"
import { BiChevronRight } from "react-icons/bi"
import { FaSquareRss } from "react-icons/fa6"

export type AutoDownloaderRuleItemProps = {
    rule: Anime_AutoDownloaderRule
    userMedia: AL_BaseAnime[] | undefined
}

export function AutoDownloaderRuleItem(props: AutoDownloaderRuleItemProps) {

    const {
        rule,
        userMedia,
        ...rest
    } = props

    const modal = useBoolean(false)

    const media = React.useMemo(() => {
        return userMedia?.find(media => media.id === rule.mediaId)
    }, [(userMedia?.length || 0), rule])

    return (
        <>
            <div className="rounded-[--radius] bg-gray-900 hover:bg-gray-800 transition-colors">
                <div className="flex justify-between p-3 gap-2 items-center cursor-pointer" onClick={() => modal.on()}>

                    <div className="space-y-1 w-full">
                        <p
                            className={cn(
                                "font-medium text-base tracking-wide line-clamp-1",
                            )}
                        ><span className="text-gray-400 italic font-normal pr-1">Rule for</span> "{rule.comparisonTitle}"</p>
                        <p className="text-sm text-gray-400 line-clamp-1 flex space-x-2 items-center divide-x divide-[--border] [&>span]:pl-2">
                            <FaSquareRss
                                className={cn(
                                    "text-xl",
                                    rule.enabled ? "text-green-500" : "text-gray-500",
                                    (!media) && "text-red-300",
                                )}
                            />
                            {!!rule.releaseGroups?.length && <span>{rule.releaseGroups.join(", ")}</span>}
                            {!!rule.resolutions?.length && <span>{rule.resolutions.join(", ")}</span>}
                            {!!rule.episodeType && <span>{getEpisodeTypeName(rule.episodeType)}</span>}
                            {!!media ? (
                                <>
                                    {media.status === "FINISHED" &&
                                        <span className="text-orange-300 opacity-70">This anime is no longer airing</span>}
                                </>
                            ) : (
                                <span className="text-red-300">This anime is not in your library</span>
                            )}
                        </p>
                    </div>

                    <div>
                        <IconButton intent="white-basic" icon={<BiChevronRight />} size="sm" />
                    </div>
                </div>
            </div>
            <Modal
                open={modal.active}
                onOpenChange={modal.off}
                title="Edit rule"
                contentClass="max-w-4xl"

            >
                <AutoDownloaderRuleForm type="edit" rule={rule} />
            </Modal>
        </>
    )
}

function getEpisodeTypeName(episodeType: Anime_AutoDownloaderRule["episodeType"]) {
    switch (episodeType) {
        case "recent":
            return "Recent releases"
        case "selected":
            return "Select episodes"
    }
}
