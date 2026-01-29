import { useServerStatus } from "@/app/(main)/_hooks/use-server-status"
import { SettingsCard, SettingsPageHeader } from "@/app/(main)/settings/_components/settings-card"
import { SettingsSubmitButton } from "@/app/(main)/settings/_components/settings-submit-button"
import { Alert } from "@/components/ui/alert"
import { Badge } from "@/components/ui/badge"
import { cn } from "@/components/ui/core/styling"
import { Field } from "@/components/ui/form"
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs"
import React, { useState } from "react"
import { useWatch } from "react-hook-form"
import { MdOutlineConnectWithoutContact } from "react-icons/md"

type Props = {
    isPending: boolean
    children?: React.ReactNode
}

const tabsRootClass = cn("w-full contents space-y-4")

const tabsTriggerClass = cn(
    "text-base px-6 rounded-[--radius-md] w-fit border-none data-[state=active]:bg-[--subtle] data-[state=active]:text-white dark:hover:text-white",
    "h-10 lg:justify-center px-3 flex-1",
)

const tabsListClass = cn(
    "w-full flex flex-row lg:flex-row flex-wrap h-fit !mt-4",
)

const tabContentClass = cn(
    "space-y-4 animate-in fade-in-0 duration-300",
)

export function NakamaSettings(props: Props) {

    const {
        isPending,
        children,
        ...rest
    } = props

    const serverStatus = useServerStatus()
    const nakamaIsHost = useWatch({ name: "nakamaIsHost" })

    const [tab, setTab] = useState("peer")

    React.useLayoutEffect(() => {
        setTab(serverStatus?.settings?.nakama?.isHost ? "host" : "peer")
    }, [serverStatus?.settings?.nakama?.isHost])


    return (
        <div className="space-y-4">

            <SettingsPageHeader
                title="Nakama"
                description="Communicate with other Seanime instances"
                icon={MdOutlineConnectWithoutContact}
            />

            <SettingsCard>
                <Field.Switch
                    side="right"
                    name="nakamaEnabled"
                    label="Enable Nakama"
                />

                <Field.Text
                    label="Username"
                    name="nakamaUsername"
                    placeholder="Username"
                    help="The username to identify this server to other instances. If empty a random ID will be assigned."
                />
            </SettingsCard>

            <Tabs
                value={tab}
                onValueChange={setTab}
                className={tabsRootClass}
                triggerClass={tabsTriggerClass}
                listClass={tabsListClass}
            >
                <TabsList className="flex-wrap max-w-full bg-[--paper] p-2 border rounded-xl">
                    <TabsTrigger value="peer">Connect as a Peer</TabsTrigger>
                    <TabsTrigger value="host">Hosting {serverStatus?.settings?.nakama?.isHost &&
                        <Badge intent="info" className="ml-3">Currently hosting</Badge>}</TabsTrigger>
                    {/*<TabsTrigger value="browser-client">Rendering</TabsTrigger>*/}
                </TabsList>

                <TabsContent value="host" className={tabContentClass}>

                    {!serverStatus?.serverHasPassword &&
                        <Alert
                            intent="warning"
                            title="Reminder"
                            description="Add a password to your config file before exposing your server to the internet."
                        />}

                    <SettingsCard className="!bg-gray-900 text-sm">
                        <div>
                            <p>
                                Host mode is intended for Seanime servers that are accessible over the internet.
                            </p>
                            <p>
                                However, you can use <strong>Cloud Rooms</strong> to host watch parties without exposing your server to the internet.
                            </p>
                        </div>
                    </SettingsCard>

                    <SettingsCard>

                        <Field.Switch
                            side="right"
                            name="nakamaIsHost"
                            label="Enable host mode"
                            // moreHelp="Password must be set in the config file"
                            help="If enabled, this server will act as a host for other clients. This requires a host password to be set."
                        />

                        <Field.Text
                            label="Passcode"
                            name="nakamaHostPassword"
                            placeholder="Passcode"
                            help="Set a passcode to secure your host mode and room. This passcode should be different than your server password."
                        />

                        {/*<Field.Switch*/}
                        {/*    side="right"*/}
                        {/*    name="nakamaHostEnablePortForwarding"*/}
                        {/*    label="Enable port forwarding"*/}
                        {/*    moreHelp="This might not work for all networks."*/}
                        {/*    help="If enabled, this server will expose its port to the internet. This might be required for other clients to connect to this server."*/}
                        {/*/>*/}
                    </SettingsCard>

                    {nakamaIsHost && <SettingsCard title="Settings">

                        <Field.Switch
                            side="right"
                            name="nakamaHostShareLocalAnimeLibrary"
                            label="Share local anime library"
                            help="If enabled, this server will share its local anime library to other clients."
                        />

                        <Field.MediaExclusionSelector
                            name="nakamaHostUnsharedAnimeIds"
                            label="Exclude anime from sharing"
                            help="Select anime that you don't want to share with other clients."
                        />
                    </SettingsCard>}
                </TabsContent>

                <TabsContent value="peer" className={tabContentClass}>
                    <SettingsCard>
                        {serverStatus?.settings?.nakama?.isHost && <Alert intent="info" description="Cannot connect to a host while in host mode." />}

                        <div
                            className={cn(
                                "space-y-4",
                                serverStatus?.settings?.nakama?.isHost ? "hidden" : "",
                            )}
                        >

                            <Field.Text
                                label="Nakama Server URL"
                                name="nakamaRemoteServerURL"
                                placeholder="https://{address} or room://{id}"
                                help="The URL of the Nakama host to connect to."
                            />

                            <Field.Text
                                label="Nakama Passcode"
                                name="nakamaRemoteServerPassword"
                                placeholder="Passcode"
                                help="The passcode to connect to the Nakama host."
                            />
                        </div>
                    </SettingsCard>

                    {!serverStatus?.settings?.nakama?.isHost && <SettingsCard title="Settings">
                        <Field.Switch
                            side="right"
                            name="includeNakamaAnimeLibrary"
                            label="Use Nakama's anime library"
                            help="If enabled, the Nakama's anime library will be used as your library if it is being shared."
                        />
                    </SettingsCard>}
                </TabsContent>

            </Tabs>

            <SettingsSubmitButton isPending={isPending} />

        </div>
    )
}
