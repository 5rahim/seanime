import Page from "@/app/(main)/custom-sources/page"
import { createLazyFileRoute } from "@tanstack/react-router"

export const Route = createLazyFileRoute("/_main/custom-sources/")({
    component: Page,
})
