import { Anime_AutoDownloaderRule, Anime_Entry } from "@/api/generated/types"
import { useGetAutoDownloaderRulesByAnime } from "@/api/hooks/auto_downloader.hooks"
import { __anilist_userAnimeMediaAtom } from "@/app/(main)/_atoms/anilist.atoms"
import { useServerStatus } from "@/app/(main)/_hooks/use-server-status"
import { AutoDownloaderRuleItem } from "@/app/(main)/auto-downloader/_components/autodownloader-rule-item"
import { AutoDownloaderRuleForm } from "@/app/(main)/auto-downloader/_containers/autodownloader-rule-form"
import { LuffyError } from "@/components/shared/luffy-error"
import { Button, IconButton } from "@/components/ui/button"
import { Modal } from "@/components/ui/modal"
import { useBoolean } from "@/hooks/use-disclosure"
import { useAtomValue } from "jotai/react"
import React from "react"
import { BiPlus } from "react-icons/bi"
import { TbWorldDownload } from "react-icons/tb"

type AnimeAutoDownloaderButtonProps = {
    entry: Anime_Entry
    size?: "sm" | "md" | "lg"
}

export function AnimeAutoDownloaderButton(props: AnimeAutoDownloaderButtonProps) {

    const {
        entry,
        size,
        ...rest
    } = props

    const serverStatus = useServerStatus()
    const { data: rules, isLoading } = useGetAutoDownloaderRulesByAnime(entry.mediaId, !!serverStatus?.settings?.autoDownloader?.enabled)

    const [isModalOpen, setIsModalOpen] = React.useState(false)

    if (
        isLoading
        || !serverStatus?.settings?.autoDownloader?.enabled
        || !entry.listData
    ) return null

    const isTracked = !!rules?.length

    return (
        <>
            <Modal
                title="Auto Downloader"
                contentClass="max-w-3xl"
                open={isModalOpen}
                onOpenChange={setIsModalOpen}
                trigger={<IconButton
                    icon={isTracked ? <TbWorldDownload /> : <TbWorldDownload />}
                    loading={isLoading}
                    intent={isTracked ? "primary-subtle" : "gray-subtle"}
                    size={size}
                    {...rest}
                />}
            >
                <Content entry={entry} rules={rules} />
            </Modal>
        </>
    )
}

type ContentProps = {
    entry: Anime_Entry
    rules: Anime_AutoDownloaderRule[] | undefined
}

export function Content(props: ContentProps) {

    const {
        entry,
        rules,
        ...rest
    } = props

    const userMedia = useAtomValue(__anilist_userAnimeMediaAtom)
    const createRuleModal = useBoolean(false)

    return (
        <div className="space-y-4">

            <div className="flex w-full">
                <div className="flex-1"></div>
                <Modal
                    open={createRuleModal.active}
                    onOpenChange={createRuleModal.set}
                    title="Create a new rule"
                    contentClass="max-w-3xl"
                    trigger={<Button
                        className="rounded-full"
                        intent="success-subtle"
                        leftIcon={<BiPlus />}
                        onClick={() => {
                            createRuleModal.on()
                        }}
                    >
                        New Rule
                    </Button>}
                >
                    <AutoDownloaderRuleForm
                        mediaId={entry.mediaId}
                        type="create"
                        onRuleCreatedOrDeleted={() => createRuleModal.off()}
                    />
                </Modal>
            </div>

            {!rules?.length && (
                <LuffyError title={null}>
                    No rules found for this anime.
                </LuffyError>
            )}

            {rules?.map(rule => (
                <AutoDownloaderRuleItem
                    key={rule.dbId}
                    rule={rule}
                    userMedia={userMedia}
                />
            ))}

        </div>
    )
}
