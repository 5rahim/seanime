import { useServerStatus } from "@/app/(main)/_hooks/use-server-status"
import { SettingsCard, SettingsPageHeader } from "@/app/(main)/settings/_components/settings-card"
import { SettingsSubmitButton } from "@/app/(main)/settings/_components/settings-submit-button"
import { Alert } from "@/components/ui/alert"
import { cn } from "@/components/ui/core/styling"
import { Field } from "@/components/ui/form"
import { Separator } from "@/components/ui/separator"
import { Tabs, TabsList, TabsContent, TabsTrigger } from "@/components/ui/tabs"
import React from "react"
import { MdOutlineConnectWithoutContact } from "react-icons/md"
import { RiInformation2Line } from "react-icons/ri"

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
    "w-full flex flex-row lg:flex-row flex-wrap h-fit mt-4",
)

export function NakamaSettings(props: Props) {

    const {
        isPending,
        children,
        ...rest
    } = props

    const serverStatus = useServerStatus()


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
                    help="The username to identify this server to other instances."
                />
            </SettingsCard>

            <Tabs
                className={tabsRootClass}
                triggerClass={tabsTriggerClass}
                listClass={tabsListClass}
                defaultValue="peer"

            >
                <TabsList className="flex-wrap max-w-full bg-[--paper] p-2 border rounded-lg">
                    <TabsTrigger value="peer">Connect to a Nakama</TabsTrigger>
                    <TabsTrigger value="host">Host</TabsTrigger>
                </TabsList>

                <TabsContent value="peer" className="space-y-4">
                    <SettingsCard title="Connect to a Nakama">
                        <div className="flex items-center gap-2 text-sm bg-gray-50 dark:bg-gray-900/30 rounded-lg p-3 border border-gray-700 border-dashed text-blue-100">
                            <RiInformation2Line className="text-base" />
                            <span>The server you're connecting to must be running in host mode and be accessible on the internet or on the same
                                  network.</span>
                        </div>

                        <Field.Text
                            label="Nakama Server URL"
                            name="nakamaRemoteServerURL"
                            placeholder="http://{address}"
                            help="The URL of the Nakama host to connect to."
                        />

                        <Field.Text
                            label="Nakama Password"
                            name="nakamaRemoteServerPassword"
                            placeholder="Password"
                            help="The password to connect to the Nakama host."
                        />

                        <Separator className="!my-6" />

                        <h3>Library</h3>

                        <Field.Switch
                            side="right"
                            name="includeNakamaAnimeLibrary"
                            label="Use Nakama's anime library"
                            help="If enabled, the Nakama's anime library will be used as your library if it is being shared."
                        />
                    </SettingsCard>
                </TabsContent>

                <TabsContent value="host" className="space-y-4">
                    <SettingsCard title="Host">
                        <div className="flex items-center gap-2 text-sm bg-gray-50 dark:bg-gray-900/30 rounded-lg p-3 border border-gray-700 border-dashed text-blue-100">
                            <RiInformation2Line className="text-base" />
                            <span>Enabling host mode does not automatically set up remote access; you must manually expose your server using your
                                  preferred method.</span>
                        </div>

                        {!serverStatus?.serverHasPassword &&
                            <Alert intent="warning" description="Your server is not password protected. Add a password to your config file." />}

                        <Field.Switch
                            side="right"
                            name="nakamaIsHost"
                            label="Enable host mode"
                            moreHelp="Password must be set in the config file"
                            help="If enabled, this server will act as a host for other clients. This requires a host password to be set."
                        />

                        <Field.Text
                            label="Host Password"
                            name="nakamaHostPassword"
                            placeholder="Password"
                            help="Set a password to secure your host mode. This password should be different than your server password."
                        />

                        <Separator className="!my-6" />

                        <h3>Host settings</h3>

                        <Field.Switch
                            side="right"
                            name="nakamaHostShareLocalAnimeLibrary"
                            label="Share local anime library"
                            help="If enabled, this server will share its local anime library to other clients."
                        />
                    </SettingsCard>
                </TabsContent>
            </Tabs>


            <SettingsSubmitButton isPending={isPending} />

        </div>
    )
}
