"use client"
import { serverStatusAtom } from "@/atoms/server-status"
import { cn } from "@/components/ui/core"
import { Divider } from "@/components/ui/divider"
import { TabPanels } from "@/components/ui/tabs"
import { Field, TypesafeForm } from "@/components/ui/typesafe-form"
import { SeaEndpoints } from "@/lib/server/endpoints"
import { getDefaultMpcSocket } from "@/lib/server/hooks/settings"
import { useSeaMutation } from "@/lib/server/queries/utils"
import { settingsSchema } from "@/lib/server/schemas"
import { ServerStatus, Settings } from "@/lib/server/types"
import { FcClapperboard } from "@react-icons/all-files/fc/FcClapperboard"
import { FcFolder } from "@react-icons/all-files/fc/FcFolder"
import { FcVideoCall } from "@react-icons/all-files/fc/FcVideoCall"
import { FcVlc } from "@react-icons/all-files/fc/FcVlc"
import { useAtom } from "jotai/react"
import React, { useEffect } from "react"
import toast from "react-hot-toast"
import { BsPlayCircleFill } from "react-icons/bs"

export default function Page() {
    const [status, setServerStatus] = useAtom(serverStatusAtom)

    const { mutate, data, isPending } = useSeaMutation<ServerStatus, Settings>({
        endpoint: SeaEndpoints.SETTINGS,
        mutationKey: ["patch-settings"],
        method: "patch",
        onSuccess: () => {
            toast.success("Settings updated")
        },
    })

    useEffect(() => {
        if (!isPending && !!data?.settings) {
            setServerStatus(data)
        }
    }, [data, isPending])

    return (
        <div className="p-12 space-y-4">
            <div className="space-y-1">
                <h2>Settings</h2>
                <p className="text-[--muted]">App version: {status?.version}</p>
            </div>
            {/*<Divider/>*/}
            <TypesafeForm
                schema={settingsSchema}
                onSubmit={data => {
                    mutate({
                        library: {
                            libraryPath: data.libraryPath,
                            autoUpdateProgress: data.autoUpdateProgress,
                        },
                        mediaPlayer: {
                            host: data.mediaPlayerHost,
                            defaultPlayer: data.defaultPlayer,
                            vlcPort: data.vlcPort,
                            vlcUsername: data.vlcUsername || "",
                            vlcPassword: data.vlcPassword,
                            vlcPath: data.vlcPath || "",
                            mpcPort: data.mpcPort,
                            mpcPath: data.mpcPath || "",
                            mpvSocket: data.mpvSocket || "",
                            mpvPath: data.mpvPath || "",
                        },
                        torrent: {
                            qbittorrentPath: data.qbittorrentPath,
                            qbittorrentHost: data.qbittorrentHost,
                            qbittorrentPort: data.qbittorrentPort,
                            qbittorrentPassword: data.qbittorrentPassword,
                            qbittorrentUsername: data.qbittorrentUsername,
                        },
                        anilist: {
                            hideAudienceScore: data.hideAudienceScore,
                        },
                    })
                }}
                defaultValues={{
                    libraryPath: status?.settings?.library?.libraryPath,
                    mediaPlayerHost: status?.settings?.mediaPlayer?.host,
                    defaultPlayer: status?.settings?.mediaPlayer?.defaultPlayer,
                    vlcPort: status?.settings?.mediaPlayer?.vlcPort,
                    vlcUsername: status?.settings?.mediaPlayer?.vlcUsername,
                    vlcPassword: status?.settings?.mediaPlayer?.vlcPassword,
                    vlcPath: status?.settings?.mediaPlayer?.vlcPath,
                    mpcPort: status?.settings?.mediaPlayer?.mpcPort,
                    mpcPath: status?.settings?.mediaPlayer?.mpcPath,
                    mpvSocket: status?.settings?.mediaPlayer?.mpvSocket,
                    mpvPath: status?.settings?.mediaPlayer?.mpvPath,
                    qbittorrentPath: status?.settings?.torrent?.qbittorrentPath,
                    qbittorrentHost: status?.settings?.torrent?.qbittorrentHost,
                    qbittorrentPort: status?.settings?.torrent?.qbittorrentPort,
                    qbittorrentPassword: status?.settings?.torrent?.qbittorrentPassword,
                    qbittorrentUsername: status?.settings?.torrent?.qbittorrentUsername,
                    hideAudienceScore: status?.settings?.anilist?.hideAudienceScore ?? false,
                    autoUpdateProgress: status?.settings?.library?.autoUpdateProgress ?? false,
                }}
                stackClassName="space-y-4"
            >

                <TabPanels
                    navClassName="border-[--border]"
                    tabClassName={cn(
                        "rounded-none border-b border-b-2 data-[selected=true]:text-white data-[selected=true]:border-brand-400",
                        "hover:bg-transparent dark:hover:bg-transparent hover:text-white text-sm",
                        "dark:border-transparent dark:hover:border-b-transparent dark:data-[selected=true]:border-brand-400 dark:data-[selected=true]:text-white",
                        "hover:bg-[--highlight] line-clamp-1 truncate",
                        "dark:data-[selected=true]:bg-[--highlight]",
                    )}
                >
                    <div className="border border-[--border] rounded-[--radius] bg-[--paper] text-lg space-y-2">
                        <TabPanels.Nav>
                            <TabPanels.Tab>Seanime</TabPanels.Tab>
                            <TabPanels.Tab>Media Players</TabPanels.Tab>
                            <TabPanels.Tab>qBittorrent</TabPanels.Tab>
                            <TabPanels.Tab>AniList</TabPanels.Tab>
                        </TabPanels.Nav>
                        <div className="p-4">
                            <TabPanels.Container>
                                <TabPanels.Panel className="space-y-4">
                                    <Field.DirectorySelector
                                        name="libraryPath"
                                        label="Library folder"
                                        leftIcon={<FcFolder />}
                                        help="Folder where your anime library is located. (Keep the casing consistent)"
                                        shouldExist
                                    />
                                    <Divider />
                                    <Field.Switch
                                        name="autoUpdateProgress"
                                        label="Automatically update progress"
                                        help="If enabled, your progress will be automatically updated without having to confirm it when you watch 90% of an episode."
                                    />

                                </TabPanels.Panel>

                                <TabPanels.Panel className="space-y-4">
                                    <Field.Select
                                        name="defaultPlayer"
                                        label="Default player"
                                        leftIcon={<FcVideoCall />}
                                        options={[
                                            { label: "VLC", value: "vlc" },
                                            { label: "MPC-HC (Windows only)", value: "mpc-hc" },
                                            { label: "MPV", value: "mpv" },
                                        ]}
                                        help="Player that will be used to open files and track your progress automatically."
                                    />

                                    <Field.Text
                                        name="mediaPlayerHost"
                                        label="Host"
                                    />

                                    <Divider />

                                    <h3 className="flex gap-2 items-center"><FcVlc /> VLC</h3>
                                    <div className="flex gap-4">
                                        <Field.Text
                                            name="vlcUsername"
                                            label="Username"
                                        />
                                        <Field.Text
                                            name="vlcPassword"
                                            label="Password"
                                        />
                                        <Field.Number
                                            name="vlcPort"
                                            label="Port"
                                            discrete
                                        />
                                    </div>
                                    <Field.Text
                                        name="vlcPath"
                                        label="Application path"
                                    />

                                    <Divider />

                                    <h3 className="flex gap-2 items-center"><FcClapperboard /> MPC-HC</h3>
                                    <div className="flex gap-4">
                                        <Field.Number
                                            name="mpcPort"
                                            label="Port"
                                            discrete
                                        />
                                    </div>
                                    <Field.Text
                                        name="mpcPath"
                                        label="Application path"
                                    />

                                    <Divider />

                                    <h3 className="flex gap-2 items-center"><BsPlayCircleFill className="mr-1" /> MPV</h3>
                                    <div className="flex gap-4">
                                        <Field.Text
                                            name="mpvSocket"
                                            label="Socket"
                                            placeholder={`Default: '${getDefaultMpcSocket(status?.os ?? "")}'`}
                                        />
                                        <Field.Text
                                            name="mpvPath"
                                            label="Application path"
                                            placeholder={"Defaults to 'mpv' command"}
                                            help={"Leave empty to automatically use the 'mpv' command"}
                                        />
                                    </div>
                                </TabPanels.Panel>

                                <TabPanels.Panel className="space-y-4">
                                    <Field.Text
                                        name="qbittorrentHost"
                                        label="Host"
                                    />
                                    <div className="flex gap-4">
                                        <Field.Text
                                            name="qbittorrentUsername"
                                            label="Username"
                                        />
                                        <Field.Text
                                            name="qbittorrentPassword"
                                            label="Password"
                                        />
                                        <Field.Number
                                            name="qbittorrentPort"
                                            label="Port"
                                            discrete
                                        />
                                    </div>
                                    <Field.Text
                                        name="qbittorrentPath"
                                        label="Application path"
                                    />
                                </TabPanels.Panel>

                                <TabPanels.Panel>
                                    <Field.Switch
                                        name="hideAudienceScore"
                                        label="Hide audience score"
                                    />
                                </TabPanels.Panel>
                                <div className="mt-4">
                                    <Field.Submit role="save" isLoading={isPending} />
                                </div>
                            </TabPanels.Container>
                        </div>
                    </div>
                </TabPanels>

            </TypesafeForm>
        </div>
    )

}
