import Page from "@/app/(main)/settings/page"
import { createLazyFileRoute } from "@tanstack/react-router"

export const Route = createLazyFileRoute("/_main/settings/")({
    component: Page,
})
