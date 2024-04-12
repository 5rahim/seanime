import { GetViewerQuery } from "@/lib/anilist/gql/graphql"
import { Settings, ThemeSettings } from "@/lib/types/settings.types"

export type ServerStatus = {
    os: string,
    user: {
        viewer: GetViewerQuery["Viewer"],
        token: string
    } | null,
    settings: Settings | null
    mal: MalInfo | null
    version: string
    themeSettings?: ThemeSettings | null
    isOffline: boolean
}

export type MalInfo = {
    username: string
    accessToken: string
    refreshToken: string
}
