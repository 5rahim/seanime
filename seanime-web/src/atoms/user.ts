import { GetViewerQuery } from "@/lib/anilist/gql/graphql"
import { useAtom } from "jotai/react"
import { serverStatusAtom } from "@/atoms/server-status"
import { atom } from "jotai"

export const userAtom = atom<GetViewerQuery["Viewer"]>((get) => {
    const data = get(serverStatusAtom)
    return data?.user?.viewer
})

export function useCurrentUser() {

    const [user, setUser] = useAtom(userAtom)

    return {
        user,
        setUser,
    }

}