import Page from "@/app/(main)/extensions/page"
import { createLazyFileRoute } from "@tanstack/react-router"

export const Route = createLazyFileRoute("/_main/extensions/")({
    component: Page,
})
