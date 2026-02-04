import Page from "@/app/(main)/search/page"
import { createLazyFileRoute } from "@tanstack/react-router"

export const Route = createLazyFileRoute("/_main/search/")({
    component: Page,
})
