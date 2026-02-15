import Page from "@/app/(main)/qbittorrent/page"
import { createLazyFileRoute } from "@tanstack/react-router"

export const Route = createLazyFileRoute("/_main/qbittorrent/")({
    component: Page,
})
