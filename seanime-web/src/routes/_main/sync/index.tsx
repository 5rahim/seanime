import { createFileRoute } from "@tanstack/react-router"
import Page from "@/app/(main)/sync/page"

export const Route = createFileRoute("/_main/sync/")({
    component: Page,
})
