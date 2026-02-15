import Page from "@/app/(main)/sync/page"
import { createLazyFileRoute } from "@tanstack/react-router"

export const Route = createLazyFileRoute("/_main/sync/")({
    component: Page,
})
