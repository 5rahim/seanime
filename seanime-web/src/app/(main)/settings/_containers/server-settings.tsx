import { useGetAnilistCacheLayerStatus, useToggleAnilistCacheLayerStatus } from "@/api/hooks/anilist.hooks"
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
import { useFormContext } from "react-hook-form"
import { FaRedo } from "react-icons/fa"
import { LuCircleAlert, LuCloudUpload } from "react-icons/lu"
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

    const [shortcuts, setShortcuts] = useAtom(__seaCommand_shortcuts)
    const f = useFormContext()

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
                />
                {/*<Separator />*/}
                <Field.Switch
                    side="right"
                    name="enableWatchContinuity"
                    label="Enable watch history"
                    help="If enabled, Seanime will remember your watch progress and resume from where you left off."
                    moreHelp="Only applies to desktop and integrated players."
                />

                <Field.Switch
                    side="right"
                    name="disableAnimeCardTrailers"
                    label="Disable anime card trailers"
                    help=""
                />

                <Separator />

                <Field.Switch
                    side="right"
                    name="hideAudienceScore"
                    label="Hide audience score"
                    help="If enabled, the audience score will be hidden until you decide to view it."
                />

                <Field.Switch
                    side="right"
                    name="enableAdultContent"
                    label="Enable adult content"
                    help="If disabled, adult content will be hidden from search results and your library."
                />
                <Field.Switch
                    side="right"
                    name="blurAdultContent"
                    label="Blur adult content"
                    help="If enabled, adult content will be blurred."
                    fieldClass={cn(
                        !f.watch("enableAdultContent") && "opacity-50",
                    )}
                />

            </SettingsCard>

            <SettingsCard
                title="Local Data"
                description="Local data is used when you're not using an AniList account."
            >
                <div className={cn(serverStatus?.user?.isSimulated && "opacity-50 pointer-events-none")}>
                    <Field.Switch
                        side="right"
                        name="autoSyncToLocalAccount"
                        label="Auto backup lists from AniList"
                        help="If enabled, your local lists will be periodically updated by using your AniList data."
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
                    label="Download metadata automatically for offline use"
                    help="If disabled, you will need to manually refresh your local metadata by clicking 'Sync now' in the offline mode page."
                    moreHelp="Will be paused if you have made changes offline and have not synced them to AniList yet."
                />

                <Field.Switch
                    side="right"
                    name="autoSaveCurrentMediaOffline"
                    label="Save all currently watched/read media for offline use"
                    help="If enabled, Seanime will automatically save all media you're currently watching/reading for offline use."
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
                    name="openWebURLOnStart"
                    label="Open localhost web URL on startup"
                />
                <Field.Switch
                    side="right"
                    name="disableNotifications"
                    label="Disable system notifications"
                    moreHelp="Notifications shown by the OS when Seanime runs the auto-downloader or auto-scanner."
                />
                <Field.Switch
                    side="right"
                    name="disableCacheLayer"
                    label="Disable AniList caching"
                    help="If enabled, Seanime will stop caching AniList requests to disk."
                    moreHelp="By default, all requests made to AniList are cached. This allows Seanime to keep being usable when AniList goes down. The cache directory is modifiable in the config file."
                />
                {!f.watch("disableCacheLayer") && (
                    <div>
                        <Switch
                            value={!isApiWorking}
                            onValueChange={v => toggleCacheLayer()}
                            disabled={isTogglingCacheLayer}
                            label="Use cache-only mode"
                            moreHelp="Seanime will use cached data instead of making API requests."
                        />
                    </div>
                )}
                <Field.Switch
                    side="right"
                    name="useFallbackMetadataProvider"
                    label="Use fallback metadata provider"
                    help="If enabled, Seanime will use an alternative source to fetch episode metadata."
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
                    label={__isElectronDesktop__ ? "Do not fetch update notes" : "Do not check for updates"}
                    help={__isElectronDesktop__ ? (<span className="flex gap-2 items-center">
                        <LuCircleAlert className="size-4 text-[--blue]" />
                        <span>If enabled, new releases won't be displayed. Seanime Denshi will still auto-update in the background.</span>
                    </span>) : "If enabled, Seanime will not check for new releases."}
                    moreHelp={__isElectronDesktop__ ? "You cannot disable auto-updates for Seanime Denshi." : undefined}
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
