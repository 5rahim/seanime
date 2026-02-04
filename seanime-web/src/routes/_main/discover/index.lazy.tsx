import Page from "@/app/(main)/discover/page"
import { createLazyFileRoute } from "@tanstack/react-router"

export const Route = createLazyFileRoute("/_main/discover/")({
    component: Page,
})
