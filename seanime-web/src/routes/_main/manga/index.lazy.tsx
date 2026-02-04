import Page from "@/app/(main)/manga/page.tsx"
import { createLazyFileRoute } from "@tanstack/react-router"

export const Route = createLazyFileRoute("/_main/manga/")({
    component: Page,
})
