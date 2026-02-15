import { createFileRoute } from "@tanstack/react-router"
import Page from "@/app/(main)/auth/callback/page"

export const Route = createFileRoute("/_main/auth/callback/")({
    component: Page,
})
