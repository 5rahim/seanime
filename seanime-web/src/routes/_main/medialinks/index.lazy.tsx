import Page from "@/app/(main)/medialinks/page"
import { createLazyFileRoute } from "@tanstack/react-router"

export const Route = createLazyFileRoute("/_main/medialinks/")({
    component: Page,
})
