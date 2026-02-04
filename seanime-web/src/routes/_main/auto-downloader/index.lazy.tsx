import Page from "@/app/(main)/auto-downloader/page"
import { createLazyFileRoute } from "@tanstack/react-router"

export const Route = createLazyFileRoute("/_main/auto-downloader/")({
    component: Page,
})
