import { createFileRoute } from "@tanstack/react-router"
import Page from "@/app/scan-log-viewer/page"

export const Route = createFileRoute("/scan-log-viewer/")({
    component: Page,
})
