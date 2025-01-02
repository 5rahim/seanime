import { SettingsSubmitButton } from "@/app/(main)/settings/_components/settings-submit-button"
import { cn } from "@/components/ui/core/styling"
import { Field } from "@/components/ui/form"
import { Separator } from "@/components/ui/separator"
import React from "react"

type ServerSettingsProps = {
    isPending: boolean
}

export function ServerSettings(props: ServerSettingsProps) {

    const {
        isPending,
        ...rest
    } = props


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

            {/*<Separator />*/}

            <Field.Switch
                name="disableUpdateCheck"
                label="Do not check for updates"
                help="If enabled, Seanime will not check for new releases."
            />
            <Field.Switch
                name="openTorrentClientOnStart"
                label="Open torrent client on startup"
            />
            <Field.Switch
                name="openWebURLOnStart"
                label="Open localhost web URL on startup"
            />

            <Separator />

            <div>
                <h3>
                    Anime tracking
                </h3>
                <p className="text-[--muted]">
                    Only applies to desktop and built-in players.
                </p>
            </div>

            <Field.Switch
                name="autoUpdateProgress"
                label="Automatically update progress"
                help="If enabled, your progress will be automatically updated without having to confirm it when you watch 80% of an episode."
            />

            <Field.Switch
                name="enableWatchContinuity"
                label="Enable watch continuity"
                help="If enabled, Seanime will remember your watch progress and resume from where you left off."
            />

            <Separator />

            <h3>Notifications</h3>

            <Field.Switch
                name="disableNotifications"
                label="Disable notifications"
            />

            <Field.Switch
                name="disableAutoDownloaderNotifications"
                label="Disable Auto Downloader notifications"
            />

            <Field.Switch
                name="disableAutoScannerNotifications"
                label="Disable Auto Scanner notifications"
            />

            <Separator />

            <h3>Offline</h3>

            <Field.Switch
                name="autoSyncOfflineLocalData"
                label="Automatically refresh local data"
                help="Automatically refresh local data with AniList data periodically if no offline changes have been made."
            />

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
