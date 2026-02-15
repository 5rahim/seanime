import Page from "@/app/(main)/mediastream/page"
import { createLazyFileRoute } from "@tanstack/react-router"

export const Route = createLazyFileRoute("/_main/mediastream/")({
    component: Page,
})
