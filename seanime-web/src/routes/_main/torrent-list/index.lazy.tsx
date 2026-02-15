import Page from "@/app/(main)/torrent-list/page"
import { createLazyFileRoute } from "@tanstack/react-router"

export const Route = createLazyFileRoute("/_main/torrent-list/")({
    component: Page,
})
