import { serverStatusAtom } from "@/app/(main)/_atoms/server-status"
import { GetViewerQuery } from "@/lib/anilist/gql/graphql"
import { atom } from "jotai"
import { useAtom } from "jotai/react"

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
