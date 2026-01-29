import { createFileRoute } from "@tanstack/react-router"
import Page from "@/app/(main)/scan-summaries/page"

export const Route = createFileRoute("/_main/scan-summaries/")({
    component: Page,
})
