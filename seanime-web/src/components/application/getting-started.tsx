import { serverStatusAtom } from "@/atoms/server-status"
import { LoadingOverlayWithLogo } from "@/components/shared/loading-overlay-with-logo"
import { AppLayoutStack } from "@/components/ui/app-layout"
import { Card } from "@/components/ui/card"
import { Field, Form } from "@/components/ui/form"
import { Separator } from "@/components/ui/separator"
import { SeaEndpoints } from "@/lib/server/endpoints"
import { useSeaMutation } from "@/lib/server/query"
import { settingsSchema, useDefaultSettingsPaths } from "@/lib/server/settings"
import { DEFAULT_TORRENT_PROVIDER, ServerStatus, Settings } from "@/lib/server/types"
import { useSetAtom } from "jotai/react"
import { useRouter } from "next/navigation"
import React, { useEffect, useMemo } from "react"
import { FcClapperboard, FcFolder, FcMindMap, FcVideoCall, FcVlc } from "react-icons/fc"

export function GettingStarted({ status }: { status: ServerStatus }) {
    const router = useRouter()
    const { getDefaultVlcPath, getDefaultQBittorrentPath } = useDefaultSettingsPaths()
    const setServerStatus = useSetAtom(serverStatusAtom)

    const { mutate, data, isPending, isSuccess } = useSeaMutation<ServerStatus, Settings>({
        endpoint: SeaEndpoints.SETTINGS,
        mutationKey: ["patch-settings"],
        method: "patch",
    })
    useEffect(() => {
        if (!isPending && !!data?.settings) {
            setServerStatus(data)
            router.push("/")
        }
    }, [data, isPending])

    useEffect(() => {
        if (isSuccess) {
            router.push("/")
        }
    }, [isSuccess])

    const vlcDefaultPath = useMemo(() => getDefaultVlcPath(status.os), [status.os])
    const qbittorrentDefaultPath = useMemo(() => getDefaultQBittorrentPath(status.os), [status.os])

    if (isPending) return <LoadingOverlayWithLogo/>

    if (!data) return (
        <div className="container max-w-5xl py-10">
            <div className="mb-4 flex justify-center w-full">
                <img src="/logo.png" alt="logo" className="w-24 h-auto" />
            </div>
            <Card className="relative p-4">
                <AppLayoutStack>
                    <div className="space-y-4">
                        <h3>Getting started</h3>
                        <em className="text-[--muted]">These settings can be modified later.</em>
                        <Form
                            schema={settingsSchema}
                            onSubmit={data => {
                                mutate({
                                    library: {
                                        libraryPath: data.libraryPath,
                                        autoUpdateProgress: false,
                                        disableUpdateCheck: false,
                                        torrentProvider: DEFAULT_TORRENT_PROVIDER,
                                        autoScan: false,
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
                                        mpvSocket: "",
                                        mpvPath: "",
                                    },
                                    torrent: {
                                        qbittorrentPath: data.qbittorrentPath,
                                        qbittorrentHost: data.qbittorrentHost,
                                        qbittorrentPort: data.qbittorrentPort,
                                        qbittorrentPassword: data.qbittorrentPassword,
                                        qbittorrentUsername: data.qbittorrentUsername,
                                    },
                                })
                            }}
                            defaultValues={{
                                mediaPlayerHost: "127.0.0.1",
                                vlcPort: 8080,
                                mpcPort: 13579,
                                vlcPath: vlcDefaultPath,
                                qbittorrentPath: qbittorrentDefaultPath,
                                qbittorrentHost: "127.0.0.1",
                                qbittorrentPort: 8081,
                                mpcPath: "C:/Program Files/MPC-HC/mpc-hc64.exe",
                                torrentProvider: DEFAULT_TORRENT_PROVIDER,
                                autoScan: false,
                            }}
                            stackClass="space-y-4"
                        >
                            <Field.DirectorySelector
                                name="libraryPath"
                                label="Library folder"
                                leftIcon={<FcFolder/>}
                                shouldExist
                            />
                            <Separator />
                            <Field.Select
                                name="defaultPlayer"
                                label="Default player"
                                leftIcon={<FcVideoCall/>}
                                options={[
                                    { label: "VLC", value: "vlc" },
                                    { label: "MPC-HC (Windows only)", value: "mpc-hc" },
                                    { label: "MPV", value: "mpv" },
                                ]}
                                help="Player that will be used to open files and track your progress automatically."
                            />
                            {/*<Separator/>*/}
                            <Field.Text
                                name="mediaPlayerHost"
                                label="Host (VLC/MPC-HC)"
                            />
                            <h3 className="flex gap-2 items-center"><FcVlc/> VLC</h3>
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
                                />
                            </div>
                            <Field.Text
                                name="vlcPath"
                                label="Executable"
                            />
                            <h3 className="flex gap-2 items-center"><FcClapperboard /> MPC-HC (Windows)</h3>
                            <div className="flex gap-4">
                                <Field.Number
                                    name="mpcPort"
                                    label="Port"
                                />
                            </div>
                            <Field.Text
                                name="mpcPath"
                                label="Executable"
                            />

                            <h3 className="flex gap-2 items-center"><FcMindMap/> qBittorrent</h3>
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
                                />
                            </div>
                            <Field.Text
                                name="qbittorrentPath"
                                label="Executable"
                            />
                            <Field.Submit
                                className="w-full"
                                role="submit"
                                showLoadingOverlayOnSuccess={true}
                                loading={isPending}
                            >Continue</Field.Submit>
                        </Form>
                    </div>
                </AppLayoutStack>
            </Card>
            <p className="text-[--muted] mt-5 text-center">Made by 5rahim</p>
        </div>
    )
}
