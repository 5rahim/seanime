import Page from "@/app/(main)/scan-summaries/page"
import { createLazyFileRoute } from "@tanstack/react-router"

export const Route = createLazyFileRoute("/_main/scan-summaries/")({
    component: Page,
})
