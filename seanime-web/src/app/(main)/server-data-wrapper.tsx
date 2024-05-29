import { useGetStatus } from "@/api/hooks/status.hooks"
import { GettingStartedPage } from "@/app/(main)/_features/getting-started/getting-started-page"
import { useServerStatus, useSetServerStatus } from "@/app/(main)/_hooks/use-server-status"
import { LoadingOverlayWithLogo } from "@/components/shared/loading-overlay-with-logo"
import { LuffyError } from "@/components/shared/luffy-error"
import { AppLayoutStack } from "@/components/ui/app-layout"
import { Button } from "@/components/ui/button"
import { Card } from "@/components/ui/card"
import { defineSchema, Field, Form } from "@/components/ui/form"
import { logger } from "@/lib/helpers/debug"
import { ANILIST_OAUTH_URL, ANILIST_PIN_URL } from "@/lib/server/config"
import { usePathname, useRouter } from "next/navigation"
import React from "react"

type ServerDataWrapperProps = {
    children?: React.ReactNode
}

export function ServerDataWrapper(props: ServerDataWrapperProps) {

    const {
        children,
        ...rest
    } = props

    const pathname = usePathname()
    const router = useRouter()
    const serverStatus = useServerStatus()
    const setServerStatus = useSetServerStatus()
    const { data: _serverStatus, isLoading } = useGetStatus()

    React.useEffect(() => {
        if (_serverStatus) {
            logger("SERVER").info("Server status", _serverStatus)
            setServerStatus(_serverStatus)
        }
    }, [_serverStatus])


    /**
     * If the server status is loading or doesn't exist, show the loading overlay
     */
    if (isLoading || !serverStatus) return <LoadingOverlayWithLogo />

    /**
     * If the pathname is /auth/callback, show the callback page
     */
    if (pathname.startsWith("/auth/callback")) return children

    /**
     * If the server status doesn't have settings, show the getting started page
     */
    if (!serverStatus?.settings) {
        return <GettingStartedPage status={serverStatus} />
    }

    /**
     * If the app is updating, show a different screen
     */
    if (serverStatus?.updating) {
        return <div className="container max-w-3xl py-10">
            <div className="mb-4 flex justify-center w-full">
                <img src="/logo_2.png" alt="logo" className="w-36 h-auto" />
            </div>
            <p className="text-center text-lg">
                The app is currently updating. Once the update is complete and the connection has been reestablished, please <strong>refresh the
                                                                                                                                     page</strong>.
            </p>
        </div>
    }

    /**
     * Check feature flag routes
     */

    if (!serverStatus?.mediastreamSettings?.transcodeEnabled && pathname.startsWith("/mediastream")) {
        return <LuffyError title="Transcoding not enabled" />
    }

    if (!serverStatus?.user && window?.location?.host === "127.0.0.1:43211") {
        return <div className="container max-w-3xl py-10">
            <Card className="md:py-10">
                <AppLayoutStack>
                    <div className="text-center space-y-4">
                        <div className="mb-4 flex justify-center w-full">
                            <img src="/logo.png" alt="logo" className="w-24 h-auto" />
                        </div>
                        <h3>Welcome!</h3>
                        <Button
                            onClick={() => {
                                const url = serverStatus?.anilistClientId
                                    ? `https://anilist.co/api/v2/oauth/authorize?client_id=${serverStatus?.anilistClientId}&response_type=token`
                                    : ANILIST_OAUTH_URL
                                window.open(url, "_self")
                            }}
                            leftIcon={<svg
                                xmlns="http://www.w3.org/2000/svg" fill="currentColor" width="24" height="24"
                                viewBox="0 0 24 24" role="img"
                            >
                                <path
                                    d="M6.361 2.943 0 21.056h4.942l1.077-3.133H11.4l1.052 3.133H22.9c.71 0 1.1-.392 1.1-1.101V17.53c0-.71-.39-1.101-1.1-1.101h-6.483V4.045c0-.71-.392-1.102-1.101-1.102h-2.422c-.71 0-1.101.392-1.101 1.102v1.064l-.758-2.166zm2.324 5.948 1.688 5.018H7.144z"
                                />
                            </svg>}
                            intent="primary"
                            size="xl"
                        >Log in with AniList</Button>
                    </div>
                </AppLayoutStack>
            </Card>
        </div>
    } else if (!serverStatus?.user && window?.location?.host !== "127.0.0.1:43211") {
        return <div className="container max-w-3xl py-10">
            <Card className="md:py-10">
                <AppLayoutStack>
                    <div className="text-center space-y-4">
                        <div className="mb-4 flex justify-center w-full">
                            <img src="/logo.png" alt="logo" className="w-24 h-auto" />
                        </div>
                        <h3>Welcome!</h3>
                        <Button
                            onClick={() => {
                                window.open(ANILIST_PIN_URL, "_blank")
                            }}
                            leftIcon={<svg
                                xmlns="http://www.w3.org/2000/svg" fill="currentColor" width="24" height="24"
                                viewBox="0 0 24 24" role="img"
                            >
                                <path
                                    d="M6.361 2.943 0 21.056h4.942l1.077-3.133H11.4l1.052 3.133H22.9c.71 0 1.1-.392 1.1-1.101V17.53c0-.71-.39-1.101-1.1-1.101h-6.483V4.045c0-.71-.392-1.102-1.101-1.102h-2.422c-.71 0-1.101.392-1.101 1.102v1.064l-.758-2.166zm2.324 5.948 1.688 5.018H7.144z"
                                />
                            </svg>}
                            intent="white"
                            size="md"
                        >Get AniList token</Button>

                        <Form
                            schema={defineSchema(({ z }) => z.object({
                                token: z.string().min(1, "Token is required"),
                            }))}
                            onSubmit={data => {
                                router.push("/auth/callback#access_token=" + data.token.trim())
                            }}
                        >
                            <Field.Textarea
                                name="token"
                                label="Enter the token"
                                fieldClass="px-4"
                            />
                            <Field.Submit showLoadingOverlayOnSuccess>Continue</Field.Submit>
                        </Form>
                    </div>
                </AppLayoutStack>
            </Card>
        </div>
    }

    return children
}
