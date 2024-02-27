import { serverStatusAtom } from "@/atoms/server-status"
import { GettingStarted } from "@/components/application/getting-started"
import { LoadingOverlayWithLogo } from "@/components/shared/loading-overlay-with-logo"
import { AppLayoutStack } from "@/components/ui/app-layout"
import { Button } from "@/components/ui/button"
import { Card } from "@/components/ui/card"
import { defineSchema, Field, Form } from "@/components/ui/form"
import { ANILIST_OAUTH_URL, ANILIST_PIN_URL } from "@/lib/anilist/config"
import { SeaEndpoints } from "@/lib/server/endpoints"
import { useSeaQuery } from "@/lib/server/query"
import { ServerStatus } from "@/lib/server/types"
import { useAtom } from "jotai/react"
import Cookies from "js-cookie"
import { usePathname, useRouter } from "next/navigation"
import React, { useEffect } from "react"

type AuthWrapperProps = {
    children?: React.ReactNode
}

export function AuthWrapper(props: AuthWrapperProps) {
    const { children } = props

    const pathname = usePathname()
    const router = useRouter()
    const [serverStatus, setServerStatus] = useAtom(serverStatusAtom)

    const { data, isLoading } = useSeaQuery<ServerStatus>({
        endpoint: SeaEndpoints.STATUS,
        queryKey: ["status"],
    })

    useEffect(() => {
        if (!isLoading) {
            console.log(data)
            if (data?.user) {
                Cookies.set("anilistToken", data?.user?.token ?? "", {
                    expires: 30 * 24 * 60 * 60,
                })
            } else {
                Cookies.remove("anilistToken")
            }
            setServerStatus(data)
        }
    }, [data])

    if (pathname.startsWith("/auth/callback")) return children

    if (isLoading || !serverStatus) return <LoadingOverlayWithLogo />

    if (!serverStatus?.settings) {
        return <GettingStarted status={serverStatus} />
    }

    if (!serverStatus?.user && window.location.host === "127.0.0.1:43211") {
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
                                window.open(ANILIST_OAUTH_URL, "_self")
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
    }


    if (!serverStatus?.user && window.location.host !== "127.0.0.1:43211") {
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
