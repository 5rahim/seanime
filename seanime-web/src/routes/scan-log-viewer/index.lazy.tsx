import Page from "@/app/scan-log-viewer/page"
import { createLazyFileRoute } from "@tanstack/react-router"

export const Route = createLazyFileRoute("/scan-log-viewer/")({
    component: Page,
})
