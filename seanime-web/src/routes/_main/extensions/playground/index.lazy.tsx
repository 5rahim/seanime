import Page from "@/app/(main)/extensions/playground/page"
import { createLazyFileRoute } from "@tanstack/react-router"

export const Route = createLazyFileRoute("/_main/extensions/playground/")({
    component: Page,
})
