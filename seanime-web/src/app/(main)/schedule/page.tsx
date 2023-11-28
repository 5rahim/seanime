"use client"

import { MissingEpisodes } from "@/app/(main)/schedule/_containers/missing-episodes/missing-episodes"
import { ComingUpNext } from "@/app/(main)/schedule/_containers/coming-up-next/coming-up-next"
import { RecentReleases } from "@/app/(main)/schedule/_containers/recent-releases/recent-releases"

export default function Page() {

    // const refetchCollection = useRefreshAnilistCollection()
    // useMount(() => {
    //     refetchCollection({ muteAlert: true })
    // })

    return (
        <div className={"px-4 pt-8 space-y-10 pb-10"}>
            <MissingEpisodes/>
            <ComingUpNext/>
            <RecentReleases/>
        </div>
    )
}
