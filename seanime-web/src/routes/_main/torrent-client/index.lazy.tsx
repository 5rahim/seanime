import Page from "@/app/(main)/torrent-client/page"
import { createLazyFileRoute } from "@tanstack/react-router"

export const Route = createLazyFileRoute("/_main/torrent-client/")({
    component: Page,
})
