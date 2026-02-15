import Page from "@/app/(main)/lists/page"
import { createLazyFileRoute } from "@tanstack/react-router"

export const Route = createLazyFileRoute("/_main/lists/")({
    component: Page,
})
