import { useGetAnilistCacheLayerStatus, useToggleAnilistCacheLayerStatus } from "@/api/hooks/anilist.hooks"
import { useListAnimeEntryEpisodeTabExtensions } from "@/api/hooks/extensions.hooks"
import { useLocalSyncSimulatedDataToAnilist } from "@/api/hooks/local.hooks"
import { __seaCommand_shortcuts } from "@/app/(main)/_features/sea-command/sea-command"
import { SettingsCard } from "@/app/(main)/settings/_components/settings-card"
import { SettingsSubmitButton } from "@/app/(main)/settings/_components/settings-submit-button"
import { ConfirmationDialog, useConfirmationDialog } from "@/components/shared/confirmation-dialog"
import { Alert } from "@/components/ui/alert"
import { Button } from "@/components/ui/button"
import { cn } from "@/components/ui/core/styling"
import { Field } from "@/components/ui/form"
import { Separator } from "@/components/ui/separator"
import { Switch } from "@/components/ui/switch"
import { __isElectronDesktop__ } from "@/types/constants"
import { useAtom } from "jotai/react"
import React from "react"
import { useFormContext, useWatch } from "react-hook-form"
import { FaRedo } from "react-icons/fa"
import { LuCircleAlert, LuCloudUpload, LuDatabaseBackup, LuEyeOff, LuImageOff, LuImages, LuShield, LuStarOff, LuUserPen } from "react-icons/lu"
import { MdDownloading } from "react-icons/md"
import { RiMovieAiLine } from "react-icons/ri"
import { TbAlertSquareRoundedOff, TbBrowserShare, TbChecklist, TbClockPlay, TbDownloadOff, TbProgressCheck, TbRating18Plus } from "react-icons/tb"
import { useServerStatus } from "../../_hooks/use-server-status"

type ServerSettingsProps = {
    isPending: boolean
}

export function ServerSettings(props: ServerSettingsProps) {

    const {
        isPending,
        ...rest
    } = props

    const serverStatus = useServerStatus()
    const { data: episodeTabExtensions } = useListAnimeEntryEpisodeTabExtensions()

    const [shortcuts, setShortcuts] = useAtom(__seaCommand_shortcuts)
    const f = useFormContext()
    const defaultPlaybackSource = useWatch({ name: "defaultPlaybackSource" })

    const defaultPlaybackSourceOptions = React.useMemo(() => {
        const pluginOptions = Array.from(new Map((episodeTabExtensions ?? []).map(ext => [
            `ext:${ext.id}`,
            {
                value: `ext:${ext.id}`,
                label: ext.tabName ? `${ext.tabName} (${ext.name})` : ext.name,
            },
        ])).values()).sort((a, b) => a.label.localeCompare(b.label))

        const options = [
            { value: "-", label: "Automatic" },
            { value: "library", label: "Local library" },
            ...(serverStatus?.debridSettings?.enabled ? [{ value: "debridstream", label: "Debrid streaming" }] : []),
            ...(serverStatus?.torrentstreamSettings?.enabled ? [{ value: "torrentstream", label: "Torrent streaming" }] : []),
            ...(serverStatus?.settings?.library?.enableOnlinestream ? [{ value: "onlinestream", label: "Online streaming" }] : []),
            ...pluginOptions,
        ]

        if (!!defaultPlaybackSource && defaultPlaybackSource.startsWith("ext:") && !options.some(option => option.value === defaultPlaybackSource)) {
            options.push({ value: defaultPlaybackSource, label: "Unavailable plugin" })
        }

        return options
    }, [episodeTabExtensions, serverStatus, defaultPlaybackSource])

    const { mutate: upload, isPending: isUploading } = useLocalSyncSimulatedDataToAnilist()

    const { data: isApiWorking, isLoading: isFetchingApiStatus } = useGetAnilistCacheLayerStatus()
    const { mutate: toggleCacheLayer, isPending: isTogglingCacheLayer } = useToggleAnilistCacheLayerStatus()

    const confirmDialog = useConfirmationDialog({
        title: "Upload to AniList",
        description: "This will upload your local Seanime collection to your AniList account. Are you sure you want to proceed?",
        actionText: "Upload",
        actionIntent: "primary",
        onConfirm: async () => {
            if (isUploading) return
            upload()
        },
    })

    return (
        <div className="space-y-4">

            {(!isApiWorking && !isFetchingApiStatus) && (
                <Alert
                    intent="warning-basic"
                    description={<div className="space-y-1">
                        <p>The AniList API is not working. All requests will be served from the cache.</p>
                        <p>You can disable this in the app settings.</p>
                    </div>}
                    className="fixed top-4 right-4 z-[50] hidden lg:block"
                />
            )}

            <SettingsCard>
                {/*<p className="text-[--muted]">*/}
                {/*    Only applies to desktop and integrated players.*/}
                {/*</p>*/}

                <Field.Switch
                    side="right"
                    name="autoUpdateProgress"
                    label="Automatically update progress"
                    help="If enabled, your progress will be automatically updated when you watch 80% of an episode."
                    moreHelp="Only applies to desktop and integrated players."
                    icon={<TbProgressCheck className="" />}
                />
                {/*<Separator />*/}
                <Field.Switch
                    side="right"
                    name="enableWatchContinuity"
                    label="Enable watch history"
                    help="If enabled, Seanime will remember your watch progress and resume from where you left off."
                    moreHelp="Only applies to desktop and integrated players."
                    icon={<TbClockPlay className="" />}
                />

                <div data-settings-default-episode-source>
                    <Field.Select
                        name="defaultPlaybackSource"
                        label="Default episode source"
                        help="Used when opening anime pages."
                        leftIcon={<RiMovieAiLine />}
                        options={defaultPlaybackSourceOptions}
                    />
                </div>

                <Separator />

                <div data-settings-hide-anime-spoilers>
                    <Field.Switch
                        side="right"
                        label="Hide anime spoilers"
                        help="Use spoiler-safe episode art and text across continue watching, entry episode lists, and missing episodes."
                        name="hideAnimeSpoilers"
                        icon={<LuEyeOff className="" />}
                    />
                </div>

                {f.watch("hideAnimeSpoilers") && (
                    <div className="space-y-1 pl-4 border-l border-[--border] ml-2">
                        <Field.Switch
                            side="right"
                            label="Hide thumbnails"
                            name="hideAnimeSpoilerThumbnails"
                        />

                        <Field.Switch
                            side="right"
                            label="Hide titles"
                            name="hideAnimeSpoilerTitles"
                        />

                        <Field.Switch
                            side="right"
                            label="Hide descriptions"
                            name="hideAnimeSpoilerDescriptions"
                        />

                        <Field.Switch
                            side="right"
                            label="Skip next episode"
                            help="Start hiding spoilers from the episode after the next one."
                            name="hideAnimeSpoilerSkipNextEpisode"
                        />
                    </div>
                )}

                <Field.Switch
                    side="right"
                    name="hideAudienceScore"
                    label="Hide audience score"
                    help="If enabled, the audience score will be hidden until you decide to view it."
                    icon={<LuStarOff className="" />}
                />


                <Field.Switch
                    side="right"
                    name="enableAdultContent"
                    label="Enable adult content"
                    help="If disabled, adult content will be hidden from search results and your library."
                    icon={<TbRating18Plus className="" />}
                />
                {f.watch("enableAdultContent") && <div className="space-y-1 pl-4 border-l border-[--border] ml-2">
                    <Field.Switch
                        side="right"
                        name="blurAdultContent"
                        label="Blur adult content"
                        fieldClass={cn(
                            !f.watch("enableAdultContent") && "opacity-50",
                        )}
                    />
                </div>}

                <Field.Switch
                    side="right"
                    name="disableAnimeCardTrailers"
                    label="Disable anime card trailers"
                    help=""
                    icon={<LuImageOff className="" />}
                />

                <Separator />

                <div data-settings-enable-extension-secure-mode>
                    <Field.Switch
                        side="right"
                        name="enableExtensionSecureMode"
                        label="Enable Extension Secure Mode"
                        help="If enabled, Seanime will prompt you for confirmation whenever an extension tries to perform a sensitive action, even if permissions have been granted."
                        icon={<LuShield className="" />}
                    />
                </div>


            </SettingsCard>

            <SettingsCard
                title="Local Account"
                description="Local account is used when you're not using an AniList account."
            >
                <div className={cn(serverStatus?.user?.isSimulated && "opacity-50 pointer-events-none")}>
                    <Field.Switch
                        side="right"
                        name="autoSyncToLocalAccount"
                        label="Automatically back up AniList lists"
                        help="If enabled, your local lists will be periodically updated by using your AniList data. This will override any local changes you've made since the last sync."
                        icon={<LuUserPen className="" />}
                    />
                </div>
                <Separator />
                <Button
                    size="sm"
                    intent="primary-subtle"
                    loading={isUploading}
                    leftIcon={<LuCloudUpload className="size-4" />}
                    onClick={() => {
                        confirmDialog.open()
                    }}
                    disabled={serverStatus?.user?.isSimulated}
                >
                    Upload local lists to AniList
                </Button>
            </SettingsCard>

            <ConfirmationDialog {...confirmDialog} />

            <SettingsCard title="Offline mode" description="Only available when authenticated with AniList.">

                <Field.Switch
                    side="right"
                    name="autoSyncOfflineLocalData"
                    label="Auto-refresh offline media"
                    help="If disabled, you will need to manually refresh your local metadata by clicking 'Sync now' in the offline mode page."
                    moreHelp="Will be paused if you have made changes offline and have not synced them to AniList yet."
                    icon={<MdDownloading className="" />}
                />

                <Field.Switch
                    side="right"
                    name="autoSaveCurrentMediaOffline"
                    label="Auto-save currently watched/read media"
                    help="If enabled, Seanime will automatically save all media you're currently watching/reading for offline use."
                    icon={<TbChecklist className="" />}
                />

            </SettingsCard>

            <SettingsCard title="Keyboard shortcuts">
                <div className="space-y-4">
                    {[
                        {
                            label: "Open command palette",
                            value: "meta+j",
                            altValue: "q",
                        },
                    ].map(item => {
                        return (
                            <div className="flex gap-2 items-center" key={item.label}>
                                <label className="text-[--gray]">
                                    <span className="font-semibold">{item.label}</span>
                                </label>
                                <div className="flex gap-2 items-center">
                                    <Button
                                        onKeyDownCapture={(e) => {
                                            e.preventDefault()
                                            e.stopPropagation()

                                            const specialKeys = ["Control", "Shift", "Meta", "Command", "Alt", "Option"]
                                            if (!specialKeys.includes(e.key)) {
                                                const keyStr = `${e.metaKey ? "meta+" : ""}${e.ctrlKey ? "ctrl+" : ""}${e.altKey
                                                    ? "alt+"
                                                    : ""}${e.shiftKey ? "shift+" : ""}${e.key.toLowerCase()
                                                    .replace("arrow", "")
                                                    .replace("insert", "ins")
                                                    .replace("delete", "del")
                                                    .replace(" ", "space")
                                                    .replace("+", "plus")}`

                                                // Update the first shortcut
                                                setShortcuts(prev => [keyStr, prev[1]])
                                            }
                                        }}
                                        className="focus:ring-2 focus:ring-[--brand] focus:ring-offset-1"
                                        size="sm"
                                        intent="white-subtle"
                                    >
                                        {shortcuts[0]}
                                    </Button>
                                    <span className="text-[--muted]">or</span>
                                    <Button
                                        onKeyDownCapture={(e) => {
                                            e.preventDefault()
                                            e.stopPropagation()

                                            const specialKeys = ["Control", "Shift", "Meta", "Command", "Alt", "Option"]
                                            if (!specialKeys.includes(e.key)) {
                                                const keyStr = `${e.metaKey ? "meta+" : ""}${e.ctrlKey ? "ctrl+" : ""}${e.altKey
                                                    ? "alt+"
                                                    : ""}${e.shiftKey ? "shift+" : ""}${e.key.toLowerCase()
                                                    .replace("arrow", "")
                                                    .replace("insert", "ins")
                                                    .replace("delete", "del")
                                                    .replace(" ", "space")
                                                    .replace("+", "plus")}`

                                                // Update the second shortcut
                                                setShortcuts(prev => [prev[0], keyStr])
                                            }
                                        }}
                                        className="focus:ring-2 focus:ring-[--brand] focus:ring-offset-1"
                                        size="sm"
                                        intent="white-subtle"
                                    >
                                        {shortcuts[1]}
                                    </Button>
                                </div>
                                {(shortcuts[0] !== "meta+j" || shortcuts[1] !== "q") && (
                                    <Button
                                        onClick={() => {
                                            setShortcuts(["meta+j", "q"])
                                        }}
                                        className="rounded-full"
                                        size="sm"
                                        intent="white-basic"
                                        leftIcon={<FaRedo />}
                                    >
                                        Reset
                                    </Button>
                                )}
                            </div>
                        )
                    })}
                </div>
            </SettingsCard>

            <SettingsCard title="App">
                {/*<Separator />*/}
                <Field.Switch
                    side="right"
                    name="disableCacheLayer"
                    label="Disable AniList caching"
                    help="If enabled, Seanime will stop caching AniList requests to disk."
                    moreHelp="By default, all requests made to AniList are cached. This allows Seanime to keep being usable when AniList goes down. The cache directory is modifiable in the config file."
                    icon={<LuDatabaseBackup className="" />}
                />
                {!f.watch("disableCacheLayer") && (
                    <div className="space-y-1 pl-4 border-l border-[--border] ml-2">
                        <Switch
                            value={!isApiWorking}
                            onValueChange={v => toggleCacheLayer()}
                            disabled={isTogglingCacheLayer}
                            label="Enable cache-only mode"
                            moreHelp="Seanime will use cached data instead of making API requests."
                        />
                    </div>
                )}
                <Field.Switch
                    side="right"
                    name="useFallbackMetadataProvider"
                    label="Use fallback metadata provider"
                    help="If enabled, Seanime will use an alternative source to fetch episode metadata."
                    icon={<LuImages className="" />}
                />
                {/*<Separator />*/}
                {/*<Field.Switch*/}
                {/*    side="right"*/}
                {/*    name="disableAutoDownloaderNotifications"*/}
                {/*    label="Disable Auto Downloader system notifications"*/}
                {/*/>*/}
                {/*/!*<Separator />*!/*/}
                {/*<Field.Switch*/}
                {/*    side="right"*/}
                {/*    name="disableAutoScannerNotifications"*/}
                {/*    label="Disable Auto Scanner system notifications"*/}
                {/*/>*/}
                <Separator />
                <Field.Switch
                    side="right"
                    name="disableUpdateCheck"
                    label={__isElectronDesktop__ ? "Do not fetch updates" : "Do not check for updates"}
                    help={__isElectronDesktop__ ? (<span className="flex gap-2 items-center">
                        <LuCircleAlert className="size-4 text-[--blue]" />
                        <span>If enabled, new releases won't be displayed. Seanime Denshi may still auto-update in the background.</span>
                    </span>) : "If enabled, Seanime will not check for new releases."}
                    moreHelp={__isElectronDesktop__ ? "You cannot disable auto-updates for Seanime Denshi." : undefined}
                    icon={<TbDownloadOff className="" />}
                />
                <Field.Select
                    label="Update Channel"
                    name="updateChannel"
                    help={__isElectronDesktop__ ? "Also applies to Seanime Denshi auto-updates." : ""}
                    options={[
                        { label: "GitHub (Default)", value: "github" },
                        { label: "Seanime", value: "seanime" },
                        { label: "Seanime (Canary)", value: "seanime_nightly" },
                    ]}
                />
                {serverStatus?.settings?.library?.updateChannel === "seanime" && (
                    <Alert intent="info" description="You are currently using a release channel hosted on Seanime." />
                )}
                {serverStatus?.settings?.library?.updateChannel === "seanime_nightly" && (
                    <Alert
                        intent="warning"
                        description="You are currently using the canary release channel hosted on Seanime. This channel may receive unstable updates without much testing."
                    />
                )}
                <Separator />
                <Field.Switch
                    side="right"
                    name="openWebURLOnStart"
                    label="Open web UI on startup"
                    icon={<TbBrowserShare className="" />}
                />
                <Field.Switch
                    side="right"
                    name="disableNotifications"
                    label="Disable system notifications"
                    moreHelp="Notifications shown by the OS"
                    icon={<TbAlertSquareRoundedOff className="" />}
                />
            </SettingsCard>

            {/*<Accordion*/}
            {/*    type="single"*/}
            {/*    collapsible*/}
            {/*    className="border rounded-[--radius-md]"*/}
            {/*    triggerClass="dark:bg-[--paper]"*/}
            {/*    contentClass="!pt-2 dark:bg-[--paper]"*/}
            {/*>*/}
            {/*    <AccordionItem value="more">*/}
            {/*        <AccordionTrigger className="bg-gray-900 rounded-[--radius-md]">*/}
            {/*            Advanced*/}
            {/*        </AccordionTrigger>*/}
            {/*        <AccordionContent className="pt-6 flex flex-col md:flex-row gap-3">*/}
            {/*            */}
            {/*        </AccordionContent>*/}
            {/*    </AccordionItem>*/}
            {/*</Accordion>*/}


            <SettingsSubmitButton isPending={isPending} />

        </div>
    )
}

const cardCheckboxStyles = {
    itemContainerClass: cn(
        "block border border-[--border] cursor-pointer transition overflow-hidden w-full",
        "bg-gray-50 hover:bg-[--subtle] dark:bg-gray-950 border-dashed",
        "data-[checked=false]:opacity-30",
        "data-[checked=true]:bg-white dark:data-[checked=true]:bg-gray-950",
        "focus:ring-2 ring-brand-100 dark:ring-brand-900 ring-offset-1 ring-offset-[--background] focus-within:ring-2 transition",
        "data-[checked=true]:border data-[checked=true]:ring-offset-0",
    ),
    itemClass: cn(
        "hidden",
    ),
    // itemLabelClass: cn(
    //     "border-transparent border data-[checked=true]:border-brand dark:bg-transparent dark:data-[state=unchecked]:bg-transparent",
    //     "data-[state=unchecked]:bg-transparent data-[state=unchecked]:hover:bg-transparent dark:data-[state=unchecked]:hover:bg-transparent",
    //     "focus-visible:ring-0 focus-visible:ring-offset-0 focus-visible:ring-offset-transparent",
    // ),
    // itemLabelClass: "font-medium flex flex-col items-center data-[state=checked]:text-[--brand] cursor-pointer",
    stackClass: "flex md:flex-row flex-col space-y-0 gap-4",
}
