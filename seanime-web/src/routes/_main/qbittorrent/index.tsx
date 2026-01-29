import { createFileRoute } from "@tanstack/react-router"
import Page from "@/app/(main)/qbittorrent/page"

export const Route = createFileRoute("/_main/qbittorrent/")({
    component: Page,
})
