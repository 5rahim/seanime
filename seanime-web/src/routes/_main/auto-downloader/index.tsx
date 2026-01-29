import { createFileRoute } from "@tanstack/react-router"
import Page from "@/app/(main)/auto-downloader/page"

export const Route = createFileRoute("/_main/auto-downloader/")({
    component: Page,
})
