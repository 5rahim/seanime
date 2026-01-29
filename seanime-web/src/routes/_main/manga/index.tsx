import Page from "@/app/(main)/manga/page.tsx"
import { createFileRoute } from "@tanstack/react-router"

export const Route = createFileRoute("/_main/manga/")({
    component: Page,
})
