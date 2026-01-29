import { createFileRoute } from "@tanstack/react-router"
import Page from "@/app/(main)/torrent-list/page"

export const Route = createFileRoute("/_main/torrent-list/")({
    component: Page,
})
