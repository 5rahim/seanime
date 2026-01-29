import { createFileRoute } from "@tanstack/react-router"
import Page from "@/app/docs/page"

export const Route = createFileRoute("/docs/")({
    component: Page,
})
