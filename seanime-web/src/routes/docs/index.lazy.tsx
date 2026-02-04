import Page from "@/app/docs/page"
import { createLazyFileRoute } from "@tanstack/react-router"

export const Route = createLazyFileRoute("/docs/")({
    component: Page,
})
