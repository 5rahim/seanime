import { __seaCommand_shortcuts } from "@/app/(main)/_features/sea-command/sea-command"
import { SettingsCard } from "@/app/(main)/settings/_components/settings-card"
import { SettingsSubmitButton } from "@/app/(main)/settings/_components/settings-submit-button"
import { Button } from "@/components/ui/button"
import { cn } from "@/components/ui/core/styling"
import { Field } from "@/components/ui/form"
import { useAtom } from "jotai/react"
import React from "react"
import { FaRedo } from "react-icons/fa"

type ServerSettingsProps = {
    isPending: boolean
}

export function ServerSettings(props: ServerSettingsProps) {

    const {
        isPending,
        ...rest
    } = props

    const [shortcuts, setShortcuts] = useAtom(__seaCommand_shortcuts)

    return (
        <div className="space-y-4">

            {/*<Field.RadioGroup*/}
            {/*    options={[*/}
            {/*        {*/}
            {/*            label: <div className="flex flex-col items-center gap-1">*/}
            {/*                <div className="text-5xl text-[--muted]">*/}
            {/*                    <IoLibrarySharp />*/}
            {/*                </div>*/}
            {/*                <p className="text-base md:text-lg">*/}
            {/*                    Download & Stream*/}
            {/*                </p>*/}
            {/*                <p className="text-sm text-[--muted] text-center">*/}
            {/*                    Anime library + Streaming<br/>This is the default mode*/}
            {/*                </p>*/}
            {/*            </div>, value: "-",*/}
            {/*        },*/}
            {/*        {*/}
            {/*            label: <div className="flex flex-col items-center gap-1">*/}
            {/*                <div className="text-5xl text-[--muted]">*/}
            {/*                    <SiBittorrent />*/}
            {/*                </div>*/}
            {/*                <p className="text-base md:text-lg">*/}
            {/*                    Torrent streaming*/}
            {/*                </p>*/}
            {/*                <p className="text-sm text-[--muted]">*/}
            {/*                    Stream torrents only*/}
            {/*                </p>*/}
            {/*            </div>, value: "torrentstream",*/}
            {/*        },*/}
            {/*        {*/}
            {/*            label: <div className="flex flex-col items-center gap-1">*/}
            {/*                <div className="text-5xl text-[--muted]">*/}
            {/*                    <HiServerStack />*/}
            {/*                </div>*/}
            {/*                <p className="text-base md:text-lg">*/}
            {/*                    Debrid streaming*/}
            {/*                </p>*/}
            {/*                <p className="text-sm text-[--muted]">*/}
            {/*                    Stream debrid torrents only*/}
            {/*                </p>*/}
            {/*            </div>, value: "debridstream",*/}
            {/*        },*/}
            {/*        {*/}
            {/*            label: <div className="flex flex-col items-center gap-1">*/}
            {/*                <div className="text-5xl text-[--muted]">*/}
            {/*                    <CgMediaPodcast />*/}
            {/*                </div>*/}
            {/*                <p className="text-base md:text-lg">*/}
            {/*                    Online streaming*/}
            {/*                </p>*/}
            {/*                <p className="text-sm text-[--muted]">*/}
            {/*                    Stream online only*/}
            {/*                </p>*/}
            {/*            </div>, value: "onlinestream",*/}
            {/*        },*/}
            {/*    ]}*/}
            {/*    name="userMode"*/}
            {/*    label={<h4 className="flex items-center gap-2">Anime experience <Tooltip trigger={<BiInfoCircle />}>Changing the mode will disable some features.</Tooltip></h4>}*/}
            {/*    stackClass="grid grid-cols-2 md:grid-cols-2 lg:grid-cols-3 2xl:grid-cols-4 min-[2000px]:grid-cols-5 gap-4 py-2"*/}
            {/*    fieldLabelClass="text-xl"*/}
            {/*    itemContainerClass={cn(*/}
            {/*        "cursor-pointer aspect-square transition border-transparent rounded-[--radius] p-4 w-full h-52 justify-center",*/}
            {/*        "bg-gray-50 hover:bg-[--subtle] dark:bg-gray-900",*/}
            {/*        "data-[state=checked]:bg-white dark:data-[state=checked]:bg-gray-950",*/}
            {/*        "focus:ring-2 ring-brand-100 dark:ring-brand-900 ring-offset-1 ring-offset-[--background] focus-within:ring-2 transition",*/}
            {/*        "data-[state=checked]:border data-[state=checked]:border-[--brand] data-[state=checked]:ring-offset-0",*/}
            {/*    )}*/}
            {/*    itemClass={cn(*/}
            {/*        "border-transparent absolute top-2 right-2 bg-transparent dark:bg-transparent dark:data-[state=unchecked]:bg-transparent",*/}
            {/*        "data-[state=unchecked]:bg-transparent data-[state=unchecked]:hover:bg-transparent dark:data-[state=unchecked]:hover:bg-transparent",*/}
            {/*        "focus-visible:ring-0 focus-visible:ring-offset-0 focus-visible:ring-offset-transparent",*/}
            {/*    )}*/}
            {/*    itemLabelClass="font-medium flex flex-col items-center data-[state=checked]:text-[--brand] cursor-pointer"*/}
            {/*    itemCheckIcon={<BiCheck className="text-white text-lg" />}*/}
            {/*    help="Choose the mode you want to fine-tune your anime experience."*/}
            {/*/>*/}

            {/*/!*<Separator />*!/*/}

            <SettingsCard title="Anime">
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

            </SettingsCard>

            <SettingsCard title="Offline">

                <Field.Switch
                    side="right"
                    name="autoSyncOfflineLocalData"
                    label="Automatically refresh local data"
                    help="If enabled, local data will be refreshed periodically using current AniList data."
                    moreHelp="Only if no offline changes have been made."
                />

            </SettingsCard>

            <SettingsCard title="Notifications">

                <Field.Switch
                    side="right"
                    name="disableNotifications"
                    label="Disable notifications"
                />
                {/*<Separator />*/}
                <Field.Switch
                    side="right"
                    name="disableAutoDownloaderNotifications"
                    label="Disable Auto Downloader notifications"
                />
                {/*<Separator />*/}
                <Field.Switch
                    side="right"
                    name="disableAutoScannerNotifications"
                    label="Disable Auto Scanner notifications"
                />

            </SettingsCard>

            <SettingsCard title="App">
                <Field.Switch
                    side="right"
                    name="disableUpdateCheck"
                    label="Do not check for updates"
                    help="If enabled, Seanime will not check for new releases."
                />
                {/*<Separator />*/}
                <Field.Switch
                    side="right"
                    name="openTorrentClientOnStart"
                    label="Open torrent client on startup"
                />
                {/*<Separator />*/}
                <Field.Switch
                    side="right"
                    name="openWebURLOnStart"
                    label="Open localhost web URL on startup"
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
