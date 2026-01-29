import { createFileRoute } from "@tanstack/react-router"
import Page from "@/app/(main)/extensions/playground/page"

export const Route = createFileRoute("/_main/extensions/playground/")({
    component: Page,
})
