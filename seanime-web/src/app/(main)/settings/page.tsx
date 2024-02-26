"use client"
import { serverStatusAtom } from "@/atoms/server-status"
import { Accordion, AccordionContent, AccordionItem, AccordionTrigger } from "@/components/ui/accordion"
import { Card } from "@/components/ui/card"
import { Field, Form } from "@/components/ui/form"
import { Separator } from "@/components/ui/separator"
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs"
import { SeaEndpoints } from "@/lib/server/endpoints"
import { useSeaMutation } from "@/lib/server/query"
import { getDefaultMpcSocket, settingsSchema } from "@/lib/server/settings"
import { DEFAULT_TORRENT_PROVIDER, ServerStatus, Settings } from "@/lib/server/types"
import { useAtom } from "jotai/react"
import React, { useEffect } from "react"
import { BsPlayCircleFill } from "react-icons/bs"
import { FcClapperboard, FcFolder, FcVideoCall, FcVlc } from "react-icons/fc"
import { toast } from "sonner"


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
        <div className="p-8 space-y-4">
            <div className="space-y-1">
                <h2>Settings</h2>
                <p className="text-[--muted]">App version: {status?.version}</p>
            </div>
            {/*<Separator/>*/}
            <Form
                schema={settingsSchema}
                onSubmit={data => {
                    mutate({
                        library: {
                            libraryPath: data.libraryPath,
                            autoUpdateProgress: data.autoUpdateProgress,
                            disableUpdateCheck: data.disableUpdateCheck,
                            torrentProvider: data.torrentProvider,
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
                    torrentProvider: status?.settings?.library?.torrentProvider || DEFAULT_TORRENT_PROVIDER, // (Backwards compatibility)
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
                    disableUpdateCheck: status?.settings?.library?.disableUpdateCheck ?? false,
                }}
                stackClass="space-y-4"
            >

                <Card className="p-0 overflow-hidden">
                    <Tabs
                        defaultValue="seanime"
                        triggerClass="w-full data-[state=active]:bg-[--subtle]"
                    >
                        <TabsList className="flex w-full border-b">
                            <TabsTrigger value="seanime">Seanime</TabsTrigger>
                            <TabsTrigger value="media-player">Media Player</TabsTrigger>
                            <TabsTrigger value="qbittorrent">qBittorrent</TabsTrigger>
                        </TabsList>

                        <div className="p-4">
                            <TabsContent value="seanime" className="space-y-4">
                                <Field.DirectorySelector
                                    name="libraryPath"
                                    label="Library folder"
                                    leftIcon={<FcFolder />}
                                    help="Folder where your anime library is located. (Keep the casing consistent)"
                                    shouldExist
                                />
                                <Separator />
                                <Field.Switch
                                    name="autoUpdateProgress"
                                    label="Automatically update progress"
                                    help="If enabled, your progress will be automatically updated without having to confirm it when you watch 90% of an episode."
                                />
                                <Separator />
                                <Field.Switch
                                    name="disableUpdateCheck"
                                    label="Do not check for updates"
                                    help="If enabled, Seanime will not check for new releases."
                                />
                                <Separator />
                                <Field.Switch
                                    name="hideAudienceScore"
                                    label="Hide audience score"
                                    help="If enabled, the audience score will be hidden on the media entry page."
                                />
                                <Separator />
                                <Field.RadioGroup
                                    options={[
                                        { label: "Nyaa", value: "nyaa" },
                                        { label: "AnimeTosho", value: "animetosho" },
                                    ]}
                                    name="torrentProvider"
                                    label="Torrent provider"
                                    help="Provider to use for searching and downloading torrents."
                                />

                            </TabsContent>

                            <TabsContent value="media-player" className="space-y-4">
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
                                    help="VLC/MPC-HC"
                                />

                                <Accordion
                                    type="single"
                                    className="border"
                                    triggerClass=""
                                >
                                    <AccordionItem value="vlc">
                                        <AccordionTrigger>
                                            <h4 className="flex gap-2 items-center"><FcVlc /> VLC</h4>
                                        </AccordionTrigger>
                                        <AccordionContent className="space-y-4">
                                            <div className="flex flex-col md:flex-row gap-4">
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
                                                    formatOptions={{
                                                        useGrouping: false,
                                                    }}
                                                    hideControls
                                                />
                                            </div>
                                            <Field.Text
                                                name="vlcPath"
                                                label="Application path"
                                            />
                                        </AccordionContent>
                                    </AccordionItem>

                                    <AccordionItem value="mpc-hc">
                                        <AccordionTrigger>
                                            <h4 className="flex gap-2 items-center"><FcClapperboard /> MPC-HC</h4>
                                        </AccordionTrigger>
                                        <AccordionContent>
                                            <div className="flex flex-col md:flex-row gap-4">
                                                <Field.Number
                                                    name="mpcPort"
                                                    label="Port"
                                                    formatOptions={{
                                                        useGrouping: false,
                                                    }}
                                                    hideControls
                                                />
                                                <Field.Text
                                                    name="mpcPath"
                                                    label="Application path"
                                                />
                                            </div>
                                        </AccordionContent>
                                    </AccordionItem>

                                    <AccordionItem value="mpv">
                                        <AccordionTrigger>
                                            <h4 className="flex gap-2 items-center"><BsPlayCircleFill className="mr-1" /> MPV</h4>
                                        </AccordionTrigger>
                                        <AccordionContent>
                                            <div className="flex gap-4">
                                                <Field.Text
                                                    name="mpvSocket"
                                                    label="Socket"
                                                    placeholder={`Default: '${getDefaultMpcSocket(status?.os ?? "")}'`}
                                                />
                                                <Field.Text
                                                    name="mpvPath"
                                                    label="Application path"
                                                    placeholder="Defaults to 'mpv' command"
                                                    help="Leave empty to automatically use the 'mpv' command"
                                                />
                                            </div>
                                        </AccordionContent>
                                    </AccordionItem>
                                </Accordion>

                            </TabsContent>

                            <TabsContent value="qbittorrent" className="space-y-4">
                                <Field.Text
                                    name="qbittorrentHost"
                                    label="Host"
                                />
                                <div className="flex flex-col md:flex-row gap-4">
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
                                        formatOptions={{
                                            useGrouping: false,
                                        }}
                                        hideControls
                                    />
                                </div>
                                <Field.Text
                                    name="qbittorrentPath"
                                    label="Application path"
                                />
                            </TabsContent>
                            <div className="mt-4">
                                <Field.Submit role="save" loading={isPending}>Save</Field.Submit>
                            </div>
                        </div>
                    </Tabs>
                </Card>

            </Form>
        </div>
    )

}
